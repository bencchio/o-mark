package internal

import (
	"embed"
	"log"
	"sync"
)

//go:embed nav/nav.css nav/nav.js
var navFS embed.FS

var navCSSOnce sync.Once
var navCSSStr string

func navCSS() string {
	navCSSOnce.Do(func() {
		css, err := navFS.ReadFile("nav/nav.css")
		if err != nil {
			log.Printf("error reading embedded nav.css: %v", err)
			return
		}
		navCSSStr = "<style>" + string(css) + "</style>"
	})
	return navCSSStr
}

var navJSOnce sync.Once
var navJSStr string

func navJS() string {
	navJSOnce.Do(func() {
		js, err := navFS.ReadFile("nav/nav.js")
		if err != nil {
			log.Printf("error reading embedded nav.js: %v", err)
			return
		}
		navJSStr = string(js)
	})
	return navJSStr
}

func init() {
	// Inline CSS: injected into <head> for every theme (via InlineHead).
	RegisterAsset(Asset{
		Kind:   AssetInline,
		Detect: func(_, _ string) bool { return true },
		Script: navCSS,
	})

	// Post-load JS: the word navigation module. This file (nav.go) sorts after
	// highlightjs.go and mermaidext.go, so this asset registers last and its
	// script runs after theirs. That ordering is not load-bearing, though: the
	// module excludes pre/code/mermaid/KaTeX containers before wrapping words.
	RegisterAsset(Asset{
		Kind:   AssetPostLoad,
		Detect: func(_, _ string) bool { return true },
		Script: navJS,
	})
}
