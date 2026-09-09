package internal

import (
	"embed"
	"log"
	"strings"
	"sync"
)

//go:embed highlightjs/*.js
var highlightJSFS embed.FS

// hasHighlightableFence reports whether the document has at least one fenced
// code block that is not a mermaid diagram. Mermaid fences are rewritten into
// <pre class="mermaid"> containers with no <code> child (renderMermaidBlocks),
// so a document holding only diagrams gives highlight.js nothing to do and does
// not need the bundle injected.
func hasHighlightableFence(raw string) bool {
	marker := byte(0) // non-zero while inside a fence
	for _, line := range strings.Split(raw, "\n") {
		trimmed := strings.TrimLeft(line, " \t")
		if len(line)-len(trimmed) > 3 {
			continue // more than 3 spaces of indent is not a fence
		}
		var open byte
		switch {
		case strings.HasPrefix(trimmed, "```"):
			open = '`'
		case strings.HasPrefix(trimmed, "~~~"):
			open = '~'
		default:
			continue
		}
		// Everything after the run of fence characters: the info string on an
		// opener, and necessarily blank on a closer.
		rest := strings.TrimSpace(strings.TrimLeft(trimmed, string(open)))
		if marker != 0 {
			// Only the same character closes, and only with nothing after it —
			// otherwise the line is still content inside the open fence.
			if open == marker && rest == "" {
				marker = 0
			}
			continue
		}
		marker = open
		if lang := strings.Fields(rest); len(lang) == 0 || lang[0] != "mermaid" {
			return true
		}
	}
	return false
}

// highlightInitJS wires both bundles together once the page has loaded:
// highlight every code block except mermaid containers, then add the
// line-number gutter to the ones that actually hold content.
const highlightInitJS = `
try {
  hljs.configure({ignoreUnescapedHTML: true});
  document.querySelectorAll('pre:not(.mermaid) > code').forEach(function (el) {
    hljs.highlightElement(el);
    if (el.textContent.trim() !== '') {
      hljs.lineNumbersBlock(el);
    }
  });
} catch (e) {
  console.error('highlight.js init failed:', e);
}
`

// HighlightJSScript returns the highlight.js bundle plus the line-numbers
// plugin for injection via QML runJavaScript after the page loads, following
// the same pattern as MermaidScript. Returns empty string if either embedded
// file cannot be read.
var HighlightJSScript = sync.OnceValue(func() string {
	core, err := highlightJSFS.ReadFile("highlightjs/highlight.min.js")
	if err != nil {
		log.Printf("error reading embedded highlight.min.js: %v", err)
		return ""
	}
	lineNumbers, err := highlightJSFS.ReadFile("highlightjs/line-numbers.min.js")
	if err != nil {
		log.Printf("error reading embedded line-numbers.min.js: %v", err)
		return ""
	}
	return string(core) + ";" + string(lineNumbers) + ";" + highlightInitJS
})

// hlPaletteVars returns the --o-mark-hl-* declarations derived from a theme
// palette, for the :root block of paletteCSS. Hues are blended off the palette
// accent the same way admonition colors are, so the highlight follows an
// Omarchy theme change at runtime while staying distinguishable per token.
func hlPaletteVars(p ThemePalette) string {
	return `
	--o-mark-hl-keyword: ` + blendHex(p.Accent, "#c678dd", 0.5) + `;
	--o-mark-hl-string: ` + blendHex(p.Accent, "#98c379", 0.5) + `;
	--o-mark-hl-comment: ` + blendHex(p.Foreground, p.Background, 0.45) + `;
	--o-mark-hl-number: ` + blendHex(p.Accent, "#d19a66", 0.5) + `;
	--o-mark-hl-title: ` + blendHex(p.Accent, "#61afef", 0.5) + `;
	--o-mark-hl-type: ` + blendHex(p.Accent, "#e5c07b", 0.5) + `;
	--o-mark-hl-builtin: ` + blendHex(p.Accent, "#56b6c2", 0.5) + `;
	--o-mark-hl-meta: ` + blendHex(p.Foreground, p.Background, 0.6) + `;`
}

// hljsCSS returns the syntax-highlighting stylesheet shared by every theme.
// It is appended after the theme CSS in RenderMarkdownWithCSS, which is the
// single funnel both render paths go through, so the selectors live in one
// place instead of being repeated per theme. Themes contribute only the
// --o-mark-hl-* variables; the literal fallbacks keep code readable in user
// disk themes, which define none of them.
func hljsCSS() string {
	return `
pre code, .hljs-ln-code { white-space: pre-wrap; overflow-wrap: anywhere; }
.hljs-ln { border-collapse: collapse; width: 100%; }
.hljs-ln td { border: none; }
/* .hljs-ln td... (0,2,1) outranks the plugin's own injected
   ".hljs-ln td { padding: 0 }" (0,1,1), which otherwise zeroes the gutter
   padding at runtime. The td class alone (0,1,1) ties and loses to it. */
.hljs-ln td.hljs-ln-numbers {
	user-select: none;
	text-align: right;
	vertical-align: top;
	width: 1%;
	white-space: nowrap;
	padding: 0 1em 0 0;
	opacity: 0.5;
}
.hljs-ln-n::before { content: attr(data-line-number); }
.hljs-ln td.hljs-ln-code { padding: 0 0 0 1em; vertical-align: top; }
.hljs-keyword, .hljs-selector-tag, .hljs-literal, .hljs-doctag, .hljs-name {
	color: var(--o-mark-hl-keyword, #9a5fd0);
}
.hljs-string, .hljs-regexp, .hljs-addition, .hljs-attribute, .hljs-symbol, .hljs-bullet, .hljs-quote {
	color: var(--o-mark-hl-string, #3f9950);
}
.hljs-comment { color: var(--o-mark-hl-comment, #7d8590); font-style: italic; }
.hljs-number, .hljs-deletion { color: var(--o-mark-hl-number, #c07a3e); }
.hljs-title, .hljs-section, .hljs-selector-id, .hljs-selector-class {
	color: var(--o-mark-hl-title, #3d7fd6);
}
.hljs-type, .hljs-class .hljs-title, .hljs-params { color: var(--o-mark-hl-type, #b08a2e); }
.hljs-built_in, .hljs-variable, .hljs-template-variable, .hljs-attr, .hljs-property {
	color: var(--o-mark-hl-builtin, #2a9d9d);
}
.hljs-meta, .hljs-subst, .hljs-punctuation, .hljs-operator, .hljs-tag {
	color: var(--o-mark-hl-meta, #7d8590);
}
.hljs-emphasis { font-style: italic; }
.hljs-strong { font-weight: 600; }
`
}

func init() {
	RegisterAsset(Asset{
		Kind: AssetPostLoad,
		Detect: func(raw, _ string) bool {
			return hasHighlightableFence(raw)
		},
		Script: HighlightJSScript,
	})
}
