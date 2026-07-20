# Recording video

Sketchy can record your animated sketches straight to video. Frames are
captured from the same recorded-frame pipeline that powers PNG/SVG saves and
piped as raw RGBA into [ffmpeg](https://ffmpeg.org/), so there is nothing to
configure beyond having `ffmpeg` on your `PATH` (`ffmpeg` is in every major
package manager: `dnf install ffmpeg`, `apt install ffmpeg`,
`brew install ffmpeg`, …).

Recordings land in `saves/video/` in the sketch working directory, named
`<prefix>_<timestamp>.<ext>`.

# The Recording rows

The **Builtins** panel has a Recording section:

- **Rec format** — the output container/codec (see the table below).
- **Rec FPS** — playback frame rate written into the file (1–240, default
  60). Recording does not resample: **one tick is always one frame**, so a
  sketch animated for 60 TPS recorded at 30 FPS plays at half speed.
- **Rec scale** — raster scale for the recording, like **Export scale** (1×
  = one raster pixel per sketch pixel). Recording renders independently of
  the display, so **Preview mode never leaks into the video** — you can
  iterate at preview resolution while recording full-quality frames.
- **Rec mode** — when the recording starts and stops:
  - **Manual** — record until you press the button (or **Ctrl+R**) again.
  - **Frames** — record exactly N frames, then stop.
  - **Loop** — *arm* the recording: capture begins at the next tick where
    `Tick % N == 0` and stops after exactly N frames. If your animation
    repeats every N ticks (e.g. `phase := float64(s.Tick%600) / 600`), the
    result is a **perfect loop**: the last frame is the one right before a
    repeat of the first.
- The button starts/arms and stops/disarms; **Ctrl+R** does the same from
  the keyboard so you can record with the panel hidden. The row underneath
  shows a live frame counter while recording and the saved path (or error)
  when done.

# Formats

| Format | File | Codec | Notes |
|--------|------|-------|-------|
| WebM (VP9) | `.webm` | libvpx-vp9, CRF 24 | Royalty-free, excellent quality-per-bit. Good default. |
| MP4 (H.264) | `.mp4` | libx264, CRF 17 slow | Plays everywhere; best for sharing/social. |
| WebP (anim) | `.webp` | libwebp, q90, infinite loop | GitHub renders animated WebP inline in Markdown — ideal for READMEs. |
| FFV1 (MKV) | `.mkv` | FFV1 level 3 | **Lossless** archival master; transcode later without generation loss. Large files. |

On the licensing front: VP9, FFV1, and WebP are royalty-free; H.264 is
covered by patent pools, which matters for shipping *encoders*, not for
encoding your own art with your own ffmpeg. Sketchy itself never links any
codec — it only talks to the `ffmpeg` binary you installed.

# Recording is frame-perfect, not real-time

Sketch animation advances by tick (`s.Tick`), not by wall-clock time. While
recording, sketchy renders each tick, hands the frame to the encoder, and
**waits if the encoder falls behind** rather than dropping frames. So a
heavy sketch at 4× scale may preview sluggishly while recording — but the
file plays back at exactly **Rec FPS** with every frame present. Slow
preview never means a slow or stuttering video.

Two consequences worth knowing:

- Sketches that don't animate (no `MarkDirty` per tick) record duplicate
  frames — ticks and frames stay 1:1, and encoders compress duplicates to
  almost nothing.
- Stopping is asynchronous: the status row shows *Finalizing…* until ffmpeg
  has flushed the file. Wait for the saved-path message before closing the
  sketch.

# Shader sketches

[Shader sketches](shaders.md) record exactly the same way — frames are read
back from the GPU each tick instead of copied from the CPU raster. Perfect
loops pair naturally with the `Time` builtin uniform: a shader periodic in
`Time` with period `P` seconds loops perfectly with Loop mode and
`N = 60 * P` frames.

# Recording from code

Everything the panel does is available on `Sketch`:

```go
// Manual:
err := s.StartRecording(sketchy.RecordingOptions{
    Format: sketchy.RecordWebM,
    FPS:    60,
    Scale:  2,
})
// ... later:
s.StopRecording()

// Fixed length — five seconds at 60fps:
err = s.StartRecording(sketchy.RecordingOptions{
    Format:    sketchy.RecordMP4,
    NumFrames: 300,
})

// Perfect loop of a 600-tick cycle (arms, then starts at Tick % 600 == 0):
err = s.ArmLoopRecording(600)
```

`IsRecording()` and `RecordingFrameCount()` report progress, e.g. for
drawing a recording indicator inside the sketch itself. Scripted sketches
that exit right after recording should call
`s.FinishRecording(timeout)` — it blocks until ffmpeg has fully written the
file, where the interactive path finalizes asynchronously.

`RecordingOptions.ExtraArgs` is appended to the ffmpeg command line after
the defaults, so ffmpeg's last-option-wins makes overrides clean:

```go
// Lossless VP9 instead of CRF 24:
ExtraArgs: []string{"-lossless", "1"},
// Faster (lower quality) VP9 encoding for long captures:
ExtraArgs: []string{"-deadline", "realtime", "-cpu-used", "8"},
// Different WebP quality:
ExtraArgs: []string{"-quality", "75"},
```

`RecordingOptions.OutPath` overrides the default `saves/video/` naming.

# Caveats

- **`DisableClearBetweenFrames` (accumulation) sketches** record the
  *display* raster, because accumulated strokes exist only there. Turn
  **Preview mode off** and leave **Export scale** alone before recording —
  changing the display raster size mid-recording aborts with an error (the
  video's frame size is fixed when recording starts). For all other
  sketches, recording replays the frame recording at **Rec scale** and is
  immune to display settings.
- A translucent `DefaultBackground` may shift recorded colors slightly
  (frames are captured premultiplied); opaque backgrounds are exact.
- MP4 requires even pixel dimensions; odd-sized sketches are padded by one
  black row/column rather than cropped.
