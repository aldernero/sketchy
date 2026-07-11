// Command vshot renders a single frame of a sketchy example or visual test headlessly and writes it
// to a PNG. It runs the target package as an Ebitengine "vmhost" guest in a hidden window, so no
// changes to the example's source are needed and no visible window ever appears.
//
// Adapted from the "run-ebitengine-app-headless" skill shipped inside the ebiten module
// (.agents/skills/run-ebitengine-app-headless/driver/main.go in github.com/hajimehoshi/ebiten/v2).
//
// Copyright 2026 The Ebitengine Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"errors"
	"flag"
	"fmt"
	"image"
	"image/png"
	"log/slog"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/vmhost"
)

var (
	pkg       = flag.String("pkg", "", "package path of the guest sketch to screenshot, e.g. ./examples/simple")
	ticks     = flag.Int("ticks", 60, "number of ticks to run before dumping the frame and exiting")
	out       = flag.String("out", "frame.png", "PNG output path for the final frame")
	logicalW  = flag.Int("w", 800, "logical screen width in device-independent pixels")
	logicalH  = flag.Int("h", 800, "logical screen height in device-independent pixels")
	seed      = flag.Int64("seed", 1, "random seed passed to the guest via -s, for deterministic output")
	clicks    = flag.Int("clicks", 0, "number of random left-click press/release cycles to inject before the final ticks")
	clickSeed = flag.Int64("click-seed", 1234567890, "seed for the RNG generating random click coordinates")
)

// driver is the host: an ebiten.Game that drives one guest and composites its final frame into its own
// (hidden) screen image. It does all its work in a single Update, then terminates.
type driver struct {
	guest    *vmhost.GuestSession
	screen   *ebiten.Image
	dumped   bool
	closed   bool
	closeErr error
}

func (d *driver) Update() error {
	defer func() {
		if err := d.guest.Close(); err != nil {
			d.closeErr = err
		}
		d.closed = true
	}()

	// Create the host-owned screen and size the guest to it. SetOutsideScreen must precede AdvanceTicks.
	scale := ebiten.Monitor().DeviceScaleFactor()
	d.screen = ebiten.NewImage(int(float64(*logicalW)*scale), int(float64(*logicalH)*scale))
	if err := d.guest.SetOutsideScreen(d.screen); err != nil {
		return err
	}

	injectClicks(d.guest, *clicks, *clickSeed, *logicalW, *logicalH)
	d.guest.AdvanceTicks(*ticks)

	// Render the final frame. WaitFrame blocks until every queued tick has run and the frame is rendered.
	d.guest.AdvanceFrame()
	if !d.guest.WaitFrame() {
		return ebiten.Termination
	}
	if !d.guest.CompositeFrame() {
		return errors.New("compositing the guest frame failed")
	}

	if err := d.dump(); err != nil {
		return err
	}
	d.dumped = true
	return ebiten.Termination
}

// injectClicks presses and releases the left mouse button at n random points within a w x h logical
// screen, using a seeded RNG so the sequence is identical across runs. Each press/release is separated
// by a tick so the guest's edge-triggered input detection (e.g. inpututil-style just-pressed checks)
// observes both events.
func injectClicks(guest *vmhost.GuestSession, n int, seed int64, w, h int) {
	if n <= 0 {
		return
	}
	rng := rand.New(rand.NewSource(seed))
	for i := 0; i < n; i++ {
		x := rng.Float64() * float64(w)
		y := rng.Float64() * float64(h)
		guest.MoveCursor(x, y)
		guest.AdvanceTicks(1)
		guest.PressMouseButton(ebiten.MouseButtonLeft)
		guest.AdvanceTicks(1)
		guest.ReleaseMouseButton(ebiten.MouseButtonLeft)
		guest.AdvanceTicks(1)
	}
}

// dump writes the current screen to a PNG. ReadPixels returns premultiplied-alpha RGBA, which is
// identical to non-premultiplied bytes for an opaque screen (the common case).
func (d *driver) dump() error {
	b := d.screen.Bounds()
	img := image.NewRGBA(b)
	d.screen.ReadPixels(img.Pix)
	f, err := os.Create(*out)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func (d *driver) Draw(screen *ebiten.Image) {
	// Nothing to present: the host window is hidden.
}

func (d *driver) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func main() {
	if err := xmain(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func xmain() error {
	flag.Parse()
	if *pkg == "" {
		return errors.New("specify the guest package with -pkg, e.g. -pkg ./examples/simple")
	}
	guestDir, err := filepath.Abs(*pkg)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(*out), 0o755); err == nil {
		if abs, err := filepath.Abs(*out); err == nil {
			*out = abs
		}
	}

	// A short temp dir keeps the unix socket path within the OS limit (~104 bytes on macOS).
	dir, err := os.MkdirTemp("", "vshot")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(dir)
	}()

	ln, err := net.Listen("unix", filepath.Join(dir, "g.sock"))
	if err != nil {
		return err
	}
	defer func() {
		_ = ln.Close()
	}()

	endpoint, err := vmhost.EndpointURLFromAddr(ln.Addr())
	if err != nil {
		return err
	}

	// Build the guest with the ebitenginevm tag so its RunGame connects to this host instead of opening
	// a window. The app's own source is unchanged.
	guestBin := filepath.Join(dir, "guest")
	build := exec.Command("go", "build", "-tags", "ebitenginevm", "-o", guestBin, *pkg)
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return fmt.Errorf("building the guest failed: %w", err)
	}

	// Run from the guest package's own directory so relative asset paths (e.g. photo_stripes/cloud.png)
	// and sketch.db/saves resolve exactly as they do when the example is run normally.
	cmd := exec.Command(guestBin, "-s", strconv.FormatInt(*seed, 10))
	cmd.Dir = guestDir
	cmd.Env = append(os.Environ(), "EBITENGINE_VM_ENDPOINT="+endpoint)
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("starting the guest failed: %w", err)
	}
	var processDone bool
	defer func() {
		if processDone {
			return
		}
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
	}()

	if dl, ok := ln.(interface{ SetDeadline(time.Time) error }); ok {
		if err := dl.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
			return err
		}
	}
	conn, err := ln.Accept()
	if err != nil {
		return fmt.Errorf("accepting the guest failed: %w", err)
	}

	guest, err := vmhost.NewGuestSession(conn, &vmhost.NewGuestSessionOptions{IdleTimeout: 10 * time.Second})
	if err != nil {
		return err
	}

	ebiten.SetWindowVisible(false)
	d := &driver{guest: guest}
	runErr := ebiten.RunGame(d)

	if d.closeErr != nil {
		slog.Warn("closing the guest session failed", "err", d.closeErr)
	}
	if !d.closed {
		_ = cmd.Process.Kill()
	}
	if err := cmd.Wait(); err != nil {
		slog.Warn("waiting for the guest process failed", "err", err)
	}
	processDone = true

	if runErr != nil {
		return fmt.Errorf("host run failed: %w", runErr)
	}
	if err := guest.Err(); err != nil && !errors.Is(err, ebiten.Termination) {
		return fmt.Errorf("guest session failed: %w", err)
	}
	if !d.dumped {
		if err := guest.Err(); err != nil {
			return fmt.Errorf("the guest ended before the frame was captured: %w", err)
		}
		return errors.New("no frame was captured; the run must advance at least one tick")
	}
	slog.Info("wrote frame", "path", *out, "ticks", *ticks)
	return nil
}
