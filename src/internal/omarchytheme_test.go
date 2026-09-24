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

func TestTreatmentNamesComeFromTheLibrary(t *testing.T) {
	names := TreatmentNames()
	if len(names) == 0 || names[0] != "Original" {
		t.Fatalf("names: %v", names)
	}
	if len(names) != omarchyTreatmentCount {
		t.Errorf("got %d names, the library counts %d", len(names), omarchyTreatmentCount)
	}
}

func TestTreatmentIndexByName(t *testing.T) {
	if i, ok := TreatmentIndex("original"); !ok || i != 0 {
		t.Errorf("original: %d %v", i, ok)
	}
	print, ok := TreatmentIndex(" PRINT ")
	if !ok || TreatmentNames()[print] != "Print" {
		t.Errorf("print: %d %v", print, ok)
	}
	if i, ok := TreatmentIndex("nonexistent"); ok || i != 0 {
		t.Errorf("an unknown name must miss and read Original, got %d %v", i, ok)
	}
}

func TestPrintPaletteIgnoresTheChosenTreatment(t *testing.T) {
	w, err := NewOmarchyThemeWatcher()
	if err != nil {
		t.Skip(err)
	}
	defer w.Close()
	inverted, _ := TreatmentIndex("Inverted")
	w.SetTreatment(inverted)
	if p := w.PrintPalette(); p.Background != "#ffffff" {
		t.Errorf("print stays white on a chosen treatment, got %q", p.Background)
	}
	w.SetTreatment(-1)
	w.SetTreatment(99)
	if w.treatment != 0 {
		t.Errorf("an out-of-range treatment reads Original, got %d", w.treatment)
	}
}

func TestPaletteFollowsTheChosenTreatment(t *testing.T) {
	w, err := NewOmarchyThemeWatcher()
	if err != nil {
		t.Skip(err)
	}
	defer w.Close()
	original := w.Palette()
	inverted, _ := TreatmentIndex("Inverted")
	w.SetTreatment(inverted)
	if w.Palette() == original {
		t.Error("Inverted must restate the theme, but the palette did not change")
	}
}
