package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// sessionCall records one state write, in order.
type sessionCall struct {
	key string
	str string
	i   int
	b   bool
}

type sessionRecorder struct{ calls []sessionCall }

func (r *sessionRecorder) SetString(k, v string) {
	r.calls = append(r.calls, sessionCall{key: k, str: v})
}
func (r *sessionRecorder) SetInt(k string, v int) {
	r.calls = append(r.calls, sessionCall{key: k, i: v})
}
func (r *sessionRecorder) SetBool(k string, v bool) {
	r.calls = append(r.calls, sessionCall{key: k, b: v})
}

func (r *sessionRecorder) index(key string) int {
	for i, c := range r.calls {
		if c.key == key {
			return i
		}
	}
	return -1
}

// fakePalette is a PaletteSource with scripted screen and print palettes.
type fakePalette struct {
	current ThemePalette
	print   ThemePalette
}

func (f *fakePalette) Palette() ThemePalette      { return f.current }
func (f *fakePalette) PrintPalette() ThemePalette { return f.print }

func sessionPalette(bg string) ThemePalette {
	return ThemePalette{
		Background: bg, Foreground: "#111111", Accent: "#222222",
		Surface: "#333333", Border: "#444444", CodeBg: "#555555",
		CodeFg: "#666666", LinkColor: "#777777", HeadingColor: "#888888",
		SelectionBg: "#999999", SelectionFg: "#aaaaaa",
	}
}

func sessionConfig(theme string) Config {
	v := false
	return Config{ViewerTheme: theme, ToolbarPosition: "bottom", ToolbarVisible: &v}
}

func isolateHome(t *testing.T) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	t.Setenv(OmarchyStateEnv, "")
}

func writeSessionDoc(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "doc.md")
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return p
}

func newSession(t *testing.T, doc string, cfg Config, diskThemes []ViewerTheme, p PaletteSource) *Session {
	t.Helper()
	s, _, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config: cfg, DiskThemes: diskThemes, Palette: p,
	})
	if err != nil {
		t.Fatalf("StartSession: %v", err)
	}
	return s
}

func TestSessionStartupPublishesChromeBeforeDocument(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	_, state, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config:  sessionConfig("github"),
		Palette: &fakePalette{current: sessionPalette("#101010")},
	})
	if err != nil {
		t.Fatal(err)
	}

	r := &sessionRecorder{}
	state.Publish(r)

	docAt := r.index("viewerThemesJson")
	if docAt < 0 {
		t.Fatal("viewerThemesJson was not published")
	}
	for _, key := range []string{"omarchyPaletteJson", "omarchyFont", "initialViewerThemeIndex", "viewerThemeLabelsJson"} {
		at := r.index(key)
		if at < 0 {
			t.Fatalf("%s was not published", key)
		}
		if at > docAt {
			t.Errorf("%s published after viewerThemesJson", key)
		}
	}
}

func TestSessionStartupIndexFollowsConfiguredTheme(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	_, state, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config:     sessionConfig("github"),
		DiskThemes: []ViewerTheme{{ID: "github", Label: "GitHub"}},
		Palette:    &fakePalette{current: sessionPalette("#101010")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Identity == nil || state.Identity.InitialViewerThemeIndex != 1 {
		t.Fatalf("index: got %+v, want 1", state.Identity)
	}
}

func TestSessionThemeUnchangedPublishesNothing(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(ThemeChanged{})
	if got.Chrome != nil || got.Layout != nil || got.Identity != nil || got.Document != nil {
		t.Errorf("an unchanged palette must publish nothing, got %+v", got)
	}
}

func TestSessionThemeChangeRendersFromNewPalette(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010")}
	s := newSession(t, doc, sessionConfig("github"), nil, fp)

	fp.current = sessionPalette("#202020")
	got := s.Apply(ThemeChanged{})
	if got.Chrome == nil || got.Document == nil {
		t.Fatalf("a changed palette must publish chrome and document, got %+v", got)
	}
	if !strings.Contains(got.Chrome.OmarchyPaletteJSON, "#202020") {
		t.Error("chrome palette was not updated")
	}
	if !strings.Contains(got.Document.ViewerThemesJSON, "#202020") {
		t.Error("document was not rendered from the new palette")
	}
}

func TestSessionConfigReloadPreservesTheme(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	body := "viewer_theme = \"night\"\ntoolbar_position = \"top\"\n"
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}

	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(ConfigChanged{})
	if s.cfg.ViewerTheme != "github" {
		t.Errorf("a config reload changed the session theme to %q", s.cfg.ViewerTheme)
	}
	if got.Layout == nil || got.Layout.ToolbarPosition != "top" {
		t.Errorf("layout was not reloaded: %+v", got.Layout)
	}
}

func TestSessionDocumentReloadBumpsSignal(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# One\n")
	s := newSession(t, doc, sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})

	if err := os.WriteFile(doc, []byte("# Two\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := s.Apply(DocumentSaved{})
	if got.Document == nil || got.Document.DocReloadSignal != 1 {
		t.Fatalf("signal: got %+v, want 1", got.Document)
	}
	if !strings.Contains(got.Document.ViewerThemesJSON, "Two") {
		t.Error("document was not re-rendered from the new content")
	}
}

func TestSessionDocumentReloadSurvivesUnreadableFile(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# One\n")
	s := newSession(t, doc, sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})

	if err := os.Remove(doc); err != nil {
		t.Fatal(err)
	}
	got := s.Apply(DocumentSaved{})
	if got.Document != nil || got.Chrome != nil || got.Layout != nil || got.Identity != nil {
		t.Errorf("an unreadable file must publish nothing, got %+v", got)
	}
	if s.reloads != 0 {
		t.Errorf("the signal bumped on a failed reload: %d", s.reloads)
	}
}

func TestSessionPersistWritesSelection(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("github"), []ViewerTheme{{ID: "github", Label: "GitHub"}}, &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(Persist{ViewerThemeIndex: 1, ToolbarVisible: true})
	if got.Chrome != nil || got.Layout != nil || got.Identity != nil || got.Document != nil {
		t.Errorf("Persist must publish nothing, got %+v", got)
	}

	body, err := os.ReadFile(filepath.Join(home, ".config", "o-mark", "config.toml"))
	if err != nil {
		t.Fatalf("config was not written: %v", err)
	}
	if !strings.Contains(string(body), `viewer_theme = "github"`) {
		t.Errorf("selected theme was not persisted:\n%s", body)
	}
	if !strings.Contains(string(body), "toolbar_visible = true") {
		t.Errorf("toolbar visibility was not persisted:\n%s", body)
	}
}

func TestSessionPreparePdfUsesPaletteSource(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010"), print: sessionPalette("#ffffff")}
	s := newSession(t, doc, sessionConfig("github"), nil, fp)

	out := s.PreparePdf()
	if out.Exists {
		t.Fatal("exists: want false before any PDF is written")
	}
	if !strings.Contains(out.HTML, "#ffffff") {
		t.Error("print HTML did not use the palette source's print palette")
	}

	if err := os.WriteFile(out.Path, []byte("%PDF"), 0644); err != nil {
		t.Fatal(err)
	}
	again := s.PreparePdf()
	if !again.Exists || again.HTML != "" {
		t.Errorf("an existing sibling PDF must short-circuit: %+v", again)
	}
}

func TestSessionStartupPublishesPageOrientation(t *testing.T) {
	isolateHome(t)
	cfg := sessionConfig("github")
	cfg.PageOrientation = "landscape"
	s, state, err := StartSession(SessionOptions{
		AbsPath: writeSessionDoc(t, "# Hi\n"), Title: "doc.md", Config: cfg,
		Palette: &fakePalette{current: sessionPalette("#101010")},
	})
	if err != nil || s == nil {
		t.Fatalf("StartSession: %v", err)
	}
	if state.Layout == nil || state.Layout.PageOrientation != "landscape" {
		t.Errorf("startup layout: %+v", state.Layout)
	}
}

func TestSessionOrientationToggleRerendersTheSheet(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(OrientationToggled{})
	if got.Layout == nil || got.Layout.PageOrientation != "landscape" || got.Document == nil {
		t.Fatalf("first toggle: %+v", got)
	}
	if !strings.Contains(got.Document.ViewerThemesJSON, "--o-mark-page-width: 297mm") {
		t.Error("the re-rendered document does not carry the landscape sheet")
	}
	if back := s.Apply(OrientationToggled{}); back.Layout.PageOrientation != "portrait" {
		t.Errorf("second toggle: %+v", back.Layout)
	}
}

func TestSessionConfigReloadPreservesOrientation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("page_orientation = \"portrait\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})
	s.Apply(OrientationToggled{})

	got := s.Apply(ConfigChanged{})
	if got.Layout.PageOrientation != "landscape" {
		t.Errorf("a config edit yanked the session orientation: %q", got.Layout.PageOrientation)
	}
}

func TestSessionPersistWritesOrientation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010")})
	s.Apply(OrientationToggled{})
	s.Apply(Persist{})

	body, err := os.ReadFile(filepath.Join(home, ".config", "o-mark", "config.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `page_orientation = "landscape"`) {
		t.Errorf("orientation was not persisted:\n%s", body)
	}
}

func TestSessionPreparePdfCarriesOrientation(t *testing.T) {
	isolateHome(t)
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("github"), nil, &fakePalette{current: sessionPalette("#101010"), print: sessionPalette("#ffffff")})
	if got := s.PreparePdf().Orientation; got != "portrait" {
		t.Errorf("default orientation: %q", got)
	}
	s.Apply(OrientationToggled{})
	if got := s.PreparePdf().Orientation; got != "landscape" {
		t.Errorf("after a toggle: %q", got)
	}
}
