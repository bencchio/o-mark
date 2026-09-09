package internal

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ThemePalette struct {
	Background   string
	Foreground   string
	Accent       string
	Surface      string
	Border       string
	CodeBg       string
	CodeFg       string
	LinkColor    string
	HeadingColor string
	SelectionBg  string
	SelectionFg  string
}

func builtinPalette(dark bool) ThemePalette {
	if dark {
		return ThemePalette{
			Background:   "#0d1117",
			Foreground:   "#c9d1d9",
			Accent:       "#58a6ff",
			Surface:      "#161b22",
			Border:       "#30363d",
			CodeBg:       "#161b22",
			CodeFg:       "#c9d1d9",
			LinkColor:    "#58a6ff",
			HeadingColor: "#c9d1d9",
			SelectionBg:  "#58a6ff",
			SelectionFg:  "#0d1117",
		}
	}
	return ThemePalette{
		Background:   "#ffffff",
		Foreground:   "#24292e",
		Accent:       "#0366d6",
		Surface:      "#f6f8fa",
		Border:       "#dfe2e5",
		CodeBg:       "#f6f8fa",
		CodeFg:       "#24292e",
		LinkColor:    "#0366d6",
		HeadingColor: "#24292e",
		SelectionBg:  "#0366d6",
		SelectionFg:  "#ffffff",
	}
}

// OmarchyStateEnv overrides the directory holding the active Omarchy theme
// state, for an install that keeps it somewhere other than the default.
const OmarchyStateEnv = "O_MARK_OMARCHY_STATE"

// OmarchyStatePath joins parts onto the directory where Omarchy 4.0 keeps the
// active theme state. Omarchy 3.x kept it under ~/.config/omarchy/current, and
// an upgraded system still resolves that path through a compatibility symlink,
// but a clean 4.0 install does not create one — so this reads the canonical
// location directly rather than depending on the symlink.
func OmarchyStatePath(parts ...string) string {
	return filepath.Join(append([]string{omarchyStateDir()}, parts...)...)
}

func omarchyStateDir() string {
	if dir := os.Getenv(OmarchyStateEnv); dir != "" {
		return dir
	}
	return filepath.Join(homeDir(), ".local/state/omarchy/current")
}

func ReadOmarchyFont() string {
	cssFile := OmarchyStatePath("theme", "hyprland-preview-share-picker.css")
	data, err := os.ReadFile(cssFile)
	if err != nil {
		log.Printf("warning: cannot read %s (%v); falling back to JetBrains Mono NF", cssFile, err)
		return "JetBrains Mono NF"
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "font-family:") {
			font := strings.TrimSpace(strings.TrimPrefix(line, "font-family:"))
			font = strings.TrimSuffix(font, ";")
			font = strings.Trim(font, `"'`)
			if font != "" {
				return font
			}
		}
	}
	log.Printf("warning: no font-family found in %s; falling back to JetBrains Mono NF", cssFile)
	return "JetBrains Mono NF"
}

// parseHex splits a #RRGGBB string into its R, G, B components.
func parseHex(s string) (int64, int64, int64) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return 0, 0, 0
	}
	r, errR := strconv.ParseInt(s[0:2], 16, 64)
	g, errG := strconv.ParseInt(s[2:4], 16, 64)
	b, errB := strconv.ParseInt(s[4:6], 16, 64)
	if errR != nil || errG != nil || errB != nil {
		log.Printf("theme: invalid hex color %q", s)
	}
	return r, g, b
}

// luminance returns relative perceived brightness of a #RRGGBB color.
func luminance(hex string) float64 {
	r, g, b := parseHex(hex)
	return 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
}

// blendHex returns a color that is t*a + (1-t)*b, where a and b are #RRGGBB strings.
// Used to derive visible mid-tone colors from palette extremes.
func blendHex(a, b string, t float64) string {
	r1, g1, b1 := parseHex(a)
	r2, g2, b2 := parseHex(b)
	blend := func(c1, c2 int64) int { return int(float64(c1)*t + float64(c2)*(1-t)) }
	return fmt.Sprintf("#%02x%02x%02x", blend(r1, r2), blend(g1, g2), blend(b1, b2))
}

// normalizedFontName strips spaces and folds case, so "JetBrains Mono" lines
// up with fontconfig's own "JetBrainsMono" family name for a patched (e.g.
// Nerd Font) variant — fontconfig treats them as the same family for
// substitution purposes, but a literal string compare would not.
func normalizedFontName(name string) string {
	return strings.ToLower(strings.ReplaceAll(name, " ", ""))
}

// VerifyFontStack checks if at least one font in the comma-separated stack is
// available on the system. Logs a warning if none are found. Non-blocking.
func VerifyFontStack(stack, themeLabel string) {
	for _, name := range strings.Split(stack, ",") {
		name = strings.Trim(strings.TrimSpace(name), `'"`)
		if name == "serif" || name == "sans-serif" || name == "monospace" {
			continue // generics always available
		}
		if name == "-apple-system" || name == "Segoe UI" || name == "Helvetica" || name == "Arial" {
			return // known safe system font, stack is OK
		}
		cmd := exec.Command("fc-match", "-f", "%{family}", "--", name)
		out, err := cmd.Output()
		if err != nil {
			continue
		}
		// fc-match can return several comma-separated aliases for the matched
		// family; any one of them starting with the requested name (modulo
		// spacing/case) counts as a real match, not just a generic fallback.
		want := normalizedFontName(name)
		for _, family := range strings.Split(string(out), ",") {
			if strings.HasPrefix(normalizedFontName(family), want) {
				return
			}
		}
	}
	log.Printf("warning: none of [%s] found on system; theme %s will use fallback", stack, themeLabel)
}

func GetThemePalette(mode string) ThemePalette {
	switch mode {
	case "dark":
		return builtinPalette(true)
	case "omarchy":
		if p, ok := paletteFromOmarchy(); ok {
			return p
		}
		return builtinPalette(false)
	default:
		return builtinPalette(false)
	}
}
