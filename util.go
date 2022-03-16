package sketchy

import (
	"fmt"
	"math"
	"math/rand"
	"time"
)

var (
	Pi    = math.Pi
	Tau   = 2 * math.Pi
	Sqrt2 = math.Sqrt2
	Sqrt3 = math.Sqrt(3)
	Smol  = 1e-6
)

// Greatest common divisor
func Gcd(a int, b int) int {
	if b == 0 {
		return a
	} else {
		return Gcd(b, a%b)
	}
}

// Linear interpolation between two values
func Lerp(a float64, b float64, i float64) float64 {
	return a + i*(b-a)
}

// Linear interpolation from one range to another
func Map(a float64, b float64, c float64, d float64, i float64) float64 {
	p := (i - a) / (b - a)
	return Lerp(c, d, p)
}

// Restrict a value to a given range
func Clamp(a float64, b float64, c float64) float64 {
	if c <= a {
		return a
	}
	if c >= b {
		return b
	}
	return c
}

// NoTinyVals sets values very close to zero to zero
func NoTinyVals(a float64) float64 {
	if math.Abs(a) < Smol {
		return 0
	}
	return a
}

// Creates a slice of linearly distributed values in a range
func Linspace(i float64, j float64, n int, b bool) []float64 {
	var result []float64
	N := float64(n)
	if b {
		N -= 1
	}
	d := (j - i) / N
	for k := 0; k < n; k++ {
		result = append(result, i+float64(k)*d)
	}
	return result
}

// Convert from degrees to radians
func Deg2Rad(f float64) float64 {
	return math.Pi * f / 180
}

// Convert from radians to degrees
func Rad2Deg(f float64) float64 {
	return 180 * f / math.Pi
}

// Convert from cartesian coordinates to screen coordinates
func CartesianToScreen(p []Point, o Point, s float64) []Point {
	points := []Point{}
	for _, i := range p {
		x := s*i.X + o.X
		y := -s*i.Y + o.Y
		points = append(points, Point{X: x, Y: y})
	}
	return points
}

// Shuffle a slice of points
func Shuffle(p *[]Point) {
	rand.Seed(time.Now().UnixMicro())
	n := len(*p)
	for i := 0; i < 3*n; i++ {
		j := rand.Intn(n)
		k := rand.Intn(n)
		(*p)[j], (*p)[k] = (*p)[k], (*p)[j]
	}
}

// Create a string based on the current time for use in filenames
func GetTimestampString() string {
	now := time.Now()
	return fmt.Sprintf("%d%02d%02d_%02d%02d%02d",
		now.Year(), now.Month(), now.Day(), now.Hour(),
		now.Minute(), now.Second())
}

func Equalf(a, b float64) bool {
	return math.Abs(b-a) <= Smol
}
