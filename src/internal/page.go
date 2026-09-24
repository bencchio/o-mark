package internal

import "strings"

// Page formats and orientations the config and the front matter accept.
const (
	PageFormatA3         = "a3"
	PageFormatA4         = "a4"
	PageFormatA5         = "a5"
	OrientationPortrait  = "portrait"
	OrientationLandscape = "landscape"
)

// pageFormats maps each format to its portrait sides.
var pageFormats = map[string]struct{ width, height string }{
	PageFormatA3: {"297mm", "420mm"},
	PageFormatA4: {"210mm", "297mm"},
	PageFormatA5: {"148mm", "210mm"},
}

// parseFormat returns the canonical format for v, ignoring case and
// surrounding space, and whether v named a known format.
func parseFormat(v string) (string, bool) {
	f := strings.ToLower(strings.TrimSpace(v))
	_, ok := pageFormats[f]
	return f, ok
}

// parseOrientation returns the canonical orientation for v, ignoring case and
// surrounding space, and whether v named one.
func parseOrientation(v string) (string, bool) {
	o := strings.ToLower(strings.TrimSpace(v))
	return o, o == OrientationPortrait || o == OrientationLandscape
}

// sheetFormat normalizes a format to a known one, falling back to A4.
func sheetFormat(v string) string {
	if f, ok := parseFormat(v); ok {
		return f
	}
	return PageFormatA4
}

// sheetOrientation normalizes an orientation, falling back to portrait.
func sheetOrientation(v string) string {
	if o, ok := parseOrientation(v); ok {
		return o
	}
	return OrientationPortrait
}

// pageSides returns the sheet's width and height as CSS lengths. An unknown
// format is A4 and anything but landscape is portrait, so a zero-value Config
// still gets a valid sheet.
func pageSides(format, orientation string) (width, height string) {
	sides := pageFormats[sheetFormat(format)]
	if sheetOrientation(orientation) == OrientationLandscape {
		return sides.height, sides.width
	}
	return sides.width, sides.height
}

// legacyOrientation maps the retired document_max_width to an orientation:
// only the A4 long side meant landscape; every other width meant portrait.
func legacyOrientation(maxWidth string) string {
	if maxWidth == "297mm" {
		return OrientationLandscape
	}
	return OrientationPortrait
}
