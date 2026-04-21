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

	cropRect := contentCropRect(img)
	cropped := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(cropRect)

	var buf bytes.Buffer
	if err := png.Encode(&buf, cropped); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to encode image: %w", err)
	}

	return buf.Bytes(), cropRect.Dx(), cropRect.Dy(), nil
}

func contentCropRect(img image.Image) image.Rectangle {
	minX, minY := img.Bounds().Max.X, img.Bounds().Max.Y
	maxX, maxY := 0, 0
	for y := img.Bounds().Min.Y; y < img.Bounds().Max.Y; y++ {
		for x := img.Bounds().Min.X; x < img.Bounds().Max.X; x++ {
			_, _, _, a := img.At(x, y).RGBA()
			if a > 0 {
				if x < minX {
					minX = x
				}
				if x > maxX {
					maxX = x
				}
				if y < minY {
					minY = y
				}
				if y > maxY {
					maxY = y
				}
			}
		}
	}

	contentW := maxX - minX + 1
	contentH := maxY - minY + 1

	targetW := contentW
	targetH := contentW * 3 / 2

	if targetH < contentH {
		targetH = contentH
		targetW = targetH * 2 / 3
	}

	cropX := minX + (contentW-targetW)/2
	cropY := minY + (contentH-targetH)/2

	return image.Rect(cropX, cropY, cropX+targetW, cropY+targetH)
}
