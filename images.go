package sketchy

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/aldernero/gaul/render"
)

// ImageAsset names an image file to load at Init. Path is relative to the sketch working
// directory unless absolute. Name is the key used with Image, DrawNamedImage, and DrawNamedImageAt.
type ImageAsset struct {
	Name string
	Path string
}

// Image returns a configured or runtime-registered image by name.
func (s *Sketch) Image(name string) image.Image {
	if s.images == nil {
		log.Fatalf("sketchy: unknown image %q (no images loaded)", name)
	}
	img, ok := s.images[name]
	if !ok {
		log.Fatalf("sketchy: unknown image %q", name)
	}
	return img
}

// DrawImage draws img with its top-left corner at the canvas origin, mapping
// one image pixel to one sketch pixel.
func (s *Sketch) DrawImage(c *render.Context, img image.Image) {
	s.DrawImageAt(c, 0, 0, img)
}

// DrawImageAt is like DrawImage but offsets the top-left of img to (x, y).
func (s *Sketch) DrawImageAt(c *render.Context, x, y float64, img image.Image) {
	if img == nil {
		return
	}
	c.DrawImage(img, x, y)
}

// DrawNamedImage draws a Config ImageAsset (or RegisterImage entry) at the origin.
func (s *Sketch) DrawNamedImage(c *render.Context, name string) {
	s.DrawNamedImageAt(c, 0, 0, name)
}

// DrawNamedImageAt draws a named asset at (x, y).
func (s *Sketch) DrawNamedImageAt(c *render.Context, x, y float64, name string) {
	s.DrawImageAt(c, x, y, s.Image(name))
}

// RegisterImage adds or replaces a named image after Init (e.g. for images created in code).
func (s *Sketch) RegisterImage(name string, img image.Image) {
	if name == "" {
		log.Fatal("sketchy: RegisterImage requires a non-empty name")
	}
	if img == nil {
		log.Fatalf("sketchy: RegisterImage(%q): image is nil", name)
	}
	if s.images == nil {
		s.images = make(map[string]image.Image)
	}
	s.images[name] = img
}

func (s *Sketch) loadImages() {
	if len(s.imageAssets) == 0 {
		s.images = nil
		return
	}
	s.images = make(map[string]image.Image, len(s.imageAssets))
	for _, asset := range s.imageAssets {
		if asset.Name == "" || asset.Path == "" {
			log.Fatal("sketchy: ImageAsset requires non-empty Name and Path")
		}
		if _, exists := s.images[asset.Name]; exists {
			log.Fatalf("sketchy: duplicate image name %q", asset.Name)
		}
		path := asset.Path
		if !filepath.IsAbs(path) {
			path = filepath.Join(s.workDir, path)
		}
		f, err := os.Open(path)
		if err != nil {
			log.Fatalf("sketchy: open image %q (%s): %v", asset.Name, path, err)
		}
		img, _, err := image.Decode(f)
		_ = f.Close()
		if err != nil {
			log.Fatalf("sketchy: decode image %q (%s): %v", asset.Name, path, err)
		}
		s.images[asset.Name] = img
	}
}

// imageAssetSummary is used only for clearer duplicate-name errors before the map exists.
func validateImageAssets(assets []ImageAsset) error {
	seen := make(map[string]struct{}, len(assets))
	for _, a := range assets {
		if a.Name == "" || a.Path == "" {
			return fmt.Errorf("ImageAsset requires non-empty Name and Path")
		}
		if _, ok := seen[a.Name]; ok {
			return fmt.Errorf("duplicate image name %q", a.Name)
		}
		seen[a.Name] = struct{}{}
	}
	return nil
}
