package sketchy

type controlEntryKind int

const (
	entryFloatSlider controlEntryKind = iota
	entryIntSlider
	entryToggle
	entryColor
	entryDropdown
)

type controlEntry struct {
	Kind   controlEntryKind
	Index  int
	Folder string
}

func folderOrder(plan []controlEntry) []string {
	var out []string
	seen := map[string]bool{}
	for _, e := range plan {
		if e.Folder == "" {
			continue
		}
		if !seen[e.Folder] {
			seen[e.Folder] = true
			out = append(out, e.Folder)
		}
	}
	return out
}

// uiFolderPlan is the control panel layout grouped by folder, precomputed at
// Init so controlWindow doesn't rebuild it every frame.
type uiFolderPlan struct {
	rootEntries   []controlEntry
	folderTitles  []string
	folderEntries map[string][]controlEntry
}

func buildFolderPlan(plan []controlEntry) uiFolderPlan {
	p := uiFolderPlan{
		folderTitles:  folderOrder(plan),
		folderEntries: make(map[string][]controlEntry),
	}
	for _, e := range plan {
		if e.Folder == "" {
			p.rootEntries = append(p.rootEntries, e)
			continue
		}
		p.folderEntries[e.Folder] = append(p.folderEntries[e.Folder], e)
	}
	return p
}
