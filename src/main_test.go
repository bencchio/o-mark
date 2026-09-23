package main

import "testing"

func TestResolveDocArg(t *testing.T) {
	cases := []struct {
		name       string
		in         string
		wantPath   string
		wantAnchor string
	}{
		{"bare path", "/home/user/doc.md", "/home/user/doc.md", ""},
		{"relative bare path", "doc.md", "doc.md", ""},
		{"bare path with literal hash", "notes#1.md", "notes#1.md", ""},
		{"file uri no anchor", "file:///home/user/doc.md", "/home/user/doc.md", ""},
		{"file uri with anchor", "file:///home/user/doc.md#section", "/home/user/doc.md", "section"},
		{"o-mark uri no anchor", "o-mark:///home/user/other.md", "/home/user/other.md", ""},
		{"o-mark uri with anchor", "o-mark:///home/user/other.md#heading-1", "/home/user/other.md", "heading-1"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			path, anchor := resolveDocArg(c.in)
			if path != c.wantPath || anchor != c.wantAnchor {
				t.Errorf("resolveDocArg(%q) = (%q, %q), want (%q, %q)",
					c.in, path, anchor, c.wantPath, c.wantAnchor)
			}
		})
	}
}
