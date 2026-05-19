package admin

import (
	"chameth.com/chameth.com/features/routing"
)

func RegisterRoutes(rm *routing.Manager) {
	rm.Admin.HandleFunc("GET /videogames", ListVideoGamesHandler())
	rm.Admin.HandleFunc("POST /videogames", CreateVideoGameHandler())
	rm.Admin.HandleFunc("GET /videogames/edit/{id}", EditVideoGameHandler())
	rm.Admin.HandleFunc("POST /videogames/edit/{id}", UpdateVideoGameHandler())
	rm.Admin.HandleFunc("POST /videogames/delete/{id}", DeleteVideoGameHandler())
	rm.Admin.HandleFunc("GET /video-game-reviews/edit/{id}", EditVideoGameReviewHandler())
	rm.Admin.HandleFunc("POST /video-game-reviews/create/{id}", CreateVideoGameReviewHandler())
	rm.Admin.HandleFunc("POST /video-game-reviews/edit/{id}", UpdateVideoGameReviewHandler())
}
