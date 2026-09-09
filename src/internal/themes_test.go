package internal

import (
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
