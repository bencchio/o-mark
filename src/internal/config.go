package internal

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

// DeferToTheme is the config sentinel that means "take the value defined by the
// active theme" instead of overriding it (see ConfigOverrideCSS).
const DeferToTheme = "default"

// cssLengthRe matches a plain CSS length or percentage (e.g. "210mm", "80ch",
// "50%", "16px") — no arbitrary CSS is allowed through, since these values are
// interpolated directly into generated CSS (see ConfigOverrideCSS).
var cssLengthRe = regexp.MustCompile(`^\d+(\.\d+)?(mm|cm|in|px|pt|pc|%|ch|em|rem|vh|vw|vmin|vmax)$`)

// bareNumberRe matches an integer without a unit (e.g. "16"); it is
// normalized to a px length by cssLength.
var bareNumberRe = regexp.MustCompile(`^\d+$`)

// maxWidthRe is kept for backward compatibility and aliases cssLengthRe.
var maxWidthRe = cssLengthRe

type Config struct {
	ViewerTheme      string `toml:"viewer_theme"`
	DocumentMaxWidth string `toml:"document_max_width"`
	FontSizeBase     string `toml:"font_size_base"`
	ShowScrollbars   bool   `toml:"show_scrollbars"`
	ToolbarPosition  string `toml:"toolbar_position"`
	ToolbarVisible   *bool  `toml:"toolbar_visible"`
	// CodeLineNumbers is a pointer because the default is on: an absent key
	// must mean "enabled", which a plain bool cannot express.
	CodeLineNumbers *bool `toml:"code_line_numbers"`
}

// isDeferredOrLength reports whether v is either the deferral sentinel, a
// valid CSS length, or a bare number (normalized to px later). Empty values
// are treated as deferred (fall through).
func isDeferredOrLength(v string) bool {
	return v == "" || v == DeferToTheme || cssLengthRe.MatchString(v) || bareNumberRe.MatchString(v)
}

func homeDir() string {
	h, err := os.UserHomeDir()
	if err != nil {
		log.Printf("warning: cannot determine home directory: %v", err)
		return ""
	}
	return h
}

func ConfigDir() string {
	return filepath.Join(homeDir(), ".config", "o-mark")
}

func ThemesDir() string {
	return filepath.Join(ConfigDir(), "themes")
}

// EnsureConfig creates the default config at the XDG config path on first run.
// Uses O_EXCL so a file created between a hypothetical stat and write is never overwritten.
func EnsureConfig(defaultTOML []byte) {
	dir := ConfigDir()
	if !filepath.IsAbs(dir) {
		return
	}
	p := filepath.Join(dir, "config.toml")
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("config: cannot create dir: %v", err)
		return
	}
	f, err := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return
		}
		log.Printf("config: cannot write default config: %v", err)
		return
	}
	defer f.Close()
	if _, err := f.Write(defaultTOML); err != nil {
		log.Printf("config: cannot write default config: %v", err)
	}
}

// configRaw mirrors Config but keeps FontSizeBase as a raw toml.Primitive so a
// legacy integer value (font_size_base = 14) doesn't abort the whole decode.
type configRaw struct {
	ViewerTheme      string         `toml:"viewer_theme"`
	DocumentMaxWidth string         `toml:"document_max_width"`
	FontSizeBase     toml.Primitive `toml:"font_size_base"`
	ShowScrollbars   bool           `toml:"show_scrollbars"`
	ToolbarPosition  string         `toml:"toolbar_position"`
	ToolbarVisible   *bool          `toml:"toolbar_visible"`
	CodeLineNumbers  *bool          `toml:"code_line_numbers"`
}

// decodeFontSizeBase resolves the raw FontSizeBase value, accepting both the
// new string form and the legacy integer form. It returns "" if absent.
func decodeFontSizeBase(p toml.Primitive) string {
	var s string
	if err := toml.PrimitiveDecode(p, &s); err == nil {
		return s
	}
	var i int
	if err := toml.PrimitiveDecode(p, &i); err == nil {
		return strconv.Itoa(i)
	}
	return ""
}

// LoadConfig reads ~/.config/o-mark/config.toml. Call EnsureConfig first so
// the file always exists. Falls back to sane values for missing or invalid fields.
func LoadConfig() Config {
	var raw configRaw
	p := filepath.Join(ConfigDir(), "config.toml")
	if _, err := toml.DecodeFile(p, &raw); err != nil {
		log.Printf("config: cannot read %s: %v", p, err)
	}
	cfg := Config{
		ViewerTheme:      raw.ViewerTheme,
		DocumentMaxWidth: raw.DocumentMaxWidth,
		FontSizeBase:     decodeFontSizeBase(raw.FontSizeBase),
		ShowScrollbars:   raw.ShowScrollbars,
		ToolbarPosition:  raw.ToolbarPosition,
		ToolbarVisible:   raw.ToolbarVisible,
		CodeLineNumbers:  raw.CodeLineNumbers,
	}
	if cfg.ViewerTheme == "" {
		cfg.ViewerTheme = "github"
	}
	if !isDeferredOrLength(cfg.DocumentMaxWidth) {
		log.Printf("config: invalid document_max_width %q, falling back to 210mm", cfg.DocumentMaxWidth)
		cfg.DocumentMaxWidth = "210mm"
	}
	if cfg.FontSizeBase == "" {
		// Absent or non-numeric (e.g. a broken value) — defer to the theme.
		cfg.FontSizeBase = DeferToTheme
	} else if !isDeferredOrLength(cfg.FontSizeBase) {
		log.Printf("config: invalid font_size_base %q, falling back to theme default", cfg.FontSizeBase)
		cfg.FontSizeBase = DeferToTheme
	}
	if cfg.ToolbarPosition != "top" && cfg.ToolbarPosition != "bottom" {
		cfg.ToolbarPosition = "bottom"
	}
	if cfg.ToolbarVisible == nil {
		v := false
		cfg.ToolbarVisible = &v
	}
	if cfg.CodeLineNumbers == nil {
		v := true
		cfg.CodeLineNumbers = &v
	}
	return cfg
}

func SaveConfig(cfg Config) {
	dir := ConfigDir()
	if !filepath.IsAbs(dir) {
		log.Printf("config: home directory unavailable, cannot save config")
		return
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("config: cannot create dir: %v", err)
		return
	}
	f, err := os.Create(filepath.Join(dir, "config.toml"))
	if err != nil {
		log.Printf("config: cannot write: %v", err)
		return
	}
	defer f.Close()
	if err := toml.NewEncoder(f).Encode(cfg); err != nil {
		log.Printf("config: encode error: %v", err)
	}
}

// cssLength normalizes a configured value to a CSS length: bare integers get a
// px unit, "default" and invalid values return "" (meaning "don't emit").
func cssLength(v string) string {
	if v == "" || v == DeferToTheme {
		return ""
	}
	if cssLengthRe.MatchString(v) {
		return v
	}
	if n, err := strconv.Atoi(v); err == nil {
		return fmt.Sprintf("%dpx", n)
	}
	return ""
}

// ConfigOverrideCSS returns a CSS snippet that applies user config values on
// top of any theme's base CSS. Appended last so it wins the cascade. A field
// set to the deferral sentinel emits nothing for that property, letting the
// active theme's own value win.
func ConfigOverrideCSS(cfg Config) string {
	props := []string{}
	if w := cssLength(cfg.DocumentMaxWidth); w != "" {
		props = append(props, "max-width: "+w)
	}
	if f := cssLength(cfg.FontSizeBase); f != "" {
		props = append(props, "font-size: "+f)
	}
	sb := &strings.Builder{}
	if len(props) > 0 {
		fmt.Fprintf(sb, "body { %s }\n", strings.Join(props, "; "))
	}
	if cfg.ShowScrollbars {
		sb.WriteString("::-webkit-scrollbar { display: block; }\n")
	}
	// Line numbers are on unless turned off explicitly. The selector is more
	// specific than the gutter rule in hljsCSS, so concatenation order between
	// the two stylesheets does not matter.
	if cfg.CodeLineNumbers != nil && !*cfg.CodeLineNumbers {
		sb.WriteString("pre .hljs-ln-numbers { display: none; }\n")
	}
	return sb.String()
}
