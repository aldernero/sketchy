package sketchy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestKDTree_Insert(t *testing.T) {
	assert := assert.New(t)
	tree := NewKDTreeWithPoint(Point{X: 50, Y: 50}, Rect{X: 0, Y: 0, W: 100, H: 100})
	assert.Equal(1, tree.Size())
}

func BenchmarkKDTree_Query(b *testing.B) {
	var seed int64 = 42
	w := 210.0
	h := 297.0
	n := 10000
	rng := NewRng(seed)
	rng.SetNoiseOctaves(2)
	rng.SetNoiseLacunarity(2)
	rng.SetNoisePersistence(0.8)
	rng.SetNoiseScaleX(0.02)
	rng.SetNoiseScaleY(0.02)
	//points := rng.NoisyRandomPoints(n, 0.4, Rect{X: 0, Y: 0, W: w, H: h})
	points := rng.UniformRandomPoints(n, Rect{X: 0, Y: 0, W: w, H: h})
	tree := NewKDTree(Rect{X: 0, Y: 0, W: w, H: h})
	for _, p := range points {
		tree.Insert(p)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = tree.Query(Rect{
			X: 0.4 * w,
			Y: 0.4 * h,
			W: 0.2 * w,
			H: 0.2 * h,
		})
	}
}
