package main

import (
	"fmt"
	"errors"
	"log"
	"github.com/eddiejessup/vec"
	"github.com/fogleman/gg"
	"github.com/golang/geo/r2"
	"math"
	"math/rand"
	"image/color"
	"bytes"
	"encoding/binary"
	"hash/crc32"
)

type Body struct {
	Core  vec.Circle
	Limbs []float64
}

type Mover struct {
	Body
	Velocity  r2.Point
	TimeToLay int
}

type Egg struct {
	Body
	TimeToHatch int
}

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

const (
	X                = 600
	Y                = 600
	R                = 50
	TimeMax          = 100
	IncubationPeriod = 10
	LayingPeriod     = 10
	V = 10
	SmallNumber = 1e-3
	Acceleration = 1.0
)

func drawWrappedCircle(dc *gg.Context, x, y, r float64) {
	dc.DrawCircle(x, y, r)
	dc.DrawCircle(x, y-Y, r)
	dc.DrawCircle(x, y+Y, r)

	dc.DrawCircle(x-X, y, r)
	dc.DrawCircle(x-X, y-Y, r)
	dc.DrawCircle(x-X, y+Y, r)

	dc.DrawCircle(x+X, y, r)
	dc.DrawCircle(x+X, y-Y, r)
	dc.DrawCircle(x+X, y+Y, r)
}

func drawWrappedCircleStruct(dc *gg.Context, c vec.Circle) {
	drawWrappedCircle(dc, c.Centre.X, c.Centre.Y, c.R)
}

func getLimbCircle(core vec.Circle, ang float64) vec.Circle {
	return vec.Circle{
		Centre: r2.Point{
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

func setRGB(dc *gg.Context, color color.Color) {
	r, g, b, _ := color.RGBA()
	dc.SetRGB(float64(r), float64(g), float64(b))
}

func moveMover(m *Mover) {
	m.Body.Core.Centre = m.Body.Core.Centre.Add(m.Velocity)
	// Wrap.
	if m.Body.Core.Centre.X > X {
		m.Body.Core.Centre.X -= X
	} else if m.Body.Core.Centre.X < 0 {
		m.Body.Core.Centre.X += X
	}
	if m.Body.Core.Centre.Y > Y {
		m.Body.Core.Centre.Y -= Y
	} else if m.Body.Core.Centre.Y < 0 {
		m.Body.Core.Centre.Y += Y
	}
}

func bodyArea(b Body) float64 {
	return float64(len(b.Limbs)) * vec.CircleArea(b.Core)
}

func moverMomentum(m Mover) float64 {
	return m.Velocity.Norm() * bodyArea(m.Body)
}


func WrappedCirclesIntersect(c1, c2 vec.Circle) bool {
	return (
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{-X, -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{-X, 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{-X, +Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{0, -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{0, 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{0, +Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{+X, -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{+X, 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{+X, +Y})))
}

func bodiesIntersect(b1, b2 Body) bool {
	for _, c1 := range bodyCircles(b1) {
		for _, c2 := range bodyCircles(b2) {
			if WrappedCirclesIntersect(c1, c2) {
				return true
			}
		}
	}
	return false
}

func update(movers *[]Mover, eggs *[]Egg) error {
	// Move the movers.
	for i := range *movers {
		moveMover(&((*movers)[i]))
	}

	// Find the killed.
	killedSet := make(map[*Mover]bool)
	for i := range *movers {
		for j := range *movers {
			if (i != j) {
				iMom := moverMomentum((*movers)[i])
				jMom := moverMomentum((*movers)[j])
				if iMom == jMom {
					return errors.New(fmt.Sprintf("Found two movers with equal momenta: %v and %v", iMom, jMom))
				}
				if bodiesIntersect((*movers)[i].Body, (*movers)[j].Body) && (iMom > jMom) {
					killedSet[&(*movers)[j]] = true
				}
			}
		}
	}
	// Kill the killed.
	var stillMovers []Mover
	for i := range *movers {
		mover := &(*movers)[i]
		if !killedSet[mover] {
			stillMovers = append(stillMovers, *mover)
		}
	}
	*movers = stillMovers

	// Accelerate the movers.
	for i := range *movers {
		mover := &(*movers)[i]
		mover.Velocity = mover.Velocity.Add(mover.Velocity.Normalize().Mul(Acceleration))
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
	hatchedSet := make(map[*Egg]bool)
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
					Velocity: initialVelocity(),
					TimeToLay: IncubationPeriod,
				}
				*movers = append(*movers, newMover)
				hatchedSet[egg] = true
			}
		}
	}

	// Clean up the born's eggs.
	var stillEggs []Egg
	for i := range *eggs {
		egg := &(*eggs)[i]
		if !hatchedSet[egg] {
			stillEggs = append(stillEggs, *egg)
		}
	}
	*eggs = stillEggs
	return nil
}

func draw(dc *gg.Context, movers []Mover, eggs []Egg, t int) {
	// Clear image for a new frame.
	dc.SetRGBA(0, 0, 0, 0)
	dc.Clear()

	// Draw the scene.
	for _, mover := range movers {
		colorIndex := hashBody(mover.Body) % 10
		setRGB(dc, palette[colorIndex])
		drawWrappedBody(dc, mover.Body)
		dc.Fill()
	}
	for _, egg := range eggs {
		colorIndex := hashBody(egg.Body) % 10
		setRGB(dc, palette[colorIndex])
		drawWrappedBody(dc, egg.Body)
		dc.Stroke()
	}

	// Write the image.
	fileName := fmt.Sprintf("img/out_%0.5d.png", t)
	dc.SavePNG(fileName)
}


func hashBody(body Body) int {
	var buf bytes.Buffer
	binary.Write(&buf, binary.LittleEndian, body.Limbs)
	return int(crc32.ChecksumIEEE(buf.Bytes()))

}


func initialVelocity() r2.Point {
	return vec.RandomUnitVector().Mul(rand.Float64() * SmallNumber)
}


func main() {
	movers := []Mover{
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{100, 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			Velocity:  initialVelocity(),
			TimeToLay: LayingPeriod,
		},
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{400, 400}, R: R},
				Limbs: []float64{0.0, -math.Pi / 2},
			},
			Velocity: initialVelocity(),
			TimeToLay: LayingPeriod,
		},
	}

	eggs := []Egg{
		Egg{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{100, 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			TimeToHatch: IncubationPeriod,
		},
	}

	dc := gg.NewContext(X, Y)
	for t := 0; t < TimeMax; t++ {
		err := update(&movers, &eggs)
		if err != nil {
	        log.Fatal(err)
		}
		draw(dc, movers, eggs, t)
	}
}
