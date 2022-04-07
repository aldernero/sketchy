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
	w := 210.0
	h := 297.0
	n := 1000
	rng := NewRng(seed)
	rng.SetNoiseOctaves(2)
	rng.SetNoiseLacunarity(2)
	rng.SetNoisePersistence(0.8)
	rng.SetNoiseScaleX(0.02)
	rng.SetNoiseScaleY(0.02)
	points := rng.NoisyRandomPoints(n, 0.4, Rect{X: 0, Y: 0, W: w, H: h})
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		qt := NewQuadTreeWithCapacity(Rect{X: 0, Y: 0, W: w, H: h}, 4)
		for _, p := range points {
			qt.Insert(p)
		}
	}
}

func BenchmarkQuadTree_Query(b *testing.B) {
	var seed int64 = 42
	w := 210.0
	h := 297.0
	n := 1000
	rng := NewRng(seed)
	rng.SetNoiseOctaves(2)
	rng.SetNoiseLacunarity(2)
	rng.SetNoisePersistence(0.8)
	rng.SetNoiseScaleX(0.02)
	rng.SetNoiseScaleY(0.02)
	//points := rng.NoisyRandomPoints(n, 0.4, Rect{X: 0, Y: 0, W: w, H: h})
	points := rng.UniformRandomPoints(n, Rect{X: 0, Y: 0, W: w, H: h})
	qt := NewQuadTreeWithCapacity(Rect{X: 0, Y: 0, W: w, H: h}, 4)
	for _, p := range points {
		qt.Insert(p)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = qt.Query(Rect{
			X: 0.4 * w,
			Y: 0.4 * h,
			W: 0.2 * w,
			H: 0.2 * h,
		})
	}
}
