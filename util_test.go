package sketchy

import (
	"math"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGcd(t *testing.T) {
	assert := assert.New(t)
	a := 60
	b := 15
	c := Gcd(a, b)
	assert.Equal(15, c)
	b = 25
	c = Gcd(a, b)
	assert.Equal(5, c)
	a = 71
	b = 73
	c = Gcd(a, b)
	assert.Equal(1, c)
}

func TestPointString(t *testing.T) {
	p := Point{X: 1, Y: 2}
	s := p.String()
	assert.Equal(t, "(1.000000, 2.000000)", s)
}

func TestLineString(t *testing.T) {
	l := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 1},
	}
	s := l.String()
	c := "(0.000000, 0.000000) -> (1.000000, 1.000000)"
	assert.Equal(t, c, s)
}

func TestLerp(t *testing.T) {
	a := 10.0
	b := 100.0
	i := 0.3
	l := Lerp(a, b, i)
	assert.Equal(t, 37.0, l)
}

func TestMap(t *testing.T) {
	a := 0.0
	b := 100.0
	c := 0.0
	d := 360.0
	i := 75.0
	l := Map(a, b, c, d, i)
	assert.Equal(t, 270.0, l)
}

func TestLinspace(t *testing.T) {
	assert := assert.New(t)
	l := Linspace(0, Tau, 10, false)
	assert.Equal(10, len(l))
	assert.LessOrEqual(l[len(l)-1], Tau)
	l = Linspace(0, Tau, 10, true)
	assert.Equal(10, len(l))
	assert.Equal(Tau, l[len(l)-1])
}

func TestDistance(t *testing.T) {
	p := Point{X: 0, Y: 0}
	q := Point{X: 1, Y: 1}
	d := Distance(p, q)
	assert.Equal(t, Sqrt2, d)
}

func TestSquaredDistance(t *testing.T) {
	p := Point{X: 0, Y: 0}
	q := Point{X: 1, Y: 1}
	d := SquaredDistance(p, q)
	assert.Equal(t, 2.0, d)
}

func TestMidpoint(t *testing.T) {
	p := Point{X: 0, Y: 0}
	q := Point{X: 1, Y: 1}
	m := Midpoint(p, q)
	if m.X != 0.5 || m.Y != 0.5 {
		t.Errorf("Incorrect midpoint, expected (0.5, 0.5), got (%f, %f)", m.X, m.Y)
	}
}

func TestLineMidpoint(t *testing.T) {
	l := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 1},
	}
	m := l.Midpoint()
	if m.X != 0.5 || m.Y != 0.5 {
		t.Errorf("Incorrect midpoint, expected (0.5, 0.5), got (%f, %f)", m.X, m.Y)
	}
}

func TestLineLength(t *testing.T) {
	l := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 1},
	}
	d := l.Length()
	assert.Equal(t, Sqrt2, d)
}

func TestCurveLength(t *testing.T) {
	p := Linspace(0, Tau, 10000, false)
	c := Curve{}
	for _, i := range p {
		point := Point{X: math.Cos(i), Y: math.Sin(i)}
		c.Points = append(c.Points, point)
	}
	c.Closed = true
	l := c.Length()
	if math.Abs(Tau-l) >= 0.001*Tau {
		t.Errorf("Curve length was incorrect, expected %f, got %f", Tau, l)
	}
}

func TestDeg2Rad(t *testing.T) {
	d := 90.0
	r := Deg2Rad(d)
	assert.Equal(t, Tau/4, r)
}

func TestRad2Deg(t *testing.T) {
	r := Tau / 4
	d := Rad2Deg(r)
	assert.Equal(t, 90.0, d)
}

func TestPointHeap(t *testing.T) {
	W := 420.0
	H := 297.0
	rand.Seed(42)
	N := 100
	k := 7
	target := Point{X: rand.Float64() * W, Y: rand.Float64() * H}
	points := make([]MetricPoint, N)
	heap := PointHeap{
		size:   k,
		points: []MetricPoint{},
	}
	for i := 0; i < N; i++ {
		point := Point{X: rand.Float64() * W, Y: rand.Float64() * H}
		dist := SquaredDistance(target, point)
		points[i] = MetricPoint{
			Metric: dist,
			Point:  point,
		}
		heap.Push(points[i])
	}
	assert.Equal(t, N, len(points))
	assert.LessOrEqual(t, k, heap.Len())
}
