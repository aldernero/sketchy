package sketchy

import (
	"github.com/stretchr/testify/assert"
	"math"
	"testing"
)

func TestSlope(t *testing.T) {
	assert := assert.New(t)
	line := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 0, Y: 1},
	}
	assert.Equal(math.Inf(1), line.Slope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 0, Y: -1},
	}
	assert.Equal(math.Inf(-1), line.Slope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 0.0000001, Y: 1},
	}
	assert.Equal(math.Inf(1), line.Slope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: -0.0000001, Y: 1},
	}
	assert.Equal(math.Inf(-1), line.Slope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 2},
	}
	assert.Equal(float64(2), line.Slope())
}

func TestInvertedSlope(t *testing.T) {
	assert := assert.New(t)
	line := Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 0, Y: 1},
	}
	assert.Equal(float64(0), line.InvertedSlope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 0, Y: -1},
	}
	assert.Equal(float64(0), line.InvertedSlope())
	line = Line{
		P: Point{X: 0, Y: 0},
		Q: Point{X: 1, Y: 2},
	}
	assert.Equal(-0.5, line.InvertedSlope())
}

func TestPerpendicularBisector(t *testing.T) {
	assert := assert.New(t)
	line := Line{
		P: Point{X: -1, Y: 0},
		Q: Point{X: 1, Y: 0},
	}
	pb := line.PerpendicularBisector(2)
	assert.Equal(
		Line{P: Point{X: 0, Y: 1}, Q: Point{X: 0, Y: -1}},
		pb,
	)
	line = Line{
		P: Point{X: 0, Y: -1},
		Q: Point{X: 0, Y: 1},
	}
	pb = line.PerpendicularBisector(2)
	assert.Equal(
		Line{P: Point{X: -1, Y: 0}, Q: Point{X: 1, Y: 0}},
		pb,
	)
}

func TestCurveLerp(t *testing.T) {
	assert := assert.New(t)
	var curve Curve
	points := []Point{
		{X: 0, Y: 0},
		{X: 1, Y: 0},
		{X: 2, Y: 0},
		{X: 10, Y: 0},
	}
	curve.Points = points
	lerp := curve.Lerp(0.0)
	assert.Equal(Point{X: 0, Y: 0}, lerp)
	lerp = curve.Lerp(1.0)
	assert.Equal(Point{X: 10, Y: 0}, lerp)
	lerp = curve.Lerp(0.25)
	assert.Equal(Point{X: 2.5, Y: 0}, lerp)
}
