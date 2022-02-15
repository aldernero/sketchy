package sketchy

import (
	"math"
	"testing"
)

func TestGcd(t *testing.T) {
	a := 60
	b := 15
	c := Gcd(a, b)
	if c != 15 {
		t.Errorf("Incorrect gcd, expected 15 got %d", c)
	}
	b = 25
	c = Gcd(a, b)
	if c != 5 {
		t.Errorf("Incorrect gcd, expected 5 got %d", c)
	}
	a = 71
	b = 73
	c = Gcd(a, b)
	if c != 1 {
		t.Errorf("Incorrect gcd, expected 1, got %d", c)
	}
}

func TestPointString(t *testing.T) {
	p := Point{X: 1, Y: 2}
	s := p.String()
	if s != "(1.000000, 2.000000)" {
		t.Errorf("Incorrect point string, expected (1.000000, 2.000000), got %s", s)
	}
}

func TestLineString(t *testing.T) {
	l := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 1},
	}
	s := l.String()
	c := "(0.000000, 0.000000) -> (1.000000, 1.000000)"
	if s != c {
		t.Errorf("Incorrect pline string, expected %s, got %s", c, s)
	}
}

func TestLerp(t *testing.T) {
	a := 10.0
	b := 100.0
	i := 0.3
	l := Lerp(a, b, i)
	if l != 37.0 {
		t.Errorf("Lerp was incorrect, expected 37.0, got %f", l)
	}
}

func TestMap(t *testing.T) {
	a := 0.0
	b := 100.0
	c := 0.0
	d := 360.0
	i := 75.0
	l := Map(a, b, c, d, i)
	if l != 270.0 {
		t.Errorf("Map was incorrect, expected 270.0, got %f", l)
	}
}

func TestLinspace(t *testing.T) {
	l := Linspace(0, Tau, 10, false)
	if len(l) != 9 {
		t.Errorf("Incorrect linspace length, expected 9, got %d", len(l))
	}
	if l[len(l)-1] >= Tau {
		t.Errorf("Last element too large, expected %f, got %f", Tau, l[len(l)-1])
	}
	l = Linspace(0, Tau, 10, true)
	if len(l) != 10 {
		t.Errorf("Incorrect linspace length, expected 10, got %d", len(l))
	}
	if l[len(l)-1] != Tau {
		t.Errorf("Incorrect last element, expected %f, got %f", Tau, l[len(l)-1])
	}
}

func TestDistance(t *testing.T) {
	p := Point{X: 0, Y: 0}
	q := Point{X: 1, Y: 1}
	d := Distance(p, q)
	if d != math.Sqrt2 {
		t.Errorf("Distance was incorrect, expected %f, got %f", d, math.Sqrt2)
	}
}

func TestSquaredDistance(t *testing.T) {
	p := Point{X: 0, Y: 0}
	q := Point{X: 1, Y: 1}
	d := SquaredDistance(p, q)
	if d != 2.0 {
		t.Errorf("Squared distance was incorrect, expected %f, got %f", d, 2.0)
	}
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
	if d != math.Sqrt2 {
		t.Errorf("Line length was incorrect, expected %f, got %f", d, math.Sqrt2)
	}
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
	if r != Tau/4 {
		t.Errorf("Radians not correct, expected %f, got %f", Tau/4, r)
	}
}

func TestRad2Deg(t *testing.T) {
	r := Tau / 4
	d := Rad2Deg(r)
	if d != 90.0 {
		t.Errorf("Degrees not correct, expected %f, got %f", 90.0, d)
	}
}
