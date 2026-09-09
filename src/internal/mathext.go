package internal

import (
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer"
	"github.com/yuin/goldmark/util"
)

var KindMathInline = ast.NewNodeKind("MathInline")
var KindMathBlock = ast.NewNodeKind("MathBlock")

type mathInlineNode struct{ ast.BaseInline }

func (n *mathInlineNode) Kind() ast.NodeKind         { return KindMathInline }
func (n *mathInlineNode) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }

type mathBlockNode struct{ ast.BaseInline }

func (n *mathBlockNode) Kind() ast.NodeKind         { return KindMathBlock }
func (n *mathBlockNode) Dump(src []byte, level int) { ast.DumpHelper(n, src, level, nil, nil) }

type mathRenderer struct{}

func (r *mathRenderer) RegisterFuncs(reg renderer.NodeRendererFuncRegisterer) {
	reg.Register(KindMathInline, renderTag(`\(`, `\)`))
	reg.Register(KindMathBlock, renderTag(`\[`, `\]`))
}

// MathExtension adds $inline$ and $$block$$ LaTeX math syntax to goldmark.
// Output uses KaTeX delimiter conventions \(...\) and \[...\], rendered
// client-side via KaTeX auto-render injected into the HTML head.
// Limitation: $$...$$ only matches single-line expressions.
type MathExtension struct{}

func (e MathExtension) Extend(m goldmark.Markdown) {
	// Multi-line $$ is a block (mathblock.go), so its LaTeX never reaches
	// inline parsing. goldmark buckets block parsers by trigger byte and tries
	// the triggered ones before the untriggered ones (paragraph among them),
	// so this beats paragraph regardless of the number below; the priority
	// only orders parsers that trigger on the same byte, and nothing else
	// triggers on '$'.
	m.Parser().AddOptions(parser.WithBlockParsers(
		util.Prioritized(&mathDisplayParser{}, 750),
	))
	m.Parser().AddOptions(parser.WithInlineParsers(
		// $$block$$ must run before $inline$ so the double-$ is consumed first.
		util.Prioritized(&delimParser{
			delim:    '$',
			double:   true,
			makeNode: func() ast.Node { return &mathBlockNode{} },
		}, 600),
		util.Prioritized(&delimParser{
			delim:    '$',
			double:   false,
			makeNode: func() ast.Node { return &mathInlineNode{} },
			// Flanking keeps stray dollars in prose (currency, $vars) literal.
			flank: true,
		}, 601),
	))
	m.Renderer().AddOptions(renderer.WithNodeRenderers(
		util.Prioritized(&mathRenderer{}, 500),
		util.Prioritized(&mathDisplayRenderer{}, 500),
	))
}
