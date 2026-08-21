package internal

import (
	"encoding/json"
	"log"
)

// ViewerTheme is one selectable style for the document viewer.
// ID names the CSS file in the themes dir, except "system" which is rendered
// dynamically from the current Omarchy palette.
type ViewerTheme struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Bg    string `json:"bg"`   // body background, used by QML chrome around the document
	HTML  string `json:"html"` // pre-rendered HTML loaded via loadHtml
}

// RenderViewerThemesJSON renders the document for every theme and returns the
// full list as JSON for QML. diskThemes are the non-system themes; themesDir
// is used to load user-edited CSS files from disk before falling back to embedded.
func RenderViewerThemesJSON(input string, p ThemePalette, docDir string, diskThemes []ViewerTheme, themesDir string, cfg Config, font string) string {
	out := make([]ViewerTheme, 0, 1+len(diskThemes))
	out = append(out, ViewerTheme{
		ID:    "system",
		Label: "System",
		Bg:    p.Background,
		HTML:  RenderMarkdownWithPalette(input, p, docDir, font, cfg),
	})
	for _, t := range diskThemes {
		t.HTML = RenderMarkdownWithCSS(input, DiskThemeCSS(t.ID, themesDir)+ConfigOverrideCSS(cfg), docDir)
		out = append(out, t)
	}
	b, err := json.Marshal(out)
	if err != nil {
		log.Printf("viewerthemes: cannot marshal viewer themes: %v", err)
	}
	return string(b)
}

// ViewerThemeLabelsJSON returns only the labels for the theme selector ComboBox model.
// Kept separate from RenderViewerThemesJSON so the model never rebuilds (and resets
// its index) when the System palette changes.
func ViewerThemeLabelsJSON(allThemes []ViewerTheme) string {
	labels := make([]string, len(allThemes))
	for i, t := range allThemes {
		labels[i] = t.Label
	}
	b, err := json.Marshal(labels)
	if err != nil {
		log.Printf("viewerthemes: cannot marshal theme labels: %v", err)
	}
	return string(b)
}
