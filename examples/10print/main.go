package main

import (
	"flag"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Define a grid of tiles and a random number generator
// for determining whether a tile is "\", "/", or empty
type Tile int

const (
	EmptyTile Tile = iota
	BackSlash
	ForwardSlash
)

type Truchet struct {
	rows  int
	cols  int
	tiles []Tile
}

func (t *Truchet) init(r int, c int, s *sketchy.Sketch) {
	t.rows = r
	t.cols = c
	t.tiles = []Tile{}
	for i := 0; i < r*c; i++ {
		// Use opensimplex noise to make it more interesting
		noise := s.Rand.Noise2D(float64(i%c), float64(i/r))
		tile := BackSlash
		if noise > 0.5 {
			tile = ForwardSlash
		}
		t.tiles = append(t.tiles, tile)
	}
}

func (t *Truchet) flip(r int, c int) {
	i := r*t.cols + c
	val := t.tiles[i]
	switch val {
	case EmptyTile:
		t.tiles[i] = BackSlash
	case BackSlash:
		t.tiles[i] = ForwardSlash
	case ForwardSlash:
		t.tiles[i] = EmptyTile
	}
}

func setup(s *sketchy.Sketch) {
	cellSize := s.Slider("cellSize")
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(int(s.Slider("octaves")))
	s.Rand.SetNoisePersistence(s.Slider("persistence"))
	s.Rand.SetNoiseLacunarity(s.Slider("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Slider("xscale"))
	s.Rand.SetNoiseScaleY(s.Slider("yscale"))
	s.Rand.SetNoiseOffsetX(s.Slider("xoffset"))
	s.Rand.SetNoiseOffsetY(s.Slider("yoffset"))
	board.init(int(s.SketchCanvas.W/cellSize), int(s.SketchCanvas.H/cellSize), s)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Need to reinitialize board
	if s.DidControlsChange {
		setup(s)
	}
	// flip one tile
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if s.PointInSketchArea(float64(x), float64(y)) {
			p := s.CanvasCoords(float64(x), float64(y))
			c := int(math.Floor(float64(board.rows) * p.X / s.SketchCanvas.W))
			r := int(math.Floor(float64(board.cols) * p.Y / s.SketchCanvas.H))
			board.flip(r, c)
		}
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(color.White)
	c.SetStrokeWidth(0.7)
	dx := c.Width() / float64(board.cols)
	dy := c.Height() / float64(board.rows)
	for i, t := range board.tiles {
		x := dx * float64(i%board.cols)
		y := dy * float64(i/board.rows)
		switch t {
		case BackSlash:
			c.MoveTo(x, y)
			c.LineTo(x+dx, y+dy)
		case ForwardSlash:
			c.MoveTo(x, y+dy)
			c.LineTo(x+dx, y)
		}
	}
	c.Stroke()
}

var board Truchet

func main() {
	var configFile string
	var prefix string
	var randomSeed int64
	flag.StringVar(&configFile, "c", "sketch.json", "Sketch config file")
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s, err := sketchy.NewSketchFromFile(configFile)
	if err != nil {
		log.Fatal(err)
	}
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	setup(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
