package main

import (
	"fmt"
	"github.com/eddiejessup/vec"
	"github.com/fogleman/gg"
	"math"
)

type Body struct {
	Core  vec.Circle
	Limbs []float64
}

type Mover struct {
	Body
	Velocity  vec.Vector
	TimeToLay int
}

type Egg struct {
	Body
	TimeToHatch int
}

type RGB struct {
	R, G, B float64
}

var rgbs = []RGB{
	{0, 0, 0},
	{0, 100, 0},
	{70, 10, 30},
}

const (
	X                = 600
	Y                = 600
	R                = 50
	TimeMax          = 100
	IncubationPeriod = 10
	LayingPeriod     = 10
)

func drawWrappedCircle(dc *gg.Context, x, y, r float64) {
	dc.DrawCircle(x, y, r)
	dc.DrawCircle(x+X, y, r)
	dc.DrawCircle(x-X, y, r)
	dc.DrawCircle(x, y+Y, r)
	dc.DrawCircle(x, y-Y, r)
}

func drawWrappedCircleStruct(dc *gg.Context, c vec.Circle) {
	drawWrappedCircle(dc, c.Centre.X, c.Centre.Y, c.R)
}

func getLimbCircle(core vec.Circle, ang float64) vec.Circle {
	return vec.Circle{
		Centre: vec.Vector{
			X: core.Centre.X + 2*core.R*math.Cos(ang),
			Y: core.Centre.Y + 2*core.R*math.Sin(ang),
		},
		R: core.R,
	}
}

func bodyCircles(b Body) []vec.Circle {
	cs := []vec.Circle{b.Core}
	for _, ang := range b.Limbs {
		cs = append(cs, getLimbCircle(b.Core, ang))
	}
	return cs
}

func drawWrappedBody(dc *gg.Context, b Body) {
	for _, c := range bodyCircles(b) {
		drawWrappedCircleStruct(dc, c)
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

func bodiesIntersect(b1, b2 Body) bool {
	for _, c1 := range bodyCircles(b1) {
		for _, c2 := range bodyCircles(b2) {
			if vec.CirclesIntersect(c1, c2) {
				return true
			}
		}
	}
	return false
}

func update(movers *[]Mover, eggs *[]Egg) {
	// Move the movers.
	for i := range *movers {
		moveMover(&((*movers)[i]))
	}

	// Find the killed.
	var moverIndsToDelete []int
	for i := range *movers {
		for j := range *movers {
			if (i != j) && bodiesIntersect((*movers)[i].Body, (*movers)[j].Body) && (moverMomentum((*movers)[i]) > moverMomentum((*movers)[j])) {
				moverIndsToDelete = append(moverIndsToDelete, j)
			}
		}
	}
	// Kill the killed.
	for _, v := range moverIndsToDelete {
		*movers = append((*movers)[:v], (*movers)[v+1:]...)
	}

	// Lay the movers' prepared eggs.
	for i := range *movers {
		mover := &(*movers)[i]
		if mover.TimeToLay > 0 {
			mover.TimeToLay--
		} else {
			// Lay an egg.
			newEgg := Egg{
				Body:        mover.Body,
				TimeToHatch: IncubationPeriod,
			}
			*eggs = append(*eggs, newEgg)
			mover.TimeToLay = LayingPeriod
		}
	}

	// Hatch the incubated eggs.
	var eggIndsToDelete []int
	for i := range *eggs {
		egg := &(*eggs)[i]
		if egg.TimeToHatch > 0 {
			egg.TimeToHatch--
		} else {
			// Check Egg has space to hatch.
			eggIntersectsAMover := false
			for _, mover := range *movers {
				if bodiesIntersect(egg.Body, mover.Body) {
					eggIntersectsAMover = true
				}
			}
			if !eggIntersectsAMover {
				newMover := Mover{
					Body:      egg.Body,
					TimeToLay: IncubationPeriod,
				}
				*movers = append(*movers, newMover)
				eggIndsToDelete = append(eggIndsToDelete, i)
			}
		}
	}
	// Clean up the born's eggs.
	for _, v := range eggIndsToDelete {
		*eggs = append((*eggs)[:v], (*eggs)[v+1:]...)
	}
}

func draw(dc *gg.Context, movers []Mover, eggs []Egg, t int) {
	// Clear image for a new frame.
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	// Draw the scene.
	for i, mover := range movers {
		setRGB(dc, rgbs[i])
		drawWrappedBody(dc, mover.Body)
		dc.Fill()
	}
	for i, egg := range eggs {
		setRGB(dc, rgbs[i])
		drawWrappedBody(dc, egg.Body)
		dc.Stroke()
	}

	// Write the image.
	fileName := fmt.Sprintf("img/out_%0.5d.png", t)
	dc.SavePNG(fileName)
}

func main() {
	movers := []Mover{
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: vec.Vector{100, 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			Velocity:  vec.Vector{X: 5, Y: 5},
			TimeToLay: LayingPeriod,
		},
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: vec.Vector{400, 400}, R: R},
				Limbs: []float64{0.0, -math.Pi / 2},
			},
			Velocity:  vec.Vector{X: -5.1, Y: -5},
			TimeToLay: LayingPeriod,
		},
	}

	eggs := []Egg{
		Egg{
			Body: Body{
				Core:  vec.Circle{Centre: vec.Vector{100, 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			TimeToHatch: IncubationPeriod,
		},
	}

	dc := gg.NewContext(X, Y)
	for t := 0; t < TimeMax; t++ {
		update(&movers, &eggs)
		draw(dc, movers, eggs, t)
	}
}
