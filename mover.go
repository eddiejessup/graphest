package main

import (
	"github.com/eddiejessup/vec"
	"github.com/golang/geo/r2"
	"math/rand"
)

type Mover struct {
	Body
	Velocity  r2.Point
	TimeToLay int
}

func (m *Mover) Move(w r2.Point) {
	for _, c := range m.Body.Circles() {
		vec.MoveWrappedPoint(&(c.Centre), m.Velocity, w)
	}
}

func (m *Mover) Rotate(ang float64) {
    m.Velocity = vec.RotateVector2d(m.Velocity, vec.RandSymmetricFloat64(ang))
}

func (m *Mover) Momentum() float64 {
	return m.Velocity.Norm() * m.Body.Area()
}

func newVelocity(v float64) *r2.Point {
	vel := vec.RandomUnitVector().Mul(rand.Float64() * v)
	return &vel
}

func NewMover(b Body, v float64, layingPeriod int) *Mover {
	return &Mover{
		Body:      b,
		Velocity:  *newVelocity(v),
		TimeToLay: layingPeriod,
	}
}
