package internal

import (
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindMathDisplay = ast.NewNodeKind("MathDisplay")

// mathDisplayNode holds a multi-line $$...$$ block. It is a block node, so its
// lines are kept as raw source and never reach the inline parsers.
type mathDisplayNode struct{ ast.BaseBlock }

func (n *mathDisplayNode) Kind() ast.NodeKind         { return KindMathDisplay }
func (n *mathDisplayNode) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }

// mathDisplayParser parses a multi-line display math block: a line holding only
// $$, every following line taken verbatim, closed by another line holding only
// $$. Parsing this as a block is what keeps the LaTeX intact — as paragraph
// text it would go through inline parsing, where CommonMark escapes collapse
// `\\` (the row separator of \begin{aligned}) into a single backslash and the
// ^ / ~ / == extensions match delimiters inside formulas, wrapping parts of the
// expression in <sup>, <sub> or <mark> tags.
//
// Single-line $$a+b$$ stays with the inline parser in mathext.go; only the
// multi-line form is a block.
type mathDisplayParser struct{}

// isMathFence reports whether the line is exactly $$ with optional indentation
// (up to 3 spaces, as CommonMark allows for block starts) and trailing spaces.
func isMathFence(line []byte) bool {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	if i > 3 || i+2 > len(line) || line[i] != '$' || line[i+1] != '$' {
		return false
	}
	for j := i + 2; j < len(line); j++ {
		if line[j] != ' ' && line[j] != '\t' && line[j] != '\n' && line[j] != '\r' {
			return false
		}
	}
	return true
}

func (p *mathDisplayParser) Trigger() []byte { return []byte{'$'} }

func (p *mathDisplayParser) Open(parent ast.Node, reader text.Reader, pc parser.Context) (ast.Node, parser.State) {
	line, _ := reader.PeekLine()
	if !isMathFence(line) {
		return nil, parser.NoChildren
	}
	// Deliberately no Advance here: goldmark moves to the next line itself
	// once Open returns a node, the same way the fenced code parser does.
	return &mathDisplayNode{}, parser.NoChildren
}

func (p *mathDisplayParser) Continue(node ast.Node, reader text.Reader, pc parser.Context) parser.State {
	line, segment := reader.PeekLine()
	if isMathFence(line) {
		reader.AdvanceToEOL()
		return parser.Close
	}
	segment.ForceNewline = true // treat EOF as a newline
	node.Lines().Append(segment)
	reader.AdvanceToEOL()
	return parser.Continue | parser.NoChildren
}

func (p *mathDisplayParser) Close(node ast.Node, reader text.Reader, pc parser.Context) {}

func (p *mathDisplayParser) CanInterruptParagraph() bool { return true }

func (p *mathDisplayParser) CanAcceptIndentedLine() bool { return false }

type mathDisplayRenderer struct{}

func (r *mathDisplayRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathDisplay, renderMathDisplay)
}

// renderMathDisplay writes the block as \[...\] for KaTeX auto-render. The
// LaTeX is HTML-escaped so the browser hands KaTeX back the original source
// through textContent.
func renderMathDisplay(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
	if !entering {
		return ast.WalkContinue, nil
	}
	if _, err := w.WriteString(`<p class="math-display">\[` + "\n"); err != nil {
		return ast.WalkStop, err
	}
	lines := node.Lines()
	for i := 0; i < lines.Len(); i++ {
		segment := lines.At(i)
		if _, err := w.Write(util.EscapeHTML(segment.Value(source))); err != nil {
			return ast.WalkStop, err
		}
	}
	if _, err := w.WriteString(`\]</p>` + "\n"); err != nil {
		return ast.WalkStop, err
	}
	return ast.WalkSkipChildren, nil
}
