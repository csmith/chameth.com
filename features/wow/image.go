package wow

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/png"
	"io"
	"net/http"

	"chameth.com/chameth.com/external/blizzard"
)

func fetchAndSaveCharacterImage(ctx context.Context, characterID int, name string, media *blizzard.CharacterMedia) error {
	imageURL := ""
	for _, a := range media.Assets {
		if a.Key == "main-raw" {
			imageURL = a.Value
			break
		}
	}
	if imageURL == "" {
		return fmt.Errorf("no main-raw asset found")
	}

	imgData, width, height, err := downloadAndCropImage(imageURL)
	if err != nil {
		return fmt.Errorf("failed to download and crop image: %w", err)
	}

	return saveCharacterImage(ctx, characterID, name, imgData, width, height)
}

func downloadAndCropImage(url string) ([]byte, int, int, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read image: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	const targetW, targetH = 500, 800
	srcBounds := img.Bounds()
	srcW := srcBounds.Dx()
	srcH := srcBounds.Dy()

	offsetX := (srcW - targetW) / 2

	gapTotal := srcH - targetH
	offsetY := (gapTotal * 2) / 3

	cropRect := image.Rect(offsetX, offsetY, offsetX+targetW, offsetY+targetH)
	cropped := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	var buf bytes.Buffer
	if err := png.Encode(&buf, cropped); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), targetW, targetH, nil
}
