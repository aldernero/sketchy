package sketchy

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestQuadTree_Insert(t *testing.T) {
	assert := assert.New(t)
	qt := NewQuadTree(Rect{
		X: 0,
		Y: 0,
		W: 100,
		H: 100,
	})
	assert.Equal(0, qt.Size())
	qt.Insert(Point{X: 0, Y: 0})
	assert.Equal(1, qt.Size())
	qt.Insert(Point{X: -1, Y: 0})
	assert.Equal(1, qt.Size())
	qt.Insert(Point{X: 10, Y: 0})
	assert.Equal(2, qt.Size())
	for i := 0; i < 5; i++ {
		x := rand.Float64() * 100
		y := rand.Float64() * 100
		qt.Insert(Point{X: x, Y: y})
	}
	assert.Equal(7, qt.Size())
}

func BenchmarkQuadTree_Insert(b *testing.B) {
	var seed int64 = 42
	rand.Seed(seed)
	w := 210.0
	h := 297.0
	points := make([]Point, 1000)
	for i := 0; i < 1000; i++ {
		x := rand.Float64() * w
		y := rand.Float64() * h
		points[i] = Point{X: x, Y: y}
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt := NewQuadTreeWithCapacity(Rect{X: 0, Y: 0, W: w, H: h}, 4)
		for _, p := range points {
			qt.Insert(p)
		}
	}
}
