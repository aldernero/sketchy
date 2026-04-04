package sketchy

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"path/filepath"
	"strings"
	"time"

	"github.com/aldernero/debugui"
	"github.com/aldernero/gaul"
)

// formatSnapshotCreatedLocal parses snapshot created_at (UTC RFC3339 from sketch.db) for display in local time.
func formatSnapshotCreatedLocal(createdAt string) string {
	s := strings.TrimSpace(createdAt)
	if s == "" {
		return createdAt
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339} {
		if t, err := time.Parse(layout, s); err == nil {
			return t.Local().Format("2006-01-02 15:04:05 MST")
		}
	}
	return createdAt
}

func (s *Sketch) entriesForFolder(folder string) []controlEntry {
	var out []controlEntry
	for _, e := range s.uiPlan {
		if e.Folder == folder {
			out = append(out, e)
		}
	}
	return out
}

func (s *Sketch) controlWindow(ctx *debugui.Context) {
	ctx.Window("Control Panel", image.Rect(DefaultControlWindowX, DefaultControlWindowY, s.ControlWidth, s.ControlHeight), func(layout debugui.ContainerLayout) {
		s.builtinsPanel(ctx)
		root := s.entriesForFolder("")
		if len(root) > 0 {
			ctx.SetGridLayout([]int{-1}, nil)
			s.drawControlEntries(ctx, root)
		}
		for _, title := range folderOrder(s.uiPlan) {
			// Header IDs are derived from the call site; the loop would reuse one ID for every folder.
			ctx.IDScope("folder:"+title, func() {
				ctx.Header(title, true, func() {
					ctx.SetGridLayout([]int{-1}, nil)
					s.drawControlEntries(ctx, s.entriesForFolder(title))
				})
			})
		}
	})
	s.dialogSaveImage(ctx)
	s.dialogSnapshot(ctx)
	s.dialogLoadSnapshot(ctx)
}

func (s *Sketch) builtinsPanel(ctx *debugui.Context) {
	ctx.Header("Builtins", true, func() {
		ctx.SetGridLayout([]int{36, -1, 44}, nil)
		ctx.Text("Seed")
		ctx.IDScope("builtinSeed", func() {
			ctx.NumberField(&s.builtinSeedInt, 1).On(func() {
				s.setRandomSeed(int64(s.builtinSeedInt))
			})
		})
		ctx.Button("Rand").On(func() { s.randomizeRandomSeed() })

		s.drawColorRow(ctx, s.builtinColorBGIdx)
		s.drawColorRow(ctx, s.builtinColorFGIdx)
		s.drawBuiltinDefaultStrokeWidthRow(ctx)

		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Button("Save Image…").On(func() {
			s.dlgSaveImageOpen = true
			s.dlgSaveImagePrefix = s.Prefix + "_" + gaul.GetTimestampString()
			s.dlgSavePNG = true
			s.dlgSaveSVG = true
		})
		ctx.Button("Take Snapshot…").On(func() {
			s.dlgSnapshotOpen = true
			s.dlgSnapshotName = s.Prefix + "_snap_" + gaul.GetTimestampString()
			s.dlgSnapshotDescription = ""
			s.dlgSnapshotPNG = false
			s.dlgSnapshotSVG = false
		})
		ctx.Button("Load Snapshot…").On(func() {
			s.dlgLoadOpen = true
			s.dlgLoadNames = s.dbListSnapshots()
			if len(s.dlgLoadNames) > 0 {
				s.dlgLoadSelected = s.dlgLoadNames[0]
				s.refreshLoadPreview()
			} else {
				s.dlgLoadSelected = ""
				s.dlgLoadPreviewRow = nil
				s.dlgLoadMissing = nil
			}
		})
	})
}

func (s *Sketch) drawControlEntries(ctx *debugui.Context, entries []controlEntry) {
	for _, e := range entries {
		switch e.Kind {
		case entryFloatSlider:
			s.drawFloatSliderRow(ctx, e.Index)
		case entryIntSlider:
			s.drawIntSliderRow(ctx, e.Index)
		case entryToggle:
			s.drawToggleRow(ctx, e.Index)
		case entryColor:
			s.drawColorRow(ctx, e.Index)
		case entryDropdown:
			s.drawDropdownRow(ctx, e.Index)
		}
	}
}

func (s *Sketch) drawBuiltinDefaultStrokeWidthRow(ctx *debugui.Context) {
	const minW, maxW, stepW = 0.05, 3.0, 0.05
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1}, nil)
	ctx.Text("Default stroke width")
	ctx.IDScope("builtinStrokeW", func() {
		ctx.NumberFieldF(&s.DefaultStrokeWidth, stepW, 2).On(func() {
			s.DefaultStrokeWidth = clampFloat(s.DefaultStrokeWidth, minW, maxW)
			s.dirty = true
		})
	})
}

func (s *Sketch) drawFloatSliderRow(ctx *debugui.Context, idx int) {
	sl := &s.FloatSliders[idx]
	sl.maybeSyncTextBufFromVal()
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1, 96}, nil)
	ctx.Text(sl.Name)
	ctx.GridCell(func(bounds image.Rectangle) {
		ctx.IDScope(fmt.Sprintf("fsld%d", idx), func() {
			disp := clampFloat(sl.Val, sl.MinVal, sl.MaxVal)
			ctx.SliderFNoValue(&disp, sl.MinVal, sl.MaxVal, sl.Incr, sl.digits).On(func() {
				sl.Val = disp
				sl.syncTextBufFromVal()
			})
		})
		if ctx.ConsumeSecondaryClick(bounds) {
			s.openFloatSliderRangeModal(idx)
		}
	})
	ctx.GridCell(func(bounds image.Rectangle) {
		ctx.IDScope(fmt.Sprintf("fsln%d", idx), func() {
			ctx.TextField(&sl.textBuf).On(func() {
				commitFloatSliderText(sl)
			})
		})
		if ctx.ConsumeSecondaryClick(bounds) {
			s.openFloatSliderRangeModal(idx)
		}
	})
}

func (s *Sketch) drawIntSliderRow(ctx *debugui.Context, idx int) {
	sl := &s.IntSliders[idx]
	sl.maybeSyncTextBufFromVal()
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1, 96}, nil)
	ctx.Text(sl.Name)
	ctx.GridCell(func(bounds image.Rectangle) {
		ctx.IDScope(fmt.Sprintf("isld%d", idx), func() {
			ctx.SliderNoValue(&sl.Val, sl.MinVal, sl.MaxVal, sl.Incr).On(func() {
				sl.syncTextBufFromVal()
			})
		})
		if ctx.ConsumeSecondaryClick(bounds) {
			s.openIntSliderRangeModal(idx)
		}
	})
	ctx.GridCell(func(bounds image.Rectangle) {
		ctx.IDScope(fmt.Sprintf("isln%d", idx), func() {
			ctx.TextField(&sl.textBuf).On(func() {
				commitIntSliderText(sl)
			})
		})
		if ctx.ConsumeSecondaryClick(bounds) {
			s.openIntSliderRangeModal(idx)
		}
	})
}

func (s *Sketch) drawToggleRow(ctx *debugui.Context, idx int) {
	t := &s.Toggles[idx]
	// debugui identifies Checkbox/Button by caller PC; all toggles share the same call site here without a scope.
	ctx.IDScope(fmt.Sprintf("toggle:%d", idx), func() {
		if t.IsButton {
			ctx.Button(t.Name).On(func() {
				t.Checked = !t.Checked
			})
			return
		}
		ctx.Checkbox(&t.Checked, t.Name)
	})
}

func (s *Sketch) drawColorRow(ctx *debugui.Context, idx int) {
	cp := &s.ColorPickers[idx]
	hexLil := "#" + strings.ToLower(strings.TrimPrefix(cp.GetHex(), "#"))

	// One row: label | wide swatch (lil-gui style) | compact hex (read-only).
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1, 72}, nil)
	ctx.Text(cp.Name)
	ctx.IDScope(fmt.Sprintf("csw%d", idx), func() {
		ctx.Clickable(func(bounds image.Rectangle) {
			ctx.DrawSolidRect(bounds, color.RGBA{uint8(cp.r), uint8(cp.g), uint8(cp.b), 255})
		}).On(func() {
			s.openColorModal(idx)
		})
	})
	ctx.Text(hexLil)
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// modalActionRow places Cancel and a primary button at the bottom-right of a dialog (debugui grid spacer + fixed columns).
func modalActionRow(ctx *debugui.Context, primaryLabel string, onCancel, onPrimary func()) {
	ctx.SetGridLayout([]int{-1, 72, 72}, nil)
	ctx.Text("")
	ctx.Button("Cancel").On(onCancel)
	ctx.Button(primaryLabel).On(onPrimary)
}

func (s *Sketch) drawDropdownRow(ctx *debugui.Context, idx int) {
	d := &s.Dropdowns[idx]
	ctx.SetGridLayout([]int{ControlLabelColumnWidth, -1}, nil)
	ctx.Text(d.Name)
	ctx.IDScope(fmt.Sprintf("dd%d", idx), func() {
		ctx.Dropdown(&d.Index, d.Options).On(func() {})
	})
}

func (s *Sketch) dialogSaveImage(ctx *debugui.Context) {
	if !s.dlgSaveImageOpen {
		return
	}
	ctx.Window("Save Image", image.Rect(200, 120, 520, 300), func(layout debugui.ContainerLayout) {
		ctx.BringRootContainerToFront()
		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("Filename prefix (no extension)")
		prefix := &s.dlgSaveImagePrefix
		ctx.TextField(prefix).On(func() {})
		ctx.Checkbox(&s.dlgSavePNG, "PNG")
		ctx.Checkbox(&s.dlgSaveSVG, "SVG")
		modalActionRow(ctx, "OK", func() { s.dlgSaveImageOpen = false }, func() {
			base := strings.TrimSpace(*prefix)
			if base == "" {
				base = s.Prefix + "_" + gaul.GetTimestampString()
			}
			if s.dlgSavePNG {
				rel := filepath.ToSlash(filepath.Join("saves", "png", base+".png"))
				s.EnqueueSave(rel, "png", s.RasterDPI, true)
			}
			if s.dlgSaveSVG {
				rel := filepath.ToSlash(filepath.Join("saves", "svg", base+".svg"))
				s.EnqueueSave(rel, "svg", 0, true)
			}
			s.dlgSaveImageOpen = false
		})
	})
}

func (s *Sketch) dialogSnapshot(ctx *debugui.Context) {
	if !s.dlgSnapshotOpen {
		return
	}
	ctx.Window("Take Snapshot", image.Rect(200, 80, 560, 460), func(layout debugui.ContainerLayout) {
		ctx.BringRootContainerToFront()
		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("Snapshot name")
		name := &s.dlgSnapshotName
		ctx.TextField(name).On(func() {})
		ctx.Text("Description")
		desc := &s.dlgSnapshotDescription
		ctx.SetGridLayout([]int{-1}, []int{88})
		ctx.GridCell(func(bounds image.Rectangle) {
			ctx.IDScope("snapDesc", func() {
				ctx.TextField(desc).On(func() {})
			})
		})
		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("Save images")
		ctx.Checkbox(&s.dlgSnapshotPNG, "PNG")
		ctx.Checkbox(&s.dlgSnapshotSVG, "SVG")
		modalActionRow(ctx, "OK", func() { s.dlgSnapshotOpen = false }, func() {
			n := strings.TrimSpace(*name)
			if n == "" {
				n = s.Prefix + "_snap_" + gaul.GetTimestampString()
			}
			data, err := s.serializeControlState()
			if err != nil {
				fmt.Println("snapshot serialize:", err)
				s.dlgSnapshotOpen = false
				return
			}
			bdata, err := s.serializeBuiltinState()
			if err != nil {
				fmt.Println("snapshot builtin serialize:", err)
				s.dlgSnapshotOpen = false
				return
			}
			var pngID, svgID *int64
			var pngVal, svgVal int64
			base := n
			if s.dlgSnapshotPNG {
				rel := filepath.ToSlash(filepath.Join("saves", "png", base+".png"))
				full := filepath.Join(s.workDir, filepath.FromSlash(rel))
				if err := writePNG(full, s); err != nil {
					fmt.Println("snapshot png:", err)
				} else if s.db != nil {
					id, ierr := s.db.InsertSave(rel, "png")
					if ierr != nil {
						fmt.Println("snapshot db png:", ierr)
					} else {
						pngVal = id
						pngID = &pngVal
					}
				}
			}
			if s.dlgSnapshotSVG {
				rel := filepath.ToSlash(filepath.Join("saves", "svg", base+".svg"))
				full := filepath.Join(s.workDir, filepath.FromSlash(rel))
				if err := writeSVG(full, s); err != nil {
					fmt.Println("snapshot svg:", err)
				} else if s.db != nil {
					id, ierr := s.db.InsertSave(rel, "svg")
					if ierr != nil {
						fmt.Println("snapshot db svg:", ierr)
					} else {
						svgVal = id
						svgID = &svgVal
					}
				}
			}
			if err := s.dbInsertSnapshot(n, strings.TrimSpace(*desc), string(data), string(bdata), pngID, svgID); err != nil {
				fmt.Println("snapshot db:", err)
			} else {
				log.Printf("sketchy: saved snapshot %q", n)
			}
			s.dlgSnapshotOpen = false
		})
	})
}

func (s *Sketch) dialogLoadSnapshot(ctx *debugui.Context) {
	if !s.dlgLoadOpen {
		return
	}
	ctx.Window("Load Snapshot", image.Rect(180, 80, 560, 420), func(layout debugui.ContainerLayout) {
		ctx.BringRootContainerToFront()
		ctx.SetGridLayout([]int{-1}, nil)
		if len(s.dlgLoadNames) == 0 {
			ctx.Text("No snapshots in sketch.db")
			ctx.Button("Close").On(func() { s.dlgLoadOpen = false })
			return
		}
		idx := 0
		for i, n := range s.dlgLoadNames {
			if n == s.dlgLoadSelected {
				idx = i
				break
			}
		}
		ctx.Dropdown(&idx, s.dlgLoadNames).On(func() {
			if idx >= 0 && idx < len(s.dlgLoadNames) {
				s.dlgLoadSelected = s.dlgLoadNames[idx]
				s.refreshLoadPreview()
			}
		})
		ctx.SetGridLayout([]int{-1}, nil)
		ctx.Text("")
		if s.dlgLoadPreviewRow != nil {
			ctx.Text("Taken: " + formatSnapshotCreatedLocal(s.dlgLoadPreviewRow.CreatedAt))
			if d := strings.TrimSpace(s.dlgLoadPreviewRow.Description); d != "" {
				ctx.Text("Description:")
				for _, line := range strings.Split(d, "\n") {
					ctx.Text(line)
				}
			}
			if s.dlgLoadPreviewRow.PNGPath != "" {
				ctx.Text("PNG: " + filepath.Base(s.dlgLoadPreviewRow.PNGPath))
			}
			if s.dlgLoadPreviewRow.SVGPath != "" {
				ctx.Text("SVG: " + filepath.Base(s.dlgLoadPreviewRow.SVGPath))
			}
		}
		if len(s.dlgLoadMissing) > 0 {
			ctx.Text("Warning: unknown keys in snapshot:")
			for _, k := range s.dlgLoadMissing {
				ctx.Text("  • " + k)
			}
		}
		if s.dlgLoadPreviewRow != nil || len(s.dlgLoadMissing) > 0 {
			ctx.SetGridLayout([]int{-1}, nil)
			ctx.Text("")
		}
		modalActionRow(ctx, "OK", func() { s.dlgLoadOpen = false }, func() {
			row := s.dbGetSnapshot(s.dlgLoadSelected)
			if row == nil {
				s.dlgLoadOpen = false
				return
			}
			if _, err := s.applyControlStateJSON([]byte(row.ControlJSON)); err != nil {
				fmt.Println("apply snapshot:", err)
			} else if err := s.applyBuiltinStateJSON([]byte(row.BuiltinJSON)); err != nil {
				fmt.Println("apply snapshot builtin:", err)
			}
			s.dlgLoadOpen = false
		})
	})
}

func (s *Sketch) refreshLoadPreview() {
	row := s.dbGetSnapshot(s.dlgLoadSelected)
	s.dlgLoadPreviewRow = row
	s.dlgLoadMissing = nil
	if row != nil {
		miss, err := s.snapshotKeysPresentInJSON([]byte(row.ControlJSON))
		if err == nil {
			s.dlgLoadMissing = miss
		}
	}
}
