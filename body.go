package main

import (
	"github.com/eddiejessup/vec"
	"github.com/golang/geo/r2"
)

type Body struct {
	Core  vec.Circle
	Limbs []vec.Circle
}

func NewBody(centre r2.Point, r float64, angs []float64) *Body {
	b := &Body{Core: vec.Circle{Centre: centre, R: r}}
	for _, ang := range angs {
		b.Limbs = append(b.Limbs, getAdjacentCircle(&b.Core, ang))
	}
	return b
}

func (b *Body) Copy() *Body {
	newCore := b.Core
	newLimbs := make([]vec.Circle, len(b.Limbs))
	copy(newLimbs, b.Limbs)
	return &Body{
		Core:  newCore,
		Limbs: newLimbs,
	}
}

func (b *Body) Circles() []*vec.Circle {
	cs := make([]*vec.Circle, len(b.Limbs)+1)
	cs[0] = &(b.Core)
	for i := range b.Limbs {
		cs[1+i] = &(b.Limbs[i])
	}
	return cs
}

func (b *Body) LimbAngles(w r2.Point) []float64 {
	angs := make([]float64, len(b.Limbs))
	for i, limbC := range b.Limbs {
		angs[i] = vec.PointAngle(vec.SubWrap(limbC.Centre, b.Core.Centre, w))
	}
	return angs
}

func (b *Body) Mutate(ang float64, w r2.Point) *Body {
	newB := b.Copy()
	for i, limbC := range newB.Limbs {
		uR := vec.SubWrap(limbC.Centre, b.Core.Centre, w)
		vR := vec.RotateVector2d(uR, vec.RandSymmetricFloat64(ang))
		v := newB.Core.Centre.Add(vR)
		newB.Limbs[i].Centre = v
		vec.WrapPoint(&(newB.Limbs[i].Centre), w)
	}
	return newB
}

func (b *Body) Area() float64 {
	return float64(len(b.Limbs)) * b.Core.Area()
}

func (b *Body) Intersects(bO *Body, w r2.Point) bool {
	maxBoundingB := b.Core.R + b.Limbs[0].R
	maxBoundingBO := bO.Core.R + bO.Limbs[0].R
	if !vec.WrappedSquaresIntersect(b.Core.Centre, bO.Core.Centre, maxBoundingB, maxBoundingBO, w) {
		return false
	}
	for _, c := range b.Limbs {
		for _, cO := range bO.Limbs {
			if c.WrappedIntersects(&cO, w) {
				return true
			}
		}

	}
	for _, c := range b.Limbs {
		if c.WrappedIntersects(&bO.Core, w) {
			return true
		}
	}
	for _, cO := range bO.Limbs {
		if cO.WrappedIntersects(&b.Core, w) {
			return true
		}
	}
	return false
}

func getAdjacentCircle(core *vec.Circle, ang float64) vec.Circle {
	return vec.Circle{
		Centre: core.Centre.Add(vec.UnitPointFromAngle(ang).Mul(2 * core.R)),
		R:      core.R,
	}
}
