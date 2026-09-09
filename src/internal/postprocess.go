package internal

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var imgTagRe = regexp.MustCompile(`(?i)<img[^>]*>`)
var imgSrcRe = regexp.MustCompile(`(?i)\bsrc=["']([^"']*)["']`)
var imgAltRe = regexp.MustCompile(`(?i)\balt=["']([^"']*)["']`)

func isExternalSrc(src string) bool {
	low := strings.ToLower(src)
	return strings.HasPrefix(low, "http://") ||
		strings.HasPrefix(low, "https://") ||
		strings.HasPrefix(low, "data:") ||
		strings.HasPrefix(low, "javascript:")
}

func mimeFromExt(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	default:
		return "image/png"
	}
}

func fallbackSpan(alt string) string {
	if alt != "" {
		return `<span class="img-external">[` + alt + `]</span>`
	}
	return `<span class="img-external">[image]</span>`
}

// resolveImages processes <img> tags: external URLs and unresolvable paths are
// replaced with a styled alt-text fallback; relative and absolute local paths
// are read and inlined as base64 data URLs.
func resolveImages(html, dir string) string {
	// Resolve dir symlinks once so per-image checks are consistent.
	canonDir := dir
	if dir != "" {
		if r, err := filepath.EvalSymlinks(dir); err == nil {
			canonDir = r
		}
	}
	return imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		alt := ""
		if m := imgAltRe.FindStringSubmatch(tag); m != nil {
			alt = m[1]
		}
		m := imgSrcRe.FindStringSubmatch(tag)
		if m == nil {
			return fallbackSpan(alt)
		}
		src := m[1]
		if isExternalSrc(src) {
			return fallbackSpan(alt)
		}
		if dir == "" {
			return fallbackSpan(alt)
		}
		abs := src
		if !filepath.IsAbs(src) {
			abs = filepath.Join(dir, src)
		}
		// Resolve symlinks and ensure the target stays within the document directory.
		resolved, err := filepath.EvalSymlinks(abs)
		if err != nil {
			return fallbackSpan(alt)
		}
		rel, err := filepath.Rel(canonDir, resolved)
		if err != nil || rel == ".." || strings.HasPrefix(rel, "../") {
			return fallbackSpan(alt)
		}
		data, err := os.ReadFile(resolved)
		if err != nil {
			return fallbackSpan(alt)
		}
		mime := mimeFromExt(filepath.Ext(resolved))
		dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
		return imgSrcRe.ReplaceAllStringFunc(tag, func(_ string) string {
			return `src="` + dataURL + `"`
		})
	})
}

var mermaidBlockRe = regexp.MustCompile(`(?s)<pre><code class="language-mermaid">(.*?)</code></pre>`)

// renderMermaidBlocks converts goldmark's fenced mermaid code blocks into the
// <pre class="mermaid"> containers expected by mermaid.js.
func renderMermaidBlocks(html string) string {
	return mermaidBlockRe.ReplaceAllString(html, `<pre class="mermaid">$1</pre>`)
}

var admonRe = regexp.MustCompile(`(?is)<blockquote>\s*<p>\[!(NOTE|TIP|WARNING|IMPORTANT|CAUTION)\][ \t]*\n?(.*?)</blockquote>`)

var admonLabels = map[string]string{
	"NOTE":      "ℹ Note",
	"TIP":       "★ Tip",
	"WARNING":   "⚠ Warning",
	"IMPORTANT": "◆ Important",
	"CAUTION":   "◉ Caution",
}

// renderAdmonitions transforms GitHub-style callout blockquotes into styled
// admonition divs. Runs as HTML post-process so goldmark needs no changes.
func renderAdmonitions(html string) string {
	return admonRe.ReplaceAllStringFunc(html, func(match string) string {
		m := admonRe.FindStringSubmatch(match)
		typeName := strings.ToUpper(m[1])
		label := admonLabels[typeName]
		class := "admonition-" + strings.ToLower(m[1])
		content := strings.TrimSpace(m[2])
		var body string
		if strings.HasPrefix(content, "</p>") {
			// [!TYPE] was alone on its line; remaining content already has block tags
			body = strings.TrimSpace(content[4:])
		} else {
			// [!TYPE] was followed by content on the same <p>
			body = "<p>" + content
		}
		return `<div class="admonition ` + class + `">` +
			`<div class="admonition-title">` + label + `</div>` +
			body +
			`</div>`
	})
}

// extractFrontmatter detects YAML frontmatter at the very start of a markdown
// document (must begin with "---\n"). Returns the YAML body (without delimiters),
// the remaining document text, and whether frontmatter was found.
func extractFrontmatter(input string) (yaml string, rest string, found bool) {
	if !strings.HasPrefix(input, "---\n") {
		return "", input, false
	}
	body := input[4:] // skip opening "---\n"
	for _, closing := range []string{"---", "..."} {
		marker := "\n" + closing + "\n"
		if idx := strings.Index(body, marker); idx >= 0 {
			return strings.TrimSpace(body[:idx]), body[idx+len(marker):], true
		}
		// Handle closing at end of file without trailing newline
		if strings.HasSuffix(body, "\n"+closing) {
			trim := len(body) - len("\n"+closing)
			return strings.TrimSpace(body[:trim]), "", true
		}
	}
	return "", input, false
}

var fmEscaper = strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;")

// renderFrontmatterBlock returns a collapsed <details> block for YAML frontmatter.
func renderFrontmatterBlock(yaml string) string {
	return `<details class="o-mark-frontmatter">` +
		`<summary class="o-mark-frontmatter-header">frontmatter</summary>` +
		`<pre class="o-mark-frontmatter-content">` + fmEscaper.Replace(yaml) + `</pre>` +
		`</details>`
}

var checkboxRe = regexp.MustCompile(`<input[^>]*type="checkbox"[^>]*>`)

// renderTaskListCheckboxes replaces goldmark's <input type="checkbox" disabled>
// with <span> elements so that CSS can style them freely without fighting the
// browser's native checkbox appearance. goldmark emits neither a class on the
// <input> nor on its parent <li>, so both are added here.
func renderTaskListCheckboxes(html string) string {
	html = checkboxRe.ReplaceAllStringFunc(html, func(match string) string {
		if strings.Contains(match, `checked`) {
			return `<span class="task-list-item-checkbox checked"></span>`
		}
		return `<span class="task-list-item-checkbox"></span>`
	})
	return strings.ReplaceAll(html,
		`<li><span class="task-list-item-checkbox`,
		`<li class="task-list-item"><span class="task-list-item-checkbox`)
}

// htmlPostProcess applies all HTML transformations after goldmark rendering.
func htmlPostProcess(html, dir string) string {
	html = renderMermaidBlocks(html)
	html = resolveImages(html, dir)
	html = renderAdmonitions(html)
	html = renderTaskListCheckboxes(html)
	return html
}
