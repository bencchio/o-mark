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
