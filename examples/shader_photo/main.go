// Command shader_photo demonstrates sketchy's //sketchy:image directive:
// loading a photo as a shader source image for GPU-based photo
// manipulation (Sobel edge detection, HSL hue/saturation/lightness shift).
// See docs/shaders.md.
package main

import (
	"log"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	s := sketchy.New(sketchy.Config{
		Title:        "Shader Photo",
		SketchWidth:  1080,
		SketchHeight: 1080,
		ShaderPath:   "fragment.kage",
	})
	s.Init()

	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetVsyncEnabled(true)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
