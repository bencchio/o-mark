package internal

import (
	"os"
	"path/filepath"
	"strings"
)

type PdfExport struct {
	Path   string
	Exists bool
	HTML   string
}

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
	out := PdfExport{Path: path, Exists: err == nil}
	if out.Exists {
		return out
	}
	p := printFallbackPalette()
	if w != nil {
		p = w.PrintPalette()
	}
	out.HTML = RenderMarkdownWithPalette(raw, p, docDir, font, cfg)
	return out
}
