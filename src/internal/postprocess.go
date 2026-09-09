package internal

import (
	"encoding/base64"
	"fmt"
	"io"
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
	if dir == "" {
		return imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
			return fallbackSpan(imgAlt(tag))
		})
	}
	// A directory handle keeps containment and reading in one operation: the OS
	// refuses any component that escapes the root, so there is no window between
	// checking a path and opening it. Closed once every image is resolved.
	root, err := os.OpenRoot(dir)
	if err != nil {
		return imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
			return fallbackSpan(imgAlt(tag))
		})
	}
	defer root.Close()

	return imgTagRe.ReplaceAllStringFunc(html, func(tag string) string {
		alt := imgAlt(tag)
		m := imgSrcRe.FindStringSubmatch(tag)
		if m == nil {
			return fallbackSpan(alt)
		}
		src := m[1]
		if isExternalSrc(src) {
			return fallbackSpan(alt)
		}
		// The root only accepts paths relative to itself; an absolute src is
		// rewritten as one, and stays subject to the same containment check.
		rel := src
		if filepath.IsAbs(src) {
			var err error
			if rel, err = filepath.Rel(dir, src); err != nil {
				return fallbackSpan(alt)
			}
		}
		data, err := readRootFile(root, rel)
		if err != nil {
			return fallbackSpan(alt)
		}
		mime := mimeFromExt(filepath.Ext(src))
		dataURL := "data:" + mime + ";base64," + base64.StdEncoding.EncodeToString(data)
		return imgSrcRe.ReplaceAllStringFunc(tag, func(_ string) string {
			return `src="` + dataURL + `"`
		})
	})
}

// imgAlt returns the alt attribute of an <img> tag, empty when it has none.
func imgAlt(tag string) string {
	if m := imgAltRe.FindStringSubmatch(tag); m != nil {
		return m[1]
	}
	return ""
}

// readRootFile reads name from root, refusing anything that is not a regular
// file so a device node or a fifo cannot stall the render.
func readRootFile(root *os.Root, name string) ([]byte, error) {
	f, err := root.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, fmt.Errorf("resolveImages: %s is not a regular file", name)
	}
	return io.ReadAll(f)
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
