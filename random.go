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

// Pseudo-random number generator data
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

// Returns a PRNG with a system and noise generator
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

// Sets the seed for both the system and opensimplex PRNG
func (r *Rng) SetSeed(i int64) {
	rand.Seed(i)
	r.seed = i
	r.noise = opensimplex.New(i)
}

func (r *Rng) Gaussian(mean float64, stdev float64) float64 {
	return rand.NormFloat64()*stdev + mean
}

// The noise scale functions scale the position values passed into the
// noise PRNG. Typically for screen coordinates scale values in the
// range of 0.001 to 0.01 produce visually appealing noise

// Scales the x position in noise calculations
func (r *Rng) SetNoiseScaleX(scale float64) {
	r.xscale = scale
}

// Scales the y position in noise calculations
func (r *Rng) SetNoiseScaleY(scale float64) {
	r.yscale = scale
}

// Scales the z position in noise calculations
func (r *Rng) SetNoiseScaleZ(scale float64) {
	r.zscale = scale
}

// The noise offset functions simple increment/decrement the
// position values before scaling

// Offsets the x position in noise calculations
func (r *Rng) SetNoiseOffsetX(offset float64) {
	r.xoffset = offset
}

// Offsets the y position in noise calculations
func (r *Rng) SetNoiseOffsetY(offset float64) {
	r.yoffset = offset
}

// Offsets the z position in noise calculations
func (r *Rng) SetNoiseOffsetZ(offset float64) {
	r.zoffset = offset
}

// Number of steps when calculating fractal noise
func (r *Rng) SetNoiseOctaves(i int) {
	r.octaves = i
}

// How amplitude scales with octaves
func (r *Rng) SetNoisePersistence(p float64) {
	r.persistence = p
}

// How frequency scales with octaves
func (r *Rng) SetNoiseLacunarity(l float64) {
	r.lacunarity = l
}

// Returns noise values in the range of [-1, 1]
func (r *Rng) SignedNoise2D(x float64, y float64) float64 {
	return r.calcNoise(x, y, 0)
}

// Returns noise values in the range of [-1, 1]
func (r *Rng) SignedNoise3D(x float64, y float64, z float64) float64 {
	return r.calcNoise(x, y, z)
}

// Returns noise values in the range of [0, 1]
func (r *Rng) Noise2D(x float64, y float64) float64 {
	return Map(-1, 1, 0, 1, r.calcNoise(x, y, 0))
}

// Returns noise values in the range of [0, 1]
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
