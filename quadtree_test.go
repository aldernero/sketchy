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
	qt.Insert(Point{X: 0, Y: 0}.ToIndexPoint(0))
	assert.Equal(1, qt.Size())
	qt.Insert(Point{X: -1, Y: 0}.ToIndexPoint(0))
	assert.Equal(1, qt.Size())
	qt.Insert(Point{X: 10, Y: 0}.ToIndexPoint(0))
	assert.Equal(2, qt.Size())
	for i := 0; i < 5; i++ {
		x := rand.Float64() * 100
		y := rand.Float64() * 100
		qt.Insert(Point{X: x, Y: y}.ToIndexPoint(0))
	}
	assert.Equal(7, qt.Size())
}

func TestQuadTree_Search(t *testing.T) {
	assert := assert.New(t)
	points := []IndexPoint{
		{
			Index: 0,
			Point: Point{X: 1, Y: 7},
		},
		{
			Index: 1,
			Point: Point{X: 2, Y: 8},
		},
		{
			Index: 2,
			Point: Point{X: 3, Y: 9},
		},
		{
			Index: 3,
			Point: Point{X: 4, Y: 10},
		},
		{
			Index: 4,
			Point: Point{X: 5, Y: 11},
		},
		{
			Index: 5,
			Point: Point{X: 6, Y: 12},
		},
	}
	tree := NewQuadTree(Rect{X: 0, Y: 0, W: 20, H: 20})
	for _, p := range points {
		tree.Insert(p)
	}
	point := tree.Search(points[2])
	assert.NotNil(point)
	assert.Equal(points[2].Index, point.Index)

	tree.UpdateIndex(points[2], -2)
	point = tree.Search(points[2])
	assert.NotNil(point)
	assert.Equal(-2, point.Index)
}

func BenchmarkQuadTree_Insert(b *testing.B) {
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
				qt := NewQuadTreeWithCapacity(rect, v.cap)
				for j, p := range points {
					qt.Insert(p.ToIndexPoint(j))
				}
			}
		})
	}
}

func BenchmarkQuadTree_Query(b *testing.B) {
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
			qt := NewQuadTreeWithCapacity(rect, v.cap)
			for i, p := range points {
				qt.Insert(p.ToIndexPoint(i))
			}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = qt.Query(Rect{
					X: 0.4 * v.w,
					Y: 0.4 * v.h,
					W: 0.2 * v.w,
					H: 0.2 * v.h,
				})
			}
		})
	}
}

func BenchmarkQuadTree_kNN(b *testing.B) {
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
			qt := NewQuadTreeWithCapacity(rect, v.cap)
			for i, p := range points {
				qt.Insert(p.ToIndexPoint(i))
			}
			targetPoint := Point{X: v.w / 2, Y: v.h / 2}.ToIndexPoint(-1)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_ = qt.NearestNeighbors(targetPoint, v.k)
			}
		})
	}
}
