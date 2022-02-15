package sketchy

import (
	"fmt"
	"math"
)

type Point struct {
	X float64
	Y float64
}

type Line struct {
	P Point
	Q Point
}

type Curve struct {
	Points []Point
	Closed bool
}

type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

func (p Point) String() string {
	return fmt.Sprintf("(%f, %f)", p.X, p.Y)
}

func (p Point) Lerp(a Point, i float64) Point {
	return Point{
		X: Lerp(p.X, a.X, i),
		Y: Lerp(p.Y, a.Y, i),
	}
}

func (l Line) String() string {
	return fmt.Sprintf("(%f, %f) -> (%f, %f)", l.P.X, l.P.Y, l.Q.X, l.Q.Y)
}

func (l Line) Lerp(i float64) Point {
	return Point{
		X: Lerp(l.P.X, l.Q.X, i),
		Y: Lerp(l.P.Y, l.Q.Y, i),
	}
}

func Distance(p Point, q Point) float64 {
	return math.Sqrt(math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2))
}

func SquaredDistance(p Point, q Point) float64 {
	return math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2)
}

func Midpoint(p Point, q Point) Point {
	return Point{X: 0.5 * (p.X + q.X), Y: 0.5 * (p.Y + q.Y)}
}

func (l Line) Midpoint() Point {
	return Midpoint(l.P, l.Q)
}

func (l Line) Length() float64 {
	return Distance(l.P, l.Q)
}

func (c *Curve) Length() float64 {
	result := 0.0
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		result += Distance(c.Points[i], c.Points[i+1])
	}
	if c.Closed {
		result += Distance(c.Points[0], c.Points[n-1])
	}
	return result
}

func (c *Curve) Last() Point {
	n := len(c.Points)
	switch n {
	case 0:
		return Point{
			X: 0,
			Y: 0,
		}
	case 1:
		return c.Points[0]
	}
	return c.Points[n-1]
}

func (r *Rect) ContainsPoint(p Point) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}
