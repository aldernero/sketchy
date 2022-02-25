package sketchy

import "math"

// Generates a Chaikin curve given a set of control points,
// a cutoff ratio, and the number of steps to use in the
// calculation.
func Chaikin(c Curve, q float64, n int) Curve {
	points := []Point{}
	// Start with control points
	points = append(points, c.Points...)
	left := q / 2
	right := 1 - (q / 2)
	for i := 0; i < n; i++ {
		newPoints := []Point{}
		for j := 0; j < len(points)-1; j++ {
			p1 := points[j]
			p2 := points[j+1]
			q := Point{
				X: right*p1.X + left*p2.X,
				Y: right*p1.Y + left*p2.Y,
			}
			r := Point{
				X: left*p1.X + right*p2.X,
				Y: left*p1.Y + right*p2.Y,
			}
			newPoints = append(newPoints, q, r)
		}
		if c.Closed {
			p1 := points[len(points)-1]
			p2 := points[0]
			q := Point{
				X: right*p1.X + left*p2.X,
				Y: right*p1.Y + left*p2.Y,
			}
			r := Point{
				X: left*p1.X + right*p2.X,
				Y: left*p1.Y + right*p2.Y,
			}
			newPoints = append(newPoints, q, r)
		}
		points = []Point{}
		points = append(points, newPoints...)
	}
	return Curve{Points: points, Closed: c.Closed}
}

// Core parametrers for a 2D Lissajous curve
type Lissajous struct {
	Nx int
	Ny int
	Px float64
	Py float64
}

// Generates a Lissajous curve given parameters, a number of points
// to use (i.e. resolution), and an offset and scale (typically to convert
// to screen coordinates)
func GenLissajous(l Lissajous, n int, offset Point, s float64) Curve {
	curve := Curve{}
	maxPhase := Tau / float64(Gcd(l.Nx, l.Ny))
	dt := maxPhase / float64(n)
	for t := 0.0; t < maxPhase; t += dt {
		xPos := s*math.Sin(float64(l.Nx)*t+l.Px) + offset.X
		yPos := s*math.Sin(float64(l.Ny)*t+l.Py) + offset.Y
		point := Point{X: xPos, Y: yPos}
		curve.Points = append(curve.Points, point)
	}
	return curve
}

// Calculates Padua points for a certain class of Lissajous curves,
// where Nx = Ny +/- 1. The correspond to intersection points and
// some of the outside points on the curve
// See https://en.wikipedia.org/wiki/Padua_points for more details.
func PaduaPoints(n int) []Point {
	points := []Point{}
	for i := 0; i <= n; i++ {
		delta := 0
		if n%2 == 1 && i%2 == 1 {
			delta = 1
		}
		for j := 1; j < (n/2)+2+delta; j++ {
			x := math.Cos(float64(i) * Pi / float64(n))
			var y float64
			if i%2 == 1 {
				y = math.Cos(float64(2*j-2) * Pi / float64(n+1))
			} else {
				y = math.Cos(float64(2*j-1) * Pi / float64(n+1))
			}
			points = append(points, Point{X: x, Y: y})
		}
	}
	return points
}
