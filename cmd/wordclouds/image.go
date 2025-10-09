package main

import (
	"bytes"
	_ "embed"
	"math/rand/v2"

	"github.com/anthonynsimon/bild/transform"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"

	"image"
	"image/color"
	"image/draw"
	"image/png"
)

//go:embed ibmplexsans.ttf
var fontBytes []byte

func generateImage(words []string) ([]byte, error) {
	im := image.NewNRGBA(image.Rect(0, 0, 500, 400))

	dark := image.NewUniform(color.NRGBA{
		R: 30,
		G: 50,
		B: 70,
		A: 255,
	})

	fg := image.NewUniform(color.NRGBA{
		R: 60,
		G: 101,
		B: 141,
		A: 255,
	})

	bg := image.NewUniform(color.NRGBA{
		R: 48,
		G: 81,
		B: 113,
		A: 255,
	})

	draw.Draw(im, im.Bounds(), bg, image.Point{}, draw.Src)

	f, err := opentype.Parse(fontBytes)
	if err != nil {
		return nil, err
	}

	face, err := opentype.NewFace(f, &opentype.FaceOptions{
		Size:    60,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		return nil, err
	}

	d := font.Drawer{
		Dst:  im,
		Src:  dark,
		Face: face,
		Dot:  fixed.P(30, 30),
	}

	nextWord := 0

	fillLine := func() {
		// Initial word
		bounds, _ := d.BoundString(words[nextWord])
		originalX := fixed.I(270)
		d.Dot.X = originalX
		d.DrawString(words[nextWord])
		nextWord++
		d.Src = fg

		// Scan right
		lastX := originalX + bounds.Max.X - bounds.Min.X + 15<<6
		d.Dot.X = lastX
		for d.Dot.X < fixed.I(450) {
			newBounds, _ := d.BoundString(words[nextWord])
			d.DrawString(words[nextWord])
			lastX += newBounds.Max.X - newBounds.Min.X + 15<<6
			d.Dot.X = lastX
			nextWord++
		}

		// Scan left
		lastX = originalX
		for d.Dot.X > fixed.I(0) {
			newBounds, _ := d.BoundString(words[nextWord])
			d.Dot.X = lastX - (newBounds.Max.X - newBounds.Min.X) - 15<<6
			lastX = d.Dot.X
			d.DrawString(words[nextWord])
			nextWord++
		}
	}

	// Key line, about 1/4 of the way down
	d.Dot.Y = fixed.I(150 + 0*65)
	fillLine()

	// Line below
	d.Dot.Y = fixed.I(150 + 1*65)
	fillLine()

	// Second line below
	d.Dot.Y = fixed.I(150 + 2*65)
	fillLine()

	// Line above
	d.Dot.Y = fixed.I(150 - 1*65)
	fillLine()

	// Third line below
	d.Dot.Y = fixed.I(150 + 3*65)
	fillLine()

	angle := -10.0
	if rand.Float64() >= 0.5 {
		angle = 10
	}
	dst := transform.Crop(transform.Rotate(im, angle, nil), image.Rect(50, 50, 450, 350))

	b := &bytes.Buffer{}
	err = png.Encode(b, dst)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
