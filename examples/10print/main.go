package main

import (
	"flag"
	"image/color"
	"log"
	"math"

	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
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
	rows     int
	cols     int
	originX  float64
	originY  float64
	cellSize float64
	tiles    []Tile
}

func (t *Truchet) init(cellSize float64, s *sketchy.Sketch) {
	t.rows = int(s.Height() / cellSize)
	t.cols = int(s.Width() / cellSize)
	t.cellSize = cellSize
	t.originX = 0.5 * (s.Width() - float64(t.cols)*cellSize)
	t.originY = 0.5 * (s.Height() - float64(t.rows)*cellSize)
	t.tiles = make([]Tile, t.rows*t.cols)
	for i := 0; i < t.rows*t.cols; i++ {
		// Use opensimplex noise to make it more interesting
		noise := s.Rand.Noise2D(float64(i%t.rows), float64(i)/float64(t.cols))
		tile := BackSlash
		if noise > 0.0 {
			tile = ForwardSlash
		}
		t.tiles[i] = tile
	}
	// Debug info removed for performance
}

func (t *Truchet) flip(x, y float64) {
	// Convert x, y to row, col
	r := int(math.Floor((y - t.originY) / t.cellSize))
	c := int(math.Floor((x - t.originX) / t.cellSize))
	i := r*t.cols + c
	if i < 0 || i >= len(t.tiles) {
		// Invalid tile index - skip
		return
	}
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

func (t *Truchet) rectForTile(i int) gaul.Rect {
	cell := t.cellSize
	x := cell*float64(i%t.cols) + t.originX
	y := cell*float64(i/t.cols) + t.originY
	return gaul.Rect{
		X: x,
		Y: y,
		W: t.cellSize,
		H: t.cellSize,
	}
}

func setup(s *sketchy.Sketch) {
	cellSize := s.GetFloat("Grid", "cellSize")
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(s.GetInt("Grid", "octaves"))
	s.Rand.SetNoisePersistence(s.GetFloat("Grid", "persistence"))
	s.Rand.SetNoiseLacunarity(s.GetFloat("Grid", "lacunarity"))
	s.Rand.SetNoiseScaleX(s.GetFloat("Grid", "xscale"))
	s.Rand.SetNoiseScaleY(s.GetFloat("Grid", "yscale"))
	s.Rand.SetNoiseOffsetX(float64(s.GetInt("Grid", "xoffset")))
	s.Rand.SetNoiseOffsetY(float64(s.GetInt("Grid", "yoffset")))
	board.init(cellSize, s)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Need to reinitialize board
	if s.DidControlsChange {
		setup(s)
		s.MarkDirty() // Mark for re-render
	}
	// flip one tile
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		p := s.CanvasCoords(float64(x), float64(y))
		board.flip(p.X, p.Y)
		s.MarkDirty() // Mark for re-render
	}
}

func draw(s *sketchy.Sketch, c *canvas.Context) {
	// Drawing code goes here
	c.SetStrokeColor(color.White)
	c.SetStrokeWidth(0.5)
	// draw board rectangle
	c.MoveTo(board.originX, board.originY)
	c.LineTo(board.originX+float64(board.cols)*board.cellSize, board.originY)
	c.LineTo(board.originX+float64(board.cols)*board.cellSize, board.originY+float64(board.rows)*board.cellSize)
	c.LineTo(board.originX, board.originY+float64(board.rows)*board.cellSize)
	c.LineTo(board.originX, board.originY)
	// draw tiles
	c.SetStrokeWidth(0.5)
	for i, t := range board.tiles {
		rect := board.rectForTile(i)
		switch t {
		case BackSlash:
			c.MoveTo(rect.X, rect.Y)
			c.LineTo(rect.X+board.cellSize, rect.Y+board.cellSize)
		case ForwardSlash:
			c.MoveTo(rect.X, rect.Y+board.cellSize)
			c.LineTo(rect.X+board.cellSize, rect.Y)
		default:
			// do nothing
		}
	}
	c.Stroke()
}

var board Truchet

func buildUI(_ *sketchy.Sketch, ui *sketchy.UI) {
	ui.Folder("Grid", func() {
		ui.FloatSlider("cellSize", 1, 60, 20, 0.5)
		ui.IntSlider("octaves", 1, 10, 2, 1)
		ui.FloatSlider("persistence", 0, 2, 0.95, 0.01)
		ui.FloatSlider("lacunarity", 0, 10, 2.7, 0.1)
		ui.FloatSlider("xscale", 0, 0.1, 0.04, 0.0001)
		ui.FloatSlider("yscale", 0, 0.1, 0.07, 0.0001)
		ui.IntSlider("xoffset", -1000, 1000, 0, 1)
		ui.IntSlider("yoffset", -1000, 1000, 0, 1)
	})
}

func main() {
	var prefix string
	var randomSeed int64
	flag.StringVar(&prefix, "p", "", "Output file prefix")
	flag.Int64Var(&randomSeed, "s", 0, "Random number generator seed")
	flag.Parse()
	s := sketchy.New(sketchy.Config{
		Title:                 "10PRINT Interaction",
		Prefix:                "sketch",
		SketchWidth:           1080,
		SketchHeight:          768,
		SketchBackgroundColor: "#1e1e1e",
		SketchOutlineColor:    "#1e1e1e",
	})
	s.BuildUI = buildUI
	if prefix != "" {
		s.Prefix = prefix
	}
	s.RandomSeed = randomSeed
	s.Updater = update
	s.Drawer = draw
	s.Init()
	setup(s)
	ww, wh := s.WindowSize()
	ebiten.SetWindowSize(ww, wh)
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
