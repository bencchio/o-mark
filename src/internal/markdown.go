package internal

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
)

//go:embed themes/*.css
var themesFS embed.FS

func ThemeCSS(name string) string {
	data, err := themesFS.ReadFile("themes/" + name + ".css")
	if err != nil {
		return ""
	}
	return string(data)
}

func paletteCSS(p ThemePalette, font string) string {
	note := p.Accent
	tip := blendHex(p.Accent, "#22c55e", 0.5)
	warning := blendHex(p.Accent, "#f59e0b", 0.5)
	important := blendHex(p.Accent, "#a855f7", 0.5)
	caution := blendHex(p.Accent, "#ef4444", 0.5)
	colorScheme := "light"
	if luminance(p.Background) < 128 {
		colorScheme = "dark"
	}
	return `html { background-color: var(--o-mark-bg); }
:root {
	color-scheme: ` + colorScheme + `;
	--o-mark-bg: ` + p.Background + `;
	--o-mark-fg: ` + p.Foreground + `;
	--o-mark-accent: ` + p.Accent + `;
	--o-mark-surface: ` + p.Surface + `;
	--o-mark-border: ` + p.Border + `;
	--o-mark-code-bg: ` + p.CodeBg + `;
	--o-mark-code-fg: ` + p.CodeFg + `;
	--o-mark-link: ` + p.LinkColor + `;
	--o-mark-heading: ` + p.HeadingColor + `;
	--o-mark-font: ` + font + `;
	--o-mark-selection-bg: ` + p.SelectionBg + `;
	--o-mark-selection-fg: ` + p.SelectionFg + `;
	--o-mark-admonition-note: ` + note + `;
	--o-mark-admonition-tip: ` + tip + `;
	--o-mark-admonition-warning: ` + warning + `;
	--o-mark-admonition-important: ` + important + `;
	--o-mark-admonition-caution: ` + caution + `;
	--o-mark-page-gap-bg: ` + blendHex(p.Background, p.Border, 0.25) + `;` + hlPaletteVars(p) + `
}
body {
	font-family: var(--o-mark-font);
	line-height: 1.6;
	color: var(--o-mark-fg);
	background-color: var(--o-mark-bg);
	font-size: 16px;
	margin: 8mm auto;
	padding: 0 10mm;
	box-sizing: border-box;
}
b, strong { font-weight: bold; }
p { margin: 0 0 1em 0; }
del { text-decoration: line-through; }
img { max-width: 100%; height: auto; }
h1 {
	font-size: 2.4em;
	font-weight: bold;
	color: var(--o-mark-heading);
	border-bottom: 1px solid var(--o-mark-border);
	padding-bottom: 8px;
}
h2 {
	font-size: 1.8em;
	font-weight: bold;
	color: var(--o-mark-heading);
	border-bottom: 1px solid var(--o-mark-border);
	padding-bottom: 6px;
}
h3 { font-size: 1.4em; font-weight: bold; color: var(--o-mark-heading); }
h4 { font-size: 1.15em; font-weight: bold; color: var(--o-mark-heading); }
h5, h6 { font-size: 1em; font-weight: bold; color: var(--o-mark-heading); }
code {
	font-family: "SFMono-Regular", Consolas, "Liberation Mono", Menlo, monospace;
	background-color: var(--o-mark-code-bg);
	border: 1px solid var(--o-mark-border);
	border-radius: 4px;
	padding: 2px 5px;
	font-size: 0.88em;
	color: var(--o-mark-code-fg);
}
pre {
	background-color: var(--o-mark-code-bg);
	border: 1px solid var(--o-mark-border);
	border-radius: 6px;
	padding: 4mm;
}
pre code { background: none; border: none; border-radius: 0; padding: 0; font-size: 0.9em; }
a { color: var(--o-mark-link); text-decoration: underline; }
.img-external { color: inherit; opacity: 0.5; font-style: italic; font-size: 0.875em; }
blockquote {
	border-left: 4px solid var(--o-mark-border);
	margin-left: 0;
	padding-left: 4mm;
	color: var(--o-mark-fg);
}
hr { border: none; border-top: 1px solid var(--o-mark-border); margin: 6mm 0; }
table { border-collapse: collapse; width: 100%; }
th, td {
	border: 1px solid var(--o-mark-border);
	padding: 8px 13px;
	text-align: left;
}
th { background-color: var(--o-mark-surface); font-weight: bold; cursor: pointer; position: relative; }
th:hover { background-color: var(--o-mark-border); }
th.ow-sorted::after { content: ''; position: absolute; right: 6px; top: 50%; transform: translateY(-50%); border: 5px solid transparent; border-bottom-color: var(--o-mark-accent); }
th.ow-sorted.ow-sorted-desc::after { border-bottom-color: transparent; border-top-color: var(--o-mark-accent); }
tbody tr:nth-child(even) td { background-color: var(--o-mark-surface); }
dt { font-weight: bold; color: var(--o-mark-heading); }
dd { margin-left: 6mm; margin-bottom: 4px; }
::-webkit-scrollbar { display: none; }
::selection { background-color: var(--o-mark-selection-bg); color: var(--o-mark-selection-fg); }
@media (max-width: 160mm) { body { padding: 0 6mm; } }
.footnotes {
	border-top: 1px solid var(--o-mark-border);
	margin-top: 8mm;
	padding-top: 4mm;
	font-size: 0.875em;
}
.footnotes ol { padding-left: 6mm; }
.footnote-ref { font-size: 0.75em; }
.footnote-backref { text-decoration: none; margin-left: 1mm; }
mark { background-color: var(--o-mark-surface); color: var(--o-mark-fg); padding: 1px 3px; }
.task-list-item { list-style: none; }
.task-list-item-checkbox { display: inline-block; width: 14px; height: 14px; margin-right: 4px; vertical-align: middle; border: 1px solid var(--o-mark-border); background-color: var(--o-mark-bg); border-radius: 2px; position: relative; }
.task-list-item-checkbox.checked { background-color: var(--o-mark-accent); border-color: var(--o-mark-accent); }
.task-list-item-checkbox.checked::after { content: ''; position: absolute; left: 50%; top: 50%; width: 4px; height: 8px; border: solid var(--o-mark-bg); border-width: 0 2px 2px 0; transform: translate(-50%, -66%) rotate(45deg); }
.katex { color: inherit; }
.katex-display { overflow-x: auto; }
.katex .mord.text { font-family: inherit; }
.katex .text > span { font-family: inherit; }
.admonition { border-left: 4px solid; margin: 4mm 0; padding: 3mm 4mm; }
.admonition-title { font-weight: bold; margin-bottom: 2mm; font-size: 0.9em; text-transform: uppercase; letter-spacing: 0.05em; }
.admonition-note { border-color: var(--o-mark-admonition-note); }
.admonition-note .admonition-title { color: var(--o-mark-admonition-note); }
.admonition-tip { border-color: var(--o-mark-admonition-tip); }
.admonition-tip .admonition-title { color: var(--o-mark-admonition-tip); }
.admonition-warning { border-color: var(--o-mark-admonition-warning); }
.admonition-warning .admonition-title { color: var(--o-mark-admonition-warning); }
.admonition-important { border-color: var(--o-mark-admonition-important); }
.admonition-important .admonition-title { color: var(--o-mark-admonition-important); }
.admonition-caution { border-color: var(--o-mark-admonition-caution); }
.admonition-caution .admonition-title { color: var(--o-mark-admonition-caution); }
pre.mermaid { background-color: var(--o-mark-code-bg); border: 1px solid var(--o-mark-border); border-radius: 6px; padding: 4mm; text-align: center; }
.o-mark-frontmatter { border-left: 3px solid var(--o-mark-border); margin: 0 0 6mm 0; padding: 2mm 0 2mm 4mm; opacity: 0.7; }
.o-mark-frontmatter summary { font-size: 0.8em; cursor: pointer; user-select: none; }
.o-mark-frontmatter-content { font-size: 0.8em; background: none; border: none; padding: 2mm 0 0 0; margin: 0; font-family: inherit; white-space: pre-wrap; color: var(--o-mark-fg); }
.o-mark-page { box-sizing: border-box; }
.o-mark-page > *:first-child { margin-top: 0; }
.o-mark-page-gap { height: 3mm; margin: 6mm 0; background-color: var(--o-mark-page-gap-bg); }
@media print {
	.o-mark-page { min-height: 0; break-after: page; }
	.o-mark-page:last-child { break-after: avoid; }
	.o-mark-page-gap { display: none; }
}
`
}

var md = goldmark.New(
	goldmark.WithExtensions(
		extension.Table,
		extension.DefinitionList,
		extension.TaskList,
		extension.Strikethrough,
		extension.Footnote,
		SubSupExtension{},
		HighlightExtension{},
		MathExtension{},
	),
)

func LoadFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func renderBody(input, dir string) string {
	fm, doc, hasFm := extractFrontmatter(input)
	html := renderPagedBody(doc, dir)
	if hasFm {
		html = renderFrontmatterBlock(fm) + html
	}
	return html
}

// renderPagedBody renders doc as one HTML block, or as one ".o-mark-page" div
// per page when splitPages finds top-level thematic breaks to split on.
func renderPagedBody(doc, dir string) string {
	if pages := splitPages(doc); pages != nil {
		var b strings.Builder
		for i, page := range pages {
			if i > 0 {
				b.WriteString(`<div class="o-mark-page-gap"></div>`)
			}
			b.WriteString(`<div class="o-mark-page">`)
			b.WriteString(htmlPostProcess(page, dir))
			b.WriteString(`</div>`)
		}
		return b.String()
	}
	var body bytes.Buffer
	if err := md.Convert([]byte(doc), &body); err != nil {
		log.Printf("markdown: cannot render document: %v", err)
		return ""
	}
	return htmlPostProcess(body.String(), dir)
}

func RenderMarkdownWithCSS(input, css, dir string) string {
	body := renderBody(input, dir)
	// hljsCSS goes last so its wrap rule wins over the pre/code rules every
	// theme defines; themes only supply the --o-mark-hl-* variables it reads.
	head := "<style>" + css + hljsCSS() + "</style>"
	head += InlineHead(input, body)
	return fmt.Sprintf("<!DOCTYPE html><html><head>%s</head><body>%s</body></html>", head, body)
}

func RenderMarkdownWithPalette(input string, p ThemePalette, dir string, font string, cfg Config) string {
	return RenderMarkdownWithCSS(input, paletteCSS(p, font)+ConfigOverrideCSS(cfg), dir)
}
