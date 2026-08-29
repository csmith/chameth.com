package workouts

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/external/pompeiband"
	"tailscale.com/tsnet"
)

var pompeibandBaseUrl = flag.String("pompeiband-base-url", "https://pb.yak-wall.ts.net", "Base URL for the Pompei Band API")

func RegisterGoroutine(ctx context.Context, ts *tsnet.Server) func() {
	return func() {
		ts.Up(ctx)
		runSync(ctx, ts.HTTPClient())
	}
}

func runSync(ctx context.Context, client *http.Client) {
	if *pompeibandBaseUrl == "" {
		return
	}

	ticker := time.NewTicker(time.Hour)
	defer ticker.Stop()

	syncWorkouts(ctx, client)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			syncWorkouts(ctx, client)
		}
	}
}

func syncWorkouts(ctx context.Context, client *http.Client) {
	since, err := cursorSince(ctx)
	if err != nil {
		slog.Error("Failed to get workout sync cursor", "error", err)
		return
	}

	sinceVersion, err := cursorMinProcessorVersion(ctx)
	if err != nil {
		slog.Error("Failed to get workout sync cursor version", "error", err)
		return
	}

	c := pompeiband.NewClient(client, *pompeibandBaseUrl)
	activities, err := c.GetActivities(ctx, since, sinceVersion)
	if err != nil {
		slog.Error("Failed to fetch activities", "error", err)
		return
	}

	imported := 0
	for _, a := range activities {
		if err := importActivity(ctx, a); err != nil {
			slog.Error("Failed to import activity", "error", err, "external_id", a.ID)
			continue
		}
		imported++
	}

	slog.Info("Workout sync complete", "fetched", len(activities), "imported", imported)
}

func importActivity(ctx context.Context, a pompeiband.Activity) error {
	w := workout{
		ExternalID:       a.ID,
		ProcessorVersion: a.ProcessorVersion,
		Activity:         a.Activity,
		ActivityGroup:    a.ActivityGroup,
		StartTime:        a.StartTime,
		EndTime:          a.EndTime,
		DurationS:        a.DurationS,
		DistanceM:        a.DistanceM,
		ActiveS:          a.ActiveS,
		ElapsedS:         a.ElapsedS,
		PausedS:          a.PausedS,
		Calories:         a.Calories,
		RunDistanceM:     a.RunDistanceM,
		WalkDistanceM:    a.WalkDistanceM,
	}

	if a.Elevation != nil {
		w.ElevationGainM = &a.Elevation.GainM
		w.ElevationLossM = &a.Elevation.LossM
	}
	if a.Gait != nil {
		w.GaitRunDistanceM = &a.Gait.RunDistanceM
		w.GaitWalkDistanceM = &a.Gait.WalkDistanceM
		w.GaitOverallPaceSPerKm = &a.Gait.OverallPaceSPerKm
		w.GaitOverallCadenceSpm = &a.Gait.OverallCadenceSpm
		w.GaitRunPaceSPerKm = &a.Gait.RunPaceSPerKm
		w.GaitRunCadenceSpm = &a.Gait.RunCadenceSpm
		w.GaitWalkPaceSPerKm = &a.Gait.WalkPaceSPerKm
		w.GaitWalkCadenceSpm = &a.Gait.WalkCadenceSpm
	}
	if a.Weather != nil {
		w.WeatherTempC = &a.Weather.TempC
		w.WeatherApparentTempC = &a.Weather.ApparentTempC
		w.WeatherHumidityPct = &a.Weather.HumidityPct
		w.WeatherCode = &a.Weather.WeatherCode
		w.WeatherLabel = &a.Weather.WeatherLabel
		w.WeatherPrecipMm = &a.Weather.PrecipMm
		w.WeatherSnowfallCm = &a.Weather.SnowfallCm
		w.WeatherCloudCoverPct = &a.Weather.CloudCoverPct
		w.WeatherVisibilityM = &a.Weather.VisibilityM
	}

	segments := make([]workoutSegment, len(a.Segments))
	for i, s := range a.Segments {
		segments[i] = workoutSegment{
			SegmentIndex:  i,
			DistanceM:     s.DistanceM,
			StartTime:     s.StartTime,
			EndTime:       s.EndTime,
			ElapsedS:      s.ElapsedS,
			PaceSPerKm:    s.PaceSPerKm,
			SpeedKmh:      s.SpeedKmh,
			GapElapsedS:   s.GapElapsedS,
			GapPaceSPerKm: s.GapPaceSPerKm,
			IsPb:          s.IsPB,
			PreviousBestS: s.PreviousBestS,
			BestS:         s.BestS,
			Rank:          s.Rank,
		}
	}

	return insertWorkoutWithSegments(ctx, w, segments)
}
