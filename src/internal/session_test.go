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
	f   float64
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

func (r *sessionRecorder) SetFloat(k string, v float64) {
	r.calls = append(r.calls, sessionCall{key: k, f: v})
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
	current   ThemePalette
	print     ThemePalette
	treatment int
}

func (f *fakePalette) Palette() ThemePalette      { return f.current }
func (f *fakePalette) PrintPalette() ThemePalette { return f.print }
func (f *fakePalette) SetTreatment(index int)     { f.treatment = index }

func sessionPalette(bg string) ThemePalette {
	return ThemePalette{
		Background: bg, Foreground: "#111111", Accent: "#222222",
		Surface: "#333333", Border: "#444444", CodeBg: "#555555",
		CodeFg: "#666666", LinkColor: "#777777", HeadingColor: "#888888",
		SelectionBg: "#999999", SelectionFg: "#aaaaaa",
	}
}

func sessionConfig(treatment string) Config {
	v := false
	return Config{ViewerTreatment: treatment, ToolbarPosition: "bottom", ToolbarVisible: &v}
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

func newSession(t *testing.T, doc string, cfg Config, p PaletteSource) *Session {
	t.Helper()
	s, _, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config: cfg, Palette: p,
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
		Config:  sessionConfig("original"),
		Palette: &fakePalette{current: sessionPalette("#101010")},
	})
	if err != nil {
		t.Fatal(err)
	}

	r := &sessionRecorder{}
	state.Publish(r)

	docAt := r.index("documentHtml")
	if docAt < 0 {
		t.Fatal("documentHtml was not published")
	}
	for _, key := range []string{"omarchyPaletteJson", "omarchyFont", "initialTreatmentIndex", "treatmentLabelsJson", "documentBg"} {
		at := r.index(key)
		if at < 0 {
			t.Fatalf("%s was not published", key)
		}
		if at > docAt {
			t.Errorf("%s published after documentHtml", key)
		}
	}
}

func TestSessionStartupTreatmentFollowsTheConfig(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010")}
	_, state, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config:  sessionConfig("Inverted"),
		Palette: fp,
	})
	if err != nil {
		t.Fatal(err)
	}
	want, _ := TreatmentIndex("inverted")
	if state.Identity == nil || state.Identity.InitialTreatmentIndex != want {
		t.Fatalf("index: got %+v, want %d", state.Identity, want)
	}
	if fp.treatment != want {
		t.Errorf("the palette source was not set to the treatment: %d", fp.treatment)
	}
}

func TestSessionUnknownTreatmentIsOriginal(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	_, state, err := StartSession(SessionOptions{
		AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md",
		Config:  sessionConfig("sepia-that-does-not-exist"),
		Palette: &fakePalette{current: sessionPalette("#101010")},
	})
	if err != nil {
		t.Fatal(err)
	}
	if state.Identity.InitialTreatmentIndex != 0 {
		t.Errorf("got %d, want Original (0)", state.Identity.InitialTreatmentIndex)
	}
}

func TestSessionWithoutAPaletteSourceOffersNoTreatments(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	s, state, err := StartSession(SessionOptions{AbsPath: doc, DocDir: filepath.Dir(doc), Title: "doc.md", Config: sessionConfig("original")})
	if err != nil {
		t.Fatal(err)
	}
	if state.Identity.TreatmentLabelsJSON != "[]" {
		t.Errorf("labels: %s", state.Identity.TreatmentLabelsJSON)
	}
	if got := s.Apply(TreatmentChosen{Index: 1}); got.Chrome != nil || got.Document != nil {
		t.Errorf("a treatment without a palette source changes nothing, got %+v", got)
	}
}

func TestSessionTreatmentChoiceRepaintsChromeAndDocument(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010")}
	s := newSession(t, doc, sessionConfig("original"), fp)

	inverted, _ := TreatmentIndex("inverted")
	fp.current = sessionPalette("#f0f0f0")
	got := s.Apply(TreatmentChosen{Index: inverted})
	if got.Chrome == nil || got.Document == nil {
		t.Fatalf("a chosen treatment must repaint the chrome and the document, got %+v", got)
	}
	if fp.treatment != inverted {
		t.Errorf("the palette source was not told: %d", fp.treatment)
	}
	if !strings.Contains(got.Document.DocumentHTML, "#f0f0f0") || got.Document.DocumentBg != "#f0f0f0" {
		t.Error("the document was not rendered from the new palette")
	}
	if again := s.Apply(TreatmentChosen{Index: inverted}); again.Chrome != nil || again.Document != nil {
		t.Errorf("choosing the current treatment publishes nothing, got %+v", again)
	}
	if bad := s.Apply(TreatmentChosen{Index: 99}); bad.Chrome != nil || bad.Document != nil {
		t.Errorf("an out-of-range index publishes nothing, got %+v", bad)
	}
}

func TestSessionThemeUnchangedPublishesNothing(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(ThemeChanged{})
	if got.Chrome != nil || got.Layout != nil || got.Identity != nil || got.Document != nil {
		t.Errorf("an unchanged palette must publish nothing, got %+v", got)
	}
}

func TestSessionThemeChangeRendersFromNewPalette(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010")}
	s := newSession(t, doc, sessionConfig("original"), fp)

	fp.current = sessionPalette("#202020")
	got := s.Apply(ThemeChanged{})
	if got.Chrome == nil || got.Document == nil {
		t.Fatalf("a changed palette must publish chrome and document, got %+v", got)
	}
	if !strings.Contains(got.Chrome.OmarchyPaletteJSON, "#202020") {
		t.Error("chrome palette was not updated")
	}
	if !strings.Contains(got.Document.DocumentHTML, "#202020") {
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
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(ConfigChanged{})
	if s.treatment != 0 {
		t.Errorf("a config reload changed the session treatment to %d", s.treatment)
	}
	if got.Layout == nil || got.Layout.ToolbarPosition != "top" {
		t.Errorf("layout was not reloaded: %+v", got.Layout)
	}
}

func TestSessionDocumentReloadBumpsSignal(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# One\n")
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})

	if err := os.WriteFile(doc, []byte("# Two\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := s.Apply(DocumentSaved{})
	if got.Document == nil || got.Document.DocReloadSignal != 1 {
		t.Fatalf("signal: got %+v, want 1", got.Document)
	}
	if !strings.Contains(got.Document.DocumentHTML, "Two") {
		t.Error("document was not re-rendered from the new content")
	}
}

func TestSessionDocumentReloadSurvivesUnreadableFile(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# One\n")
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})

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
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	highContrast, _ := TreatmentIndex("highcontrast")
	s.Apply(TreatmentChosen{Index: highContrast})

	got := s.Apply(Persist{ToolbarVisible: true})
	if got.Chrome != nil || got.Layout != nil || got.Identity != nil || got.Document != nil {
		t.Errorf("Persist must publish nothing, got %+v", got)
	}

	body, err := os.ReadFile(filepath.Join(home, ".config", "o-mark", "config.toml"))
	if err != nil {
		t.Fatalf("config was not written: %v", err)
	}
	if !strings.Contains(string(body), `viewer_treatment = "highcontrast"`) {
		t.Errorf("selected treatment was not persisted:\n%s", body)
	}
	if !strings.Contains(string(body), "toolbar_visible = true") {
		t.Errorf("toolbar visibility was not persisted:\n%s", body)
	}
}

func TestSessionDocumentUsesTheConfiguredFont(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	cfg := sessionConfig("original")
	cfg.Font = "Liberation Serif, serif"
	s := newSession(t, doc, cfg, &fakePalette{current: sessionPalette("#101010")})
	if !strings.Contains(s.documentState().DocumentHTML, "--o-mark-font: Liberation Serif, serif;") {
		t.Error("the config font did not reach the document")
	}
	cfg.Font = MonoFont
	s = newSession(t, doc, cfg, &fakePalette{current: sessionPalette("#101010")})
	if !strings.Contains(s.documentState().DocumentHTML, "--o-mark-font: "+s.font+";") {
		t.Error("mono must resolve to the Omarchy font")
	}
}

func TestSessionPreparePdfUsesPaletteSource(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	fp := &fakePalette{current: sessionPalette("#101010"), print: sessionPalette("#ffffff")}
	s := newSession(t, doc, sessionConfig("original"), fp)

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
	cfg := sessionConfig("original")
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
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})

	got := s.Apply(OrientationToggled{})
	if got.Layout == nil || got.Layout.PageOrientation != "landscape" || got.Document == nil {
		t.Fatalf("first toggle: %+v", got)
	}
	if !strings.Contains(got.Document.DocumentHTML, "--o-mark-page-width: 297mm") {
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
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	s.Apply(OrientationToggled{})

	got := s.Apply(ConfigChanged{})
	if got.Layout.PageOrientation != "landscape" {
		t.Errorf("a config edit yanked the session orientation: %q", got.Layout.PageOrientation)
	}
}

// The orientation of a session is a view, not a default: only the config
// edited by hand sets it.
func TestSessionPersistLeavesTheOrientationDefault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte("page_orientation = \"portrait\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	s.Apply(OrientationToggled{})
	s.Apply(Persist{})

	body, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(body), `page_orientation = "portrait"`) {
		t.Errorf("the session orientation leaked into the default:\n%s", body)
	}
}

func TestSessionPreparePdfCarriesOrientation(t *testing.T) {
	isolateHome(t)
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), sessionConfig("original"), &fakePalette{current: sessionPalette("#101010"), print: sessionPalette("#ffffff")})
	if got := s.PreparePdf().Orientation; got != "portrait" {
		t.Errorf("default orientation: %q", got)
	}
	s.Apply(OrientationToggled{})
	if got := s.PreparePdf().Orientation; got != "landscape" {
		t.Errorf("after a toggle: %q", got)
	}
}

const declaredDoc = "---\npage_format: a5\npage_orientation: landscape\n---\n# Hi\n"

func TestSessionDocumentDeclaresItsSheet(t *testing.T) {
	isolateHome(t)
	s := newSession(t, writeSessionDoc(t, declaredDoc), sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	got := s.layout()
	if got.PageOrientation != "landscape" {
		t.Errorf("layout orientation: %q", got.PageOrientation)
	}
	pdf := s.PreparePdf()
	if pdf.Format != "a5" || pdf.Orientation != "landscape" {
		t.Errorf("pdf sheet: %q %q", pdf.Format, pdf.Orientation)
	}
	doc := s.documentState().DocumentHTML
	if !strings.Contains(doc, "--o-mark-page-width: 210mm") || !strings.Contains(doc, "--o-mark-page-height: 148mm") {
		t.Error("the screen render does not carry the A5 landscape sheet")
	}
}

// Ctrl+R is a view of this session: it wins over the declaration, dies with the
// session, and never touches the defaults.
func TestSessionOverrideWinsOverTheDeclaration(t *testing.T) {
	isolateHome(t)
	s := newSession(t, writeSessionDoc(t, declaredDoc), sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	got := s.Apply(OrientationToggled{})
	if got.Layout.PageOrientation != "portrait" {
		t.Errorf("a flip from the declared landscape should be portrait, got %q", got.Layout.PageOrientation)
	}
	if s.cfg.PageOrientation != "" && s.cfg.PageOrientation != "portrait" {
		t.Errorf("the default was changed: %q", s.cfg.PageOrientation)
	}
	if again := s.Apply(OrientationToggled{}); again.Layout.PageOrientation != "landscape" {
		t.Errorf("second flip: %q", again.Layout.PageOrientation)
	}
}

func TestSessionOverrideSurvivesReloadsEvenWhenTheDeclarationChanges(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, declaredDoc)
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	s.Apply(OrientationToggled{}) // portrait, over the declared landscape

	if err := os.WriteFile(doc, []byte("---\npage_orientation: landscape\n---\n# Edited\n"), 0644); err != nil {
		t.Fatal(err)
	}
	got := s.Apply(DocumentSaved{})
	if got.Layout == nil || got.Layout.PageOrientation != "portrait" {
		t.Errorf("the override should survive a reload, got %+v", got.Layout)
	}

	if err := os.WriteFile(doc, []byte("---\npage_orientation: portrait\n---\n# Edited again\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := s.Apply(DocumentSaved{}); got.Layout.PageOrientation != "portrait" {
		t.Errorf("still the override, got %q", got.Layout.PageOrientation)
	}
}

func TestSessionReloadPublishesTheNewDeclaration(t *testing.T) {
	isolateHome(t)
	doc := writeSessionDoc(t, "# Hi\n")
	s := newSession(t, doc, sessionConfig("original"), &fakePalette{current: sessionPalette("#101010")})
	if err := os.WriteFile(doc, []byte(declaredDoc), 0644); err != nil {
		t.Fatal(err)
	}
	if got := s.Apply(DocumentSaved{}); got.Layout == nil || got.Layout.PageOrientation != "landscape" {
		t.Errorf("the toolbar must follow a declaration added by a reload, got %+v", got.Layout)
	}
}

func TestSessionConfigEditAppliesUntilTheUserFlips(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv(OmarchyStateEnv, "")
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	write := func(o string) {
		if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte("page_orientation = \""+o+"\"\n"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	write("portrait")
	s := newSession(t, writeSessionDoc(t, "# Hi\n"), LoadConfig(), &fakePalette{current: sessionPalette("#101010")})

	write("landscape")
	if got := s.Apply(ConfigChanged{}); got.Layout.PageOrientation != "landscape" {
		t.Errorf("an edit of the default should apply while nothing was flipped, got %q", got.Layout.PageOrientation)
	}
	s.Apply(OrientationToggled{}) // portrait
	write("landscape")
	if got := s.Apply(ConfigChanged{}); got.Layout.PageOrientation != "portrait" {
		t.Errorf("after a flip the session view wins, got %q", got.Layout.PageOrientation)
	}
}

func TestSessionLayoutCarriesTheZoomDefault(t *testing.T) {
	isolateHome(t)
	cfg := sessionConfig("original")
	if got := newSession(t, writeSessionDoc(t, "# Hi\n"), cfg, &fakePalette{current: sessionPalette("#101010")}).layout().ZoomDefault; got != 1.0 {
		t.Errorf("a config that sets none is 100 %%, got %v", got)
	}
	cfg.ZoomDefault = 1.2
	if got := newSession(t, writeSessionDoc(t, "# Hi\n"), cfg, &fakePalette{current: sessionPalette("#101010")}).layout().ZoomDefault; got != 1.2 {
		t.Errorf("got %v, want 1.2", got)
	}
}
