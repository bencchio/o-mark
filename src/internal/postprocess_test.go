package internal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A symlink inside the document directory pointing outside it must not be
// inlined: the referenced file would end up embedded in the rendered HTML.
func TestResolveImagesRefusesSymlinkEscape(t *testing.T) {
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.png")
	if err := os.WriteFile(secret, []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}

	dir := t.TempDir()
	if err := os.Symlink(secret, filepath.Join(dir, "escape.png")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	got := resolveImages(`<img src="escape.png" alt="x">`, dir)
	if strings.Contains(got, "base64") {
		t.Errorf("symlink escaping the document directory was inlined: %s", got)
	}
	if !strings.Contains(got, "img-external") {
		t.Errorf("expected the alt-text fallback, got: %s", got)
	}
}

// Parent traversal must be refused whether it is spelled relatively or as an
// absolute path.
func TestResolveImagesRefusesTraversal(t *testing.T) {
	outside := t.TempDir()
	secret := filepath.Join(outside, "secret.png")
	if err := os.WriteFile(secret, []byte("private"), 0644); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(outside, "doc")
	if err := os.Mkdir(dir, 0755); err != nil {
		t.Fatal(err)
	}

	for _, src := range []string{"../secret.png", secret} {
		got := resolveImages(`<img src="`+src+`" alt="x">`, dir)
		if strings.Contains(got, "base64") {
			t.Errorf("src %q escaped the document directory: %s", src, got)
		}
	}
}

// A file inside the directory is still inlined, including one whose name starts
// with the traversal prefix — the check is on the path, not on the name.
func TestResolveImagesInlinesLocalFile(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"pic.png", "..hidden.png"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("bytes"), 0644); err != nil {
			t.Fatal(err)
		}
		got := resolveImages(`<img src="`+name+`" alt="x">`, dir)
		if !strings.Contains(got, "data:image/png;base64,") {
			t.Errorf("%s was not inlined: %s", name, got)
		}
	}
}

// A symlink that stays inside the directory is legitimate and still resolves.
func TestResolveImagesFollowsInternalSymlink(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "real.png"), []byte("bytes"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink("real.png", filepath.Join(dir, "link.png")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	got := resolveImages(`<img src="link.png" alt="x">`, dir)
	if !strings.Contains(got, "data:image/png;base64,") {
		t.Errorf("internal symlink was not inlined: %s", got)
	}
}

// Without a document directory nothing can be resolved, so every image falls
// back rather than reaching the filesystem.
func TestResolveImagesWithoutDirectory(t *testing.T) {
	got := resolveImages(`<img src="pic.png" alt="x">`, "")
	if !strings.Contains(got, "img-external") {
		t.Errorf("expected the alt-text fallback, got: %s", got)
	}
}

// External and script-bearing sources never reach the filesystem.
func TestResolveImagesRejectsExternalSources(t *testing.T) {
	dir := t.TempDir()
	for _, src := range []string{"https://example.com/a.png", "javascript:alert(1)", "data:image/png;base64,AAAA"} {
		got := resolveImages(`<img src="`+src+`" alt="x">`, dir)
		if !strings.Contains(got, "img-external") {
			t.Errorf("src %q was not replaced by the fallback: %s", src, got)
		}
	}
}
