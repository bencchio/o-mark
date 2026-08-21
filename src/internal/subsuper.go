package internal

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

var KindSub = ast.NewNodeKind("Subscript")
var KindSup = ast.NewNodeKind("Superscript")

type subNode struct{ ast.BaseInline }

func (n *subNode) Kind() ast.NodeKind            { return KindSub }
func (n *subNode) Dump(src []byte, level int)    { ast.DumpHelper(n, src, level, nil, nil) }

type supNode struct{ ast.BaseInline }

func (n *supNode) Kind() ast.NodeKind            { return KindSup }
func (n *supNode) Dump(src []byte, level int)    { ast.DumpHelper(n, src, level, nil, nil) }

// delimParser parses inline syntax: Xcontentx → <tag>content</tag>.
// Does not span newlines. When double=false, rejects doubled delimiters at open
// or close (e.g. ~~) to avoid conflicts with extensions using the same byte doubled.
// When double=true, requires exactly two delimiters at open and close (e.g. ==text==).
//
// This same struct is reused by highlight.go (0.3.3) for ==text== and by
// mathext.go (0.3.6) for the inline $math$ parsing side.
type delimParser struct {
	delim    byte
	double   bool
	makeNode func() ast.Node
}

func (p *delimParser) Trigger() []byte { return []byte{p.delim} }

func (p *delimParser) Parse(parent ast.Node, reader text.Reader, pc parser.Context) ast.Node {
	line, segment := reader.PeekLine()
	if p.double {
		if len(line) < 5 || line[0] != p.delim || line[1] != p.delim {
			return nil
		}
		end := -1
		for i := 2; i < len(line)-1; i++ {
			if line[i] == '\n' || line[i] == '\r' {
				break
			}
			if line[i] == p.delim && line[i+1] == p.delim {
				end = i
				break
			}
		}
		if end < 2 {
			return nil
		}
		reader.Advance(end + 2)
		node := p.makeNode()
		node.AppendChild(node, ast.NewTextSegment(
			text.NewSegment(segment.Start+2, segment.Start+end),
		))
		return node
	}
	if len(line) < 3 || line[0] != p.delim {
		return nil
	}
	// Reject doubled opening delimiter (e.g. ~~ strikethrough).
	if line[1] == p.delim {
		return nil
	}
	end := -1
	for i := 1; i < len(line); i++ {
		if line[i] == '\n' || line[i] == '\r' {
			break
		}
		if line[i] == p.delim {
			end = i
			break
		}
	}
	if end < 1 {
		return nil
	}
	// Reject doubled closing delimiter (e.g. content ends in ~~).
	if end+1 < len(line) && line[end+1] == p.delim {
		return nil
	}
	reader.Advance(end + 1)
	node := p.makeNode()
	node.AppendChild(node, ast.NewTextSegment(
		text.NewSegment(segment.Start+1, segment.Start+end),
	))
	return node
}

type subSupRenderer struct{}

func (r *subSupRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindSub, renderTag("<sub>", "</sub>"))
	reg.Register(KindSup, renderTag("<sup>", "</sup>"))
}

func renderTag(open, close string) renderer.NodeRendererFunc {
	return func(w util.BufWriter, source []byte, node ast.Node, entering bool) (ast.WalkStatus, error) {
		var err error
		if entering {
			_, err = w.WriteString(open)
		} else {
			_, err = w.WriteString(close)
		}
		return ast.WalkContinue, err
	}
}

// SubSupExtension adds ~subscript~ and ^superscript^ inline syntax to goldmark.
type SubSupExtension struct{}

func (e SubSupExtension) Extend(m goldmark.Markdown) {
	// goldmark sorts parsers in ascending priority order (lower number = runs first).
	// Strikethrough is at 500, so ~ must be below 500 to win the same trigger byte.
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&delimParser{
			delim:    '~',
			makeNode: func() ast.Node { return &subNode{} },
		}, 499),
		util.Prioritized(&delimParser{
			delim:    '^',
			makeNode: func() ast.Node { return &supNode{} },
		}, 700),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&subSupRenderer{}, 500),
	))
}
