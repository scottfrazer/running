package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"os/exec"

	"github.com/fogleman/gg"
	"github.com/scottfrazer/running/strava"
)

func ResizePNG(pngPath string, width, height int) []byte {
	out, err := exec.Command("convert", "-resize", fmt.Sprintf("%dx%d", width, height), pngPath, "-").Output()
	check(err)
	return out
}

func rectangle(rgba *image.RGBA, topX, topY, bottomX, bottomY, width int, c color.Color) {
	for j := 0; j < width; j++ {
		for i := 0; i <= bottomX-topX+(width-1); i++ {
			rgba.Set(topX+i, topY+j, c)
			rgba.Set(topX+i, bottomY+j, c)
		}
		for i := 0; i <= bottomY-topY+(width-1); i++ {
			rgba.Set(topX+j, topY+i, c)
			rgba.Set(bottomX+j, topY+i, c)
		}
	}
}

type TilePoster struct {
	activities      []strava.SummaryActivity
	rows            int
	cols            int
	width           int
	height          int
	borderWidth     int
	space           int
	rightLeftMargin int
	topBottomMargin int
	mapWidth        int
	mapHeight       int
	mapPath         func(int64) string // activity id => string
}

func NewTilePoster(activities []strava.SummaryActivity, ratio float64, mapWidth, mapHeight, borderWidth, space int, mapPath func(int64) string) *TilePoster {
	var r, c, w, h, margin int

	dimensions := func(c int) (int, int, int) {
		r := int(math.Ceil(float64(len(activities)) / float64(c)))
		return r + 1,
			(c * (mapWidth + (2 * borderWidth))) + ((c + 1) * space),
			(r+1)*(mapHeight+(2*borderWidth)) + (r * space)
	}

	for c = 1; ; c++ {
		// r = len(activities) / c
		// w = c * (mapWidth + borderWidth)
		// h = (r + 1) * (mapHeight + borderWidth)
		r, w, h = dimensions(c)
		fmt.Printf("newCompute(): r=%d, c=%d, w=%d, h=%d, ratio=%f\n", r, c, w, h, float64(w)/float64(h))

		_, w1, h1 := dimensions(c + 1)
		if float64(w1)/float64(h1) > ratio {
			// (w+margin)/h = ratio; solve for `margin`
			margin = int(ratio*float64(h) - float64(w))
			w += margin
			fmt.Printf("margin=%d, w=%d, h=%d, r=%f\n", margin, w, h, float64(w)/float64(h))
			break
		}
	}

	return &TilePoster{
		activities,
		r,
		c,
		w,
		h,
		borderWidth,
		space,
		margin / 2,
		0,
		mapWidth,
		mapHeight,
		mapPath,
	}
}

func (poster *TilePoster) Generate() []byte {
	var partition [][]strava.SummaryActivity
	for r := 0; r < poster.rows; r++ {
		var row []strava.SummaryActivity
		for c := 0; c < poster.cols; c++ {
			index := r*poster.cols + c
			if index >= len(poster.activities) {
				break
			}
			row = append(row, poster.activities[index])
		}
		if len(row) > 0 {
			partition = append(partition, row)
		}
	}

	fmt.Printf("TilePoster:\n")
	fmt.Printf("  activites=%d\n", len(poster.activities))
	fmt.Printf("  rows=%d\n", poster.rows)
	fmt.Printf("  cols=%d\n", poster.cols)
	fmt.Printf("  width=%d\n", poster.width)
	fmt.Printf("  height=%d\n", poster.height)
	fmt.Printf("  borderWidth=%d\n", poster.borderWidth)
	fmt.Printf("  space=%d\n", poster.space)
	fmt.Printf("  rightLeftMargin=%d\n", poster.rightLeftMargin)
	fmt.Printf("  topBottomMargin=%d\n", poster.topBottomMargin)
	fmt.Printf("  mapWidth=%d\n", poster.mapWidth)
	fmt.Printf("  mapHeight=%d\n", poster.mapHeight)

	rgba := image.NewRGBA(image.Rect(0, 0, poster.width, poster.height))

	for i := 0; i < poster.width; i++ {
		for j := 0; j < poster.height; j++ {
			rgba.Set(i, j, color.White)
		}
	}

	for r, row := range partition {
		for c, activity := range row {
			centerAdjust := 0
			if len(row) < poster.cols {
				centerAdjust = (poster.cols - len(row)) * (poster.space + poster.mapWidth)
				centerAdjust /= 2
			}

			topX := poster.rightLeftMargin + centerAdjust + poster.space + (c * (poster.space + poster.mapWidth))
			topY := poster.space + (r * (poster.space + poster.mapHeight))
			bottomX := topX + poster.mapWidth + poster.borderWidth
			bottomY := topY + poster.mapHeight + poster.borderWidth

			// draw border
			rectangle(rgba, topX, topY, bottomX, bottomY, poster.borderWidth, color.RGBA{0, 0, 0, 255})

			// load image
			imageFile, err := os.ReadFile(poster.mapPath(activity.Id))
			check(err)
			activityMap, _, err := image.Decode(bytes.NewReader(imageFile))
			check(err)

			// draw the image within the border
			draw.Draw(
				rgba,
				image.Rect(
					topX+poster.borderWidth,
					topY+poster.borderWidth,
					topX+poster.borderWidth+poster.mapWidth,
					topY+poster.borderWidth+poster.mapHeight,
				),
				activityMap,
				image.Point{0, 0},
				draw.Src,
			)
		}
	}

	/////////////////////////////////////////////////////

	dc := gg.NewContextForRGBA(rgba)
	err := dc.LoadFontFace("Roboto-Regular.ttf", 1000)
	check(err)
	dc.SetRGB(0, 0, 0)
	rowHeight := float64(poster.height) / float64(poster.rows)
	//sw, sh := dc.MeasureString("2020")
	//fmt.Printf("%fx%f\n", sw, sh)
	dc.DrawStringAnchored(
		"2022",
		float64(poster.width)/2,
		float64((poster.rows-1)*(poster.mapHeight+2*poster.borderWidth))+(float64(poster.rows)*float64(poster.space))+rowHeight/2.0-(float64(poster.space)*2),
		0.5,
		0.5,
	)

	// fontBytes, err := ioutil.ReadFile("Roboto-Regular.ttf")
	// check(err)
	// font, err := freetype.ParseFont(fontBytes)
	// check(err)
	// size := 1000.0

	// c := freetype.NewContext()
	// c.SetDPI(72)
	// c.SetFont(font)
	// c.SetFontSize(size)
	// c.SetClip(rgba.Bounds())
	// c.SetDst(rgba)
	// c.SetSrc(image.Black)
	// fmt.Printf("%v\n", rgba.Bounds())

	// x := (poster.width / 2) + int(c.PointToFixed(size)>>6)
	// y := (poster.rows - 2) * poster.mapHeight
	// fmt.Printf("%d, %d\n", x, y)
	// pt := freetype.Pt(x, y)
	// _, err = c.DrawString("2020", pt)
	// check(err)
	// pt.Y += c.PointToFixed(size * 1.5)
	/////////////////////////////////////////////////////

	var buf bytes.Buffer
	dc.EncodePNG(&buf)
	//png.Encode(&buf, rgba)
	return buf.Bytes()
}
