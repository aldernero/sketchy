package main

import (
	"flag"
	"fmt"
	"github.com/aldernero/gaul"
	"github.com/aldernero/sketchy"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tdewolff/canvas"
	"image/color"
	"log"
	"math"
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
		if noise > 0.5 {
			tile = ForwardSlash
		}
		t.tiles[i] = tile
	}
	fmt.Println(t.rows, t.cols, t.originX, t.originY, t.cellSize)
}

func (t *Truchet) flip(x, y float64) {
	// Convert x, y to row, col
	r := int(math.Floor((y - t.originY) / t.cellSize))
	c := int(math.Floor((x - t.originX) / t.cellSize))
	i := r*t.cols + c
	if i < 0 || i >= len(t.tiles) {
		fmt.Println(x, y, r, c, i)
		fmt.Printf("Invalid tile index: %d Tile count: %d", i, len(t.tiles))
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
	cellSize := s.Slider("cellSize")
	s.Rand.SetSeed(s.RandomSeed)
	s.Rand.SetNoiseOctaves(int(s.Slider("octaves")))
	s.Rand.SetNoisePersistence(s.Slider("persistence"))
	s.Rand.SetNoiseLacunarity(s.Slider("lacunarity"))
	s.Rand.SetNoiseScaleX(s.Slider("xscale"))
	s.Rand.SetNoiseScaleY(s.Slider("yscale"))
	s.Rand.SetNoiseOffsetX(s.Slider("xoffset"))
	s.Rand.SetNoiseOffsetY(s.Slider("yoffset"))
	board.init(cellSize, s)
}

func update(s *sketchy.Sketch) {
	// Update logic goes here
	// Need to reinitialize board
	if s.DidControlsChange {
		setup(s)
	}
	// flip one tile
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()
		p := s.CanvasCoords(float64(x), float64(y))
		board.flip(p.X, p.Y)
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
	fmt.Println(s.SketchWidth, s.SketchHeight)
	fmt.Println(s.Width(), s.Height())
	ebiten.SetWindowSize(int(s.SketchWidth), int(s.SketchHeight))
	ebiten.SetWindowTitle("Sketchy - " + s.Title)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeDisabled)
	if err := ebiten.RunGame(s); err != nil {
		log.Fatal(err)
	}
}
