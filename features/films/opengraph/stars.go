package opengraph

import (
	"image"
	"image/color"
	"math"
)

const (
	starSize   = 48.0
	starStep   = starSize - 4
	starRadius = starSize * 0.47
)

// Each star is tilted a little differently, matching the sticker-style
// stars used by the rating shortcode.
var starRotations = [5]float64{-3, 5, 10, -10, 15}

var starPoints = func() [10][2]float64 {
	var pts [10][2]float64
	for i := range pts {
		a := -math.Pi/2 + float64(i)*math.Pi/5
		r := starRadius
		if i%2 == 1 {
			r *= 0.48
		}
		pts[i] = [2]float64{r * math.Cos(a), 3 + r*math.Sin(a)}
	}
	return pts
}()

type starLayer struct {
	colour color.RGBA
	// coverage returns how much of the pixel the layer covers, given the
	// signed distance from the pixel to the star's outline.
	coverage func(sd float64) float64
	offsetX  float64
	offsetY  float64
}

var (
	white = color.RGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	black = color.RGBA{A: 0xff}

	filledStar = []starLayer{
		{colour: withAlpha(black, 0.45), coverage: fill, offsetX: 2, offsetY: 3},
		{colour: withAlpha(black, 0.45), coverage: stroke(starSize * 0.14), offsetX: 2, offsetY: 3},
		{colour: white, coverage: stroke(starSize * 0.11)},
		{colour: starColor, coverage: fill},
	}
	emptyStar = []starLayer{
		{colour: withAlpha(white, 0.10), coverage: fill},
		{colour: withAlpha(white, 0.28), coverage: stroke(starSize * 0.07)},
	}
)

// drawStars draws five stars showing a rating out of ten, with the top-left
// of the row at (x, y).
func drawStars(dst *image.RGBA, x, y float64, rating int) {
	for i := range 5 {
		cx := x + 4 + starSize/2 + float64(i)*starStep
		cy := y + 28
		switch {
		case rating >= (i+1)*2:
			drawStar(dst, cx, cy, starRotations[i], filledStar, false)
		case rating == i*2+1:
			drawStar(dst, cx, cy, starRotations[i], emptyStar, false)
			drawStar(dst, cx, cy, starRotations[i], filledStar, true)
		default:
			drawStar(dst, cx, cy, starRotations[i], emptyStar, false)
		}
	}
}

// drawStar draws a star centred on (cx, cy), rotated by the given degrees.
// If half is set, only the left half (in the star's own frame) is drawn.
func drawStar(dst *image.RGBA, cx, cy, rotation float64, layers []starLayer, half bool) {
	sin, cos := math.Sincos(-rotation * math.Pi / 180)
	extent := int(math.Ceil(starRadius + starSize*0.14 + 6))
	for py := int(cy) - extent; py <= int(cy)+extent; py++ {
		for px := int(cx) - extent; px <= int(cx)+extent; px++ {
			dx, dy := float64(px)+0.5-cx, float64(py)+0.5-cy
			lx, ly := dx*cos-dy*sin, dx*sin+dy*cos

			clip := 1.0
			if half {
				clip = clamp01(0.5 - lx)
			}

			for _, l := range layers {
				blend(dst, px, py, l.colour, l.coverage(starDistance(lx-l.offsetX, ly-l.offsetY))*clip)
			}
		}
	}
}

// starDistance is the signed distance from a point to the star's outline,
// negative inside.
func starDistance(x, y float64) float64 {
	n := len(starPoints)
	d := math.Inf(1)
	inside := false
	for i, j := 0, n-1; i < n; j, i = i, i+1 {
		ax, ay := starPoints[i][0], starPoints[i][1]
		bx, by := starPoints[j][0], starPoints[j][1]

		ex, ey := bx-ax, by-ay
		wx, wy := x-ax, y-ay
		t := clamp01((wx*ex + wy*ey) / (ex*ex + ey*ey))
		d = math.Min(d, math.Hypot(wx-ex*t, wy-ey*t))

		if (ay > y) != (by > y) && x < ax+(y-ay)*ex/ey {
			inside = !inside
		}
	}
	if inside {
		return -d
	}
	return d
}

func fill(sd float64) float64 {
	return clamp01(0.5 - sd)
}

func stroke(width float64) func(float64) float64 {
	return func(sd float64) float64 {
		return clamp01(width/2 + 0.5 - math.Abs(sd))
	}
}

func withAlpha(c color.RGBA, alpha float64) color.RGBA {
	c.A = uint8(math.Round(alpha * 255))
	return c
}
