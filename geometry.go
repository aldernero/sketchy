package sketchy

import (
	"fmt"
	"github.com/tdewolff/canvas"
	"log"
	"math"
)

// Primitive types

// Point is a simple point in 2D space
type Point struct {
	X float64
	Y float64
}

// Line is two points that form a line
type Line struct {
	P Point
	Q Point
}

// Curve A curve is a list of points, may be closed
type Curve struct {
	Points []Point
	Closed bool
}

// A Circle represented by a center point and radius
type Circle struct {
	Center Point
	Radius float64
}

// Rect is a simple rectangle
type Rect struct {
	X float64
	Y float64
	W float64
	H float64
}

// A Triangle specified by vertices as points
type Triangle struct {
	A Point
	B Point
	C Point
}

// Point functions

// Tuple representation of a point, useful for debugging
func (p Point) String() string {
	return fmt.Sprintf("(%f, %f)", p.X, p.Y)
}

// Lerp is a linear interpolation between two points
func (p Point) Lerp(a Point, i float64) Point {
	return Point{
		X: Lerp(p.X, a.X, i),
		Y: Lerp(p.Y, a.Y, i),
	}
}

// IsEqual determines if two points are equal
func (p Point) IsEqual(q Point) bool {
	return p.X == q.X && p.Y == q.Y
}

func (p Point) Draw(s float64, ctx *canvas.Context) {
	ctx.DrawPath(p.X, p.Y, canvas.Circle(s))
}

// Distance between two points
func Distance(p Point, q Point) float64 {
	return math.Sqrt(math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2))
}

// SquaredDistance is the square of the distance between two points
func SquaredDistance(p Point, q Point) float64 {
	return math.Pow(q.X-p.X, 2) + math.Pow(q.Y-p.Y, 2)
}

// Line functions
// String representation of a line, useful for debugging
func (l Line) String() string {
	return fmt.Sprintf("(%f, %f) -> (%f, %f)", l.P.X, l.P.Y, l.Q.X, l.Q.Y)
}

func (l Line) IsEqual(k Line) bool {
	return l.P.IsEqual(k.P) && l.Q.IsEqual(k.Q)
}

func (l Line) Angle() float64 {
	dy := l.Q.Y - l.P.Y
	dx := l.Q.X - l.P.X
	angle := math.Atan(dy / dx)
	return angle
}

// Slope computes the slope of the line
func (l Line) Slope() float64 {
	dy := l.Q.Y - l.P.Y
	dx := l.Q.X - l.P.X
	if math.Abs(dx) < Smol {
		if dx < 0 {
			if dy > 0 {
				return math.Inf(-1)
			} else {
				return math.Inf(1)
			}
		} else {
			if dy > 0 {
				return math.Inf(1)
			} else {
				return math.Inf(-1)
			}
		}
	}
	return dy / dx
}

func (l Line) InvertedSlope() float64 {
	slope := l.Slope()
	if math.IsInf(slope, 1) || math.IsInf(slope, -1) {
		return 0
	}
	return -1 / slope
}

func (l Line) PerpendicularAt(percentage float64, length float64) Line {
	angle := l.Angle()
	point := l.P.Lerp(l.Q, percentage)
	sinOffset := 0.5 * length * math.Sin(angle)
	cosOffset := 0.5 * length * math.Cos(angle)
	p := Point{
		X: NoTinyVals(point.X - sinOffset),
		Y: NoTinyVals(point.Y + cosOffset),
	}
	q := Point{
		X: NoTinyVals(point.X + sinOffset),
		Y: NoTinyVals(point.Y - cosOffset),
	}
	return Line{
		P: p,
		Q: q,
	}
}

func (l Line) PerpendicularBisector(length float64) Line {
	return l.PerpendicularAt(0.5, length)
}

// Lerp is an interpolation between the two points of a line
func (l Line) Lerp(i float64) Point {
	return Point{
		X: Lerp(l.P.X, l.Q.X, i),
		Y: Lerp(l.P.Y, l.Q.Y, i),
	}
}

func (l Line) Draw(ctx *canvas.Context) {
	ctx.MoveTo(l.P.X, l.P.Y)
	ctx.LineTo(l.Q.X, l.Q.Y)
	ctx.Stroke()
}

// Midpoint Calculates the midpoint between two points
func Midpoint(p Point, q Point) Point {
	return Point{X: 0.5 * (p.X + q.X), Y: 0.5 * (p.Y + q.Y)}
}

// Midpoint Calculates the midpoint of a line
func (l Line) Midpoint() Point {
	return Midpoint(l.P, l.Q)
}

// Length Calculates the length of a line
func (l Line) Length() float64 {
	return Distance(l.P, l.Q)
}

func (l Line) Intersects(k Line) bool {
	a1 := l.Q.X - l.P.X
	b1 := k.P.X - k.Q.X
	c1 := k.P.X - l.P.X
	a2 := l.Q.Y - l.P.Y
	b2 := k.P.Y - k.Q.Y
	c2 := k.P.Y - l.P.Y
	d := a1*b2 - a2*b1
	if d == 0 {
		// lines are parallel
		return false
	}
	// Cramer's rule
	s := (c1*b2 - c2*b1) / d
	t := (a1*c2 - a2*c1) / d
	return s >= 0 && t >= 0 && s <= 1 && t <= 1
}

func (l Line) ParallelTo(k Line) bool {
	a1 := l.Q.X - l.P.X
	b1 := k.P.X - k.Q.X
	a2 := l.Q.Y - l.P.Y
	b2 := k.P.Y - k.Q.Y
	d := a1*b2 - a2*b1
	return d == 0
}

// Curve functions

// Length Calculates the length of the line segments of a curve
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

// Last returns the last point in a curve
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
	if c.Closed {
		return c.Points[0]
	}
	return c.Points[n-1]
}

func (c *Curve) LastLine() Line {
	n := len(c.Points)
	switch n {
	case 0:
		return Line{
			P: Point{X: 0, Y: 0},
			Q: Point{X: 0, Y: 0},
		}
	case 1:
		return Line{
			P: c.Points[0],
			Q: c.Points[0],
		}
	}
	if c.Closed {
		return Line{
			P: c.Points[n-1],
			Q: c.Points[0],
		}
	}
	return Line{
		P: c.Points[n-2],
		Q: c.Points[n-1],
	}
}

// Lerp calculates a point a given percentage along a curve
func (c *Curve) Lerp(percentage float64) Point {
	var point Point
	if percentage < 0 || percentage > 1 {
		log.Fatalf("percentage in Lerp not between 0 and 1: %v\n", percentage)
	}
	if NoTinyVals(percentage) == 0 {
		return c.Points[0]
	}
	if math.Abs(percentage-1) < Smol {
		return c.Last()
	}
	totalDist := c.Length()
	targetDist := percentage * totalDist
	partialDist := 0.0
	var foundPoint bool
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		dist := Distance(c.Points[i], c.Points[i+1])
		if partialDist+dist >= targetDist {
			remainderDist := targetDist - partialDist
			pct := remainderDist / dist
			point = c.Points[i].Lerp(c.Points[i+1], pct)
			foundPoint = true
			break
		}
		partialDist += dist
	}
	if !foundPoint {
		if c.Closed {
			dist := Distance(c.Points[n-1], c.Points[0])
			remainderDist := targetDist - partialDist
			pct := remainderDist / dist
			point = c.Points[n-1].Lerp(c.Points[0], pct)
		} else {
			panic("couldn't find curve lerp point")
		}
	}
	return point
}

func (c *Curve) LineAt(percentage float64) (Line, float64) {
	var line Line
	var linePct float64
	if percentage < 0 || percentage > 1 {
		log.Fatalf("percentage in Lerp not between 0 and 1: %v\n", percentage)
	}
	if NoTinyVals(percentage) == 0 {
		return Line{P: c.Points[0], Q: c.Points[1]}, 0
	}
	if math.Abs(percentage-1) < Smol {
		return c.LastLine(), 1
	}
	totalDist := c.Length()
	targetDist := percentage * totalDist
	partialDist := 0.0
	var foundPoint bool
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		dist := Distance(c.Points[i], c.Points[i+1])
		if partialDist+dist >= targetDist {
			remainderDist := targetDist - partialDist
			linePct = remainderDist / dist
			line.P = c.Points[i]
			line.Q = c.Points[i+1]
			foundPoint = true
			break
		}
		partialDist += dist
	}
	if !foundPoint {
		if c.Closed {
			dist := Distance(c.Points[n-1], c.Points[0])
			remainderDist := targetDist - partialDist
			linePct = remainderDist / dist
			line.P = c.Points[n-1]
			line.Q = c.Points[0]
		} else {
			panic("couldn't find curve lerp point")
		}
	}
	return line, linePct
}

func (c *Curve) PerpendicularAt(percentage float64, length float64) Line {
	line, linePct := c.LineAt(percentage)
	return line.PerpendicularAt(linePct, length)
}

func (c *Curve) Draw(ctx *canvas.Context) {
	n := len(c.Points)
	for i := 0; i < n-1; i++ {
		ctx.MoveTo(c.Points[i].X, c.Points[i].Y)
		ctx.LineTo(c.Points[i+1].X, c.Points[i+1].Y)
	}
	if c.Closed {
		ctx.MoveTo(c.Points[n-1].X, c.Points[n-1].Y)
		ctx.LineTo(c.Points[0].X, c.Points[0].Y)
	}
	ctx.Stroke()
}

// Circle functions

func (c *Circle) Draw(ctx *canvas.Context) {
	ctx.DrawPath(c.Center.X, c.Center.Y, canvas.Circle(c.Radius))
}

func (c *Circle) ToCurve(resolution int) Curve {
	points := make([]Point, resolution)
	theta := Linspace(0, Tau, resolution, false)
	for i, t := range theta {
		x := c.Center.X + c.Radius*math.Cos(t)
		y := c.Center.Y + c.Radius*math.Sin(t)
		points[i] = Point{X: x, Y: y}
	}
	return Curve{Points: points, Closed: true}
}

func (c *Circle) ContainsPoint(p Point) bool {
	return Distance(c.Center, p) <= c.Radius
}

func (c *Circle) PointOnEdge(p Point) bool {
	return Equalf(Distance(c.Center, p), c.Radius)
}

// Rect functions

// ContainsPoint determines if a point lies within a rectangle
func (r *Rect) ContainsPoint(p Point) bool {
	return p.X >= r.X && p.X <= r.X+r.W && p.Y >= r.Y && p.Y <= r.Y+r.H
}

func (r *Rect) Contains(rect Rect) bool {
	a := Point{X: r.X, Y: r.Y}
	b := Point{X: r.X + r.W, Y: r.Y + r.H}
	c := Point{X: rect.X, Y: rect.Y}
	d := Point{X: rect.X + rect.W, Y: rect.Y + rect.H}
	return a.X < c.X && a.Y < c.Y && b.X > d.X && b.Y > d.Y
}

func (r *Rect) IsDisjoint(rect Rect) bool {
	aLeft := r.X
	aRight := r.X + r.W
	aTop := r.Y + r.H
	aBottom := r.Y
	bLeft := rect.X
	bRight := rect.X + rect.W
	bTop := rect.Y + rect.H
	bBottom := rect.Y

	if aLeft > bRight || aBottom > bTop || aRight < bLeft || aTop < bBottom {
		return true
	}
	return false
}

func (r *Rect) Overlaps(rect Rect) bool {
	return !r.IsDisjoint(rect)
}

func (r *Rect) Intersects(rect Rect) bool {
	a := Point{X: r.X, Y: r.Y}
	b := Point{X: r.X + r.W, Y: r.Y + r.H}
	c := Point{X: rect.X, Y: rect.Y}
	d := Point{X: rect.X + rect.W, Y: rect.Y + rect.H}

	if a.X >= d.X || c.X >= b.X {
		return false
	}

	if b.Y >= c.Y || d.Y >= a.Y {
		return false
	}

	return true
}

func (r *Rect) Draw(ctx *canvas.Context) {
	rect := canvas.Rectangle(r.W, r.H)
	ctx.DrawPath(r.X, r.Y, rect)
}

// Triangle functions

func (t *Triangle) Draw(ctx *canvas.Context) {
	ctx.MoveTo(t.A.X, t.A.Y)
	ctx.LineTo(t.B.X, t.B.Y)
	ctx.LineTo(t.C.X, t.C.Y)
	ctx.Close()
}

func (t *Triangle) Area() float64 {
	// Heron's formula
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	s := (a + b + c) / 2
	return math.Sqrt(s * (s - a) * (s - b) * (s - c))
}

func (t *Triangle) Perimeter() float64 {
	a := Line{P: t.A, Q: t.B}.Length()
	b := Line{P: t.B, Q: t.C}.Length()
	c := Line{P: t.C, Q: t.A}.Length()
	return a + b + c
}

func (t *Triangle) Centroid() Point {
	x := (t.A.X + t.B.X + t.C.X) / 3
	y := (t.A.Y + t.B.Y + t.C.Y) / 3
	return Point{X: x, Y: y}
}
