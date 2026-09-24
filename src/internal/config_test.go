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
	SaveConfig(Config{ViewerTheme: "night", ToolbarVisible: &visible}, nil)

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

	SaveConfig(Config{ViewerTheme: "sepia"}, nil)

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

func TestSaveConfigMigratesTheRetiredWidth(t *testing.T) {
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
	SaveConfig(Config{ViewerTheme: "night"}, []byte(testDefaults+"\n# Orientation.\npage_orientation = \"portrait\"\n"))
	got, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	out := string(got)
	for _, want := range []string{"viewer_theme = \"night\"", "# Format.\n# a4 or a5.\npage_format = \"a4\"", "# Orientation.\npage_orientation = \"landscape\"", "# Theme.", "# Font.", "font_size_base = \"default\""} {
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

func TestLoadConfigPageFormats(t *testing.T) {
	tests := []struct{ in, want string }{
		{"a3", "a3"}, {"a4", "a4"}, {"a5", "a5"}, {"A5", "a5"}, {"letter", "a4"}, {"", "a4"},
	}
	for _, tc := range tests {
		body := ""
		if tc.in != "" {
			body = "page_format = \"" + tc.in + "\"\n"
		}
		if got := loadConfigFrom(t, body).PageFormat; got != tc.want {
			t.Errorf("page_format %q: got %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPageSidesPerFormat(t *testing.T) {
	tests := []struct{ format, orientation, w, h string }{
		{"a3", "portrait", "297mm", "420mm"},
		{"a3", "landscape", "420mm", "297mm"},
		{"a4", "portrait", "210mm", "297mm"},
		{"a5", "portrait", "148mm", "210mm"},
		{"a5", "landscape", "210mm", "148mm"},
		{"", "", "210mm", "297mm"},
	}
	for _, tc := range tests {
		if w, h := pageSides(tc.format, tc.orientation); w != tc.w || h != tc.h {
			t.Errorf("pageSides(%q, %q) = %s x %s, want %s x %s", tc.format, tc.orientation, w, h, tc.w, tc.h)
		}
	}
}

func TestLoadConfigPdfMargins(t *testing.T) {
	def := loadConfigFrom(t, "")
	if def.PdfMarginVertical != "25mm" || def.PdfMarginHorizontal != "10mm" {
		t.Errorf("defaults: got %q %q", def.PdfMarginVertical, def.PdfMarginHorizontal)
	}
	set := loadConfigFrom(t, "pdf_margin_vertical = \"30mm\"\npdf_margin_horizontal = \"12.5mm\"\n")
	if set.PdfMarginVertical != "30mm" || set.PdfMarginHorizontal != "12.5mm" {
		t.Errorf("set: got %q %q", set.PdfMarginVertical, set.PdfMarginHorizontal)
	}
	for _, bad := range []string{"2cm", "25", "999mm", "-5mm", "auto"} {
		got := loadConfigFrom(t, "pdf_margin_vertical = \""+bad+"\"\n")
		if got.PdfMarginVertical != "25mm" {
			t.Errorf("margin %q should fall back to 25mm, got %q", bad, got.PdfMarginVertical)
		}
	}
}

func TestLoadConfigZoomDefault(t *testing.T) {
	tests := []struct {
		body string
		want float64
	}{
		{"", 1.0}, {"zoom_default = 1.2\n", 1.2}, {"zoom_default = 2\n", 2.0},
		{"zoom_default = 5.0\n", 2.0}, {"zoom_default = 0.1\n", 0.5}, {"zoom_default = \"big\"\n", 1.0},
	}
	for _, tc := range tests {
		if got := loadConfigFrom(t, tc.body).ZoomDefault; got != tc.want {
			t.Errorf("%q: got %v, want %v", tc.body, got, tc.want)
		}
	}
}

// An integer font_size_base used to be accepted; now it is invalid, but it must
// not take the rest of the file down with it.
func TestLoadConfigIntegerFontSizeIsInvalidNotFatal(t *testing.T) {
	cfg := loadConfigFrom(t, "font_size_base = 14\ntoolbar_position = \"top\"\npage_orientation = \"landscape\"\n")
	if cfg.FontSizeBase != DeferToTheme {
		t.Errorf("font_size_base: got %q, want the theme's size", cfg.FontSizeBase)
	}
	if cfg.ToolbarPosition != "top" || cfg.PageOrientation != "landscape" {
		t.Errorf("the other keys were lost: %q %q", cfg.ToolbarPosition, cfg.PageOrientation)
	}
	if got := loadConfigFrom(t, "font_size_base = \"14\"\n").FontSizeBase; got != "14" {
		t.Errorf("a quoted number is still valid, got %q", got)
	}
}

func TestSaveConfigLeavesThePageKeysToTheUser(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.toml")
	seed := "viewer_theme = \"github\"\npage_format = \"a5\"\npage_orientation = \"portrait\"\n"
	if err := os.WriteFile(p, []byte(seed), 0644); err != nil {
		t.Fatal(err)
	}
	SaveConfig(Config{ViewerTheme: "night", PageFormat: "a3", PageOrientation: "landscape"}, nil)
	got, _ := os.ReadFile(p)
	for _, want := range []string{"page_format = \"a5\"", "page_orientation = \"portrait\""} {
		if !strings.Contains(string(got), want) {
			t.Errorf("the session changed a default it must not write: missing %q in\n%s", want, got)
		}
	}
}

const testDefaults = "# Theme.\nviewer_theme = \"github\"\n\n# Format.\n# a4 or a5.\npage_format = \"a4\"\n\n# Zoom.\nzoom_default = 1.0\n"

func TestCompleteMissingKeys(t *testing.T) {
	body := "# My own note\nviewer_theme = \"night\"\n"
	got := completeMissingKeys(body, testDefaults)
	for _, want := range []string{"# My own note", "viewer_theme = \"night\"", "# Format.\n# a4 or a5.\npage_format = \"a4\"", "# Zoom.\nzoom_default = 1.0"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Count(got, "viewer_theme") != 1 {
		t.Errorf("a present key was duplicated:\n%s", got)
	}
	if again := completeMissingKeys(got, testDefaults); again != got {
		t.Errorf("a second call must change nothing:\n%s", again)
	}
	if same := completeMissingKeys(testDefaults, testDefaults); same != testDefaults {
		t.Errorf("a complete file must stay as it is:\n%s", same)
	}
}

func TestSaveConfigCompletesAnOldFileOnce(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte("# Mine.\nviewer_theme = \"github\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	SaveConfig(Config{ViewerTheme: "night"}, []byte(testDefaults))
	first, _ := os.ReadFile(p)
	SaveConfig(Config{ViewerTheme: "night"}, []byte(testDefaults))
	second, _ := os.ReadFile(p)
	if string(first) != string(second) {
		t.Errorf("a second save changed the file:\n%s\n---\n%s", first, second)
	}
	if !strings.Contains(string(first), "# Mine.") || !strings.Contains(string(first), "zoom_default = 1.0") {
		t.Errorf("own comment kept and missing key added, got:\n%s", first)
	}
}

// The shipped default must document and set every key Config reads.
func TestShippedConfigHasEveryKey(t *testing.T) {
	data, err := os.ReadFile("../resources/config.toml")
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"viewer_theme", "page_format", "page_orientation", "pdf_margin_vertical", "pdf_margin_horizontal", "zoom_default", "font_size_base", "show_scrollbars", "code_line_numbers", "toolbar_position", "toolbar_visible"} {
		if _, ok := tomlValue(string(data), key); !ok {
			t.Errorf("the shipped config.toml does not set %q", key)
		}
	}
	// Completing an empty file with it must yield a config that loads to the defaults.
	if got := completeMissingKeys("", string(data)); !strings.Contains(got, "zoom_default = 1.0") {
		t.Errorf("completing an empty file lost a key:\n%s", got)
	}
}

func TestSaveConfigMigrationKeepsAnExistingOrientation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".config", "o-mark")
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(dir, "config.toml")
	if err := os.WriteFile(p, []byte("document_max_width = \"297mm\"\npage_orientation = \"portrait\"\n"), 0644); err != nil {
		t.Fatal(err)
	}
	SaveConfig(Config{ViewerTheme: "github"}, nil)
	got, _ := os.ReadFile(p)
	if !strings.Contains(string(got), `page_orientation = "portrait"`) || strings.Contains(string(got), "document_max_width") {
		t.Errorf("the explicit orientation must stay and the old width go:\n%s", got)
	}
}
