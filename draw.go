package main

import (
	"fmt"
    "image/color"
    "bytes"
    "encoding/binary"
    "hash/crc32"
    "math"
	"github.com/fogleman/gg"
    "github.com/golang/geo/r2"
    "github.com/eddiejessup/vec"
)

var palette = color.Palette{
	color.RGBA{141, 211, 199, 0},
	color.RGBA{255, 255, 179, 0},
	color.RGBA{190, 186, 218, 0},
	color.RGBA{251, 128, 114, 0},
	color.RGBA{128, 177, 211, 0},
	color.RGBA{253, 180, 98, 0},
	color.RGBA{179, 222, 105, 0},
	color.RGBA{252, 205, 229, 0},
	color.RGBA{217, 217, 217, 0},
	color.RGBA{188, 128, 189, 0},
	color.RGBA{204, 235, 197, 0},
	color.RGBA{255, 237, 111, 0},
}

func setRGB(dc *gg.Context, color color.Color) {
	r, g, b, _ := color.RGBA()
	dc.SetRGB(float64(r), float64(g), float64(b))
}

func Hash(b Body, w r2.Point) int {
    angs := b.LimbAngles(w)
    angsRound := make([]int64, len(angs))
    for i, ang := range angs {
        // Get a number between zero and one.
        angRound := ((ang / (2 * math.Pi)) + 0.5)
        // Get a number between zero and ten.
        angsRound[i] = int64(vec.Round(10 * angRound))
    }
    var buf bytes.Buffer
    binary.Write(&buf, binary.LittleEndian, angsRound)
    // fmt.Printf("Angsround: %v\n", angsRound)
    return int(crc32.ChecksumIEEE(buf.Bytes()))
}

func drawWrappedCircle(dc *gg.Context, c *vec.Circle, w r2.Point) {
    dc.DrawCircle(c.Centre.X, c.Centre.Y, c.R)
    dc.DrawCircle(c.Centre.X, c.Centre.Y-w.Y, c.R)
    dc.DrawCircle(c.Centre.X, c.Centre.Y+w.Y, c.R)

    dc.DrawCircle(c.Centre.X-w.X, c.Centre.Y, c.R)
    dc.DrawCircle(c.Centre.X-w.X, c.Centre.Y-w.Y, c.R)
    dc.DrawCircle(c.Centre.X-w.X, c.Centre.Y+w.Y, c.R)

    dc.DrawCircle(c.Centre.X+w.X, c.Centre.Y, c.R)
    dc.DrawCircle(c.Centre.X+w.X, c.Centre.Y-w.Y, c.R)
    dc.DrawCircle(c.Centre.X+w.X, c.Centre.Y+w.Y, c.R)
}

func drawWrappedBody(dc *gg.Context, b *Body, w r2.Point) {
    for _, c := range b.Circles() {
        drawWrappedCircle(dc, c, w)
    }
}



func draw(dc *gg.Context, movers []Mover, eggs []Egg, w r2.Point, t int) {
	// Clear image for a new frame.
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	// Draw the scene.
	for _, mover := range movers {
        colorIndex := Hash(mover.Body, w) % 10
        // fmt.Printf("Hash: %v\n", Hash(mover.Body, w))
        // fmt.Printf("Color index: %v\n", colorIndex)
		setRGB(dc, palette[colorIndex])
		drawWrappedBody(dc, &mover.Body, w)
		dc.Fill()
	}
	for _, egg := range eggs {
		colorIndex := Hash(egg.Body, w) % 10
		setRGB(dc, palette[colorIndex])
		drawWrappedBody(dc, &egg.Body, w)
		dc.Stroke()
	}

	// Write the image.
	fileName := fmt.Sprintf("img/out_%0.5d.png", t)
	dc.SavePNG(fileName)
}
    // dc := gg.NewContext(int(W.X), int(W.Y))
    // "github.com/fogleman/gg"
