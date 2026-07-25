// Command reaction_diffusion demonstrates sketchy's ping-pong shader mode
// (Config.StatePath): a Gray-Scott reaction-diffusion simulation running
// entirely on the GPU, split into a simulation pass (state.kage) and a
// display pass (fragment.kage). See docs/shaders.md.
package main

import (
	"image/color"
	"log"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

const resetFolder = "ReactionDiffusion"

func vec3(v gaul.Vec3) []float32 {
	return []float32{float32(v.X), float32(v.Y), float32(v.Z)}
}

func colorRGB(c color.Color) (r, g, b float32) {
	rr, gg, bb, _ := c.RGBA()
	return float32(rr) / 65535, float32(gg) / 65535, float32(bb) / 65535
}

func main() {
	// resetPulse is true for exactly one tick after a "Reset" click.
	// state.kage has no momentary-button directive, so the click is a real
	// UI button forwarded via ExtraUniforms; we also ClearState() so the
	// reseed cannot be confused with a frozen buffer.
	var resetPulse, resetWasChecked bool

	s := sketchy.New(sketchy.Config{
		Title:        "Reaction-Diffusion",
		SketchWidth:  1080,
		SketchHeight: 1080,
		ShaderPath:   "fragment.kage",
		StatePath:    "state.kage",
	})
	s.BuildUI = func(_ *sketchy.Sketch, ui *sketchy.UI) {
		ui.Folder(resetFolder, func() {
			ui.Button("Reset")
		})
	}
	s.Updater = func(s *sketchy.Sketch) {
		checked := s.GetBool(resetFolder, "Reset")
		resetPulse = checked && !resetWasChecked
		resetWasChecked = checked
		if checked {
			s.SetBool(resetFolder, "Reset", false)
		}
		if resetPulse {
			s.ClearState()
		}
	}
	s.ExtraUniforms = func(s *sketchy.Sketch) map[string]any {
		reset := 0.0
		if resetPulse {
			reset = 1.0
		}
		sp := s.SinePalette
		hsv := 0.0
		if sp.Space == gaul.ColorSpaceHSV {
			hsv = 1.0
		}
		// Fragment can't call DiscretePalette.ColorAt; ship 32 samples.
		disc := make([]float32, 32*3)
		for i := 0; i < 32; i++ {
			r, g, b := colorRGB(s.DiscretePalette.ColorAt(float64(i) / 31.0))
			disc[i*3+0] = r
			disc[i*3+1] = g
			disc[i*3+2] = b
		}
		return map[string]any{
			"Reset":   reset,
			"SineA":   vec3(sp.A),
			"SineB":   vec3(sp.B),
			"SineC":   vec3(sp.C),
			"SineD":   vec3(sp.D),
			"SineHSV": hsv,
			"Disc":    disc,
		}
	}
	s.Init()

	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(144)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
