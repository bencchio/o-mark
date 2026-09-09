package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var builtinThemes = []string{"github", "writer", "night", "sepia", "mono", "minimal"}

// The shared highlight rules read these variables; a theme missing one falls
// back to a literal default that ignores its palette.
func TestThemesDefineHighlightVariables(t *testing.T) {
	roles := []string{"keyword", "string", "comment", "number", "title", "type", "builtin", "meta"}
	for _, name := range builtinThemes {
		css := ThemeCSS(name)
		if css == "" {
			t.Fatalf("theme %s has no CSS", name)
		}
		for _, role := range roles {
			if !strings.Contains(css, "--o-mark-hl-"+role+":") {
				t.Errorf("theme %s is missing --o-mark-hl-%s", name, role)
			}
		}
	}
}

// Every theme styles KaTeX's own classes. The math library prefixed its
// structural class names in a recent release, so these are worth pinning.
func TestThemesStyleKatexClasses(t *testing.T) {
	for _, name := range builtinThemes {
		css := ThemeCSS(name)
		for _, sel := range []string{".katex ", ".katex-display", ".katex .mord.text", ".katex .text > span"} {
			if !strings.Contains(css, sel) {
				t.Errorf("theme %s is missing the selector %q", name, sel)
			}
		}
	}
}

// The header used to be nearly invisible against the page, with the zebra
// stripes darker than the header itself.
func TestGithubTableHeaderContrast(t *testing.T) {
	css := ThemeCSS("github")
	if !strings.Contains(css, "th { background-color: #eaeef2") {
		t.Error("github table header lost its darker background")
	}
	if !strings.Contains(css, "border-bottom: 2px solid #d0d7de") {
		t.Error("github table header lost its definition border")
	}
	if !strings.Contains(css, "tbody tr:nth-child(even) td { background-color: #f6f8fa; }") {
		t.Error("github zebra stripes must sit lighter than the header")
	}
}

func TestWriterHeadingSize(t *testing.T) {
	if !strings.Contains(ThemeCSS("writer"), "font-size: 3.2em") {
		t.Error("writer top heading size changed")
	}
}

// A built-in CSS the user edited by hand outlives the next export; a missing
// one comes back, so deleting a file is how the shipped theme is restored.
func TestExportBuiltinThemesKeepsUserEdits(t *testing.T) {
	dir := t.TempDir()
	ExportBuiltinThemes(dir)

	edited := filepath.Join(dir, "github.css")
	if err := os.WriteFile(edited, []byte("/* mine */"), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "night.css")); err != nil {
		t.Fatalf("remove: %v", err)
	}

	ExportBuiltinThemes(dir)

	got, err := os.ReadFile(edited)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(got) != "/* mine */" {
		t.Errorf("github.css was overwritten, want the user edit to survive")
	}
	if css, err := os.ReadFile(filepath.Join(dir, "night.css")); err != nil {
		t.Errorf("night.css was not restored: %v", err)
	} else if string(css) != ThemeCSS("night") {
		t.Errorf("night.css does not match the embedded theme")
	}
}

// While a built-in file is absent the theme still resolves, from the embedded CSS.
func TestDiskThemeCSSFallsBackToEmbedded(t *testing.T) {
	dir := t.TempDir()
	if got := DiskThemeCSS("night", dir); got != ThemeCSS("night") {
		t.Errorf("DiskThemeCSS did not fall back to the embedded theme")
	}
}

// An install that keeps the Omarchy state elsewhere can point o-mark at it,
// instead of the path being fixed at compile time.
func TestOmarchyStatePathHonoursOverride(t *testing.T) {
	dir := t.TempDir()
	t.Setenv(OmarchyStateEnv, dir)

	if got, want := OmarchyStatePath("theme.name"), filepath.Join(dir, "theme.name"); got != want {
		t.Fatalf("got %q, want %q", got, want)
	}

	themeDir := filepath.Join(dir, "theme")
	if err := os.MkdirAll(themeDir, 0755); err != nil {
		t.Fatal(err)
	}
	css := "* {\n  font-family: \"Iosevka Custom\";\n}\n"
	if err := os.WriteFile(filepath.Join(themeDir, "hyprland-preview-share-picker.css"), []byte(css), 0644); err != nil {
		t.Fatal(err)
	}
	if got := ReadOmarchyFont(); got != "Iosevka Custom" {
		t.Errorf("font read from the override root: got %q, want %q", got, "Iosevka Custom")
	}
}

// With no override the default location under the home directory still wins.
func TestOmarchyStatePathDefaultsToHome(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")

	want := filepath.Join(home, ".local/state/omarchy/current", "theme.name")
	if got := OmarchyStatePath("theme.name"); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
