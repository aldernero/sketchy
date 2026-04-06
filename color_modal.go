package sketchy

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/aldernero/debugui"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"github.com/lucasb-eyer/go-colorful"
)

// modalHueStripHeight is the layout height of the hue bar and the final-color preview strip.
const modalHueStripHeight = 22

func (s *Sketch) openColorModal(i int) {
	if i < 0 || i >= len(s.ColorPickers) {
		return
	}
	s.colorModalIdx = i
	cp := &s.ColorPickers[i]
	s.modalR, s.modalG, s.modalB = cp.r, cp.g, cp.b
	c := colorful.Color{
		R: float64(cp.r) / 255,
		G: float64(cp.g) / 255,
		B: float64(cp.b) / 255,
	}
	s.modalH, s.modalS, s.modalV = c.Hsv()
	s.modalHexBuf = cp.GetHex()
	s.modalErr = ""
}

func (s *Sketch) closeColorModal() {
	s.colorModalIdx = -1
}

func (s *Sketch) syncModalRGBFromHSV() {
	c := colorful.Hsv(s.modalH, s.modalS, s.modalV)
	r16, g16, b16, _ := c.RGBA()
	s.modalR = int(r16 >> 8)
	s.modalG = int(g16 >> 8)
	s.modalB = int(b16 >> 8)
	s.modalHexBuf = fmt.Sprintf("#%02X%02X%02X", s.modalR, s.modalG, s.modalB)
}

func (s *Sketch) syncModalHSVFromRGB() {
	s.modalR = clampInt(s.modalR, 0, 255)
	s.modalG = clampInt(s.modalG, 0, 255)
	s.modalB = clampInt(s.modalB, 0, 255)
	c := colorful.Color{
		R: float64(s.modalR) / 255,
		G: float64(s.modalG) / 255,
		B: float64(s.modalB) / 255,
	}
	s.modalH, s.modalS, s.modalV = c.Hsv()
	s.modalHexBuf = fmt.Sprintf("#%02X%02X%02X", s.modalR, s.modalG, s.modalB)
}

// tryApplyModalHexFromBuf updates modal RGB and HSV when modalHexBuf contains a parseable hex
// color, so sliders and pickers stay in sync while the user types (not only on blur/Enter).
func (s *Sketch) tryApplyModalHexFromBuf() {
	if s.colorModalIdx < 0 {
		return
	}
	h := strings.TrimSpace(s.modalHexBuf)
	if h == "" {
		return
	}
	cl, err := colorful.Hex(h)
	if err != nil {
		return
	}
	r16, g16, b16, _ := cl.RGBA()
	nr, ng, nb := int(r16>>8), int(g16>>8), int(b16>>8)
	if nr == s.modalR && ng == s.modalG && nb == s.modalB {
		return
	}
	s.modalR, s.modalG, s.modalB = nr, ng, nb
	c := colorful.Color{
		R: float64(s.modalR) / 255,
		G: float64(s.modalG) / 255,
		B: float64(s.modalB) / 255,
	}
	s.modalH, s.modalS, s.modalV = c.Hsv()
	s.modalErr = ""
}

func (s *Sketch) drawColorModal(ctx *debugui.Context) {
	if s.colorModalIdx < 0 || s.colorModalIdx >= len(s.ColorPickers) {
		return
	}
	ctx.Window("Color picker", image.Rect(280, 60, 620, 500), func(layout debugui.ContainerLayout) {
		ctx.BringRootContainerToFront()
		drawModalCrosshair := func(screen *ebiten.Image, cx, cy int, scale float32) {
			vector.StrokeRect(
				screen,
				float32(cx)*scale-2*scale,
				float32(cy)*scale-2*scale,
				float32(5)*scale,
				float32(5)*scale,
				scale,
				color.RGBA{255, 255, 255, 220},
				false,
			)
		}

		s.tryApplyModalHexFromBuf()

		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("Hue / Saturation / Value")
		ctx.SetGridLayout([]int{-1, -1, -1}, nil)
		ctx.IDScope("mh", func() {
			ctx.SliderF(&s.modalH, 0, 360, 1, 0).On(func() {
				s.syncModalRGBFromHSV()
				s.modalErr = ""
			})
		})
		ctx.IDScope("ms", func() {
			ctx.SliderF(&s.modalS, 0, 1, 0.01, 2).On(func() {
				s.syncModalRGBFromHSV()
				s.modalErr = ""
			})
		})
		ctx.IDScope("mv", func() {
			ctx.SliderF(&s.modalV, 0, 1, 0.01, 2).On(func() {
				s.syncModalRGBFromHSV()
				s.modalErr = ""
			})
		})

		ctx.SetGridLayout([]int{-1}, []int{modalHueStripHeight})
		ctx.IDScope("huebar", func() {
			ctx.DragArea(
				func(screen *ebiten.Image, bounds image.Rectangle) {
					scale := ctx.Scale()
					dx := bounds.Dx()
					dy := bounds.Dy()
					if dx <= 0 || dy <= 0 {
						return
					}
					for x := 0; x < dx; x++ {
						hue := float64(x) / float64(max(1, dx-1)) * 360
						col := colorful.Hsv(hue, 1, 1)
						r16, g16, b16, _ := col.RGBA()
						cl := color.RGBA{uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8), 255}
						vector.DrawFilledRect(
							screen,
							float32((bounds.Min.X+x)*scale),
							float32(bounds.Min.Y*scale),
							float32(scale),
							float32(bounds.Dy()*scale),
							cl,
							false,
						)
					}
					hx := int(s.modalH / 360 * float64(max(0, dx-1)))
					cx := bounds.Min.X + hx
					cy := bounds.Min.Y + dy/2
					drawModalCrosshair(screen, cx, cy, float32(scale))
				},
				func(bounds image.Rectangle, pos image.Point) bool {
					dx := bounds.Dx()
					if dx <= 1 {
						return false
					}
					nh := float64(pos.X-bounds.Min.X) / float64(dx-1) * 360
					if nh < 0 {
						nh = 0
					}
					if nh > 360 {
						nh = 360
					}
					const eps = 1e-4
					if math.Abs(s.modalH-nh) < eps {
						return false
					}
					s.modalH = nh
					return true
				},
			).On(func() {
				s.syncModalRGBFromHSV()
				s.modalErr = ""
			})
		})

		ctx.SetGridLayout([]int{-1}, []int{120})
		ctx.IDScope("svplane", func() {
			ctx.DragArea(
				func(screen *ebiten.Image, bounds image.Rectangle) {
					scale := ctx.Scale()
					dx := bounds.Dx()
					dy := bounds.Dy()
					if dx <= 0 || dy <= 0 {
						return
					}
					for y := 0; y < dy; y++ {
						v := 1 - float64(y)/float64(max(1, dy-1))
						for x := 0; x < dx; x++ {
							sat := float64(x) / float64(max(1, dx-1))
							col := colorful.Hsv(s.modalH, sat, v)
							r16, g16, b16, _ := col.RGBA()
							cl := color.RGBA{uint8(r16 >> 8), uint8(g16 >> 8), uint8(b16 >> 8), 255}
							vector.DrawFilledRect(
								screen,
								float32((bounds.Min.X+x)*scale),
								float32((bounds.Min.Y+y)*scale),
								float32(scale),
								float32(scale),
								cl, false,
							)
						}
					}
					sx := int(s.modalS * float64(max(0, dx-1)))
					sy := int((1 - s.modalV) * float64(max(0, dy-1)))
					cx := bounds.Min.X + sx
					cy := bounds.Min.Y + sy
					drawModalCrosshair(screen, cx, cy, float32(scale))
				},
				func(bounds image.Rectangle, pos image.Point) bool {
					dx := bounds.Dx()
					dy := bounds.Dy()
					if dx <= 1 || dy <= 1 {
						return false
					}
					ns := float64(pos.X-bounds.Min.X) / float64(dx-1)
					nv := 1 - float64(pos.Y-bounds.Min.Y)/float64(dy-1)
					if ns < 0 {
						ns = 0
					}
					if ns > 1 {
						ns = 1
					}
					if nv < 0 {
						nv = 0
					}
					if nv > 1 {
						nv = 1
					}
					const eps = 1e-5
					if math.Abs(s.modalS-ns) < eps && math.Abs(s.modalV-nv) < eps {
						return false
					}
					s.modalS, s.modalV = ns, nv
					return true
				},
			).On(func() {
				s.syncModalRGBFromHSV()
				s.modalErr = ""
			})
		})

		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("RGB")
		ctx.SetGridLayout([]int{-1, -1, -1}, nil)
		ctx.IDScope("mr", func() {
			ctx.NumberField(&s.modalR, 1).On(func() {
				s.syncModalHSVFromRGB()
				s.modalErr = ""
			})
		})
		ctx.IDScope("mg", func() {
			ctx.NumberField(&s.modalG, 1).On(func() {
				s.syncModalHSVFromRGB()
				s.modalErr = ""
			})
		})
		ctx.IDScope("mb", func() {
			ctx.NumberField(&s.modalB, 1).On(func() {
				s.syncModalHSVFromRGB()
				s.modalErr = ""
			})
		})

		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("Hex")
		hb := &s.modalHexBuf
		ctx.TextField(hb).On(func() {
			h := strings.TrimSpace(*hb)
			if _, err := colorful.Hex(h); err != nil {
				s.modalErr = "invalid hex"
				*hb = fmt.Sprintf("#%02X%02X%02X", s.modalR, s.modalG, s.modalB)
				return
			}
			s.modalErr = ""
			s.syncModalHSVFromRGB()
		})
		s.tryApplyModalHexFromBuf()

		ctx.SetGridLayout([]int{-1}, []int{modalHueStripHeight})
		ctx.IDScope("mprev", func() {
			ctx.Clickable(func(bounds image.Rectangle) {
				ctx.DrawSolidRect(bounds, color.RGBA{
					uint8(s.modalR), uint8(s.modalG), uint8(s.modalB), 255,
				})
			}).On(func() {})
		})

		if s.modalErr != "" {
			ctx.Text(s.modalErr)
		}

		modalActionRow(ctx, "OK", func() { s.closeColorModal() }, func() {
			i := s.colorModalIdx
			if i >= 0 && i < len(s.ColorPickers) {
				cp := &s.ColorPickers[i]
				cp.r = clampInt(s.modalR, 0, 255)
				cp.g = clampInt(s.modalG, 0, 255)
				cp.b = clampInt(s.modalB, 0, 255)
				cp.syncFromRGB()
				s.dirty = true
				s.DidColorPickersChange = true
				if i == s.builtinColorBGIdx {
					s.DefaultBackground = cp.GetColor()
				}
				if i == s.builtinColorFGIdx {
					s.DefaultForeground = cp.GetColor()
				}
			}
			s.closeColorModal()
		})
	})
}
