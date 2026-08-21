package internal

import (
	"embed"
	"log"
	"regexp"
	"sync"
)

//go:embed mermaid/mermaid.min.js
var mermaidFS embed.FS

// MermaidScript returns the full mermaid.js bundle content for injection via
// QML runJavaScript after the page loads. This avoids the 2MB loadHtml limit.
// Returns empty string if the embedded file cannot be read.
// mermaidFenceRe matches a real mermaid fenced code block opener at the start
// of a line (0-3 spaces indent), avoiding false positives when ```mermaid
// appears as literal text inside another code block.
var mermaidFenceRe = regexp.MustCompile(`(?m)^[ \t]{0,3}(` + "```" + `|~~~)mermaid`)

func init() {
	RegisterAsset(Asset{
		Kind: AssetPostLoad,
		Detect: func(raw, _ string) bool {
			return mermaidFenceRe.MatchString(raw)
		},
		Script: MermaidScript,
	})
}

var MermaidScript = sync.OnceValue(func() string {
	js, err := mermaidFS.ReadFile("mermaid/mermaid.min.js")
	if err != nil {
		log.Printf("error reading embedded mermaid.min.js: %v", err)
		return ""
	}
	return string(js) + ";mermaid.initialize({startOnLoad:false,theme:'neutral'});mermaid.run();"
})
