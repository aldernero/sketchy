package sketchy

import (
	"fmt"
	"image"

	"github.com/aldernero/debugui"
)

func (s *Sketch) openFloatSliderRangeModal(idx int) {
	if idx < 0 || idx >= len(s.FloatSliders) {
		return
	}
	sl := &s.FloatSliders[idx]
	s.sliderRangeModalOpen = true
	s.sliderRangeModalFloat = true
	s.sliderRangeModalIdx = idx
	s.sliderRangeEditMinF = sl.MinVal
	s.sliderRangeEditMaxF = sl.MaxVal
	s.sliderRangeEditIncrF = sl.Incr
	s.sliderRangeModalErr = ""
}

func (s *Sketch) openIntSliderRangeModal(idx int) {
	if idx < 0 || idx >= len(s.IntSliders) {
		return
	}
	sl := &s.IntSliders[idx]
	s.sliderRangeModalOpen = true
	s.sliderRangeModalFloat = false
	s.sliderRangeModalIdx = idx
	s.sliderRangeEditMinI = sl.MinVal
	s.sliderRangeEditMaxI = sl.MaxVal
	s.sliderRangeEditIncrI = sl.Incr
	s.sliderRangeModalErr = ""
}

func (s *Sketch) closeSliderRangeModal() {
	s.sliderRangeModalOpen = false
	s.sliderRangeModalErr = ""
}

func (s *Sketch) applyFloatSliderRangeOK() {
	i := s.sliderRangeModalIdx
	if i < 0 || i >= len(s.FloatSliders) {
		s.closeSliderRangeModal()
		return
	}
	min := s.sliderRangeEditMinF
	max := s.sliderRangeEditMaxF
	incr := s.sliderRangeEditIncrF
	if incr <= 0 {
		s.sliderRangeModalErr = "step must be > 0"
		return
	}
	if min > max {
		s.sliderRangeModalErr = "min must be <= max"
		return
	}
	sl := &s.FloatSliders[i]
	sl.MinVal = min
	sl.MaxVal = max
	sl.Incr = incr
	sl.CalcDigits()
	sl.Val = clampFloat(sl.Val, min, max)
	sl.syncTextBufFromVal()
	s.dirty = true
	s.closeSliderRangeModal()
}

func (s *Sketch) applyIntSliderRangeOK() {
	i := s.sliderRangeModalIdx
	if i < 0 || i >= len(s.IntSliders) {
		s.closeSliderRangeModal()
		return
	}
	min := s.sliderRangeEditMinI
	max := s.sliderRangeEditMaxI
	incr := s.sliderRangeEditIncrI
	if incr <= 0 {
		s.sliderRangeModalErr = "step must be > 0"
		return
	}
	if min > max {
		s.sliderRangeModalErr = "min must be <= max"
		return
	}
	sl := &s.IntSliders[i]
	sl.MinVal = min
	sl.MaxVal = max
	sl.Incr = incr
	sl.Val = clampInt(sl.Val, min, max)
	sl.syncTextBufFromVal()
	s.dirty = true
	s.closeSliderRangeModal()
}

func (s *Sketch) drawSliderRangeModal(ctx *debugui.Context) {
	if !s.sliderRangeModalOpen {
		return
	}
	var title string
	if s.sliderRangeModalFloat {
		if s.sliderRangeModalIdx < 0 || s.sliderRangeModalIdx >= len(s.FloatSliders) {
			s.closeSliderRangeModal()
			return
		}
		title = s.FloatSliders[s.sliderRangeModalIdx].Name
	} else {
		if s.sliderRangeModalIdx < 0 || s.sliderRangeModalIdx >= len(s.IntSliders) {
			s.closeSliderRangeModal()
			return
		}
		title = s.IntSliders[s.sliderRangeModalIdx].Name
	}

	ctx.Window("Slider range", image.Rect(240, 120, 540, 340), func(layout debugui.ContainerLayout) {
		ctx.BringRootContainerToFront()
		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text(fmt.Sprintf("%s - min, max, step", title))

		if s.sliderRangeModalFloat {
			ctx.SetGridLayout([]int{40, -1}, nil)
			ctx.Text("Min")
			ctx.IDScope("srfmin", func() {
				ctx.NumberFieldF(&s.sliderRangeEditMinF, 0.1, 8).On(func() {})
			})
			ctx.Text("Max")
			ctx.IDScope("srfmax", func() {
				ctx.NumberFieldF(&s.sliderRangeEditMaxF, 0.1, 8).On(func() {})
			})
			ctx.Text("Step")
			ctx.IDScope("srfinc", func() {
				ctx.NumberFieldF(&s.sliderRangeEditIncrF, 0.1, 8).On(func() {})
			})
		} else {
			ctx.SetGridLayout([]int{40, -1}, nil)
			ctx.Text("Min")
			ctx.IDScope("srimin", func() {
				ctx.NumberField(&s.sliderRangeEditMinI, 1).On(func() {})
			})
			ctx.Text("Max")
			ctx.IDScope("srimax", func() {
				ctx.NumberField(&s.sliderRangeEditMaxI, 1).On(func() {})
			})
			ctx.Text("Step")
			ctx.IDScope("sriinc", func() {
				ctx.NumberField(&s.sliderRangeEditIncrI, 1).On(func() {})
			})
		}

		if s.sliderRangeModalErr != "" {
			ctx.SetGridLayout([]int{-1}, nil)
			ctx.Text(s.sliderRangeModalErr)
		}

		modalActionRow(ctx, "OK", func() { s.closeSliderRangeModal() }, func() {
			if s.sliderRangeModalFloat {
				s.applyFloatSliderRangeOK()
			} else {
				s.applyIntSliderRangeOK()
			}
		})
	})
}
