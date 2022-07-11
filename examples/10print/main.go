package main

import (
	"flag"
	gaul "github.com/aldernero/gaul"
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
	rng   gaul.Rng
}

func (t *Truchet) init(r int, c int) {
	t.rows = r
	t.cols = c
	t.tiles = []Tile{}
	for i := 0; i < r*c; i++ {
		// Use opensimplex noise to make it more interesting
		noise := t.rng.Noise2D(100*float64(i%c), 100*float64(i/r))
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

func reset(s *sketchy.Sketch) {
	cellSize := s.Slider("cellSize")
	board.rng.SetSeed(s.RandomSeed)
	board.rng.SetNoiseOctaves(int(s.Slider("octaves")))
	board.rng.SetNoisePersistence(s.Slider("persistence"))
	board.rng.SetNoiseLacunarity(s.Slider("lacunarity"))
	board.rng.SetNoiseScaleX(s.Slider("xscale"))
	board.rng.SetNoiseScaleY(s.Slider("yscale"))
	board.rng.SetNoiseOffsetX(s.Slider("xoffset"))
	board.rng.SetNoiseOffsetY(s.Slider("yoffset"))
	board.init(int(s.SketchCanvas.W/cellSize), int(s.SketchCanvas.H/cellSize))
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Need to reinitialize board
	if s.DidControlsChange {
		reset(s)
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
	reset(s)
	ebiten.SetWindowSize(int(s.ControlWidth+s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
