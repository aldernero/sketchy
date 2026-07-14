package sketchy

import (
	"image"
	"image/draw"
	"math"

	"github.com/tdewolff/canvas"
	"github.com/tdewolff/canvas/renderers/rasterizer"
)

// fastImageRasterizer wraps the canvas rasterizer to blit images whose
// device-space transform is a pure integer translation (no rotation, shear,
// or scaling) with image/draw. The rasterizer routes every image through
// CatmullRom resampling, which is ~13x slower and lossless-equivalent at 1:1,
// so full-frame pixel sketches (DrawImage of a sketch-sized image) pay a large
// per-frame cost for nothing. Anything else falls through to the rasterizer.
// Only valid for the linear color space, where the rasterizer applies no
// gamma conversion to source images.
type fastImageRasterizer struct {
	*rasterizer.Rasterizer
	dst *image.RGBA
	res canvas.Resolution
}

func (r *fastImageRasterizer) RenderImage(img image.Image, m canvas.Matrix) {
	if !tryFastImageBlit(r.dst, img, m, r.res) {
		r.Rasterizer.RenderImage(img, m)
	}
}

// tryFastImageBlit draws img with image/draw when m maps image pixels 1:1
// onto raster pixels at an integer offset, mirroring the rasterizer's
// coordinate math (canvas y-up in mm, image anchored at its top-left).
// It reports false when the transform needs real resampling.
func tryFastImageBlit(dst *image.RGBA, img image.Image, m canvas.Matrix, res canvas.Resolution) bool {
	const scaleEps = 1e-6 // unit scale within float noise
	const posEps = 1e-3   // offsets this close to a pixel corner snap losslessly
	if math.Abs(m[0][1]) > scaleEps || math.Abs(m[1][0]) > scaleEps {
		return false
	}
	dpmm := res.DPMM()
	if math.Abs(m[0][0]*dpmm-1) > scaleEps || math.Abs(m[1][1]*dpmm-1) > scaleEps {
		return false
	}
	size := img.Bounds().Size()
	origin := m.Dot(canvas.Point{X: 0, Y: float64(size.Y)}).Mul(dpmm)
	tx := origin.X
	ty := float64(dst.Bounds().Dy()) - origin.Y
	rx, ry := math.Round(tx), math.Round(ty)
	if math.Abs(tx-rx) > posEps || math.Abs(ty-ry) > posEps {
		return false
	}
	off := image.Pt(int(rx), int(ry))
	draw.Draw(dst, image.Rectangle{Min: off, Max: off.Add(size)}, img, img.Bounds().Min, draw.Over)
	return true
}
