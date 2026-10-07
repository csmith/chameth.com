package contact

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"time"

	"chameth.com/chameth.com/features/metrics"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

const (
	taskQueue    = "chameth.com"
	workflowName = "chameth.com.Contact"

	checksActivityName = "chameth.com.contactChecks"
	recordActivityName = "chameth.com.contactRecord"
	sendActivityName   = "chameth.com.contactSend"

	germanMallTaskQueue = "germanmall"
	germanMallWorkflow  = "germanmall.Execute"
	germanMallClient    = "chameth.com"
	spamCheckWorkflow   = "check-blog-spam"

	// germanMallTimeout bounds the whole child execution, queueing included:
	// above the five minutes German Mall allows its provider call.
	germanMallTimeout = 10 * time.Minute

	workerRetryInterval = time.Minute
)

// startWorkflow hands a submission over to Temporal; everything from the
// spam checks onwards happens in the workflow.
func startWorkflow(ctx context.Context, tc client.Client, sub submission) error {
	_, err := tc.ExecuteWorkflow(ctx, client.StartWorkflowOptions{TaskQueue: taskQueue}, workflowName, sub)
	return err
}

func RegisterGoroutine(ctx context.Context, tc client.Client) func() {
	return func() {
		w := worker.New(tc, taskQueue, worker.Options{})
		w.RegisterWorkflowWithOptions(contactWorkflow, workflow.RegisterOptions{Name: workflowName})
		w.RegisterActivityWithOptions(checksActivity, activity.RegisterOptions{Name: checksActivityName})
		w.RegisterActivityWithOptions(recordActivity, activity.RegisterOptions{Name: recordActivityName})
		w.RegisterActivityWithOptions(sendActivity, activity.RegisterOptions{Name: sendActivityName})

		// The frontend may be unreachable at boot (tailnet or Temporal down),
		// so keep trying rather than leaving the form without a worker.
		for {
			err := w.Start()
			if err == nil {
				break
			}
			slog.Error("Failed to start contact form worker", "error", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(workerRetryInterval):
			}
		}

		<-ctx.Done()
		w.Stop()
	}
}

func contactWorkflow(ctx workflow.Context, sub submission) error {
	logger := workflow.GetLogger(ctx)

	// Never retried: the rate limiter records the address on the first
	// attempt, so a retry would reject its own submission.
	checksCtx := workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 1},
		},
	)
	var failed []cause
	if err := workflow.ExecuteActivity(checksCtx, checksActivityName, sub).Get(ctx, &failed); err != nil {
		return err
	}

	var llmReason *string
	if len(failed) == 0 {
		verdict, err := checkSpam(ctx, sub)
		if err != nil {
			// Fail open: losing a genuine message is worse than letting
			// one through.
			logger.Error("LLM spam check failed, sending anyway", "error", err)
		} else {
			llmReason = &verdict.Reason
			if !verdict.Send {
				logger.Info("LLM rejected contact form message", "reason", verdict.Reason)
				failed = append(failed, causeLLM)
			}
		}
	}

	recordCtx := workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: 30 * time.Second,
			RetryPolicy:         &temporal.RetryPolicy{MaximumAttempts: 5},
		},
	)
	if err := workflow.ExecuteActivity(recordCtx, recordActivityName, sub, failed, llmReason).Get(ctx, nil); err != nil {
		logger.Error("Failed to record contact submission", "error", err)
	}

	if len(failed) > 0 {
		return nil
	}

	sendCtx := workflow.WithActivityOptions(
		ctx,
		workflow.ActivityOptions{
			StartToCloseTimeout: time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    30 * time.Second,
				BackoffCoefficient: 2,
				MaximumInterval:    30 * time.Minute,
				MaximumAttempts:    10,
			},
		},
	)
	return workflow.ExecuteActivity(sendCtx, sendActivityName, sub).Get(ctx, nil)
}

// checkSpam asks German Mall's check-blog-spam workflow for a verdict. The
// child is never retried, as each run is a billable model call.
func checkSpam(ctx workflow.Context, sub submission) (spamVerdict, error) {
	ctx = workflow.WithChildOptions(
		ctx,
		workflow.ChildWorkflowOptions{
			TaskQueue:                germanMallTaskQueue,
			WorkflowExecutionTimeout: germanMallTimeout,
			RetryPolicy:              &temporal.RetryPolicy{MaximumAttempts: 1},
			ParentClosePolicy:        enumspb.PARENT_CLOSE_POLICY_TERMINATE,
		},
	)

	type gmRequest struct {
		Workflow string            `json:"workflow"`
		Client   string            `json:"client"`
		Args     map[string]string `json:"args"`
	}
	var result struct {
		Status string      `json:"status"`
		Output spamVerdict `json:"output"`
	}
	err := workflow.ExecuteChildWorkflow(
		ctx,
		germanMallWorkflow,
		gmRequest{
			Workflow: spamCheckWorkflow,
			Client:   germanMallClient,
			Args: map[string]string{
				"page":    sub.Request.Page,
				"name":    sub.Request.SenderName,
				"email":   sub.Request.SenderEmail,
				"message": sub.Request.Message,
			},
		},
	).Get(ctx, &result)
	if err != nil {
		return spamVerdict{}, err
	}
	if result.Status != "ok" {
		return spamVerdict{}, fmt.Errorf("german mall returned status %q", result.Status)
	}
	return result.Output, nil
}

func checksActivity(ctx context.Context, sub submission) ([]cause, error) {
	host, _, err := net.SplitHostPort(sub.RemoteAddr)
	if err != nil {
		host = sub.RemoteAddr
	}

	var failed []cause
	for _, check := range checks {
		if err := check(ctx, sub, host); err != nil {
			if rej, ok := errors.AsType[*rejection](err); ok {
				failed = append(failed, rej.cause)
			} else {
				slog.Error("Error checking contact form for spam", "request", sub.Request, "error", err)
			}
		}
	}
	return failed, nil
}

func recordActivity(ctx context.Context, sub submission, failed []cause, llmReason *string) error {
	failedStrings := make([]string, len(failed))
	for i, c := range failed {
		failedStrings[i] = string(c)
	}
	return metrics.RecordContactSubmission(
		ctx,
		metrics.ContactSubmission{
			Method:       string(sub.Method),
			UserAgent:    sub.UserAgent,
			RemoteAddr:   sub.RemoteAddr,
			FailedChecks: failedStrings,
			Page:         sub.Request.Page,
			SenderName:   sub.Request.SenderName,
			SenderEmail:  sub.Request.SenderEmail,
			Message:      sub.Request.Message,
			LLMReason:    llmReason,
		},
	)
}

func sendActivity(_ context.Context, sub submission) error {
	return sendContact(sub.Request, messageBody(sub.Request, sub.Method, sub.RemoteAddr))
}
