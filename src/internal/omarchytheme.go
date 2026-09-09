package internal

/*
#cgo pkg-config: omarchy-theme
#include <stdlib.h>
#include "omarchy_theme.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"sync"
	"unsafe"
)

// omarchyRGB mirrors OmarchyColor: one byte per channel.
type omarchyRGB struct {
	R, G, B uint8
}

// omarchyRole mirrors OmarchyRole: a role's own color plus the three content
// weights that read on it.
type omarchyRole struct {
	Color            omarchyRGB
	ContentMain      omarchyRGB
	ContentSecondary omarchyRGB
	ContentDisabled  omarchyRGB
}

// omarchyShade mirrors OmarchyShade: one shade and the content that reads on it.
type omarchyShade struct {
	Color   omarchyRGB
	Content omarchyRGB
}

// omarchyColorFamily mirrors OmarchyColorFamily: a theme color as its
// original/dark/bright triad. Declared is false for orange and brown when
// the theme omits them.
type omarchyColorFamily struct {
	Original omarchyShade
	Dark     omarchyShade
	Bright   omarchyShade
	Declared bool
}

// omarchyRoleCount and omarchyColorCount mirror OMARCHY_ROLE_COUNT and
// OMARCHY_COLOR_COUNT — read from the header's own macros instead of
// hardcoding them, so a library update that grows either array is caught by
// a length mismatch instead of silently truncating.
const (
	omarchyRoleCount  = C.OMARCHY_ROLE_COUNT
	omarchyColorCount = C.OMARCHY_COLOR_COUNT
)

// omarchySnapshot mirrors OmarchySnapshot: every role resolved against the
// theme in force, every color the theme declares, and whether it sits on a
// dark ground.
type omarchySnapshot struct {
	Roles  [omarchyRoleCount]omarchyRole
	Colors [omarchyColorCount]omarchyColorFamily
	Dark   bool
}

func fromCColor(c C.OmarchyColor) omarchyRGB {
	return omarchyRGB{R: uint8(c.r), G: uint8(c.g), B: uint8(c.b)}
}

func fromCShade(c C.OmarchyShade) omarchyShade {
	return omarchyShade{Color: fromCColor(c.color), Content: fromCColor(c.content)}
}

func fromCSnapshot(c *C.OmarchySnapshot) omarchySnapshot {
	var s omarchySnapshot
	for i := range s.Roles {
		r := c.roles[i]
		s.Roles[i] = omarchyRole{
			Color:            fromCColor(r.color),
			ContentMain:      fromCColor(r.content_main),
			ContentSecondary: fromCColor(r.content_secondary),
			ContentDisabled:  fromCColor(r.content_disabled),
		}
	}
	for i := range s.Colors {
		f := c.colors[i]
		s.Colors[i] = omarchyColorFamily{
			Original: fromCShade(f.original),
			Dark:     fromCShade(f.dark),
			Bright:   fromCShade(f.bright),
			Declared: bool(f.declared),
		}
	}
	s.Dark = bool(c.dark)
	return s
}

// omarchyWatcher wraps an OmarchyWatcher handle. Never copy by value — free
// exactly once via close(), never concurrently with another call on the
// same watcher (BRIDGE.md § Thread safety).
type omarchyWatcher struct {
	ptr *C.OmarchyWatcher
}

// newOmarchyWatcher starts watching a theme: path a colors.toml, or "" for
// the system's active theme.
func newOmarchyWatcher(path string) (*omarchyWatcher, error) {
	var cPath *C.char
	if path != "" {
		cPath = C.CString(path)
		defer C.free(unsafe.Pointer(cPath))
	}

	errBuf := make([]C.char, 256)
	ptr := C.omarchy_watch(cPath, &errBuf[0], C.size_t(len(errBuf)))
	if ptr == nil {
		return nil, errors.New(C.GoString(&errBuf[0]))
	}
	return &omarchyWatcher{ptr: ptr}, nil
}

// current writes the theme currently in force, through the given
// representation (an out-of-range index falls back to Original).
func (w *omarchyWatcher) current(representation int) omarchySnapshot {
	var out C.OmarchySnapshot
	C.omarchy_current(w.ptr, C.size_t(representation), &out)
	return fromCSnapshot(&out)
}

// pollChanged reports whether a new theme arrived since the last call,
// through the given representation.
func (w *omarchyWatcher) pollChanged(representation int) (omarchySnapshot, bool) {
	var out C.OmarchySnapshot
	if !bool(C.omarchy_poll_changed(w.ptr, C.size_t(representation), &out)) {
		return omarchySnapshot{}, false
	}
	return fromCSnapshot(&out), true
}

// signalFD returns a file descriptor readable whenever a valid theme change
// landed — integrate it into an event loop (e.g. QSocketNotifier) instead of
// polling on a timer. Reading it never blocks; draining it fully before the
// next wait is the caller's job.
func (w *omarchyWatcher) signalFD() int {
	return int(C.omarchy_signal_fd(w.ptr))
}

// close stops watching and releases the handle. A no-op on an already-closed
// watcher's nil ptr never happens here — callers own exactly one close.
func (w *omarchyWatcher) close() {
	C.omarchy_free(w.ptr)
	w.ptr = nil
}

// roleIndexes caches where each role ThemePalette needs sits in
// OmarchySnapshot.Roles. Resolved by name, not by a hardcoded position:
// BRIDGE.md § Versioning warns the layout — including role order — may
// change between betas without the SONAME moving.
type roleIndexes struct {
	background, elevated, border, primary, selection int
}

var (
	roleIdx     roleIndexes
	roleIdxOnce sync.Once
	roleIdxErr  error
)

func resolvedRoleIndexes() (roleIndexes, error) {
	roleIdxOnce.Do(func() {
		want := map[string]*int{
			"Background": &roleIdx.background,
			"Elevated":   &roleIdx.elevated,
			"Border":     &roleIdx.border,
			"Primary":    &roleIdx.primary,
			"Selection":  &roleIdx.selection,
		}
		for i := 0; i < omarchyRoleCount; i++ {
			name := C.GoString(C.omarchy_role_name(C.size_t(i)))
			if dst, ok := want[name]; ok {
				*dst = i
				delete(want, name)
			}
		}
		if len(want) > 0 {
			missing := make([]string, 0, len(want))
			for name := range want {
				missing = append(missing, name)
			}
			roleIdxErr = fmt.Errorf("omarchytheme: role(s) not found in this library version: %v", missing)
		}
	})
	return roleIdx, roleIdxErr
}

// paletteFromSnapshot maps a snapshot's roles onto ThemePalette. Accent and
// LinkColor share Primary, Surface and CodeBg share Elevated, and
// HeadingColor shares Background's content — the same value reuse
// ThemePalette already had before this library.
func paletteFromSnapshot(s omarchySnapshot, idx roleIndexes) ThemePalette {
	hex := func(c omarchyRGB) string {
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	bg := s.Roles[idx.background]
	el := s.Roles[idx.elevated]
	pr := s.Roles[idx.primary]
	se := s.Roles[idx.selection]
	bo := s.Roles[idx.border]

	return ThemePalette{
		Background:   hex(bg.Color),
		Foreground:   hex(bg.ContentMain),
		Accent:       hex(pr.Color),
		Surface:      hex(el.Color),
		Border:       hex(bo.Color),
		CodeBg:       hex(el.Color),
		CodeFg:       hex(el.ContentMain),
		LinkColor:    hex(pr.Color),
		HeadingColor: hex(bg.ContentMain),
		SelectionBg:  hex(se.Color),
		SelectionFg:  hex(se.ContentMain),
	}
}

// paletteFromOmarchy reads the active theme through libomarchy_theme and
// maps it onto ThemePalette. ok is false when the library could not resolve
// a theme (no watcher) or this library version dropped a role o-mark needs —
// callers fall back to the built-in palette either way.
func paletteFromOmarchy() (palette ThemePalette, ok bool) {
	w, err := NewOmarchyThemeWatcher()
	if err != nil {
		return ThemePalette{}, false
	}
	defer w.Close()
	return w.Palette(), true
}

// OmarchyThemeWatcher follows the active Omarchy theme through
// libomarchy_theme for the life of the process, resolving it directly to
// ThemePalette instead of the raw snapshot — callers never touch role
// indexes or C types.
type OmarchyThemeWatcher struct {
	w   *omarchyWatcher
	idx roleIndexes
}

// NewOmarchyThemeWatcher starts following the system's active theme. An
// error means the library found no active theme to watch, or this library
// version dropped a role o-mark needs — the caller decides the fallback.
func NewOmarchyThemeWatcher() (*OmarchyThemeWatcher, error) {
	idx, err := resolvedRoleIndexes()
	if err != nil {
		return nil, err
	}
	w, err := newOmarchyWatcher("")
	if err != nil {
		return nil, err
	}
	return &OmarchyThemeWatcher{w: w, idx: idx}, nil
}

// Palette reads the theme currently in force.
func (t *OmarchyThemeWatcher) Palette() ThemePalette {
	return paletteFromSnapshot(t.w.current(0), t.idx)
}

// PollChanged reports whether a new theme arrived since the last call
// (either to PollChanged or Palette).
func (t *OmarchyThemeWatcher) PollChanged() (ThemePalette, bool) {
	s, changed := t.w.pollChanged(0)
	if !changed {
		return ThemePalette{}, false
	}
	return paletteFromSnapshot(s, t.idx), true
}

// SignalFD returns a file descriptor readable whenever a valid theme change
// landed — integrate it into an event loop (e.g. QSocketNotifier) instead of
// polling on a timer.
func (t *OmarchyThemeWatcher) SignalFD() int {
	return t.w.signalFD()
}

// Close stops watching and releases the underlying handle.
func (t *OmarchyThemeWatcher) Close() {
	t.w.close()
}
