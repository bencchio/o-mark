package internal

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// ThemesOldDir is where the CSS sheets of the retired theme selector are kept.
func ThemesOldDir() string {
	return filepath.Join(ConfigDir(), "themes-old")
}

// RetireThemeSheets moves the .css files of the retired theme selector from dir
// into oldDir, so they stop looking like themes without being lost. A file that
// oldDir already holds is left where it is, dir is removed once it is empty,
// and running it again finds nothing to move. It returns how many files moved.
func RetireThemeSheets(dir, oldDir string) int {
	if !filepath.IsAbs(dir) || !filepath.IsAbs(oldDir) {
		return 0
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("themes: cannot read %s: %v", dir, err)
		}
		return 0
	}
	moved := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".css") {
			continue
		}
		dest := filepath.Join(oldDir, e.Name())
		if _, err := os.Stat(dest); err == nil {
			log.Printf("themes: %s already exists, leaving %s in place", dest, e.Name())
			continue
		}
		if err := os.MkdirAll(oldDir, 0755); err != nil {
			log.Printf("themes: cannot create %s: %v", oldDir, err)
			return moved
		}
		if err := os.Rename(filepath.Join(dir, e.Name()), dest); err != nil {
			log.Printf("themes: cannot move %s: %v", e.Name(), err)
			continue
		}
		moved++
	}
	os.Remove(dir) // fails, harmlessly, while the directory still holds anything
	return moved
}
