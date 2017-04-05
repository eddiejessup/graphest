package main

import (
	"bytes"
	"encoding/binary"
	"errors"
	"flag"
	"fmt"
	"github.com/eddiejessup/vec"
	"github.com/fogleman/gg"
	"github.com/golang/geo/r2"
	"hash/crc32"
	"image/color"
	"log"
	"math"
	"math/rand"
	"os"
	"runtime/pprof"
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
	R                = 25
	TimeMax          = 10000
	DrawEvery        = 1
	doDrawing        = false
	IncubationPeriod = 20
	LayingPeriod     = 20
	InitialSpeed     = 10.0
	Acceleration     = 0.0
	MutationAngle    = 0.5
	RotationAngle    = 0.1
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

func getLimbCircle(core vec.Circle, ang float64) vec.Circle {
	return vec.Circle{
		Centre: core.Centre.Add(UnitPointFromAngle(ang).Mul(2 * core.R)),
		R:      core.R,
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
		drawWrappedCircle(dc, c.Centre.X, c.Centre.Y, c.R)
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

func RandSymmetricFloat64(scale float64) float64 {
	return (rand.Float64() - 0.5) * 2 * scale
}

func UnitPointFromAngle(a float64) r2.Point {
	return r2.Point{X: math.Cos(a), Y: math.Sin(a)}
}

func rotateMover(m *Mover) {
	ang := math.Atan2(m.Velocity.Y, m.Velocity.X)
	ang += RandSymmetricFloat64(RotationAngle)
	m.Velocity = UnitPointFromAngle(ang).Mul(m.Velocity.Norm())
}

func copyBody(b Body) Body {
	newLimbs := make([]float64, len(b.Limbs))
	copy(newLimbs, b.Limbs)

	return Body{
		Core:  b.Core,
		Limbs: newLimbs,
	}
}

func mutate(b Body) Body {
	newB := copyBody(b)
	for i := range newB.Limbs {
		newB.Limbs[i] += RandSymmetricFloat64(MutationAngle)
	}
	return newB
}

func bodyArea(b Body) float64 {
	return float64(len(b.Limbs)) * vec.CircleArea(b.Core)
}

func moverMomentum(m Mover) float64 {
	return m.Velocity.Norm() * bodyArea(m.Body)
}

func wrappedCirclesIntersect(c1, c2 vec.Circle) bool {
	return (vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: -X, Y: -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: -X, Y: 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: -X, Y: +Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: 0, Y: -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: 0, Y: 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: 0, Y: +Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: +X, Y: -Y})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: +X, Y: 0})) ||
		vec.CirclesIntersect(c1, c2.Offset(r2.Point{X: +X, Y: +Y})))
}

func bodiesIntersect(b1, b2 Body) bool {
	for _, c1 := range bodyCircles(b1) {
		for _, c2 := range bodyCircles(b2) {
			if wrappedCirclesIntersect(c1, c2) {
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
		rotateMover(&((*movers)[i]))
	}

	// Find the killed.
	killedSet := make(map[*Mover]bool)
	for i := range *movers {
		for j := range *movers {
			iMover := (*movers)[i]
			jMover := (*movers)[j]
			if i != j && bodiesIntersect(iMover.Body, jMover.Body) {
				unitSep := jMover.Body.Core.Centre.Sub(iMover.Body.Core.Centre).Normalize()
				iVPar := iMover.Velocity.Dot(unitSep)
				jVPar := -jMover.Velocity.Dot(unitSep)
				// iMomPar := bodyArea(iMover.Body) * iVPar
				// jMomPar := bodyArea(jMover.Body) * jVPar
				// fmt.Printf("i Velocity: %v\n", iMover.Velocity)
				// fmt.Printf("j Velocity: %v\n", jMover.Velocity)
				// fmt.Printf("i to j unit separation vector: %v\n", unitSep)
				// fmt.Printf("i Velocity along sep: %v\n", iVPar)
				// fmt.Printf("j Velocity along sep: %v\n", jVPar)
				// fmt.Printf("\n")
				if iVPar == jVPar {
					return errors.New(fmt.Sprintf("Found two movers colliding with equal parallel speeds: %v and %v", iMover, jMover))
				}
				if iVPar > jVPar {
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
				Body:        mutate(mover.Body),
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
					Velocity:  initialVelocity(),
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
	return vec.RandomUnitVector().Mul(rand.Float64() * InitialSpeed)
}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	movers := []Mover{
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{X: 100, Y: 100}, R: R},
				Limbs: []float64{0.0, math.Pi / 2},
			},
			Velocity:  initialVelocity(),
			TimeToLay: LayingPeriod,
		},
		Mover{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{X: 400, Y: 400}, R: R},
				Limbs: []float64{0.0, -math.Pi / 2},
			},
			Velocity:  initialVelocity(),
			TimeToLay: LayingPeriod,
		},
	}

	eggs := []Egg{
		Egg{
			Body: Body{
				Core:  vec.Circle{Centre: r2.Point{X: 100, Y: 100}, R: R},
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
		if doDrawing && t%DrawEvery == 0 {
			draw(dc, movers, eggs, t)
		}
	}
}
