package internal

import (
	"regexp"
	"strings"
	"testing"
)

// Display math is parsed as a block precisely so its LaTeX never reaches the
// inline parsers. These cases are the two ways that used to corrupt it.
func TestDisplayMathKeepsContentVerbatim(t *testing.T) {
	got := renderDoc(t, "$$\n\\begin{aligned}\na &= b \\\\\nc &= d_1 \\\\\ne &= f^2\n\\end{aligned}\n$$\n")

	// CommonMark escaping used to collapse the row separator into one backslash.
	for _, want := range []string{`\begin{aligned}`, `\\`, `d_1`, `f^2`} {
		if !strings.Contains(got, want) {
			t.Errorf("block lost %q:\n%s", want, got)
		}
	}
	// The ^ / ~ / == extensions used to match inside formulas and wrap parts of
	// the expression in markup.
	for _, bad := range []string{"<sup>", "<sub>", "<mark>"} {
		if strings.Contains(got, bad) {
			t.Errorf("inline parsing mangled the block with %s:\n%s", bad, got)
		}
	}
}

func TestDisplayMathInsideWrappers(t *testing.T) {
	cases := map[string]string{
		"blockquote": "> quote:\n>\n> $$\n> x = \\frac{1}{2} \\\\ y\n> $$",
		"loose list": "- item:\n\n  $$\n  \\sum_{k=1}^{n} k \\\\ z\n  $$\n\n- second",
		"tight list": "- $$\n  a_1 + b^2\n  $$\n- second",
		"paragraph":  "before\n\n$$\nE = mc^2\n$$\n\nafter",
	}
	for name, in := range cases {
		t.Run(name, func(t *testing.T) {
			got := renderDoc(t, in)
			if !strings.Contains(got, `\[`) {
				t.Errorf("block not converted:\n%s", got)
			}
			if strings.Contains(got, "<sup>") || strings.Contains(got, "<sub>") {
				t.Errorf("block content mangled:\n%s", got)
			}
		})
	}
}

// Code fences win over display math: while a fenced code block is open it
// consumes every line, so no new block is opened inside it.
func TestDisplayMathIgnoresCodeFences(t *testing.T) {
	got := renderDoc(t, "```\n$$\nnot math\n$$\n```")
	if strings.Contains(got, `\[`) {
		t.Errorf("code fence was converted to math:\n%s", got)
	}
	if !strings.Contains(got, "$$") {
		t.Errorf("code fence lost its $$ lines:\n%s", got)
	}
}

func TestInlineAndSingleLineMath(t *testing.T) {
	if got := renderDoc(t, "text $x^2$ here"); !strings.Contains(got, `\(`) {
		t.Errorf("inline math not rendered:\n%s", got)
	}
	if got := renderDoc(t, "$$a+b$$"); !strings.Contains(got, `\[a+b\]`) {
		t.Errorf("single-line display math not rendered:\n%s", got)
	}
}

// codeRegionRe matches the regions where a literal $$ is legitimate prose.
var codeRegionRe = regexp.MustCompile(`(?s)<pre>.*?</pre>|<code>.*?</code>`)

func TestMathExampleHasNoUnconvertedBlocks(t *testing.T) {
	full := renderDoc(t, readExample(t, "extensions/math.md"))
	stripped := codeRegionRe.ReplaceAllString(full, "")
	if i := strings.Index(stripped, "$$"); i >= 0 {
		t.Errorf("unconverted $$ outside code: %q", stripped[max(0, i-60):min(len(stripped), i+60)])
	}
	if n := strings.Count(full, `\[`); n < 10 {
		t.Errorf("expected the example's display blocks to convert, found %d", n)
	}
}
