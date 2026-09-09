package internal

import (
	"strings"
	"testing"
)

// The inline $ opts into flanking rules so ordinary prose about prices or
// shell variables is not read as math.
func TestInlineDollarFlanking(t *testing.T) {
	literal := []string{
		"The service costs $5 and $10 for the premium tier.",
		"Our prices are $1, $2 and $3 depending on volume.",
		"prose mentions of $PATH or $variable names",
		"spaced $ x $ should stay plain",
	}
	for _, in := range literal {
		if got := renderDoc(t, in); strings.Contains(got, `\(`) {
			t.Errorf("stray dollars parsed as math: %q ->\n%s", in, got)
		}
	}

	math := []string{
		`Euler: $e^{i\pi}+1=0$ inline.`,
		`like $x^2 + 1$ still renders`,
		`mixed literal then math: win $100 now but $x$ solves it`,
	}
	for _, in := range math {
		if got := renderDoc(t, in); !strings.Contains(got, `\(`) {
			t.Errorf("inline math lost: %q ->\n%s", in, got)
		}
	}

	// Aborting on a space-preceded closer keeps mixed prose from being
	// captured greedily as one long formula.
	if got := renderDoc(t, `win $100 now but $x$ solves it`); strings.Contains(got, `\(100`) {
		t.Errorf("greedy capture across prose:\n%s", got)
	}
}

// Flanking is opt-in: the other extensions sharing delimParser must behave as
// before.
func TestSharedDelimitersUnaffectedByFlanking(t *testing.T) {
	got := renderDoc(t, "H~2~O and E=mc^2^ and ==marked==")
	for _, want := range []string{"<sub>2</sub>", "<sup>2</sup>", "<mark>marked</mark>"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %s:\n%s", want, got)
		}
	}
}

func TestUnclosedDelimitersStayLiteral(t *testing.T) {
	for _, in := range []string{"a~b with no closer", "a^b with no closer"} {
		got := renderDoc(t, in)
		if strings.Contains(got, "<sub>") || strings.Contains(got, "<sup>") {
			t.Errorf("unclosed delimiter consumed the line: %q ->\n%s", in, got)
		}
	}
}
