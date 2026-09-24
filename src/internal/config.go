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

// cssLengthRe matches a plain CSS length or percentage (e.g. "16px", "80ch",
// "50%") — no arbitrary CSS is allowed through, since these values are
// interpolated directly into generated CSS (see ConfigOverrideCSS).
var cssLengthRe = regexp.MustCompile(`^\d+(\.\d+)?(mm|cm|in|px|pt|pc|%|ch|em|rem|vh|vw|vmin|vmax)$`)

// bareNumberRe matches an integer without a unit (e.g. "16"); it is
// normalized to a px length by cssLength.
var bareNumberRe = regexp.MustCompile(`^\d+$`)

type Config struct {
	ViewerTheme     string `toml:"viewer_theme"`
	FontSizeBase    string `toml:"font_size_base"`
	ShowScrollbars  bool   `toml:"show_scrollbars"`
	ToolbarPosition string `toml:"toolbar_position"`
	ToolbarVisible  *bool  `toml:"toolbar_visible"`
	PageFormat      string `toml:"page_format"`
	PageOrientation string `toml:"page_orientation"`
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
	ViewerTheme string `toml:"viewer_theme"`
	// DocumentMaxWidth is the retired width key, read only to migrate an old
	// config to an orientation; it never reaches Config.
	DocumentMaxWidth string         `toml:"document_max_width"`
	PageFormat       string         `toml:"page_format"`
	PageOrientation  string         `toml:"page_orientation"`
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
		ViewerTheme:     raw.ViewerTheme,
		FontSizeBase:    decodeFontSizeBase(raw.FontSizeBase),
		ShowScrollbars:  raw.ShowScrollbars,
		ToolbarPosition: raw.ToolbarPosition,
		ToolbarVisible:  raw.ToolbarVisible,
		CodeLineNumbers: raw.CodeLineNumbers,
	}
	if cfg.ViewerTheme == "" {
		cfg.ViewerTheme = "github"
	}
	cfg.PageFormat = PageFormatA4
	if raw.PageFormat != "" && raw.PageFormat != PageFormatA4 {
		log.Printf("config: unknown page_format %q, falling back to %s", raw.PageFormat, PageFormatA4)
	}
	switch raw.PageOrientation {
	case OrientationPortrait, OrientationLandscape:
		cfg.PageOrientation = raw.PageOrientation
	case "":
		cfg.PageOrientation = legacyOrientation(raw.DocumentMaxWidth)
	default:
		log.Printf("config: invalid page_orientation %q, falling back to %s", raw.PageOrientation, OrientationPortrait)
		cfg.PageOrientation = OrientationPortrait
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

// assignmentRe matches a top-level `key = value` line, ignoring indentation and
// skipping comments, so setTOMLValue never rewrites a commented-out example.
var assignmentRe = regexp.MustCompile(`^(\s*)([A-Za-z0-9_-]+)(\s*=\s*)`)

// setTOMLValue replaces the value assigned to key, leaving comments, ordering
// and every other line untouched. A key that is absent is appended at the end,
// so no setting is lost to a config the user trimmed down.
func setTOMLValue(body, key, value string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		m := assignmentRe.FindStringSubmatch(line)
		if m != nil && m[2] == key {
			lines[i] = m[1] + key + m[3] + value
			return strings.Join(lines, "\n")
		}
	}
	if body != "" && !strings.HasSuffix(body, "\n") {
		body += "\n"
	}
	return body + key + " = " + value + "\n"
}

// removeTOMLKey drops the assignment of key, along with the comment lines
// directly above it that documented it. A key that is absent leaves body as is.
func removeTOMLKey(body, key string) string {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		m := assignmentRe.FindStringSubmatch(line)
		if m == nil || m[2] != key {
			continue
		}
		start := i
		for start > 0 && strings.HasPrefix(strings.TrimSpace(lines[start-1]), "#") {
			start--
		}
		lines = append(lines[:start], lines[i+1:]...)
		// A blank line left between two blocks is the gap the removed block had.
		if start < len(lines) && strings.TrimSpace(lines[start]) == "" && (start == 0 || strings.TrimSpace(lines[start-1]) == "") {
			lines = append(lines[:start], lines[start+1:]...)
		}
		return strings.Join(lines, "\n")
	}
	return body
}

// tomlBool renders a Go bool the way TOML spells it.
func tomlBool(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

// SaveConfig persists the settings the session can change — the active theme,
// the toolbar visibility and the page orientation — by substituting them in
// place, so the comments that document every option survive the write. It also
// drops the retired document_max_width so no loose width stays beside the format. The whole file is
// encoded from scratch only when there is nothing on disk to edit.
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
	p := filepath.Join(dir, "config.toml")

	body, err := os.ReadFile(p)
	if err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			log.Printf("config: cannot read %s, rewriting it whole: %v", p, err)
		}
		encodeConfig(p, cfg)
		return
	}

	updated := setTOMLValue(string(body), "viewer_theme", strconv.Quote(cfg.ViewerTheme))
	if cfg.ToolbarVisible != nil {
		updated = setTOMLValue(updated, "toolbar_visible", tomlBool(*cfg.ToolbarVisible))
	}
	updated = setTOMLValue(updated, "page_format", strconv.Quote(PageFormatA4))
	orientation := OrientationPortrait
	if cfg.PageOrientation == OrientationLandscape {
		orientation = OrientationLandscape
	}
	updated = setTOMLValue(updated, "page_orientation", strconv.Quote(orientation))
	updated = removeTOMLKey(updated, "document_max_width")
	if err := os.WriteFile(p, []byte(updated), 0644); err != nil {
		log.Printf("config: cannot write: %v", err)
	}
}

func encodeConfig(path string, cfg Config) {
	f, err := os.Create(path)
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
// active theme's own value win; the page sheet is the exception, since the
// sheet is not the theme's to define.
func ConfigOverrideCSS(cfg Config) string {
	// The sheet always wins over the theme: a user-edited theme on disk that
	// still fixes a width or a page height loses to these rules.
	width, height := pageSides(cfg.PageOrientation)
	sb := &strings.Builder{}
	fmt.Fprintf(sb, ":root { --o-mark-page-width: %s; --o-mark-page-height: %s; }\n", width, height)
	props := []string{"max-width: var(--o-mark-page-width)"}
	if f := cssLength(cfg.FontSizeBase); f != "" {
		props = append(props, "font-size: "+f)
	}
	fmt.Fprintf(sb, "body { %s }\n", strings.Join(props, "; "))
	// Screen only: printing paginates by content and breaks between sheets, and a
	// fixed height there would spill each sheet onto a blank page.
	sb.WriteString("@media screen { .o-mark-page { min-height: var(--o-mark-page-height); } }\n")
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
