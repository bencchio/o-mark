package internal

import (
	"embed"
	"encoding/base64"
	"log"
	"regexp"
	"strings"
	"sync"
)

//go:embed katex/katex.min.css katex/katex.min.js katex/auto-render.min.js katex/fonts/*.woff2
var katexFS embed.FS

func init() {
	RegisterAsset(Asset{
		Kind: AssetInline,
		Detect: func(_, body string) bool {
			return strings.Contains(body, `\(`) || strings.Contains(body, `\[`)
		},
		Script: katexHead,
	})
}

var katexOnce sync.Once
var katexHeadStr string

// katexHead returns the HTML <head> fragment that loads KaTeX from embedded
// files. The CSS is processed once to replace font url() references with
// base64 data URIs so the document is fully self-contained offline.
// Returns "" if embedded assets are unavailable, rendering math without KaTeX.
func katexHead() string {
	katexOnce.Do(buildKatexHead)
	return katexHeadStr
}

func buildKatexHead() {
	css, err := buildKatexCSS()
	if err != nil {
		log.Printf("katex: embedded assets unavailable: %v", err)
		return
	}
	js, err := katexFS.ReadFile("katex/katex.min.js")
	if err != nil {
		log.Printf("katex: embedded assets unavailable: %v", err)
		return
	}
	ar, err := katexFS.ReadFile("katex/auto-render.min.js")
	if err != nil {
		log.Printf("katex: embedded assets unavailable: %v", err)
		return
	}
	katexHeadStr = "<style>" + css + "</style>" +
		"<script>" + string(js) + "</script>" +
		"<script>" + string(ar) + `</script>` +
		`<script>document.addEventListener('DOMContentLoaded',function(){` +
		`renderMathInElement(document.body,{delimiters:[` +
		`{left:'\\[',right:'\\]',display:true},` +
		`{left:'\\(',right:'\\)',display:false}` +
		`]});});</script>`
}

// woff2Re matches url(fonts/name.woff2) in all quoting styles (none, single, double).
var woff2Re = regexp.MustCompile(`url\(['"]?fonts/([^'")\s]+\.woff2)['"]?\)`)
var woffFallRe = regexp.MustCompile(`,url\(fonts/[^)]+\.(woff|ttf)\) format\("[^"]+"\)`)

func buildKatexCSS() (string, error) {
	raw, err := katexFS.ReadFile("katex/katex.min.css")
	if err != nil {
		return "", err
	}
	css := string(raw)
	css = woff2Re.ReplaceAllStringFunc(css, func(m string) string {
		name := woff2Re.FindStringSubmatch(m)[1]
		data, err := katexFS.ReadFile("katex/fonts/" + name)
		if err != nil {
			return m
		}
		return "url(data:font/woff2;base64," + base64.StdEncoding.EncodeToString(data) + ")"
	})
	css = woffFallRe.ReplaceAllString(css, "")
	return css, nil
}
