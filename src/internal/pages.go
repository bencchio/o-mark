package internal

import (
	"bytes"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/text"
)

// splitPages walks the parsed markdown AST and splits its top-level content
// into pages at each top-level thematic break (`---`, `***` or `___`). A
// thematic break inside a fence or a ```mermaid block is parsed as literal
// text, never as a ThematicBreak node, so no fence-aware scanning is needed
// here — goldmark's own parser already keeps the split out of them.
//
// A page with no content (a leading, trailing, or doubled separator) is
// dropped. Returns nil when the document has no top-level thematic break at
// all, so callers fall back to rendering the whole document as a single
// block, byte-identical to before this split existed.
func splitPages(doc string) []string {
	source := []byte(doc)
	root := md.Parser().Parse(text.NewReader(source))

	var pages []string
	var sawBreak bool
	current := ast.NewDocument()

	flush := func() {
		if current.ChildCount() == 0 {
			return
		}
		var buf bytes.Buffer
		if err := md.Renderer().Render(&buf, source, current); err == nil {
			pages = append(pages, buf.String())
		}
	}

	for n := root.FirstChild(); n != nil; {
		next := n.NextSibling()
		root.RemoveChild(root, n)
		if n.Kind() == ast.KindThematicBreak {
			sawBreak = true
			flush()
			current = ast.NewDocument()
		} else {
			current.AppendChild(current, n)
		}
		n = next
	}
	flush()

	// No break found, or every page was empty (e.g. the document is nothing
	// but separators): fall back to the normal single-block render.
	if !sawBreak || len(pages) == 0 {
		return nil
	}
	return pages
}
