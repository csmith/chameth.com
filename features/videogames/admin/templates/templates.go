package templates

import _ "embed"

//go:embed list-video-games.html.gotpl
var listVideoGamesGotpl string

//go:embed edit-video-game.html.gotpl
var editVideoGameGotpl string

//go:embed edit-video-game-review.html.gotpl
var editVideoGameReviewGotpl string
