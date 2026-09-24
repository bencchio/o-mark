package internal

import "strings"

// pageDeclaration reads the sheet a document declares in its front matter:
// page_format and page_orientation. It is not a YAML parser — it scans the
// front matter's `key: value` lines for those two keys, ignoring case and
// quotes, and returns "" for a key that is absent or has an unknown value.
func pageDeclaration(input string) (format, orientation string) {
	yaml, _, found := extractFrontmatter(input)
	if !found {
		return "", ""
	}
	for _, line := range strings.Split(yaml, "\n") {
		key, value, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		value = declaredValue(value)
		switch strings.TrimSpace(key) {
		case "page_format":
			if f, ok := parseFormat(value); ok {
				format = f
			}
		case "page_orientation":
			if o, ok := parseOrientation(value); ok {
				orientation = o
			}
		}
	}
	return format, orientation
}

// declaredValue trims a front matter value of a trailing comment and of the
// quotes around it.
func declaredValue(v string) string {
	if i := strings.Index(v, " #"); i >= 0 {
		v = v[:i]
	}
	return strings.Trim(strings.TrimSpace(v), `"'`)
}

// forDocument returns cfg with the sheet the document declares, when it
// declares one, replacing the config's defaults.
func (c Config) forDocument(raw string) Config {
	format, orientation := pageDeclaration(raw)
	if format != "" {
		c.PageFormat = format
	}
	if orientation != "" {
		c.PageOrientation = orientation
	}
	return c
}
