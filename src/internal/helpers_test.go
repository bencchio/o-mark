package internal

import (
	"bytes"
	"os"
	"testing"
)

// renderDoc converts markdown through the same pipeline the viewer uses:
// goldmark plus the HTML post-processing pass.
func renderDoc(t *testing.T, input string) string {
	t.Helper()
	var buf bytes.Buffer
	if err := md.Convert([]byte(input), &buf); err != nil {
		t.Fatalf("convert: %v", err)
	}
	return htmlPostProcess(buf.String(), ".")
}

// readExample loads a file from examples/, so the shipped examples double as
// regression fixtures.
func readExample(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("../../examples/" + name)
	if err != nil {
		t.Fatalf("read example %s: %v", name, err)
	}
	return string(data)
}
