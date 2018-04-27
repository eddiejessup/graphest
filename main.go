package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang/geo/r2"
	"log"
	"math"
	"os"
	"runtime/pprof"
)

type Egg struct {
	Body
	TimeToHatch int
}

const (
	TimeMax          = 100
	R = 25
	IncubationPeriod = 20
	LayingPeriod     = 20
	InitialSpeed     = 10
	Acceleration     = 0
	MutationAngle    = 0.5
	RotationAngle    = 0.1
	OutputEvery = 5
)

// var W = r2.Point{X: 600, Y: 600}
var W = r2.Point{X: 6000, Y: 6000}

func update(movers *[]Mover, eggs *[]Egg) error {
	// Move the movers.
	for i := range *movers {
		m := &((*movers)[i])
		m.Move(W)
		m.Rotate(RotationAngle)
	}

	// Find the killed.
	killedSet := make(map[*Mover]bool)
	for i := range *movers {
		for j := range *movers {
			iMover := (*movers)[i]
			jMover := (*movers)[j]
			if i != j && iMover.Body.Intersects(&jMover.Body, W) {
				unitSep := jMover.Body.Core.Centre.Sub(iMover.Body.Core.Centre).Normalize()
				iVPar := iMover.Velocity.Dot(unitSep)
				jVPar := -jMover.Velocity.Dot(unitSep)
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
				Body:        *mover.Body.Mutate(MutationAngle, W),
				TimeToHatch: IncubationPeriod,
			}
			*eggs = append(*eggs, newEgg)
			mover.TimeToLay = LayingPeriod
		}
	}

	// Hatch the incubated eggs.
	// TODO: Can just implement this as a slice.
	hatchedSet := make(map[*Egg]bool)
	for i := range *eggs {
		egg := &(*eggs)[i]
		if egg.TimeToHatch > 0 {
			egg.TimeToHatch--
		} else {
			// Check Egg has space to hatch.
			eggIntersectsAMover := false
			for _, mover := range *movers {
				if egg.Body.Intersects(&mover.Body, W) {
					eggIntersectsAMover = true
				}
			}
			if !eggIntersectsAMover {
				newMover := NewMover(
					egg.Body,
					InitialSpeed,
					LayingPeriod,
				)
				*movers = append(*movers, *newMover)
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
		*NewMover(
			*NewBody(
				r2.Point{X: 100, Y: 100},
				R,
				[]float64{0.0, math.Pi / 2},
			),
			InitialSpeed,
			LayingPeriod,
		),
	}

	eggs := []Egg{}

	for t := 0; t < TimeMax; t++ {
		fmt.Printf("Time: %v\n", t)
		fmt.Printf("Number of movers: %v\n", len(movers))
		err := update(&movers, &eggs)
		if err != nil {
			log.Fatal(err)
		}
		if t%OutputEvery == 0 {
			fileName := fmt.Sprintf("dat/out_%0.5d.tsv", t)
		    f, err := os.Create(fileName)
		    if err != nil {
		    	log.Fatal(err)
		    }
			output(f, movers, eggs, W, fileName)
		    f.Close()
		}
	}
}
