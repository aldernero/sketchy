# Builtin Goodies

Some of Sketchy's best features are builtin and designed to make iterating on your designs quick and easy. This
document covers those builtin features.

# Saving designs as images and screenshots

There are 3 different kinds of screenshots, each with a builtin keybinding:

- "s" key: saves the current frame as a PNG
- "q" key: saves the sketch area as a PNG
- "Esc" key: saves the entire Sketchy window as a PNG.

The first two options create a file with the format `"<prefix>_<timestamp>.png"`. The last one use a builtin Ebiten 
screenshot feature and creates a file with the format `"screenshot_<timestamp>.png"`

Most of the time you would use the first option, the "s" key. This saves the current `gg` context as a png image.
The `gg` context is created each frame, so this would save the current frame and ignore things like the sketch
area outline (if there is one) and keep transparency.

The second option, the "q" key, is useful when you want to capture the all the drawing done so far, i.e. if you
are using the `DisableClearBetweenFrames` feature. In this case you should also set `SketchOutlineColor` to `""` or
to the same values as `SketchBackgroundColor`, as this screenshot captures all pixels in the sketch area.

The last option would be useful if you want to capture the entire Sketchy window for some reason, for example to
use in a blog post or something.

# Random number generator (with noise)

The sketch struct has a builtin random number generator `s.Rand` in the `update` and `draw` functions. It's seeded
with either `RandomSeed` in the configuration file or `-s <seed>` from the command line. The CLI takes precedent, 
and if neither is specified it's seeded with 0.

The PRNG is built on top of [opensimplex-go](https://github.com/ojrac/opensimplex-go) and extends the functionality
to include fractal noise and comes with a lot of helper functions that are useful in generative art. See
[random.go](../random.go) for the full specification. However, here's a listing of some of the functions:

```
SetSeed(i int64)
SetNoiseScaleX(scale float64)
SetNoiseScaleY(scale float64)
SetNoiseScaleZ(scale float64)
SetNoiseOffsetX(offset float64)
SetNoiseOffsetY(offset float64)
SetNoiseOffsetZ(offset float64)
SetNoiseOctaves(i int)
SetNoisePersistence(p float64)
SignedNoise2D(x float64, y float64) float64
SignedNoise3D(x float64, y float64, z float64) float64
Noise2D(x float64, y float64) float64
Noise3D(x float64, y float64, z float64) float64
```

There are 3 builtin keybindings to control the seed:
- "Up Arrow": increments the seed
- "Down Arrow": decrements the seed
- "/": sets a random seed

This builtin capability was inspired by [vsketch](https://github.com/abey79/vsketch), which has a GUI control with
similar functionality that I found invaluable when experimenting with designs.

# Miscellaneous Key Bindings

There are two other builtin key bindings:

- "Space Bar" : dumps the current state to the terminal. The output
includes the current value for all GUI controls, as well as the current value of the PRNG seed.
- "c" Key : saves the current sketch configuration as a JSON file, with the format `"<prefix>_config_<timestamp>.json`. This can then be used to get back to the same state at started using the `-c <config file>` CLI flag.