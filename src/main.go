package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"

	"o-mark/internal"

	qt "github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/qml"
)

//go:embed ui/*.qml
var uiFiles embed.FS

//go:embed resources/config.toml
var defaultConfigTOML []byte

var version = "0.5.3"

func paletteJSON(p internal.ThemePalette) string {
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
		log.Printf("main: cannot marshal palette: %v", err)
	}
	return string(b)
}

// extractUI extracts the embedded QML files to a temp directory and returns
// its path. The caller must defer os.RemoveAll on the returned path.
func extractUI() (string, error) {
	dir, err := os.MkdirTemp("", "o-mark-*")
	if err != nil {
		return "", err
	}
	entries, err := fs.ReadDir(uiFiles, "ui")
	if err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	for _, e := range entries {
		data, err := uiFiles.ReadFile("ui/" + e.Name())
		if err != nil {
			os.RemoveAll(dir)
			return "", err
		}
		if err := os.WriteFile(filepath.Join(dir, e.Name()), data, 0644); err != nil {
			os.RemoveAll(dir)
			return "", err
		}
	}
	return dir, nil
}

// allThemes returns the full ordered theme list: system first, then disk themes.
func allThemes(diskThemes []internal.ViewerTheme) []internal.ViewerTheme {
	all := make([]internal.ViewerTheme, 0, 1+len(diskThemes))
	all = append(all, internal.ViewerTheme{ID: "system", Label: "System"})
	return append(all, diskThemes...)
}

// themeIndex returns the index of themeID in all, defaulting to 0.
func themeIndex(all []internal.ViewerTheme, themeID string) int {
	for i, t := range all {
		if t.ID == themeID {
			return i
		}
	}
	return 0
}

// setContextProperties passes all Go state to QML via context properties.
func setContextProperties(ctx *qml.QQmlContext, title, raw, docDir string, p internal.ThemePalette, diskThemes []internal.ViewerTheme, themesDir string, cfg internal.Config, initialIdx int, font string) {
	ctx.SetContextProperty2("documentTitle", qt.NewQVariant14(title))
	ctx.SetContextProperty2("docDir", qt.NewQVariant14(docDir))
	ctx.SetContextProperty2("viewerThemesJson", qt.NewQVariant14(internal.RenderViewerThemesJSON(raw, p, docDir, diskThemes, themesDir, cfg, font)))
	ctx.SetContextProperty2("viewerThemeLabelsJson", qt.NewQVariant14(internal.ViewerThemeLabelsJSON(allThemes(diskThemes))))
	ctx.SetContextProperty2("omarchyPaletteJson", qt.NewQVariant14(paletteJSON(p)))
	ctx.SetContextProperty2("omarchyFont", qt.NewQVariant14(font))
	ctx.SetContextProperty2("initialViewerThemeIndex", qt.NewQVariant4(initialIdx))
	ctx.SetContextProperty2("showScrollbars", qt.NewQVariant8(cfg.ShowScrollbars))
	ctx.SetContextProperty2("toolbarPositionConfig", qt.NewQVariant14(cfg.ToolbarPosition))
	ctx.SetContextProperty2("toolbarVisibleConfig", qt.NewQVariant8(*cfg.ToolbarVisible))
	postLoad, err := json.Marshal(internal.PostLoadScripts(raw))
	if err != nil {
		log.Printf("main: cannot marshal post-load scripts: %v", err)
	}
	ctx.SetContextProperty2("postLoadScripts", qt.NewQVariant14(string(postLoad)))
}

// startThemeWatcher watches the Omarchy theme files and updates context properties
// on change. cfg is a pointer so re-renders always use the latest config values.
// A singleShot(0) fires on the first event loop tick to self-correct a failed
// startup palette read without waiting for a file event.
func startThemeWatcher(ctx *qml.QQmlContext, rawPtr *string, docDir string, p internal.ThemePalette, diskThemes []internal.ViewerTheme, themesDir string, cfg *internal.Config, font string) {
	currentP := p
	currentFont := font

	update := func() {
		newP := internal.GetThemePalette("omarchy")
		newFont := internal.ReadOmarchyFont()
		if newP == currentP && newFont == currentFont {
			return
		}
		currentP = newP
		currentFont = newFont
		ctx.SetContextProperty2("omarchyPaletteJson", qt.NewQVariant14(paletteJSON(newP)))
		ctx.SetContextProperty2("omarchyFont", qt.NewQVariant14(newFont))
		ctx.SetContextProperty2("viewerThemesJson", qt.NewQVariant14(internal.RenderViewerThemesJSON(*rawPtr, newP, docDir, diskThemes, themesDir, *cfg, newFont)))
	}

	initShot := qt.NewQTimer()
	initShot.SetSingleShot(true)
	initShot.SetInterval(0)
	initShot.OnTimeout(update)
	initShot.Start2()

	if _, err := os.UserHomeDir(); err != nil {
		log.Printf("warning: cannot determine home directory, theme file watcher disabled: %v", err)
		return
	}

	watcher := qt.NewQFileSystemWatcher()
	watcher.AddPath(internal.OmarchyStatePath("theme", "colors.toml"))
	watcher.AddPath(internal.OmarchyStatePath("theme.name"))
	watcher.OnFileChanged(func(path string) {
		// Re-add: editors that atomically replace files (rename-over) change the
		// inode, causing inotify to drop the watch after the first event.
		watcher.AddPath(path)
		update()
	})
}

// startConfigWatcher watches config.toml and the known theme CSS files.
// On any change it reloads the config (preserving the session's viewer_theme),
// then re-renders all themes so CSS changes take effect immediately.
func startConfigWatcher(ctx *qml.QQmlContext, rawPtr *string, docDir string, diskThemes []internal.ViewerTheme, themesDir string, cfg *internal.Config, font string) {
	watcher := qt.NewQFileSystemWatcher()
	watcher.AddPath(filepath.Join(internal.ConfigDir(), "config.toml"))
	for _, t := range diskThemes {
		cssPath := filepath.Join(themesDir, t.ID+".css")
		if _, err := os.Stat(cssPath); err == nil {
			watcher.AddPath(cssPath)
		}
	}

	watcher.OnFileChanged(func(path string) {
		// Re-add after inotify drops the watch on atomic-replace saves.
		watcher.AddPath(path)
		newCfg := internal.LoadConfig()
		newCfg.ViewerTheme = cfg.ViewerTheme // session theme is not overridden by config edits
		*cfg = newCfg
		p := internal.GetThemePalette("omarchy")
		ctx.SetContextProperty2("viewerThemesJson", qt.NewQVariant14(
			internal.RenderViewerThemesJSON(*rawPtr, p, docDir, diskThemes, themesDir, *cfg, font),
		))
		ctx.SetContextProperty2("showScrollbars", qt.NewQVariant8(newCfg.ShowScrollbars))
		ctx.SetContextProperty2("toolbarPositionConfig", qt.NewQVariant14(newCfg.ToolbarPosition))
		ctx.SetContextProperty2("toolbarVisibleConfig", qt.NewQVariant8(*newCfg.ToolbarVisible))
	})
}

// startDocWatcher watches the open document for changes and re-renders on each save.
// rawPtr is updated in-place so subsequent theme/config watcher fires use the new content.
func startDocWatcher(ctx *qml.QQmlContext, absPath, docDir string, diskThemes []internal.ViewerTheme, themesDir string, cfg *internal.Config, rawPtr *string, reloadCount *int) {
	reload := func() {
		newRaw, err := internal.LoadFile(absPath)
		if err != nil {
			log.Printf("doc watcher: cannot re-read %s: %v", absPath, err)
			return
		}
		*rawPtr = newRaw
		p := internal.GetThemePalette("omarchy")
		font := internal.ReadOmarchyFont()
		ctx.SetContextProperty2("viewerThemesJson", qt.NewQVariant14(
			internal.RenderViewerThemesJSON(*rawPtr, p, docDir, diskThemes, themesDir, *cfg, font),
		))
		postLoad, err := json.Marshal(internal.PostLoadScripts(*rawPtr))
		if err != nil {
			log.Printf("doc watcher: cannot marshal post-load scripts: %v", err)
		}
		ctx.SetContextProperty2("postLoadScripts", qt.NewQVariant14(string(postLoad)))
		*reloadCount++
		ctx.SetContextProperty2("docReloadSignal", qt.NewQVariant4(*reloadCount))
	}

	// 100 ms delay handles atomic saves (delete+rename pattern) and debounces
	// rapid consecutive writes.
	debounce := qt.NewQTimer()
	debounce.SetSingleShot(true)
	debounce.SetInterval(100)
	debounce.OnTimeout(reload)

	watcher := qt.NewQFileSystemWatcher()
	watcher.AddPath(absPath)
	watcher.OnFileChanged(func(path string) {
		watcher.AddPath(path)
		debounce.Start2()
	})
}

// extractFontFamily parses the first font-family declaration from a CSS string.
// Used for font verification at startup.
func extractFontFamily(css string) string {
	for _, line := range strings.Split(css, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "font-family:") {
			val := strings.TrimSuffix(strings.TrimPrefix(line, "font-family:"), ";")
			return strings.TrimSpace(val)
		}
	}
	return ""
}

func main() {
	if len(os.Args) == 2 && os.Args[1] == "--version" {
		fmt.Println("o-mark", version)
		return
	}
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: o-mark <file.md>")
		os.Exit(1)
	}

	absPath, err := filepath.Abs(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	raw, err := internal.LoadFile(absPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	docDir := filepath.Dir(absPath)

	displayPath := os.Args[1]

	internal.EnsureConfig(defaultConfigTOML)
	cfg := internal.LoadConfig()
	themesDir := internal.ThemesDir()
	internal.CleanupLegacyThemes(themesDir)
	internal.ExportBuiltinThemes(themesDir)
	diskThemes := internal.DiskThemes(themesDir)

	// Verify fonts for each static theme at startup.
	for _, t := range diskThemes {
		css := internal.DiskThemeCSS(t.ID, themesDir)
		if font := extractFontFamily(css); font != "" {
			internal.VerifyFontStack(font, t.Label)
		}
	}

	all := allThemes(diskThemes)
	initialIdx := themeIndex(all, cfg.ViewerTheme)
	omarchyFont := internal.ReadOmarchyFont()

	qmlDir, err := extractUI()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error extracting UI:", err)
		os.Exit(1)
	}
	defer os.RemoveAll(qmlDir)

	os.Setenv("QT_QUICK_CONTROLS_STYLE", "Basic")
	os.Setenv("QT_LOGGING_RULES", "qt.qpa.wayland=false")
	// Software compositing avoids GPU surface invalidation when a Wayland
	// compositor unmaps/remaps the window (e.g. workspace switch in Hyprland),
	// which otherwise leaves the WebEngineView black on return.
	if os.Getenv("QTWEBENGINE_CHROMIUM_FLAGS") == "" {
		os.Setenv("QTWEBENGINE_CHROMIUM_FLAGS", "--disable-gpu-compositing")
	}
	qt.NewQApplication(os.Args)

	engine := qml.NewQQmlApplicationEngine()
	ctx := engine.RootContext()
	omarchyP := internal.GetThemePalette("omarchy")
	setContextProperties(ctx, displayPath, raw, docDir, omarchyP, diskThemes, themesDir, cfg, initialIdx, omarchyFont)
	ctx.SetContextProperty2("docReloadSignal", qt.NewQVariant4(0))
	engine.Load(qt.QUrl_FromLocalFile(filepath.Join(qmlDir, "main.qml")))
	reloadCount := 0
	startThemeWatcher(ctx, &raw, docDir, omarchyP, diskThemes, themesDir, &cfg, omarchyFont)
	startConfigWatcher(ctx, &raw, docDir, diskThemes, themesDir, &cfg, omarchyFont)
	startDocWatcher(ctx, absPath, docDir, diskThemes, themesDir, &cfg, &raw, &reloadCount)

	qt.QApplication_Exec()

	// Save the active theme to config on exit.
	if roots := engine.RootObjects(); len(roots) > 0 {
		idx := roots[0].Property("viewerThemeIndex").ToInt()
		if idx >= 0 && idx < len(all) {
			cfg.ViewerTheme = all[idx].ID
			v := roots[0].Property("toolbarVisible").ToBool()
			cfg.ToolbarVisible = &v
		}
	}
	internal.SaveConfig(cfg)
}
