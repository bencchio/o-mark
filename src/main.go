package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/url"
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

var version = "0.7.4"

// qtSink writes the session's output to QML context properties. It is the only
// Qt-aware part of publication; the order the properties must land in lives in
// internal.State.Publish.
type qtSink struct{ ctx *qml.QQmlContext }

func (q qtSink) SetString(key, value string) {
	q.ctx.SetContextProperty2(key, qt.NewQVariant14(value))
}
func (q qtSink) SetInt(key string, value int) {
	q.ctx.SetContextProperty2(key, qt.NewQVariant4(value))
}
func (q qtSink) SetBool(key string, value bool) {
	q.ctx.SetContextProperty2(key, qt.NewQVariant8(value))
}

// extractUI extracts the embedded QML files to a temp directory and returns
// its path. The caller must defer os.RemoveAll on the returned path.
func extractUI() (string, error) {
	dir, err := os.MkdirTemp("", "o-mark-*")
	if err != nil {
		return "", err
	}
	// The RemoveAll results below are deliberately unchecked: they clean up a
	// temp dir on a path that is already returning an error, and reporting a
	// second failure would only bury the one that matters.
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

// watchPath adds path to watcher, reporting the failure instead of leaving a
// watcher that silently never fires. A dropped watch means live reloading stops
// working with nothing on screen to say so.
func watchPath(watcher *qt.QFileSystemWatcher, path string) {
	if !watcher.AddPath(path) {
		log.Printf("warning: cannot watch %s; changes to it will not be picked up until restart", path)
	}
}

// wirePdfExport connects the QML property map's `prepare` handshake to the
// session's print render, so it reuses the live document and palette.
func wirePdfExport(pm *qml.QQmlPropertyMap, sess *internal.Session) {
	pm.Insert("path", qt.NewQVariant14(""))
	pm.Insert("html", qt.NewQVariant14(""))
	pm.Insert("exists", qt.NewQVariant8(false))
	pm.Insert("orientation", qt.NewQVariant14(internal.OrientationPortrait))
	pm.OnValueChanged(func(key string, value *qt.QVariant) {
		if key != "prepare" {
			return
		}
		prep := sess.PreparePdf()
		pm.Insert("path", qt.NewQVariant14(prep.Path))
		pm.Insert("html", qt.NewQVariant14(prep.HTML))
		pm.Insert("exists", qt.NewQVariant8(prep.Exists))
		pm.Insert("orientation", qt.NewQVariant14(prep.Orientation))
	})
}

// wirePageControl connects the QML property map's `toggle` handshake to the
// session: each flip of the sheet orientation re-renders and republishes.
func wirePageControl(pm *qml.QQmlPropertyMap, sess *internal.Session, sink internal.StateSink) {
	pm.OnValueChanged(func(key string, value *qt.QVariant) {
		if key != "toggle" {
			return
		}
		sess.Apply(internal.OrientationToggled{}).Publish(sink)
	})
}

// startThemeWatcher follows the Omarchy theme (color, via libomarchy-lib-theme)
// and font (file, via inotify — the library is colors-only) and asks the
// session to refresh. A singleShot(0) fires on the first event loop tick to
// self-correct a failed startup palette read without waiting for a theme or
// file event.
func startThemeWatcher(sess *internal.Session, sink internal.StateSink, omarchyWatcher *internal.OmarchyThemeWatcher) {
	update := func() {
		sess.Apply(internal.ThemeChanged{}).Publish(sink)
	}

	initShot := qt.NewQTimer()
	initShot.SetSingleShot(true)
	initShot.SetInterval(0)
	initShot.OnTimeout(update)
	initShot.Start2()

	if omarchyWatcher != nil {
		notifier := qt.NewQSocketNotifier2(uintptr(omarchyWatcher.SignalFD()), qt.QSocketNotifier__Read)
		notifier.OnActivated(func(socket qt.QSocketDescriptor, activationEvent qt.QSocketNotifier__Type) {
			// The poll itself decides whether a valid change landed; its result is
			// discarded in favor of update()'s own read so both watchers publish
			// through the exact same path.
			if _, changed := omarchyWatcher.PollChanged(); changed {
				update()
			}
		})
	}

	if _, err := os.UserHomeDir(); err != nil {
		log.Printf("warning: cannot determine home directory, font file watcher disabled: %v", err)
		return
	}

	fontWatcher := qt.NewQFileSystemWatcher()
	watchPath(fontWatcher, internal.OmarchyStatePath("theme", "hyprland-preview-share-picker.css"))
	fontWatcher.OnFileChanged(func(path string) {
		// Re-add: editors that atomically replace files (rename-over) change the
		// inode, causing inotify to drop the watch after the first event.
		watchPath(fontWatcher, path)
		update()
	})
}

// startConfigWatcher watches config.toml and the known theme CSS files and
// asks the session to reload. The session preserves the active theme, so a
// config edit does not change what the user is reading.
func startConfigWatcher(sess *internal.Session, sink internal.StateSink, diskThemes []internal.ViewerTheme, themesDir string) {
	watcher := qt.NewQFileSystemWatcher()
	watchPath(watcher, filepath.Join(internal.ConfigDir(), "config.toml"))
	for _, t := range diskThemes {
		cssPath := filepath.Join(themesDir, t.ID+".css")
		if _, err := os.Stat(cssPath); err == nil {
			watchPath(watcher, cssPath)
		}
	}

	watcher.OnFileChanged(func(path string) {
		// Re-add after inotify drops the watch on atomic-replace saves.
		watchPath(watcher, path)
		sess.Apply(internal.ConfigChanged{}).Publish(sink)
	})
}

// startDocWatcher watches the open document and asks the session to reload it.
// The 100 ms delay handles atomic saves (delete+rename pattern) and debounces
// rapid consecutive writes.
func startDocWatcher(sess *internal.Session, sink internal.StateSink, absPath string) {
	debounce := qt.NewQTimer()
	debounce.SetSingleShot(true)
	debounce.SetInterval(100)
	debounce.OnTimeout(func() {
		sess.Apply(internal.DocumentSaved{}).Publish(sink)
	})

	watcher := qt.NewQFileSystemWatcher()
	watchPath(watcher, absPath)
	watcher.OnFileChanged(func(path string) {
		watchPath(watcher, path)
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

// resolveDocArg accepts either a bare local path (direct CLI use, e.g.
// "o-mark file.md") or a URI: "file://..." from a file manager, or the
// "o-mark:" scheme cross-document links use to reach a fresh instance
// through the desktop's %u handoff. Only URI-prefixed input goes through
// url.Parse, so a bare filename containing a literal '#' is never
// misread as carrying a fragment.
func resolveDocArg(arg string) (path, anchor string) {
	if strings.HasPrefix(arg, "file://") || strings.HasPrefix(arg, "o-mark:") {
		if u, err := url.Parse(arg); err == nil {
			p := u.Path
			if p == "" {
				p = u.Opaque
			}
			if decoded, derr := url.PathUnescape(p); derr == nil {
				p = decoded
			}
			return p, u.Fragment
		}
	}
	return arg, ""
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

	docArg, initialAnchor := resolveDocArg(os.Args[1])

	absPath, err := filepath.Abs(docArg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	// Fail before the GUI starts when the document cannot be read; the session
	// reads it again once Qt is up.
	if _, err := internal.LoadFile(absPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	docDir := filepath.Dir(absPath)

	displayPath := docArg

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

	var omarchyWatcher *internal.OmarchyThemeWatcher
	var paletteSrc internal.PaletteSource
	if w, err := internal.NewOmarchyThemeWatcher(); err != nil {
		log.Printf("warning: cannot watch the Omarchy theme, colors will not update live: %v", err)
	} else {
		omarchyWatcher = w
		paletteSrc = w // a typed nil would defeat the session's fallback check
	}

	sess, state, err := internal.StartSession(internal.SessionOptions{
		AbsPath:    absPath,
		DocDir:     docDir,
		Title:      displayPath,
		Anchor:     initialAnchor,
		Config:     cfg,
		ThemesDir:  themesDir,
		DiskThemes: diskThemes,
		Palette:    paletteSrc,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sink := qtSink{ctx}
	state.Publish(sink)

	pdfExport := qml.NewQQmlPropertyMap()
	wirePdfExport(pdfExport, sess)
	ctx.SetContextProperty("pdfExport", pdfExport.QObject)
	pageControl := qml.NewQQmlPropertyMap()
	wirePageControl(pageControl, sess, sink)
	ctx.SetContextProperty("pageControl", pageControl.QObject)
	engine.Load(qt.QUrl_FromLocalFile(filepath.Join(qmlDir, "main.qml")))
	startThemeWatcher(sess, sink, omarchyWatcher)
	startConfigWatcher(sess, sink, diskThemes, themesDir)
	startDocWatcher(sess, sink, absPath)

	qt.QApplication_Exec()

	// Save the active theme, toolbar visibility and page orientation to config on exit.
	if roots := engine.RootObjects(); len(roots) > 0 {
		sess.Apply(internal.Persist{
			ViewerThemeIndex: roots[0].Property("viewerThemeIndex").ToInt(),
			ToolbarVisible:   roots[0].Property("toolbarVisible").ToBool(),
		})
	}
}
