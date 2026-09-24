package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"
)

// LoadConfig decodes through configRaw, so a key declared only on Config would
// silently never be read from the file.
func TestConfigRawDecodesCodeLineNumbers(t *testing.T) {
	var off configRaw
	if _, err := toml.Decode("code_line_numbers = false\n", &off); err != nil {
		t.Fatal(err)
	}
	if off.CodeLineNumbers == nil || *off.CodeLineNumbers {
		t.Error("code_line_numbers = false was not decoded")
	}

	var absent configRaw
	if _, err := toml.Decode("viewer_theme = \"night\"\n", &absent); err != nil {
		t.Fatal(err)
	}
	if absent.CodeLineNumbers != nil {
		t.Error("an absent key must stay nil so the default can apply")
	}
}

func TestShippedConfigParses(t *testing.T) {
	data, err := os.ReadFile("../resources/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	var shipped configRaw
	if _, err := toml.Decode(string(data), &shipped); err != nil {
		t.Fatalf("the shipped config.toml does not parse: %v", err)
	}
	if shipped.CodeLineNumbers == nil || !*shipped.CodeLineNumbers {
		t.Error("the shipped config should enable code_line_numbers")
	}
}

func TestConfigOverrideCSSLineNumbers(t *testing.T) {
	const hide = "pre .hljs-ln-numbers { display: none; }"
	off, on := false, true

	if got := ConfigOverrideCSS(Config{CodeLineNumbers: &off}); !strings.Contains(got, hide) {
		t.Error("code_line_numbers = false should hide the gutter")
	}
	if got := ConfigOverrideCSS(Config{CodeLineNumbers: &on}); strings.Contains(got, "display: none") {
		t.Error("code_line_numbers = true should not hide the gutter")
	}
	if got := ConfigOverrideCSS(Config{}); strings.Contains(got, "display: none") {
		t.Error("an absent key must default to showing the gutter")
	}
}

// Saving on exit used to re-encode the whole file, dropping the comments that
// document what every option accepts.
func TestSaveConfigPreservesComments(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	original := "# Tema del visor al arrancar.\nviewer_theme = \"github\"\n\n# Visibilidad del toolbar.\ntoolbar_visible = false\n"
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	visible := true
	SaveConfig(Config{ViewerTheme: "night", ToolbarVisible: &visible})

	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	body := string(got)
	for _, comment := range []string{"# Tema del visor al arrancar.", "# Visibilidad del toolbar."} {
		if !strings.Contains(body, comment) {
			t.Errorf("comment %q was dropped", comment)
		}
	}
	if !strings.Contains(body, `viewer_theme = "night"`) {
		t.Errorf("viewer_theme was not updated, got:\n%s", body)
	}
	if !strings.Contains(body, "toolbar_visible = true") {
		t.Errorf("toolbar_visible was not updated, got:\n%s", body)
	}
}

// With nothing on disk to edit, the whole file still gets written.
func TestSaveConfigWritesWholeFileWhenMissing(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	SaveConfig(Config{ViewerTheme: "sepia"})

	got, err := os.ReadFile(filepath.Join(home, ".config", "o-mark", "config.toml"))
	if err != nil {
		t.Fatalf("config was not created: %v", err)
	}
	if !strings.Contains(string(got), `viewer_theme = "sepia"`) {
		t.Errorf("viewer_theme missing, got:\n%s", got)
	}
}

func TestSetTOMLValue(t *testing.T) {
	cases := []struct {
		name, body, key, value, want string
	}{
		{"replaces the value", "viewer_theme = \"github\"\n", "viewer_theme", `"night"`, "viewer_theme = \"night\"\n"},
		{"appends a missing key", "# only a comment\n", "toolbar_visible", "true", "# only a comment\ntoolbar_visible = true\n"},
		{"leaves a commented example alone", "# viewer_theme = \"mono\"\nviewer_theme = \"github\"\n", "viewer_theme", `"night"`, "# viewer_theme = \"mono\"\nviewer_theme = \"night\"\n"},
		{"keeps the surrounding spacing", "  viewer_theme   =   \"github\"\n", "viewer_theme", `"night"`, "  viewer_theme   =   \"night\"\n"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := setTOMLValue(c.body, c.key, c.value); got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}

// loadConfigFrom runs LoadConfig against a config.toml holding body.
func loadConfigFrom(t *testing.T, body string) Config {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "config.toml"), []byte(body), 0644); err != nil {
		t.Fatal(err)
	}
	return LoadConfig()
}

func TestLoadConfigPageDefaults(t *testing.T) {
	cfg := loadConfigFrom(t, "")
	if cfg.PageFormat != "a4" || cfg.PageOrientation != "portrait" {
		t.Errorf("defaults: got %q %q, want a4 portrait", cfg.PageFormat, cfg.PageOrientation)
	}
}

func TestLoadConfigPageInvalid(t *testing.T) {
	cfg := loadConfigFrom(t, "page_format = \"letter\"\npage_orientation = \"sideways\"\n")
	if cfg.PageFormat != "a4" || cfg.PageOrientation != "portrait" {
		t.Errorf("invalid values: got %q %q, want a4 portrait", cfg.PageFormat, cfg.PageOrientation)
	}
}

func TestLoadConfigMigratesDocumentMaxWidth(t *testing.T) {
	tests := []struct{ width, want string }{
		{"default", "portrait"},
		{"210mm", "portrait"},
		{"297mm", "landscape"},
		{"50%", "portrait"},
		{"80ch", "portrait"},
	}
	for _, tc := range tests {
		cfg := loadConfigFrom(t, "document_max_width = \""+tc.width+"\"\n")
		if cfg.PageOrientation != tc.want {
			t.Errorf("document_max_width %q: got %q, want %q", tc.width, cfg.PageOrientation, tc.want)
		}
	}
}

func TestLoadConfigOrientationBeatsLegacyWidth(t *testing.T) {
	cfg := loadConfigFrom(t, "document_max_width = \"297mm\"\npage_orientation = \"portrait\"\n")
	if cfg.PageOrientation != "portrait" {
		t.Errorf("an explicit orientation must win over the legacy width, got %q", cfg.PageOrientation)
	}
}

func TestConfigOverrideCSSPageSheet(t *testing.T) {
	portrait := ConfigOverrideCSS(Config{PageOrientation: "portrait"})
	for _, want := range []string{"--o-mark-page-width: 210mm", "--o-mark-page-height: 297mm", "max-width: var(--o-mark-page-width)", "min-height: var(--o-mark-page-height)"} {
		if !strings.Contains(portrait, want) {
			t.Errorf("portrait CSS is missing %q:\n%s", want, portrait)
		}
	}
	landscape := ConfigOverrideCSS(Config{PageOrientation: "landscape"})
	if !strings.Contains(landscape, "--o-mark-page-width: 297mm") || !strings.Contains(landscape, "--o-mark-page-height: 210mm") {
		t.Errorf("landscape CSS swaps the sides:\n%s", landscape)
	}
	if !strings.Contains(ConfigOverrideCSS(Config{}), "--o-mark-page-width: 210mm") {
		t.Error("a zero-value Config must still get the portrait sheet")
	}
}

func TestSaveConfigWritesPageAndDropsWidth(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.toml")
	seed := "# Theme.\nviewer_theme = \"github\"\n\n# Maximum width.\ndocument_max_width = \"297mm\"\n\n# Font.\nfont_size_base = \"default\"\n"
	if err := os.WriteFile(p, []byte(seed), 0644); err != nil {
		t.Fatal(err)
	}
	SaveConfig(Config{ViewerTheme: "night", PageOrientation: "landscape"})
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	for _, want := range []string{"viewer_theme = \"night\"", "page_format = \"a4\"", "page_orientation = \"landscape\"", "# Theme.", "# Font.", "font_size_base = \"default\""} {
		if !strings.Contains(out, want) {
			t.Errorf("saved config is missing %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "document_max_width") || strings.Contains(out, "Maximum width") {
		t.Errorf("the retired width and its comment must be gone:\n%s", out)
	}
	if again := LoadConfig(); again.PageOrientation != "landscape" {
		t.Errorf("reload after save: got %q, want landscape", again.PageOrientation)
	}
}

func TestRemoveTOMLKey(t *testing.T) {
	if got := removeTOMLKey("a = 1\nb = 2\n", "c"); got != "a = 1\nb = 2\n" {
		t.Errorf("an absent key must leave the body alone, got %q", got)
	}
	if got := removeTOMLKey("a = 1\n\n# about b\nb = 2\n\nc = 3\n", "b"); got != "a = 1\n\nc = 3\n" {
		t.Errorf("removing b: got %q", got)
	}
}
