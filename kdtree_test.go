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

func BenchmarkKDTree_Insert(b *testing.B) {
	var table = []struct {
		name        string
		seed        int64
		w           float64
		h           float64
		n           int
		octaves     int
		lacunarity  float64
		persistence float64
		scalex      float64
		scaley      float64
		thresh      float64
		cap         int
		uniform     bool
	}{
		{
			name:        "a3_n1000_uniform",
			seed:        42,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n10000_uniform",
			seed:        43,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n100000_uniform",
			seed:        44,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n1000_noisy",
			seed:        45,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
		{
			name:        "a3_n10000_noisy",
			seed:        46,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
		{
			name:        "a3_n100000_noisy",
			seed:        47,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
	}
	for _, v := range table {
		b.Run(v.name, func(b *testing.B) {
			rng := NewRng(v.seed)
			rng.SetNoiseOctaves(v.octaves)
			rng.SetNoiseLacunarity(v.lacunarity)
			rng.SetNoisePersistence(v.persistence)
			rng.SetNoiseScaleX(v.scalex)
			rng.SetNoiseScaleY(v.scaley)
			rect := Rect{X: 0, Y: 0, W: v.w, H: v.h}
			var points []Point
			if v.uniform {
				points = rng.UniformRandomPoints(v.n, rect)
			} else {
				points = rng.NoisyRandomPoints(v.n, v.thresh, rect)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				kdtree := NewKDTree(rect)
				for _, p := range points {
					kdtree.Insert(p)
				}
			}
		})
	}
}

func BenchmarkKDTree_Query(b *testing.B) {
	var table = []struct {
		name        string
		seed        int64
		w           float64
		h           float64
		n           int
		octaves     int
		lacunarity  float64
		persistence float64
		scalex      float64
		scaley      float64
		thresh      float64
		cap         int
		uniform     bool
	}{
		{
			name:        "a3_n1000_uniform",
			seed:        48,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n10000_uniform",
			seed:        49,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n100000_uniform",
			seed:        50,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
		},
		{
			name:        "a3_n1000_noisy",
			seed:        51,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
		{
			name:        "a3_n10000_noisy",
			seed:        52,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
		{
			name:        "a3_n100000_noisy",
			seed:        53,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
		},
	}
	for _, v := range table {
		b.Run(v.name, func(b *testing.B) {
			rng := NewRng(v.seed)
			rng.SetNoiseOctaves(v.octaves)
			rng.SetNoiseLacunarity(v.lacunarity)
			rng.SetNoisePersistence(v.persistence)
			rng.SetNoiseScaleX(v.scalex)
			rng.SetNoiseScaleY(v.scaley)
			rect := Rect{X: 0, Y: 0, W: v.w, H: v.h}
			var points []Point
			if v.uniform {
				points = rng.UniformRandomPoints(v.n, rect)
			} else {
				points = rng.NoisyRandomPoints(v.n, v.thresh, rect)
			}
			tree := NewKDTree(rect)
			for _, p := range points {
				tree.Insert(p)
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tree.Query(Rect{
					X: 0.4 * v.w,
					Y: 0.4 * v.h,
					W: 0.2 * v.w,
					H: 0.2 * v.h,
				})
			}
		})
	}
}

func BenchmarkKDTree_kNN(b *testing.B) {
	var table = []struct {
		name        string
		seed        int64
		w           float64
		h           float64
		n           int
		octaves     int
		lacunarity  float64
		persistence float64
		scalex      float64
		scaley      float64
		thresh      float64
		cap         int
		uniform     bool
		k           int
	}{
		{
			name:        "a3_n1000_k1_uniform",
			seed:        54,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           1,
		},
		{
			name:        "a3_n10000_k1_uniform",
			seed:        55,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           1,
		},
		{
			name:        "a3_n100000_k1_uniform",
			seed:        56,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           1,
		},
		{
			name:        "a3_n1000_k1_noisy",
			seed:        57,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           1,
		},
		{
			name:        "a3_n10000_k1_noisy",
			seed:        58,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           1,
		},
		{
			name:        "a3_n100000_k1_noisy",
			seed:        59,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           1,
		},
		{
			name:        "a3_n1000_k10_uniform",
			seed:        60,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           10,
		},
		{
			name:        "a3_n10000_k10_uniform",
			seed:        61,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           10,
		},
		{
			name:        "a3_n100000_k10_uniform",
			seed:        62,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     0,
			lacunarity:  0,
			persistence: 0,
			scalex:      0,
			scaley:      0,
			thresh:      0,
			cap:         4,
			uniform:     true,
			k:           10,
		},
		{
			name:        "a3_n1000_k10_noisy",
			seed:        63,
			w:           420.0,
			h:           297.0,
			n:           1000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           10,
		},
		{
			name:        "a3_n10000_k10_noisy",
			seed:        64,
			w:           420.0,
			h:           297.0,
			n:           10000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           10,
		},
		{
			name:        "a3_n100000_k10_noisy",
			seed:        65,
			w:           420.0,
			h:           297.0,
			n:           100000,
			octaves:     2,
			lacunarity:  2,
			persistence: 0.8,
			scalex:      0.02,
			scaley:      0.02,
			thresh:      0.4,
			cap:         4,
			uniform:     false,
			k:           10,
		},
	}
	for _, v := range table {
		b.Run(v.name, func(b *testing.B) {
			rng := NewRng(v.seed)
			rng.SetNoiseOctaves(v.octaves)
			rng.SetNoiseLacunarity(v.lacunarity)
			rng.SetNoisePersistence(v.persistence)
			rng.SetNoiseScaleX(v.scalex)
			rng.SetNoiseScaleY(v.scaley)
			rect := Rect{X: 0, Y: 0, W: v.w, H: v.h}
			var points []Point
			if v.uniform {
				points = rng.UniformRandomPoints(v.n, rect)
			} else {
				points = rng.NoisyRandomPoints(v.n, v.thresh, rect)
			}
			tree := NewKDTree(rect)
			for _, p := range points {
				tree.Insert(p)
			}
			targetPoint := Point{X: v.w / 2, Y: v.h / 2}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = tree.NearestNeighbors(targetPoint, v.k)
			}
		})
	}
}
