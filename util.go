package sketchy

import (
	"container/heap"
	_ "container/heap"
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

type MetricPoint struct {
	Metric float64
	Point
}

// PointHeap implements a max heap of MetricPoint objects
type PointHeap struct {
	size   int
	points []MetricPoint
}

func (p PointHeap) Len() int {
	return len(p.points)
}

func (p PointHeap) Less(i, j int) bool {
	return p.points[i].Metric > p.points[j].Metric
}

func (p PointHeap) Swap(i, j int) {
	p.points[i], p.points[j] = p.points[j], p.points[i]
}

func (p *PointHeap) Push(x interface{}) {
	if p.size <= 0 {
		return
	}
	if len(p.points) < p.size {
		p.points = append(p.points, x.(MetricPoint))
		return
	}
	if x.(MetricPoint).Metric < p.Peek().(MetricPoint).Metric {
		heap.Pop(p)
		p.points = append(p.points, x.(MetricPoint))
	}
}

func (p *PointHeap) Pop() interface{} {
	old := *p
	n := len(old.points)
	x := old.points[n-1]
	p.points = old.points[0 : n-1]
	return x
}

func (p *PointHeap) Peek() interface{} {
	old := *p
	n := len(old.points)
	x := old.points[n-1]
	return x
}

func (p *PointHeap) Report() []Point {
	n := len(p.points)
	result := make([]Point, n)
	for i := 0; i < n; i++ {
		q := heap.Pop(p)
		result[n-i-1] = q.(MetricPoint).Point
	}
	return result
}
