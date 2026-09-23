package internal

import (
	"strings"
	"testing"
)

func TestSplitPagesNoSeparatorReturnsNil(t *testing.T) {
	if pages := splitPages("# Title\n\nSome text.\n"); pages != nil {
		t.Errorf("expected nil for a document without a thematic break, got %d pages", len(pages))
	}
}

func TestSplitPagesSplitsOnSeparator(t *testing.T) {
	pages := splitPages("# One\n\n---\n\n# Two\n")
	if len(pages) != 2 {
		t.Fatalf("expected 2 pages, got %d: %v", len(pages), pages)
	}
}

func TestSplitPagesAnyThematicBreakSyntax(t *testing.T) {
	for _, sep := range []string{"---", "***", "___"} {
		doc := "# One\n\n" + sep + "\n\n# Two\n"
		pages := splitPages(doc)
		if len(pages) != 2 {
			t.Errorf("separator %q: expected 2 pages, got %d", sep, len(pages))
		}
	}
}

func TestSplitPagesDropsLeadingAndTrailingEmptyPage(t *testing.T) {
	pages := splitPages("---\n\n# Only\n\n---\n")
	if len(pages) != 1 {
		t.Fatalf("expected the empty leading/trailing pages dropped, leaving 1 page, got %d: %v", len(pages), pages)
	}
}

func TestSplitPagesDropsDoubledSeparator(t *testing.T) {
	pages := splitPages("# One\n\n---\n\n---\n\n# Two\n")
	if len(pages) != 2 {
		t.Fatalf("expected the empty page between two separators dropped, got %d: %v", len(pages), pages)
	}
}

func TestSplitPagesAllSeparatorFallsBackToNil(t *testing.T) {
	if pages := splitPages("---\n"); pages != nil {
		t.Errorf("a document with only a separator and no content has nothing to split, expected nil, got %d pages", len(pages))
	}
}

func TestSplitPagesIgnoresSeparatorInFencedCode(t *testing.T) {
	doc := "# One\n\n```\n---\n```\n"
	if pages := splitPages(doc); pages != nil {
		t.Errorf("a thematic break inside a fenced code block must not split, got %d pages", len(pages))
	}
}

func TestSplitPagesIgnoresSeparatorInMermaidBlock(t *testing.T) {
	doc := "# One\n\n```mermaid\ngraph TD\nA-->B\n---\n```\n"
	if pages := splitPages(doc); pages != nil {
		t.Errorf("a thematic break inside a mermaid block must not split, got %d pages", len(pages))
	}
}

func TestRenderBodyUnsplitDocumentUnchanged(t *testing.T) {
	doc := "# Title\n\nSome text.\n"
	got := renderBody(doc, "")
	if strings.Contains(got, "o-mark-page") {
		t.Errorf("a document without a separator must not be wrapped in page containers: %s", got)
	}
}

func TestRenderBodySplitDocumentWrapsPages(t *testing.T) {
	doc := "# One\n\n---\n\n# Two\n"
	got := renderBody(doc, "")
	if !strings.Contains(got, `class="o-mark-page"`) {
		t.Errorf("expected page containers in split output: %s", got)
	}
	if !strings.Contains(got, `class="o-mark-page-gap"`) {
		t.Errorf("expected a gap element between pages: %s", got)
	}
}

func TestRenderBodyFrontmatterNotTreatedAsPageSeparator(t *testing.T) {
	doc := "---\ntitle: Doc\n---\n\n# Body\n\nJust one page.\n"
	got := renderBody(doc, "")
	if strings.Contains(got, "o-mark-page") {
		t.Errorf("a document with only a leading frontmatter block must not be split into pages: %s", got)
	}
	if !strings.Contains(got, "o-mark-frontmatter") {
		t.Errorf("expected the frontmatter block to still render: %s", got)
	}
}

func TestRenderBodyFrontmatterPlusSeparatorSplitsBodyOnly(t *testing.T) {
	doc := "---\ntitle: Doc\n---\n\n# One\n\n---\n\n# Two\n"
	got := renderBody(doc, "")
	if !strings.Contains(got, "o-mark-frontmatter") {
		t.Errorf("expected the frontmatter block to still render: %s", got)
	}
	if !strings.Contains(got, `class="o-mark-page"`) {
		t.Errorf("expected the body's own separator to still split into pages: %s", got)
	}
}

