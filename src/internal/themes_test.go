package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The shared highlight rules read these variables; a base layout missing one
// falls back to a literal default that ignores its palette.
func TestBaseLayoutDefinesHighlightVariables(t *testing.T) {
	css := paletteCSS(builtinPalette(false), "monospace")
	for _, role := range []string{"keyword", "string", "comment", "number", "title", "type", "builtin", "meta"} {
		if !strings.Contains(css, "--o-mark-hl-"+role+":") {
			t.Errorf("the base layout is missing --o-mark-hl-%s", role)
		}
	}
}

// The layout styles KaTeX's own classes. The math library prefixed its
// structural class names in a recent release, so these are worth pinning.
func TestBaseLayoutStylesKatexClasses(t *testing.T) {
	css := paletteCSS(builtinPalette(false), "monospace")
	for _, sel := range []string{".katex ", ".katex-display", ".katex .mord.text", ".katex .text > span"} {
		if !strings.Contains(css, sel) {
			t.Errorf("the base layout is missing the selector %q", sel)
		}
	}
}

// The sheet belongs to the page format, not to the layout.
func TestBaseLayoutLeavesThePageSizeToTheConfig(t *testing.T) {
	css := paletteCSS(builtinPalette(false), "monospace")
	for _, side := range []string{"210mm", "297mm"} {
		if strings.Contains(css, side) {
			t.Errorf("the base layout fixes a page side (%s)", side)
		}
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

// The base layout is the one Mono's sheet had: 15px, a generous line height and
// the smaller headings, until the library carries fonts and spacing.
func TestBaseLayoutTakesMonosSizes(t *testing.T) {
	css := paletteCSS(builtinPalette(false), "monospace")
	for _, want := range []string{"font-size: 15px;", "line-height: 1.9;", "h1 {\n\tfont-size: 2em;", "h2 {\n\tfont-size: 1.6em;"} {
		if !strings.Contains(css, want) {
			t.Errorf("the base layout is missing %q", want)
		}
	}
}
