package internal

import (
	"os"
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
