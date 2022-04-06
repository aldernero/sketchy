package sketchy

import (
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestKDTree_Insert(t *testing.T) {
	assert := assert.New(t)
	tree := NewKDTree(Point{X: 50, Y: 50}, Rect{X: 0, Y: 0, W: 100, H: 100})
	assert.Equal(1, tree.Size())
}

func BenchmarkKDTree_Insert(b *testing.B) {
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
		tree := NewKDTree(points[0], Rect{X: 0, Y: 0, W: w, H: h})
		for j := 1; j < len(points); j++ {
			tree.Insert(points[j])
		}
	}
}
