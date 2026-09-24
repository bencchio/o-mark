package internal

import (
	"encoding/json"
	"log"
)

// PaletteSource is the live Omarchy palette the session renders and prints
// with. *OmarchyThemeWatcher satisfies it. A nil source means the built-in
// palette fallback, the same one GetThemePalette("omarchy") resolves to.
type PaletteSource interface {
	Palette() ThemePalette
	PrintPalette() ThemePalette
}

// StateSink is the Qt seam: the only place the session writes context
// properties. main implements it over QQmlContext; tests record the calls.
type StateSink interface {
	SetString(key, value string)
	SetInt(key string, value int)
	SetBool(key string, value bool)
	SetFloat(key string, value float64)
}

// State groups the QML context properties so Publish can apply them in the
// order the UI requires. A nil group means "leave these properties alone".
type State struct {
	Chrome   *ChromeState
	Layout   *LayoutState
	Identity *IdentityState
	Document *DocumentState
}

// ChromeState feeds pure QML bindings; writing these only repaints chrome.
type ChromeState struct {
	OmarchyPaletteJSON string // "omarchyPaletteJson"
	OmarchyFont        string // "omarchyFont"
}

// LayoutState feeds pure QML bindings for the toolbar and scrollbars.
type LayoutState struct {
	ShowScrollbars  bool    // "showScrollbars"
	ToolbarPosition string  // "toolbarPositionConfig"
	ToolbarVisible  bool    // "toolbarVisibleConfig"
	PageOrientation string  // "pageOrientationConfig"
	ZoomDefault     float64 // "zoomDefaultConfig"
}

// IdentityState is the startup-only identity. Apply never populates it, so a
// refresh can never rebuild the theme combo model or reset its index.
type IdentityState struct {
	DocumentTitle           string // "documentTitle"
	DocDir                  string // "docDir"
	InitialAnchor           string // "initialAnchor"
	ViewerThemeLabelsJSON   string // "viewerThemeLabelsJson"
	InitialViewerThemeIndex int    // "initialViewerThemeIndex"
}

// DocumentState carries the rendered document. ViewerThemesJSON is the only
// property with a side effect: its QML binding reloads the page, so Publish
// writes it after the chrome it paints against is already current.
type DocumentState struct {
	ViewerThemesJSON string // "viewerThemesJson"
	PostLoadScripts  string // "postLoadScripts"
	DocReloadSignal  int    // "docReloadSignal"
}

// Publish writes each non-nil group. The order is load-bearing: Chrome and
// Layout repaint, Identity sets the startup index and labels, and Document
// ends with viewerThemesJson, whose binding reloads the document.
func (s State) Publish(to StateSink) {
	if s.Chrome != nil {
		to.SetString("omarchyPaletteJson", s.Chrome.OmarchyPaletteJSON)
		to.SetString("omarchyFont", s.Chrome.OmarchyFont)
	}
	if s.Layout != nil {
		to.SetBool("showScrollbars", s.Layout.ShowScrollbars)
		to.SetString("toolbarPositionConfig", s.Layout.ToolbarPosition)
		to.SetBool("toolbarVisibleConfig", s.Layout.ToolbarVisible)
		to.SetString("pageOrientationConfig", s.Layout.PageOrientation)
		to.SetFloat("zoomDefaultConfig", s.Layout.ZoomDefault)
	}
	if s.Identity != nil {
		to.SetString("documentTitle", s.Identity.DocumentTitle)
		to.SetString("docDir", s.Identity.DocDir)
		to.SetString("initialAnchor", s.Identity.InitialAnchor)
		to.SetString("viewerThemeLabelsJson", s.Identity.ViewerThemeLabelsJSON)
		to.SetInt("initialViewerThemeIndex", s.Identity.InitialViewerThemeIndex)
	}
	if s.Document != nil {
		to.SetString("viewerThemesJson", s.Document.ViewerThemesJSON)
		to.SetString("postLoadScripts", s.Document.PostLoadScripts)
		to.SetInt("docReloadSignal", s.Document.DocReloadSignal)
	}
}

// Change is a sealed trigger. Apply performs the recompute; the event type
// carries the only variation between the old watcher bodies.
type Change interface{ isChange() }

// ThemeChanged reports the Omarchy palette or font watcher fired.
type ThemeChanged struct{}

// ConfigChanged reports config.toml or a theme CSS file changed.
type ConfigChanged struct{}

// DocumentSaved reports the open document was rewritten.
type DocumentSaved struct{}

// OrientationToggled reports the user flipped the sheet between portrait and
// landscape.
type OrientationToggled struct{}

// Persist carries the UI's final selection to disk and publishes nothing.
type Persist struct {
	ViewerThemeIndex int
	ToolbarVisible   bool
}

func (ThemeChanged) isChange()       {}
func (ConfigChanged) isChange()      {}
func (DocumentSaved) isChange()      {}
func (Persist) isChange()            {}
func (OrientationToggled) isChange() {}

// SessionOptions is the startup state main hands the session.
type SessionOptions struct {
	AbsPath    string
	DocDir     string
	Title      string
	Anchor     string
	Config     Config
	ThemesDir  string
	DiskThemes []ViewerTheme
	Palette    PaletteSource
	// DefaultConfig is the embedded config.toml, used to complete the user's
	// file on quit.
	DefaultConfig []byte
}

// Session owns one open document: its raw text, config, palette, font, theme
// list, and reload counter. It recomputes and publishes UI state; it holds no
// Qt types, so it runs on the GUI thread only because its callers do.
type Session struct {
	absPath     string
	raw         string
	docDir      string
	title       string
	anchor      string
	cfg         Config
	themesDir   string
	diskThemes  []ViewerTheme
	paletteSrc  PaletteSource
	defaultTOML []byte
	// orientationOverride is the orientation the user picked with Ctrl+R: a view
	// held in memory until the document closes, never a default. Empty means none.
	orientationOverride string

	palette ThemePalette
	font    string
	reloads int
}

// StartSession reads the document and returns the full startup State. It
// fails only when the document cannot be read; palette, font, and config
// problems degrade to their defaults, as before.
func StartSession(o SessionOptions) (*Session, State, error) {
	raw, err := LoadFile(o.AbsPath)
	if err != nil {
		return nil, State{}, err
	}
	s := &Session{
		absPath:     o.AbsPath,
		raw:         raw,
		docDir:      o.DocDir,
		title:       o.Title,
		anchor:      o.Anchor,
		cfg:         o.Config,
		themesDir:   o.ThemesDir,
		diskThemes:  o.DiskThemes,
		paletteSrc:  o.Palette,
		defaultTOML: o.DefaultConfig,
	}
	s.palette = s.readPalette()
	s.font = ReadOmarchyFont()

	all := allThemes(s.diskThemes)
	state := State{
		Chrome: &ChromeState{
			OmarchyPaletteJSON: paletteJSON(s.palette),
			OmarchyFont:        s.font,
		},
		Layout: s.layout(),
		Identity: &IdentityState{
			DocumentTitle:           s.title,
			DocDir:                  s.docDir,
			InitialAnchor:           s.anchor,
			ViewerThemeLabelsJSON:   ViewerThemeLabelsJSON(all),
			InitialViewerThemeIndex: themeIndex(all, s.cfg.ViewerTheme),
		},
		Document: s.documentState(),
	}
	return s, state, nil
}

// effective returns the config the document is shown with: the defaults, then
// the sheet the document declares in its front matter, then the orientation
// the user picked in this session.
func (s *Session) effective() Config {
	c := s.cfg.forDocument(s.raw)
	if s.orientationOverride != "" {
		c.PageOrientation = s.orientationOverride
	}
	return c
}

// layout is the Layout group for the session's current config and sheet.
func (s *Session) layout() *LayoutState {
	c := s.effective()
	return &LayoutState{
		ShowScrollbars:  c.ShowScrollbars,
		ToolbarPosition: c.ToolbarPosition,
		ToolbarVisible:  toolbarVisible(c),
		PageOrientation: sheetOrientation(c.PageOrientation),
		ZoomDefault:     zoomDefault(c),
	}
}

// Apply moves the session on one trigger and returns only the State groups
// that changed. A zero State means "publish nothing". It never sets Identity.
func (s *Session) Apply(c Change) State {
	switch ch := c.(type) {
	case ThemeChanged:
		p := s.readPalette()
		font := ReadOmarchyFont()
		if p == s.palette && font == s.font {
			return State{}
		}
		s.palette = p
		s.font = font
		return State{
			Chrome: &ChromeState{
				OmarchyPaletteJSON: paletteJSON(s.palette),
				OmarchyFont:        s.font,
			},
			Document: s.documentState(),
		}
	case ConfigChanged:
		cfg := LoadConfig()
		cfg.ViewerTheme = s.cfg.ViewerTheme // a config edit never yanks the session theme
		s.cfg = cfg
		return State{Layout: s.layout(), Document: s.documentState()}
	case OrientationToggled:
		if sheetOrientation(s.effective().PageOrientation) == OrientationLandscape {
			s.orientationOverride = OrientationPortrait
		} else {
			s.orientationOverride = OrientationLandscape
		}
		return State{Layout: s.layout(), Document: s.documentState()}
	case DocumentSaved:
		raw, err := LoadFile(s.absPath)
		if err != nil {
			log.Printf("session: cannot re-read %s: %v", s.absPath, err)
			return State{}
		}
		s.raw = raw
		s.reloads++
		return State{Layout: s.layout(), Document: s.documentState()}
	case Persist:
		all := allThemes(s.diskThemes)
		cfg := s.cfg
		if ch.ViewerThemeIndex >= 0 && ch.ViewerThemeIndex < len(all) {
			cfg.ViewerTheme = all[ch.ViewerThemeIndex].ID
		}
		v := ch.ToolbarVisible
		cfg.ToolbarVisible = &v
		SaveConfig(cfg, s.defaultTOML)
		return State{}
	default:
		return State{}
	}
}

// PreparePdf renders the print document against the session's current raw
// text, font, and config, and the same palette source used for the screen.
func (s *Session) PreparePdf() PdfExport {
	return PreparePdfExport(s.absPath, s.raw, s.docDir, s.font, s.effective(), s.paletteSrc)
}

// readPalette resolves the live palette, falling back to the built-in one when
// no watcher is available.
func (s *Session) readPalette() ThemePalette {
	if s.paletteSrc != nil {
		return s.paletteSrc.Palette()
	}
	return GetThemePalette("omarchy")
}

// documentState renders the document for every theme from the session's
// current raw text, palette, font, and config.
func (s *Session) documentState() *DocumentState {
	scripts, err := json.Marshal(PostLoadScripts(s.raw))
	if err != nil {
		log.Printf("session: cannot marshal post-load scripts: %v", err)
	}
	return &DocumentState{
		ViewerThemesJSON: RenderViewerThemesJSON(s.raw, s.palette, s.docDir, s.diskThemes, s.themesDir, s.effective(), s.font),
		PostLoadScripts:  string(scripts),
		DocReloadSignal:  s.reloads,
	}
}

// zoomDefault reads the config's initial zoom, taking a config that never set
// it (a zero value) as 100 %.
func zoomDefault(cfg Config) float64 {
	if cfg.ZoomDefault == 0 {
		return 1.0
	}
	return cfg.ZoomDefault
}

// toolbarVisible reads the config's tri-state visibility, defaulting to hidden.
func toolbarVisible(cfg Config) bool {
	if cfg.ToolbarVisible == nil {
		return false
	}
	return *cfg.ToolbarVisible
}

// allThemes returns the full ordered theme list: system first, then disk themes.
func allThemes(diskThemes []ViewerTheme) []ViewerTheme {
	all := make([]ViewerTheme, 0, 1+len(diskThemes))
	all = append(all, ViewerTheme{ID: "system", Label: "System"})
	return append(all, diskThemes...)
}

// themeIndex returns the index of themeID in all, defaulting to 0.
func themeIndex(all []ViewerTheme, themeID string) int {
	for i, t := range all {
		if t.ID == themeID {
			return i
		}
	}
	return 0
}

// paletteJSON serializes the palette main passes to QML as context data.
func paletteJSON(p ThemePalette) string {
	m := map[string]string{
		"background":   p.Background,
		"foreground":   p.Foreground,
		"accent":       p.Accent,
		"surface":      p.Surface,
		"border":       p.Border,
		"codeBg":       p.CodeBg,
		"codeFg":       p.CodeFg,
		"linkColor":    p.LinkColor,
		"headingColor": p.HeadingColor,
		"selectionBg":  p.SelectionBg,
		"selectionFg":  p.SelectionFg,
	}
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("session: cannot marshal palette: %v", err)
	}
	return string(b)
}
