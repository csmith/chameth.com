package opengraph

import (
	"bytes"
	_ "embed"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"strings"

	"github.com/anthonynsimon/bild/blur"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/f64"
	"golang.org/x/image/math/fixed"
)

const (
	Width  = 1200
	Height = 630

	// Version identifies the design; bump it to regenerate every image.
	Version = 1
)

const (
	marginX     = 80
	textWidth   = 710
	posterX     = 860
	posterWidth = 260
	posterAngle = 4.0

	quoteSize       = 52
	quoteLineHeight = 1.3
	quoteMaxLines   = 3
	titleSize       = 64
	titleMaxLines   = 2

	bylineGap      = 36
	bylineHeight   = 56
	starsWidth     = 232
	starsTextGap   = 24
	bylineMonoSize = 26
	bylineBoldSize = 34
	// Titles too long for the byline shrink to this size before truncating.
	bylineMinBoldSize = 24
	brandSize         = 22
	brandBottom       = 36
)

var (
	background = color.RGBA{R: 0x15, G: 0x15, B: 0x1a, A: 0xff}
	quoteColor = color.RGBA{R: 0xef, G: 0xef, B: 0xff, A: 0xff}
	boldColor  = color.RGBA{R: 0xd4, G: 0xd4, B: 0xde, A: 0xff}
	altColor   = color.RGBA{R: 0x90, G: 0x90, B: 0x9a, A: 0xff}
	starColor  = color.RGBA{R: 0xbd, G: 0x9d, B: 0x07, A: 0xff}
	edgeColor  = color.RGBA{R: 0x20, G: 0x20, B: 0x26, A: 0xff}
)

var (
	//go:embed Charis-Italic.ttf
	italicFontData []byte
	//go:embed Charis-Bold.ttf
	boldFontData []byte
	//go:embed BerkeleyMono-Regular.ttf
	monoFontData []byte

	italicFont = mustParseFont(italicFontData)
	boldFont   = mustParseFont(boldFontData)
	monoFont   = mustParseFont(monoFontData)
)

type Film struct {
	Title string
	Year  string
	// Poster is optional.
	Poster       image.Image
	TimesWatched int
	// Rating is out of 10; zero means unrated.
	Rating int
	// Excerpt is plain text, truncated to fit. When empty the title is
	// shown in its place.
	Excerpt string
}

// Render draws the OpenGraph image for a film, returning it as a JPEG.
func Render(f Film) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, Width, Height))
	draw.Draw(img, img.Bounds(), image.NewUniform(background), image.Point{}, draw.Src)

	if f.Poster != nil {
		drawBackdrop(img, f.Poster)
	}

	drawText(img, f)

	if f.Poster != nil {
		drawPoster(img, f.Poster)
	}

	b := &bytes.Buffer{}
	if err := jpeg.Encode(b, img, &jpeg.Options{Quality: 90}); err != nil {
		return nil, fmt.Errorf("failed to encode image: %w", err)
	}
	return b.Bytes(), nil
}

// drawBackdrop fills the image with a heavily blurred and darkened copy of
// the poster. The blur happens at a tenth of the size, which is
// indistinguishable at this radius and far cheaper.
func drawBackdrop(dst *image.RGBA, poster image.Image) {
	const (
		scale = 10
		bleed = 80
	)

	small := image.NewRGBA(image.Rect(0, 0, (Width+2*bleed)/scale, (Height+2*bleed)/scale))
	draw.ApproxBiLinear.Scale(small, small.Bounds(), poster, coverRect(poster.Bounds(), small.Bounds()), draw.Src, nil)

	// bild's kernel has a variance of 2*radius, so this is a 60px standard
	// deviation at full size.
	sigma := 60.0 / scale
	blurred := blur.Gaussian(small, sigma*sigma/2)

	inner := image.Rect(bleed/scale, bleed/scale, (Width+bleed)/scale, (Height+bleed)/scale)
	draw.BiLinear.Scale(dst, dst.Bounds(), blurred, inner, draw.Src, nil)

	const (
		// CSS's brightness and saturate filters.
		brightness = 0.22
		s          = 1.2
	)
	for i := 0; i < len(dst.Pix); i += 4 {
		r := float64(dst.Pix[i]) * brightness
		g := float64(dst.Pix[i+1]) * brightness
		b := float64(dst.Pix[i+2]) * brightness
		dst.Pix[i] = clampByte((0.213+0.787*s)*r + (0.715-0.715*s)*g + (0.072-0.072*s)*b)
		dst.Pix[i+1] = clampByte((0.213-0.213*s)*r + (0.715+0.285*s)*g + (0.072-0.072*s)*b)
		dst.Pix[i+2] = clampByte((0.213-0.213*s)*r + (0.715-0.715*s)*g + (0.072+0.928*s)*b)
		dst.Pix[i+3] = 0xff
	}
}

// drawPoster draws the poster in the right column, tilted, with rounded
// corners, a thin border and a drop shadow.
func drawPoster(dst *image.RGBA, poster image.Image) {
	const (
		radius = 6.0
		pad    = 40
	)

	pb := poster.Bounds()
	w := posterWidth
	h := int(math.Round(float64(w) * float64(pb.Dy()) / float64(pb.Dx())))

	layer := image.NewRGBA(image.Rect(0, 0, w+2*pad, h+2*pad))

	shadow := image.NewRGBA(layer.Bounds())
	paintRoundedRect(
		shadow,
		pad+3,
		pad+3,
		w,
		h,
		radius,
		func(int, int) color.RGBA {
			return color.RGBA{A: 0xff}
		},
	)
	draw.Draw(layer, layer.Bounds(), blur.Gaussian(shadow, 6*6/2), image.Point{}, draw.Src)

	scaled := image.NewRGBA(image.Rect(0, 0, w-2, h-2))
	draw.CatmullRom.Scale(scaled, scaled.Bounds(), poster, pb, draw.Src, nil)
	paintRoundedRect(
		layer,
		pad,
		pad,
		w,
		h,
		radius,
		func(x, y int) color.RGBA {
			if x < 1 || y < 1 || x >= w-1 || y >= h-1 {
				return edgeColor
			}
			return scaled.RGBAAt(x-1, y-1)
		},
	)

	cx := float64(posterX) + float64(w)/2
	cy := float64(Height) / 2
	lx := float64(layer.Bounds().Dx()) / 2
	ly := float64(layer.Bounds().Dy()) / 2
	sin, cos := math.Sincos(posterAngle * math.Pi / 180)
	draw.BiLinear.Transform(
		dst,
		f64.Aff3{
			cos, -sin, cx - cos*lx + sin*ly,
			sin, cos, cy - sin*lx - cos*ly,
		},
		layer,
		layer.Bounds(),
		draw.Over,
		nil,
	)
}

// paintRoundedRect composites an anti-aliased rounded rectangle at (x0, y0)
// onto dst, coloured per-pixel by fill (given rectangle-relative coords).
func paintRoundedRect(dst *image.RGBA, x0, y0, w, h int, radius float64, fill func(x, y int) color.RGBA) {
	hw, hh := float64(w)/2, float64(h)/2
	for y := range h {
		for x := range w {
			// Signed distance from the pixel centre to the rounded edge.
			qx := math.Abs(float64(x)+0.5-hw) - (hw - radius)
			qy := math.Abs(float64(y)+0.5-hh) - (hh - radius)
			sd := math.Hypot(math.Max(qx, 0), math.Max(qy, 0)) + math.Min(math.Max(qx, qy), 0) - radius
			blend(dst, x0+x, y0+y, fill(x, y), clamp01(0.5-sd))
		}
	}
}

// drawText draws the review excerpt (or the title, if there's no excerpt)
// with a byline beneath it, vertically centred, and the site footer.
func drawText(dst *image.RGBA, f Film) {
	face := newFace(italicFont, quoteSize)
	lineHeight := quoteSize * quoteLineHeight
	lines := wrap(face, f.Excerpt, quoteMaxLines, "“", "”")
	quoted := len(lines) > 0
	if !quoted {
		face = newFace(boldFont, titleSize)
		lineHeight = fixedToFloat(face.Metrics().Height)
		lines = wrap(face, f.Title, titleMaxLines, "", "")
	}

	top := (Height - (float64(len(lines))*lineHeight + bylineGap + bylineHeight)) / 2

	m := face.Metrics()
	halfLeading := (lineHeight - fixedToFloat(m.Ascent+m.Descent)) / 2
	for i, line := range lines {
		drawString(
			dst,
			face,
			quoteColor,
			marginX,
			top+float64(i)*lineHeight+halfLeading+fixedToFloat(m.Ascent),
			line,
		)
	}

	drawByline(dst, f, top+float64(len(lines))*lineHeight+bylineGap, quoted)

	brand := "chameth.com"
	if f.TimesWatched > 0 {
		brand += " · watched " + timesWatched(f.TimesWatched)
	}
	brandFace := newFace(monoFont, brandSize)
	drawString(
		dst,
		brandFace,
		altColor,
		marginX,
		Height-brandBottom-fixedToFloat(brandFace.Metrics().Descent),
		brand,
	)
}

// drawByline draws the stars and, if the main text is a quote, the film's
// title. Otherwise the title is already shown so just the year follows.
func drawByline(dst *image.RGBA, f Film, top float64, quoted bool) {
	x := float64(marginX)
	if f.Rating > 0 {
		drawStars(dst, x, top, f.Rating)
		x += starsWidth + starsTextGap
	}

	monoFace := newFace(monoFont, bylineMonoSize)
	year := ""
	if f.Year != "" {
		year = "(" + f.Year + ")"
	}

	if !quoted {
		drawString(dst, monoFace, altColor, x, bylineBaseline(top, monoFace), year)
		return
	}

	if year != "" {
		year = " " + year
	}
	prefix := "on "
	titleFace, title := fitTitle(f.Title, marginX+textWidth-x-measure(monoFace, prefix+year))

	baseline := bylineBaseline(top, monoFace, titleFace)
	x = drawString(dst, monoFace, altColor, x, baseline, prefix)
	x = drawString(dst, titleFace, boldColor, x, baseline, title)
	drawString(dst, monoFace, altColor, x, baseline, year)
}

// bylineBaseline positions the baseline as a browser would for a line of
// text containing each face, vertically centred in the byline row.
func bylineBaseline(top float64, faces ...font.Face) float64 {
	var above, below float64
	for _, face := range faces {
		m := face.Metrics()
		gap := fixedToFloat(m.Height-m.Ascent-m.Descent) / 2
		above = math.Max(above, fixedToFloat(m.Ascent)+gap)
		below = math.Max(below, fixedToFloat(m.Descent)+gap)
	}
	return top + bylineHeight/2 + (above-below)/2
}

// wrap breaks text into at most maxLines lines that fit the text column,
// adding prefix and suffix around the whole text. If the text doesn't fit,
// the last line ends with an ellipsis instead.
func wrap(face font.Face, text string, maxLines int, prefix, suffix string) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	words[0] = prefix + words[0]
	words[len(words)-1] += suffix

	var lines []string
	line := ""
	for _, word := range words {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if line == "" || measure(face, candidate) <= textWidth {
			line = candidate
			continue
		}

		if len(lines) == maxLines-1 {
			return append(lines, ellipsise(face, line, suffix))
		}
		lines = append(lines, line)
		line = word
	}
	return append(lines, line)
}

// ellipsise ends line with an ellipsis and suffix, dropping words from the
// end of the line until it fits.
func ellipsise(face font.Face, line, suffix string) string {
	for {
		trimmed := strings.TrimRight(line, " ,;:.-–—")
		candidate := trimmed + "…" + suffix
		last := strings.LastIndex(trimmed, " ")
		if measure(face, candidate) <= textWidth || last < 0 {
			return candidate
		}
		line = trimmed[:last]
	}
}

// fitTitle shrinks the byline title until it fits the space available,
// then truncates it if it still doesn't.
func fitTitle(title string, available float64) (font.Face, string) {
	for size := float64(bylineBoldSize); ; size -= 2 {
		face := newFace(boldFont, size)
		if measure(face, title) <= available {
			return face, title
		}
		if size <= bylineMinBoldSize {
			runes := []rune(title)
			for len(runes) > 1 && measure(face, string(runes)+"…") > available {
				runes = runes[:len(runes)-1]
			}
			return face, strings.TrimSpace(string(runes)) + "…"
		}
	}
}

func timesWatched(n int) string {
	switch n {
	case 1:
		return "once"
	case 2:
		return "twice"
	default:
		return fmt.Sprintf("%d times", n)
	}
}

func drawString(dst *image.RGBA, face font.Face, c color.RGBA, x, baseline float64, s string) float64 {
	d := font.Drawer{
		Dst:  dst,
		Src:  image.NewUniform(c),
		Face: face,
		Dot:  fixed.Point26_6{X: floatToFixed(x), Y: floatToFixed(baseline)},
	}
	d.DrawString(s)
	return fixedToFloat(d.Dot.X)
}

func measure(face font.Face, s string) float64 {
	return fixedToFloat(font.MeasureString(face, s))
}

// newFace panics on error, which only happens for invalid options.
func newFace(f *opentype.Font, size float64) font.Face {
	face, err := opentype.NewFace(f, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingNone})
	if err != nil {
		panic(err)
	}
	return face
}

func mustParseFont(data []byte) *opentype.Font {
	f, err := opentype.Parse(data)
	if err != nil {
		panic(err)
	}
	return f
}

// coverRect returns the centred region of src with the same aspect ratio as
// dst, like CSS's object-fit: cover.
func coverRect(src, dst image.Rectangle) image.Rectangle {
	sw, sh := float64(src.Dx()), float64(src.Dy())
	target := float64(dst.Dx()) / float64(dst.Dy())
	if sw/sh > target {
		w := int(sh * target)
		x := src.Min.X + (src.Dx()-w)/2
		return image.Rect(x, src.Min.Y, x+w, src.Max.Y)
	}
	h := int(sw / target)
	y := src.Min.Y + (src.Dy()-h)/2
	return image.Rect(src.Min.X, y, src.Max.X, y+h)
}

// blend composites c over the pixel at (x, y) with the given coverage.
func blend(dst *image.RGBA, x, y int, c color.RGBA, coverage float64) {
	if coverage <= 0 || !(image.Point{X: x, Y: y}.In(dst.Rect)) {
		return
	}
	a := float64(c.A) / 255 * coverage
	i := dst.PixOffset(x, y)
	// c is non-premultiplied here; dst is premultiplied.
	dst.Pix[i] = clampByte(float64(c.R)*a + float64(dst.Pix[i])*(1-a))
	dst.Pix[i+1] = clampByte(float64(c.G)*a + float64(dst.Pix[i+1])*(1-a))
	dst.Pix[i+2] = clampByte(float64(c.B)*a + float64(dst.Pix[i+2])*(1-a))
	dst.Pix[i+3] = clampByte(255*a + float64(dst.Pix[i+3])*(1-a))
}

func clamp01(v float64) float64 {
	return math.Max(0, math.Min(1, v))
}

func clampByte(v float64) uint8 {
	return uint8(math.Max(0, math.Min(255, math.Round(v))))
}

func fixedToFloat(v fixed.Int26_6) float64 {
	return float64(v) / 64
}

func floatToFixed(v float64) fixed.Int26_6 {
	return fixed.Int26_6(math.Round(v * 64))
}
