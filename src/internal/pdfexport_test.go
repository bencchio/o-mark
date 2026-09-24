package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSiblingPdfPath(t *testing.T) {
	tests := []struct{ in, want string }{
		{"/tmp/foo.md", "/tmp/foo.pdf"},
		{"/tmp/foo.markdown", "/tmp/foo.pdf"},
		{"/tmp/foo", "/tmp/foo.pdf"},
	}
	for _, tc := range tests {
		if got := SiblingPdfPath(tc.in); got != tc.want {
			t.Errorf("SiblingPdfPath(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestPreparePdfExportExists(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	pdf := filepath.Join(dir, "doc.pdf")
	if err := os.WriteFile(pdf, []byte("%PDF"), 0644); err != nil {
		t.Fatal(err)
	}
	got := PreparePdfExport(md, "# Hi\n", dir, "sans-serif", Config{}, nil)
	if !got.Exists {
		t.Error("exists: want true")
	}
	if got.Path != pdf {
		t.Errorf("path: got %q, want %q", got.Path, pdf)
	}
	if got.HTML != "" {
		t.Error("html should be empty when the PDF already exists")
	}
}

func TestPreparePdfExportMissingWatcher(t *testing.T) {
	dir := t.TempDir()
	md := filepath.Join(dir, "doc.md")
	got := PreparePdfExport(md, "# Hi\n", dir, "sans-serif", Config{}, nil)
	if got.Exists {
		t.Error("exists: want false")
	}
	if got.Path != filepath.Join(dir, "doc.pdf") {
		t.Errorf("path: got %q", got.Path)
	}
	if !strings.Contains(got.HTML, "--o-mark-bg: #ffffff") {
		t.Error("print HTML missing white background")
	}
	if !strings.Contains(got.HTML, "--o-mark-fg: #000000") {
		t.Error("print HTML missing black foreground")
	}
}

func TestPrintMarginsWrapsTheBodyInARepeatingSpacerTable(t *testing.T) {
	html := "<!DOCTYPE html><html><head><style>x</style></head><body><p>hi</p></body></html>"
	got := printMargins(html, Config{})
	for _, want := range []string{"<thead>", "<tfoot>", "height:25mm", "<p>hi</p>", "margin: 0 auto", "padding: 0 10mm"} {
		if !strings.Contains(got, want) {
			t.Errorf("print HTML is missing %q:\n%s", want, got)
		}
	}
	if strings.Count(got, "height:25mm") != 2 {
		t.Error("expected one spacer in the header and one in the footer")
	}
	if !strings.HasSuffix(got, "</body></html>") {
		t.Error("the document must still close normally")
	}
}

func TestPrintMarginsFollowTheConfig(t *testing.T) {
	html := "<!DOCTYPE html><html><head></head><body><p>hi</p></body></html>"
	got := printMargins(html, Config{PdfMarginVertical: "30mm", PdfMarginHorizontal: "20mm"})
	if strings.Count(got, "height:30mm") != 2 || !strings.Contains(got, "padding: 0 20mm") {
		t.Errorf("margins did not follow the config:\n%s", got)
	}
	bad := printMargins(html, Config{PdfMarginVertical: "url(x)", PdfMarginHorizontal: "1cm"})
	if strings.Contains(bad, "url(x)") || strings.Contains(bad, "1cm") {
		t.Errorf("an invalid margin reached the CSS:\n%s", bad)
	}
}

func TestPrintMarginsLeavesUnexpectedHTMLAlone(t *testing.T) {
	if got := printMargins("no body here", Config{}); got != "no body here" {
		t.Errorf("got %q", got)
	}
}

func TestPreparePdfExportCarriesTheSheet(t *testing.T) {
	dir := t.TempDir()
	got := PreparePdfExport(filepath.Join(dir, "doc.md"), "# Hi\n", dir, "sans-serif", Config{PageFormat: "a5", PageOrientation: "landscape"}, nil)
	if !strings.Contains(got.HTML, "o-mark-print") {
		t.Error("the print document has no margin table")
	}
	if got.Orientation != "landscape" || got.Format != "a5" {
		t.Errorf("sheet: %q %q", got.Format, got.Orientation)
	}
}
