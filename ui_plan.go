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
