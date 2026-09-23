package internal

import "testing"

func TestPrintPaletteWhiteBlack(t *testing.T) {
	w, err := NewOmarchyThemeWatcher()
	if err != nil {
		t.Skip(err)
	}
	defer w.Close()

	p := w.PrintPalette()
	if p.Background != "#ffffff" {
		t.Errorf("background: got %q, want #ffffff", p.Background)
	}
	if p.Foreground != "#000000" {
		t.Errorf("foreground: got %q, want #000000", p.Foreground)
	}
}
