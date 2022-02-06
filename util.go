package genart

import (
	"math"
	"math/rand"
	"time"
)

var (
	Pi    = math.Pi
	Tau   = 2 * math.Pi
	Sqrt2 = math.Sqrt2
	Sqrt3 = math.Sqrt(3)
)

func Gcd(a int, b int) int {
	if b == 0 {
		return a
	} else {
		return Gcd(b, a%b)
	}
}

func Lerp(a float64, b float64, i float64) float64 {
	return a + i*(b-a)
}

func Map(a float64, b float64, c float64, d float64, i float64) float64 {
	p := (i - a) / (b - a)
	return Lerp(c, d, p)
}

func Linspace(i float64, j float64, n int, b bool) []float64 {
	d := (j - i) / float64(n-1)
	var result []float64
	m := j - d
	if b {
		m = j
	}
	k := i
	for {
		result = append(result, k)
		k += d
		if k > m {
			break
		}
	}
	return result
}

func Deg2Rad(f float64) float64 {
	return math.Pi * f / 180
}

func Rad2Deg(f float64) float64 {
	return 180 * f / math.Pi
}

func CartesianToScreen(p []Point, o Point, s float64) []Point {
	points := []Point{}
	for _, i := range p {
		x := s*i.X + o.X
		y := -s*i.Y + o.Y
		points = append(points, Point{X: x, Y: y})
	}
	return points
}

func Shuffle(p *[]Point) {
	rand.Seed(time.Now().UnixMicro())
	n := len(*p)
	for i := 0; i < 3*n; i++ {
		j := rand.Intn(n)
		k := rand.Intn(n)
		(*p)[j], (*p)[k] = (*p)[k], (*p)[j]
	}
}
