package admin

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/admin/handlers"
	"github.com/csmith/middleware"
	"tailscale.com/tsnet"
)

var (
	tailscaleHost = flag.String("tailscale-host", "website-admin", "Tailscale host")
	tailscaleDir  = flag.String("tailscale-dir", "tsdata", "Tailscale directory")
)

func Start() error {
	s := new(tsnet.Server)
	s.Hostname = *tailscaleHost
	s.Dir = *tailscaleDir
	s.UserLogf = func(s string, v ...any) {
		slog.Info(fmt.Sprintf(s, v...), "source", "tailscale")
	}
	s.Logf = func(s string, v ...any) {
		slog.Debug(fmt.Sprintf(s, v...), "source", "tailscale")
	}

	if err := s.Start(); err != nil {
		return err
	}

	httpListener, err := s.Listen("tcp", ":80")
	if err != nil {
		return err
	}

	httpsListener, err := s.ListenTLS("tcp", ":443")
	if err != nil {
		return err
	}

	httpServer := &http.Server{
		Handler: http.HandlerFunc(handlers.RedirectHandler(func() string {
			return fullHostName(s)
		})),
	}

	httpsMux := http.NewServeMux()
	httpsMux.Handle("GET /assets/", http.StripPrefix("/assets/", handlers.AssetsHandler()))
	httpsMux.HandleFunc("GET /{$}", handlers.IndexHandler())
	httpsMux.HandleFunc("GET /posts", handlers.ListPostsHandler())
	httpsMux.HandleFunc("POST /posts", handlers.CreatePostHandler())
	httpsMux.HandleFunc("GET /posts/edit/{id}", handlers.EditPostHandler())
	httpsMux.HandleFunc("POST /posts/edit/{id}", handlers.UpdatePostHandler())
	httpsMux.HandleFunc("POST /posts/generate-wordcloud/{id}", handlers.GenerateWordcloudHandler())
	httpsMux.HandleFunc("GET /pages", handlers.ListPagesHandler())
	httpsMux.HandleFunc("POST /pages", handlers.CreatePageHandler())
	httpsMux.HandleFunc("GET /pages/edit/{id}", handlers.EditPageHandler())
	httpsMux.HandleFunc("POST /pages/edit/{id}", handlers.UpdatePageHandler())
	httpsMux.HandleFunc("GET /snippets", handlers.ListSnippetsHandler())
	httpsMux.HandleFunc("POST /snippets", handlers.CreateSnippetHandler())
	httpsMux.HandleFunc("GET /snippets/edit/{id}", handlers.EditSnippetHandler())
	httpsMux.HandleFunc("POST /snippets/edit/{id}", handlers.UpdateSnippetHandler())
	httpsMux.HandleFunc("GET /poems", handlers.ListPoemsHandler())
	httpsMux.HandleFunc("POST /poems", handlers.CreatePoemHandler())
	httpsMux.HandleFunc("GET /poems/edit/{id}", handlers.EditPoemHandler())
	httpsMux.HandleFunc("POST /poems/edit/{id}", handlers.UpdatePoemHandler())
	httpsMux.HandleFunc("GET /pastes", handlers.ListPastesHandler())
	httpsMux.HandleFunc("POST /pastes", handlers.CreatePasteHandler())
	httpsMux.HandleFunc("GET /pastes/edit/{id}", handlers.EditPasteHandler())
	httpsMux.HandleFunc("POST /pastes/edit/{id}", handlers.UpdatePasteHandler())
	httpsMux.HandleFunc("GET /projects", handlers.ListProjectsHandler())
	httpsMux.HandleFunc("POST /projects", handlers.CreateProjectHandler())
	httpsMux.HandleFunc("GET /projects/edit/{id}", handlers.EditProjectHandler())
	httpsMux.HandleFunc("POST /projects/edit/{id}", handlers.UpdateProjectHandler())
	httpsMux.HandleFunc("GET /media", handlers.MediaHandler())
	httpsMux.HandleFunc("POST /media/upload", handlers.UploadMediaHandler())
	httpsMux.HandleFunc("GET /media/view/{id}", handlers.ViewMediaHandler())
	httpsMux.HandleFunc("GET /media-relations/edit", handlers.EditMediaRelationsHandler())
	httpsMux.HandleFunc("POST /media-relations/update", handlers.UpdateMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/remove", handlers.RemoveMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/add", handlers.AddMediaRelationsHandler())
	httpsMux.HandleFunc("GET /goimports", handlers.ListGoImportsHandler())
	httpsMux.HandleFunc("POST /goimports", handlers.CreateGoImportHandler())
	httpsMux.HandleFunc("GET /goimports/edit/{id}", handlers.EditGoImportHandler())
	httpsMux.HandleFunc("POST /goimports/edit/{id}", handlers.UpdateGoImportHandler())
	httpsMux.HandleFunc("GET /films", handlers.ListFilmsHandler())
	httpsMux.HandleFunc("GET /films/search", handlers.SearchFilmsHandler())
	httpsMux.HandleFunc("POST /films", handlers.CreateFilmHandler())
	httpsMux.HandleFunc("GET /films/edit/{id}", handlers.EditFilmHandler())
	httpsMux.HandleFunc("POST /films/edit/{id}", handlers.UpdateFilmHandler())
	httpsMux.HandleFunc("POST /films/delete/{id}", handlers.DeleteFilmHandler())
	httpsMux.HandleFunc("POST /films/fetch-poster/{id}", handlers.FetchFilmPosterHandler())
	httpsMux.HandleFunc("GET /film-reviews/edit/{id}", handlers.EditFilmReviewHandler())
	httpsMux.HandleFunc("POST /film-reviews/create/{id}", handlers.CreateFilmReviewHandler())
	httpsMux.HandleFunc("POST /film-reviews/edit/{id}", handlers.UpdateFilmReviewHandler())
	httpsMux.HandleFunc("GET /film-lists", handlers.ListFilmListsHandler())
	httpsMux.HandleFunc("POST /film-lists", handlers.CreateFilmListHandler())
	httpsMux.HandleFunc("GET /film-lists/{id}/edit", handlers.EditFilmListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/edit", handlers.UpdateFilmListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries", handlers.AddFilmToListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/remove/{entryId}", handlers.RemoveFilmFromListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/position/{entryId}", handlers.UpdateEntryPositionHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/reorder", handlers.ReorderFilmListEntriesHandler())
	httpsMux.HandleFunc("GET /api/films/reviews/", handlers.GetFilmsWithReviewsHandler())
	httpsMux.HandleFunc("POST /api/walks/import", handlers.ImportWalksHandler())
	httpsMux.HandleFunc("POST /api/boardgames/import", handlers.ImportBoardgamesHandler())
	httpsMux.HandleFunc("GET /films/workflow/step/1", handlers.FilmReviewWorkflowStep1Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/1", handlers.FilmReviewWorkflowStep1Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/2", handlers.FilmReviewWorkflowStep2Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/2", handlers.FilmReviewWorkflowStep2Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/3", handlers.FilmReviewWorkflowStep3Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/3", handlers.FilmReviewWorkflowStep3Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/4", handlers.FilmReviewWorkflowStep4Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/4", handlers.FilmReviewWorkflowStep4Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/5", handlers.FilmReviewWorkflowStep5Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/5", handlers.FilmReviewWorkflowStep5Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/6", handlers.FilmReviewWorkflowStep6Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/6", handlers.FilmReviewWorkflowStep6Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/7", handlers.FilmReviewWorkflowStep7Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/7", handlers.FilmReviewWorkflowStep7Handler())
	httpsMux.HandleFunc("GET /videogames", handlers.ListVideoGamesHandler())
	httpsMux.HandleFunc("POST /videogames", handlers.CreateVideoGameHandler())
	httpsMux.HandleFunc("GET /videogames/edit/{id}", handlers.EditVideoGameHandler())
	httpsMux.HandleFunc("POST /videogames/edit/{id}", handlers.UpdateVideoGameHandler())
	httpsMux.HandleFunc("POST /videogames/delete/{id}", handlers.DeleteVideoGameHandler())
	httpsMux.HandleFunc("GET /video-game-reviews/edit/{id}", handlers.EditVideoGameReviewHandler())
	httpsMux.HandleFunc("POST /video-game-reviews/create/{id}", handlers.CreateVideoGameReviewHandler())
	httpsMux.HandleFunc("POST /video-game-reviews/edit/{id}", handlers.UpdateVideoGameReviewHandler())
	httpsMux.HandleFunc("GET /syndications", handlers.ListSyndicationsHandler())
	httpsMux.HandleFunc("POST /syndications/create", handlers.CreateSyndicationHandler())
	httpsMux.HandleFunc("GET /syndications/edit/{id}", handlers.EditSyndicationHandler())
	httpsMux.HandleFunc("POST /syndications/edit/{id}", handlers.UpdateSyndicationHandler())
	httpsMux.HandleFunc("POST /syndications/delete/{id}", handlers.DeleteSyndicationHandler())
	httpsMux.HandleFunc("GET /", handlers.StaticAsset)

	httpsServer := &http.Server{
		Handler: middleware.Chain(
			middleware.WithMiddleware(
				middleware.CacheControl(
					middleware.WithCacheTimes(map[string]time.Duration{
						"application/*":   time.Hour * 24 * 365,
						"font/*":          time.Hour * 24 * 365,
						"image/*":         time.Hour * 24 * 365,
						"text/css":        time.Hour * 24 * 365,
						"text/javascript": time.Hour * 24 * 365,
					}),
				),
				middleware.CrossOriginProtection(),
				middleware.Recover(middleware.WithPanicLogger(func(r *http.Request, err any) {
					slog.Error("Panic serving admin site", "url", r.RequestURI, "error", err)
				})),
			),
		)(httpsMux),
	}

	go httpServer.Serve(httpListener)
	go httpsServer.Serve(httpsListener)

	return nil
}

func fullHostName(s *tsnet.Server) string {
	lc, err := s.LocalClient()
	if err != nil {
		slog.Error("Failed to get local client", "error", err)
		return ""
	}

	status, err := lc.Status(context.Background())
	if err != nil {
		slog.Error("Failed to get local status", "error", err)
		return ""
	}

	fullHostname := status.Self.DNSName
	if len(fullHostname) > 0 && fullHostname[len(fullHostname)-1] == '.' {
		fullHostname = fullHostname[:len(fullHostname)-1]
	}

	return fullHostname
}
