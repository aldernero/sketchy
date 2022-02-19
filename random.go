package sketchy

import (
	"math/rand"

	"github.com/ojrac/opensimplex-go"
)

const (
	defaultScale       = 0.001
	defaultOctaves     = 1
	defaultPersistence = 0.9
	defaultLacunarity  = 2.0
)

type Rng struct {
	seed        int64
	noise       opensimplex.Noise
	octaves     int
	persistence float64
	lacunarity  float64
	xscale      float64
	yscale      float64
	zscale      float64
	xoffset     float64
	yoffset     float64
	zoffset     float64
}

func NewRng(i int64) Rng {
	rand.Seed(i)
	return Rng{
		seed:        i,
		noise:       opensimplex.New(i),
		octaves:     defaultOctaves,
		persistence: defaultPersistence,
		lacunarity:  defaultLacunarity,
		xscale:      defaultScale,
		yscale:      defaultScale,
		zscale:      defaultScale,
		xoffset:     0,
		yoffset:     0,
		zoffset:     0,
	}
}

func (r *Rng) SetSeed(i int64) {
	rand.Seed(i)
	r.seed = i
	r.noise = opensimplex.New(i)
}

func (r *Rng) SetNoiseScaleX(scale float64) {
	r.xscale = scale
}

func (r *Rng) SetNoiseScaleY(scale float64) {
	r.yscale = scale
}

func (r *Rng) SetNoiseScaleZ(scale float64) {
	r.zscale = scale
}

func (r *Rng) SetNoiseOffsetX(offset float64) {
	r.xoffset = offset
}

func (r *Rng) SetNoiseOffsetY(offset float64) {
	r.yoffset = offset
}

func (r *Rng) SetNoiseOffsetZ(offset float64) {
	r.zoffset = offset
}

func (r *Rng) SetNoiseOctaves(i int) {
	r.octaves = i
}

func (r *Rng) SetNoisePersistence(p float64) {
	r.persistence = p
}

func (r *Rng) SetNoiseLacunarity(l float64) {
	r.lacunarity = l
}

func (r *Rng) SignedNoise2D(x float64, y float64) float64 {
	return r.calcNoise(x, y, 0)
}

func (r *Rng) SignedNoise3D(x float64, y float64, z float64) float64 {
	return r.calcNoise(x, y, z)
}

func (r *Rng) Noise2D(x float64, y float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, y, 0))
}

func (r *Rng) Noise3D(x float64, y float64, z float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, y, z))
}

func (r *Rng) calcNoise(x, y, z float64) float64 {
	totalNoise := 0.0
	totalAmp := 0.0
	amp := 1.0
	freq := 1.0
	for i := 0; i < r.octaves; i++ {
		totalNoise += r.noise.Eval3(
			(x+r.xoffset)*r.xscale*freq,
			(y+r.yoffset)*r.yscale*freq,
			(z+r.zoffset)*r.zscale*freq,
		)
		totalAmp += amp
		amp *= r.persistence
		freq *= r.lacunarity
	}
	return totalNoise / totalAmp
}
