package shortcodes

import "chameth.com/chameth.com/cmd/serve/db"

func Render(input string, media []db.MediaRelationWithDetails) (string, error) {
	var res = input
	var err error

	res, err = renderSideNote(res)
	if err != nil {
		return "", err
	}

	res, err = renderUpdate(res)
	if err != nil {
		return "", err
	}

	res, err = renderWarning(res)
	if err != nil {
		return "", err
	}

	res, err = renderAudio(res, media)
	if err != nil {
		return "", err
	}

	res, err = renderVideo(res, media)
	if err != nil {
		return "", err
	}

	res, err = renderFigure(res, media)
	if err != nil {
		return "", err
	}

	res, err = renderFilmReview(res)
	if err != nil {
		return "", err
	}

	res, err = renderFilmReviews(res)
	if err != nil {
		return "", err
	}

	return res, nil
}
