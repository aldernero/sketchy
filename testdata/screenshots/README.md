# Baseline screenshots

One golden frame per package in `examples/` and `visual_tests/`, rendered
headlessly by [`tools/vshot`](../../tools/vshot). They exist to catch
*unintended* visual change — a dependency bump, a regression in the render
path — by comparison rather than by memory.

These are not the same thing as the images in `assets/images/`, which are
documentation for the README and guides. Nothing user-facing links here.

## Using them

```sh
make check-baselines   # re-render, byte-compare, non-zero exit on any difference
make baselines         # regenerate after an intended change
```

`check-baselines` renders into a temporary directory, so it never touches the
committed files. When it reports a difference, look at the frame before
regenerating: every changed pixel is either a fix or a regression, and only you
can tell which.

## Settings

Regenerating with different settings will rewrite every file, so these are
pinned in the `Makefile`:

| | |
|---|---|
| Size | 800x600 (`BASELINE_W` / `BASELINE_H`) |
| Ticks | 60 (`SHOT_TICKS`) |
| Seed | 1 (`SHOT_SEED`) |
| Clicks | 40 at click-seed 1234567890, for `kdtree_mouse` and `quadtree_mouse` only |

800x600 rather than the 1920x1080 that `make screenshots` uses: these live in
git forever, and the smaller size still catches anything worth catching. The
full set is about 1.6 MB.

## Three frames are not byte-comparable

`make check-baselines` skips these, listed as `NONDETERMINISTIC_DIRS` in the
`Makefile`:

| Package | Why | Run-to-run drift |
|---|---|---|
| `examples/reaction_diffusion` | GPU ping-pong shader passes | ~190 px |
| `examples/shader_photo` | GPU shader pass | ~220 px |
| `visual_tests/nearest_neighbor` | animates off wall-clock time | ~3000 px |

Their baselines are committed anyway, as a visual reference — just don't read
meaning into a diff. The figures above are two runs of the *same* build, so
anything at that magnitude is noise. If one of them changes by far more than
its drift figure, that is worth a look.

## Adding a package

`examples/*` and `visual_tests/*` are picked up by wildcard, so a new package
needs no Makefile change — run `make baselines` and commit the new frame. If it
turns out to be nondeterministic, add it to `NONDETERMINISTIC_DIRS`; confirm
first by rendering it twice and comparing, rather than assuming.
