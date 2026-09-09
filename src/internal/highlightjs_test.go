package internal

import (
	"strings"
	"testing"
)

// The bundle is only injected for documents that actually have code to
// highlight; mermaid fences become diagram containers with no <code> child.
func TestHasHighlightableFence(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  bool
	}{
		{"prose only", "no fences here at all", false},
		{"mermaid only", "```mermaid\ngraph TD\nA-->B\n```", false},
		{"go fence", "```go\nx := 1\n```", true},
		{"untagged fence", "```\nplain\n```", true},
		{"tilde fence", "~~~python\nx = 1\n~~~", true},
		{"mermaid plus go", "```mermaid\ngraph TD\n```\n\n```go\nx := 1\n```", true},
		// A line with content after the backticks is not a closing fence per
		// CommonMark, so this stays inside the mermaid block.
		{"fence opener inside mermaid", "```mermaid\n```go still inside\n```", false},
		{"tilde does not close backtick", "```mermaid\n~~~\n```", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasHighlightableFence(c.input); got != c.want {
				t.Errorf("hasHighlightableFence = %v, want %v", got, c.want)
			}
		})
	}
}

func TestPostLoadScriptsForCodeDocument(t *testing.T) {
	code := readExample(t, "code/code.md")
	scripts := PostLoadScripts(code)
	if len(scripts) != 1 {
		t.Fatalf("want 1 post-load script for a code document, got %d", len(scripts))
	}
	if !strings.Contains(scripts[0], "lineNumbersBlock") {
		t.Error("script is missing the line-numbers plugin")
	}
}

// The shared highlight stylesheet must come after the theme so its wrap rule
// wins over the pre/code rules every theme defines.
func TestHighlightCSSFollowsTheme(t *testing.T) {
	html := RenderMarkdownWithCSS(readExample(t, "code/code.md"), ThemeCSS("github"), ".")

	// Without this rule the gutter renders empty: the plugin emits the number
	// in a data attribute, not as text.
	for _, want := range []string{
		".hljs-ln-n::before { content: attr(data-line-number); }",
		"white-space: pre-wrap",
		"var(--o-mark-hl-keyword, #9a5fd0)",
	} {
		if !strings.Contains(html, want) {
			t.Errorf("shared highlight CSS missing %q", want)
		}
	}

	themeVars := strings.Index(html, "--o-mark-hl-keyword: #cf222e")
	sharedRules := strings.Index(html, "var(--o-mark-hl-keyword")
	if themeVars < 0 || sharedRules < 0 || themeVars > sharedRules {
		t.Error("shared highlight CSS must be appended after the theme CSS")
	}
}
