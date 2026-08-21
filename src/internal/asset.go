package internal

import "strings"

// AssetKind distinguishes how a JS asset is injected into the document.
type AssetKind int

const (
	// AssetInline is injected into <head> at HTML render time (e.g. KaTeX).
	AssetInline AssetKind = iota
	// AssetPostLoad is injected via runJavaScript after the page loads (e.g. Mermaid).
	AssetPostLoad
)

// Asset describes a self-contained JS (or CSS+JS) bundle to be conditionally
// loaded into the viewer. Register new assets via RegisterAsset in init().
type Asset struct {
	Kind   AssetKind
	Detect func(rawMD, body string) bool // body="" when called for PostLoad
	Script func() string                 // lazy; implementations should cache with sync.Once
}

var registry []Asset

// RegisterAsset adds an asset to the global registry. Call from init().
func RegisterAsset(a Asset) {
	registry = append(registry, a)
}

// InlineHead returns the combined <head> HTML for all AssetInline assets
// whose Detect function returns true. Called during HTML rendering.
func InlineHead(rawMD, body string) string {
	var sb strings.Builder
	for _, a := range registry {
		if a.Kind == AssetInline && a.Detect(rawMD, body) {
			sb.WriteString(a.Script())
		}
	}
	return sb.String()
}

// PostLoadScripts returns the scripts for all AssetPostLoad assets whose
// Detect function returns true for the given raw markdown. Called at startup.
func PostLoadScripts(rawMD string) []string {
	scripts := []string{}
	for _, a := range registry {
		if a.Kind == AssetPostLoad && a.Detect(rawMD, "") {
			scripts = append(scripts, a.Script())
		}
	}
	return scripts
}
