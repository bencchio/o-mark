package internal

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PdfExport struct {
	Path        string
	Exists      bool
	HTML        string
	Orientation string // "portrait" | "landscape", the session's sheet
}

// pdfMarginMM is the top and bottom margin of every printed page. The QML
// printToPdf takes no margins and Qt prints with none, and Chromium ignores an
// @page margin there, so printMargins builds them into the document instead.
const pdfMarginMM = 25

func SiblingPdfPath(mdPath string) string {
	ext := filepath.Ext(mdPath)
	if ext == "" {
		return mdPath + ".pdf"
	}
	return strings.TrimSuffix(mdPath, ext) + ".pdf"
}

func printFallbackPalette() ThemePalette {
	return ThemePalette{
		Background:   "#ffffff",
		Foreground:   "#000000",
		Accent:       "#444444",
		Surface:      "#f0f0f0",
		Border:       "#cccccc",
		CodeBg:       "#f0f0f0",
		CodeFg:       "#000000",
		LinkColor:    "#444444",
		HeadingColor: "#000000",
		SelectionBg:  "#cccccc",
		SelectionFg:  "#000000",
	}
}

func PreparePdfExport(mdPath, raw, docDir, font string, cfg Config, w PaletteSource) PdfExport {
	path := SiblingPdfPath(mdPath)
	_, err := os.Stat(path)
	out := PdfExport{Path: path, Exists: err == nil, Orientation: sheetOrientation(cfg.PageOrientation)}
	if out.Exists {
		return out
	}
	p := printFallbackPalette()
	if w != nil {
		p = w.PrintPalette()
	}
	out.HTML = printMargins(RenderMarkdownWithPalette(raw, p, docDir, font, cfg))
	return out
}

// sheetOrientation normalizes an orientation to one of the two valid values.
func sheetOrientation(o string) string {
	if o == OrientationLandscape {
		return OrientationLandscape
	}
	return OrientationPortrait
}

// printMargins wraps the document body in a table whose header and footer are
// blank rows of pdfMarginMM: a table repeats both on every printed page, which
// gives each page the same top and bottom margin wherever the content breaks.
// The body's own vertical margin is zeroed for print so it does not add to it.
// A forced break inside a table cell drops the repeated footer, so the break
// after each sheet becomes a break before the next one.
func printMargins(html string) string {
	const open, close = "</head><body>", "</body></html>"
	start := strings.Index(html, open)
	end := strings.LastIndex(html, close)
	if start < 0 || end < start {
		return html
	}
	spacer := fmt.Sprintf(`<div style="height:%dmm"></div>`, pdfMarginMM)
	style := `<style>@media print { body { margin: 0 auto; } .o-mark-page { break-after: auto !important; } .o-mark-page ~ .o-mark-page { break-before: page; } table.o-mark-print { width: 100%; border-collapse: collapse; } table.o-mark-print td { padding: 0; } }</style>`
	return html[:start] + style + open +
		`<table class="o-mark-print"><thead><tr><td>` + spacer + `</td></tr></thead><tbody><tr><td>` +
		html[start+len(open):end] +
		`</td></tr></tbody><tfoot><tr><td>` + spacer + `</td></tr></tfoot></table>` + html[end:]
}
