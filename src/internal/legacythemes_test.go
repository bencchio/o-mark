package internal

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
}

func TestRetireThemeSheetsMovesThemAndRemovesTheDirectory(t *testing.T) {
	root := t.TempDir()
	dir, old := filepath.Join(root, "themes"), filepath.Join(root, "themes-old")
	writeFile(t, filepath.Join(dir, "night.css"), "night")
	writeFile(t, filepath.Join(dir, "mine.css"), "mine")
	if got := RetireThemeSheets(dir, old); got != 2 {
		t.Fatalf("moved %d, want 2", got)
	}
	for name, body := range map[string]string{"night.css": "night", "mine.css": "mine"} {
		data, err := os.ReadFile(filepath.Join(old, name))
		if err != nil || string(data) != body {
			t.Errorf("%s not kept in themes-old: %v %q", name, err, data)
		}
	}
	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		t.Error("the emptied themes directory should be removed")
	}
	if got := RetireThemeSheets(dir, old); got != 0 {
		t.Errorf("a second run must move nothing, moved %d", got)
	}
}

func TestRetireThemeSheetsNeverOverwrites(t *testing.T) {
	root := t.TempDir()
	dir, old := filepath.Join(root, "themes"), filepath.Join(root, "themes-old")
	writeFile(t, filepath.Join(dir, "night.css"), "new")
	writeFile(t, filepath.Join(old, "night.css"), "kept")
	if got := RetireThemeSheets(dir, old); got != 0 {
		t.Errorf("moved %d, want 0", got)
	}
	if data, _ := os.ReadFile(filepath.Join(old, "night.css")); string(data) != "kept" {
		t.Errorf("themes-old was overwritten: %q", data)
	}
	if _, err := os.Stat(filepath.Join(dir, "night.css")); err != nil {
		t.Error("the file that could not move must stay where it was")
	}
}

func TestRetireThemeSheetsLeavesOtherFilesAlone(t *testing.T) {
	root := t.TempDir()
	dir, old := filepath.Join(root, "themes"), filepath.Join(root, "themes-old")
	writeFile(t, filepath.Join(dir, "night.css"), "night")
	writeFile(t, filepath.Join(dir, "notes.txt"), "notes")
	RetireThemeSheets(dir, old)
	if _, err := os.Stat(filepath.Join(dir, "notes.txt")); err != nil {
		t.Error("a file that is not a sheet must stay")
	}
	if _, err := os.Stat(dir); err != nil {
		t.Error("a directory that still holds files must stay")
	}
}

func TestRetireThemeSheetsWithoutADirectory(t *testing.T) {
	root := t.TempDir()
	if got := RetireThemeSheets(filepath.Join(root, "missing"), filepath.Join(root, "old")); got != 0 {
		t.Errorf("moved %d", got)
	}
	if got := RetireThemeSheets("relative", "alsorelative"); got != 0 {
		t.Errorf("relative paths must be refused, moved %d", got)
	}
}
