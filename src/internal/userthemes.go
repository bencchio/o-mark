package internal

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

var builtinOrder = []string{"github", "writer", "night", "sepia", "mono", "minimal"}

var builtinLabels = map[string]string{
	"github":  "GitHub",
	"writer":  "Writer",
	"night":   "Night",
	"sepia":   "Sepia",
	"mono":    "Mono",
	"minimal": "Minimal",
}

var builtinBgs = map[string]string{
	"github":  "#ffffff",
	"writer":  "#e8e0d0",
	"night":   "#15171a",
	"sepia":   "#faf0e0",
	"mono":    "#1a1a1a",
	"minimal": "#fafafa",
}

// ExportBuiltinThemes writes the embedded CSS for each built-in theme to dir,
// never overwriting a file that already exists: a theme the user edited by hand
// wins over the embedded version. Deleting a file restores it on the next run,
// and DiskThemeCSS falls back to the embedded CSS while it is missing.
func ExportBuiltinThemes(dir string) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("themes: cannot create %s: %v", dir, err)
		return
	}
	for _, id := range builtinOrder {
		css := ThemeCSS(id)
		if css == "" {
			continue
		}
		// O_EXCL leaves no window between checking for the file and writing it.
		f, err := os.OpenFile(filepath.Join(dir, id+".css"), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
		if err != nil {
			if !errors.Is(err, os.ErrExist) {
				log.Printf("themes: cannot export %s: %v", id, err)
			}
			continue
		}
		if _, err := f.Write([]byte(css)); err != nil {
			log.Printf("themes: cannot export %s: %v", id, err)
		}
		if err := f.Close(); err != nil {
			log.Printf("themes: cannot export %s: %v", id, err)
		}
	}
}

// legacyThemeIDs are built-in IDs from older versions that are no longer part of
// o-mark. They are removed from the themes dir on startup so they don't appear
// as spurious user themes after an upgrade.
var legacyThemeIDs = []string{"printed", "bw-light", "bw-dark", "omarchy", "light"}

// CleanupLegacyThemes removes CSS files from dir whose IDs were built-ins in
// older versions and are now obsolete. Safe to call even if the files don't exist.
func CleanupLegacyThemes(dir string) {
	for _, id := range legacyThemeIDs {
		if err := os.Remove(filepath.Join(dir, id+".css")); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("themes: cannot remove legacy theme %s: %v", id, err)
		}
	}
}

// DiskThemes scans dir for .css files and returns them as a ViewerTheme list.
// Built-in themes appear first in canonical order; user extras follow alphabetically.
func DiskThemes(dir string) []ViewerTheme {
	entries, err := os.ReadDir(dir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		log.Printf("themes: cannot read dir %s: %v", dir, err)
	}
	diskIDs := map[string]bool{}
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".css") {
			diskIDs[strings.TrimSuffix(e.Name(), ".css")] = true
		}
	}

	var themes []ViewerTheme
	for _, id := range builtinOrder {
		// Always include built-ins (fall back to embedded if disk export failed).
		if diskIDs[id] || ThemeCSS(id) != "" {
			themes = append(themes, ViewerTheme{
				ID:    id,
				Label: builtinLabels[id],
				Bg:    builtinBgs[id],
			})
		}
		delete(diskIDs, id)
	}

	var extras []string
	for id := range diskIDs {
		if slices.Contains(legacyThemeIDs, id) {
			continue
		}
		extras = append(extras, id)
	}
	sort.Strings(extras)
	for _, id := range extras {
		themes = append(themes, ViewerTheme{ID: id, Label: themeLabel(id), Bg: "#ffffff"})
	}
	return themes
}

// DiskThemeCSS loads CSS from dir/id.css, falling back to the embedded FS.
func DiskThemeCSS(id, dir string) string {
	if data, err := os.ReadFile(filepath.Join(dir, id+".css")); err == nil {
		return string(data)
	}
	return ThemeCSS(id)
}

// themeLabel derives a display label from a CSS filename stem.
// "my-custom-theme" → "My Custom Theme"
func themeLabel(id string) string {
	parts := strings.Split(id, "-")
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, " ")
}
