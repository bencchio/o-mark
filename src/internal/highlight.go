package internal

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindMark = ast.NewNodeKind("Mark")

type markNode struct{ ast.BaseInline }

func (n *markNode) Kind() ast.NodeKind         { return KindMark }
func (n *markNode) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }

type markRenderer struct{}

func (r *markRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMark, renderTag("<mark>", "</mark>"))
}

// HighlightExtension adds ==highlight== inline syntax to goldmark.
type HighlightExtension struct{}

func (e HighlightExtension) Extend(m goldmark.Markdown) {
	m.Parser().AddOptions(parser.WithInlineParsers(
		util.Prioritized(&delimParser{
			delim:    '=',
			double:   true,
			makeNode: func() ast.Node { return &markNode{} },
		}, 698),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&markRenderer{}, 500),
	))
}
