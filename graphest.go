package main

import (
	"fmt"
	"math"
	"github.com/fogleman/gg"
	"github.com/eddiejessup/vec"
)

type Body struct {
	Core vec.Circle
	Limbs []float64
}

type Mover struct {
	Body    Body
	Velocity vec.Vector
}

type RGB struct {
	R, G, B float64
}

var rgbs = []RGB{
	{0, 0, 0},
	{0, 100, 0},
}

const (
	X       = 600
	Y       = 600
	R       = 50
	TimeMax = 100
)

func drawWrappedCircle(dc *gg.Context, x, y, r float64) {
	dc.DrawCircle(x, y, r)
	dc.DrawCircle(x+X, y, r)
	dc.DrawCircle(x-X, y, r)
	dc.DrawCircle(x, y+Y, r)
	dc.DrawCircle(x, y-Y, r)
	dc.Fill()
}

func drawWrappedCircleStruct(dc *gg.Context, c vec.Circle) {
	drawWrappedCircle(dc, c.Centre.X, c.Centre.Y, c.R)
}

func getLimbCircle(core vec.Circle, ang float64) vec.Circle {
	return vec.Circle{
		Centre: vec.Vector{
			X: core.Centre.X + 2 * core.R * math.Cos(ang),
			Y: core.Centre.Y + 2 * core.R * math.Sin(ang),
		},
		R: core.R,
	}
}

func drawWrappedBody(dc *gg.Context, b Body) {
	drawWrappedCircleStruct(dc, b.Core)
	for _, ang := range b.Limbs {
		drawWrappedCircleStruct(dc, getLimbCircle(b.Core, ang))
	}
}

func setRGB(dc *gg.Context, rgb RGB) {
	dc.SetRGB(rgb.R, rgb.G, rgb.B)
}

func moveMover(m *Mover) {
	m.Body.Core.Centre.X += m.Velocity.X
	m.Body.Core.Centre.Y += m.Velocity.Y
	if m.Body.Core.Centre.X > X {
		m.Body.Core.Centre.X -= X
	}
	if m.Body.Core.Centre.Y > Y {
		m.Body.Core.Centre.Y -= Y
	}
}

func bodyArea(b Body) float64 {
	return float64(len(b.Limbs)) * vec.CircleArea(b.Core)
}

func moverMomentum(m Mover) float64 {
	return vec.VectorMag(m.Velocity) * bodyArea(m.Body)
}

func main() {
	var movers = []Mover{
		Mover{
			Body: Body{
				Core: vec.Circle{Centre: vec.Vector{100, 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			Velocity: vec.Vector{X: 5, Y: 5},
		},
		Mover{
			Body: Body{
				Core: vec.Circle{Centre: vec.Vector{400, 400}, R: R},
				Limbs: []float64{0.0, -math.Pi / 2},
			},
			Velocity: vec.Vector{X: -5.1, Y: -5},
		},
	}
	var moverIndsToDelete []int

	ggContext := gg.NewContext(X, Y)
	for t := 0; t < TimeMax; t++ {
		// Clear image for a new frame.
		ggContext.SetRGBA(0, 0, 0, 0)
		ggContext.Clear()

		for i := range movers {
			setRGB(ggContext, rgbs[i])
			drawWrappedBody(ggContext, movers[i].Body)
		}

		// Write the image.
		fileName := fmt.Sprintf("img/out_%0.5d.png", t)
		ggContext.SavePNG(fileName)

		// Update the scene.
		for i := range movers {
			moveMover(&(movers[i]))
		}
		moverIndsToDelete = moverIndsToDelete[:0]
		for i := 0; i < len(movers); i++ {
			for j := 0; j < len(movers); j++ {
				if (i != j) && vec.CirclesIntersect(movers[i].Body.Core, movers[j].Body.Core) && (moverMomentum(movers[i]) > moverMomentum(movers[j])) {
					moverIndsToDelete = append(moverIndsToDelete, j)
				}
			}
		}
		// fmt.Printf("t: %v, nr movers: %v\n", t, len(movers))
		for _, v := range moverIndsToDelete {
			movers = append(movers[:v], movers[v+1:]...)
		}
	}
}
