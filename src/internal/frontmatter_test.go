package internal

import "testing"

func TestPageDeclaration(t *testing.T) {
	tests := []struct {
		name, doc, format, orientation string
	}{
		{"none", "# Hi\n", "", ""},
		{"both", "---\npage_format: a5\npage_orientation: landscape\n---\n# Hi\n", "a5", "landscape"},
		{"quoted and capitalized", "---\npage_format: \"A3\"\npage_orientation: 'Landscape'\n---\n", "a3", "landscape"},
		{"trailing comment", "---\npage_format: a4 # the default\n---\n", "a4", ""},
		{"unknown value ignored", "---\npage_format: letter\npage_orientation: sideways\n---\n", "", ""},
		{"unknown key ignored", "---\ntitle: Report\ntheme: night\n---\n", "", ""},
		{"only orientation", "---\ntitle: x\npage_orientation: landscape\n---\n", "", "landscape"},
		{"no closing marker", "---\npage_format: a5\n# Hi\n", "", ""},
		{"a separator later is not front matter", "# Hi\n\n---\npage_format: a5\n---\n", "", ""},
		{"windows line endings", "---\r\npage_format: a5\r\n---\r\n", "", ""},
	}
	for _, tc := range tests {
		f, o := pageDeclaration(tc.doc)
		if f != tc.format || o != tc.orientation {
			t.Errorf("%s: got (%q, %q), want (%q, %q)", tc.name, f, o, tc.format, tc.orientation)
		}
	}
}

func TestForDocumentPrefersTheDeclaration(t *testing.T) {
	cfg := Config{PageFormat: "a4", PageOrientation: "portrait"}
	got := cfg.forDocument("---\npage_orientation: landscape\n---\n")
	if got.PageFormat != "a4" || got.PageOrientation != "landscape" {
		t.Errorf("got %q %q: the declared orientation wins and the format stays the default", got.PageFormat, got.PageOrientation)
	}
	if cfg.PageOrientation != "portrait" {
		t.Error("forDocument must not change the config it is called on")
	}
}
