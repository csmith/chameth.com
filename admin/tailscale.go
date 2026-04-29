package admin

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"chameth.com/chameth.com/admin/handlers"
	bgadmin "chameth.com/chameth.com/features/boardgames/admin"
	filmadmin "chameth.com/chameth.com/features/films/admin"
	goimportadmin "chameth.com/chameth.com/features/goimports/admin"
	pageadmin "chameth.com/chameth.com/features/pages/admin"
	postadmin "chameth.com/chameth.com/features/posts/admin"
	pasteadmin "chameth.com/chameth.com/features/pastes/admin"
	poemadmin "chameth.com/chameth.com/features/poems/admin"
	projectadmin "chameth.com/chameth.com/features/projects/admin"
	snippetadmin "chameth.com/chameth.com/features/snippets/admin"
	walksadmin "chameth.com/chameth.com/features/walks/admin"
	wowadmin "chameth.com/chameth.com/features/wow/admin"
	"github.com/csmith/middleware"
	"tailscale.com/tsnet"
)

func Start(s *tsnet.Server) error {
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
	httpsMux.HandleFunc("GET /posts", postadmin.ListPostsHandler())
	httpsMux.HandleFunc("POST /posts", postadmin.CreatePostHandler())
	httpsMux.HandleFunc("GET /posts/edit/{id}", postadmin.EditPostHandler())
	httpsMux.HandleFunc("POST /posts/edit/{id}", postadmin.UpdatePostHandler())
	httpsMux.HandleFunc("POST /posts/generate-wordcloud/{id}", postadmin.GenerateWordcloudHandler())
	httpsMux.HandleFunc("GET /pages", pageadmin.ListPagesHandler())
	httpsMux.HandleFunc("POST /pages", pageadmin.CreatePageHandler())
	httpsMux.HandleFunc("GET /pages/edit/{id}", pageadmin.EditPageHandler())
	httpsMux.HandleFunc("POST /pages/edit/{id}", pageadmin.UpdatePageHandler())
	httpsMux.HandleFunc("GET /snippets", snippetadmin.ListSnippetsHandler())
	httpsMux.HandleFunc("POST /snippets", snippetadmin.CreateSnippetHandler())
	httpsMux.HandleFunc("GET /snippets/edit/{id}", snippetadmin.EditSnippetHandler())
	httpsMux.HandleFunc("POST /snippets/edit/{id}", snippetadmin.UpdateSnippetHandler())
	httpsMux.HandleFunc("GET /poems", poemadmin.ListPoemsHandler())
	httpsMux.HandleFunc("POST /poems", poemadmin.CreatePoemHandler())
	httpsMux.HandleFunc("GET /poems/edit/{id}", poemadmin.EditPoemHandler())
	httpsMux.HandleFunc("POST /poems/edit/{id}", poemadmin.UpdatePoemHandler())
	httpsMux.HandleFunc("GET /pastes", pasteadmin.ListPastesHandler())
	httpsMux.HandleFunc("POST /pastes", pasteadmin.CreatePasteHandler())
	httpsMux.HandleFunc("GET /pastes/edit/{id}", pasteadmin.EditPasteHandler())
	httpsMux.HandleFunc("POST /pastes/edit/{id}", pasteadmin.UpdatePasteHandler())
	httpsMux.HandleFunc("GET /projects", projectadmin.ListProjectsHandler())
	httpsMux.HandleFunc("POST /projects", projectadmin.CreateProjectHandler())
	httpsMux.HandleFunc("GET /projects/edit/{id}", projectadmin.EditProjectHandler())
	httpsMux.HandleFunc("POST /projects/edit/{id}", projectadmin.UpdateProjectHandler())
	httpsMux.HandleFunc("GET /media", handlers.MediaHandler())
	httpsMux.HandleFunc("POST /media/upload", handlers.UploadMediaHandler())
	httpsMux.HandleFunc("GET /media/view/{id}", handlers.ViewMediaHandler())
	httpsMux.HandleFunc("GET /media/edit/{id}", handlers.EditMediaHandler())
	httpsMux.HandleFunc("POST /media/edit/{id}", handlers.ReplaceMediaHandler())
	httpsMux.HandleFunc("GET /media-relations/edit", handlers.EditMediaRelationsHandler())
	httpsMux.HandleFunc("POST /media-relations/update", handlers.UpdateMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/remove", handlers.RemoveMediaRelationHandler())
	httpsMux.HandleFunc("POST /media-relations/add", handlers.AddMediaRelationsHandler())
	httpsMux.HandleFunc("GET /goimports", goimportadmin.ListGoImportsHandler())
	httpsMux.HandleFunc("POST /goimports", goimportadmin.CreateGoImportHandler())
	httpsMux.HandleFunc("GET /goimports/edit/{id}", goimportadmin.EditGoImportHandler())
	httpsMux.HandleFunc("POST /goimports/edit/{id}", goimportadmin.UpdateGoImportHandler())
	httpsMux.HandleFunc("GET /films", filmadmin.ListFilmsHandler())
	httpsMux.HandleFunc("GET /films/search", filmadmin.SearchFilmsHandler())
	httpsMux.HandleFunc("POST /films", filmadmin.CreateFilmHandler())
	httpsMux.HandleFunc("GET /films/edit/{id}", filmadmin.EditFilmHandler())
	httpsMux.HandleFunc("POST /films/edit/{id}", filmadmin.UpdateFilmHandler())
	httpsMux.HandleFunc("POST /films/delete/{id}", filmadmin.DeleteFilmHandler())
	httpsMux.HandleFunc("POST /films/fetch-poster/{id}", filmadmin.FetchFilmPosterHandler())
	httpsMux.HandleFunc("GET /film-reviews/edit/{id}", filmadmin.EditFilmReviewHandler())
	httpsMux.HandleFunc("POST /film-reviews/create/{id}", filmadmin.CreateFilmReviewHandler())
	httpsMux.HandleFunc("POST /film-reviews/edit/{id}", filmadmin.UpdateFilmReviewHandler())
	httpsMux.HandleFunc("GET /film-lists", filmadmin.ListFilmListsHandler())
	httpsMux.HandleFunc("POST /film-lists", filmadmin.CreateFilmListHandler())
	httpsMux.HandleFunc("GET /film-lists/{id}/edit", filmadmin.EditFilmListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/edit", filmadmin.UpdateFilmListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries", filmadmin.AddFilmToListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/remove/{entryId}", filmadmin.RemoveFilmFromListHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/position/{entryId}", filmadmin.UpdateEntryPositionHandler())
	httpsMux.HandleFunc("POST /film-lists/{id}/entries/reorder", filmadmin.ReorderFilmListEntriesHandler())
	httpsMux.HandleFunc("GET /api/films/reviews/", filmadmin.GetFilmsWithReviewsHandler())
	httpsMux.HandleFunc("POST /api/walks/import", walksadmin.ImportWalksHandler())
	httpsMux.HandleFunc("POST /api/boardgames/import", bgadmin.ImportBoardgamesHandler())
	httpsMux.HandleFunc("GET /wow", wowadmin.ListCharactersHandler())
	httpsMux.HandleFunc("POST /wow/import", wowadmin.ImportCharacterHandler())
	httpsMux.HandleFunc("GET /wow/refresh/{id}", wowadmin.RefreshCharacterHandler())
	httpsMux.HandleFunc("GET /films/workflow/step/1", filmadmin.FilmReviewWorkflowStep1Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/1", filmadmin.FilmReviewWorkflowStep1Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/2", filmadmin.FilmReviewWorkflowStep2Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/2", filmadmin.FilmReviewWorkflowStep2Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/3", filmadmin.FilmReviewWorkflowStep3Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/3", filmadmin.FilmReviewWorkflowStep3Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/4", filmadmin.FilmReviewWorkflowStep4Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/4", filmadmin.FilmReviewWorkflowStep4Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/5", filmadmin.FilmReviewWorkflowStep5Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/5", filmadmin.FilmReviewWorkflowStep5Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/6", filmadmin.FilmReviewWorkflowStep6Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/6", filmadmin.FilmReviewWorkflowStep6Handler())
	httpsMux.HandleFunc("GET /films/workflow/step/7", filmadmin.FilmReviewWorkflowStep7Handler())
	httpsMux.HandleFunc("POST /films/workflow/step/7", filmadmin.FilmReviewWorkflowStep7Handler())
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
