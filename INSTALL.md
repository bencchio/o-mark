# Installing O'Mark

On Arch, prefer [README.md § Install](README.md): install
[omarchy-lib-theme](https://github.com/bencchio/omarchy-lib-theme), then either
`pacman -U` the 0.7.0 release package or `git checkout 0.7.0` and
`makepkg -si`. This page is the copy-to-prefix path.

## 1. Build

```bash
cd src
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -o target/o-mark .
```

Requirements: Go 1.26+, Qt6 (`qt6-base`, `qt6-declarative`, `qt6-webengine`), CGO enabled.

For a smaller binary to actually install, add `-ldflags="-s -w"`: it strips
the symbol table and debug info, cutting the size by roughly a third. Keep
a build without those flags around (or just rebuild from the same commit)
if you ever need to symbolize a crash — see the `diagnose-crash` skill.

```bash
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -ldflags="-s -w" -o target/o-mark .
```

## 2. Install

The binary is self-contained — QML files are embedded at compile time.
Copy the binary to any directory in your PATH.

### User install (`~/.local/bin`)

```bash
cp src/target/o-mark ~/.local/bin/o-mark
```

Add `~/.local/bin` to your PATH if needed:

```bash
# ~/.bashrc or ~/.zshrc
export PATH="$HOME/.local/bin:$PATH"
```

### System install (`/usr/local/bin`)

```bash
sudo cp src/target/o-mark /usr/local/bin/o-mark
```

## 3. Register as default `.md` viewer (optional)

Required for links between markdown files to open in a new O'Mark instance.
O'Mark uses `xdg-open` internally, so it needs to be the system default for
`text/markdown`.

**1. Declare the MIME type for `.md` files:**

Many systems (including Arch) classify `.md` files as `text/plain` by default,
which means `xdg-open` never reaches the `text/markdown` handler. This step
adds a user-level override that maps `*.md` → `text/markdown`.

```bash
mkdir -p ~/.local/share/mime/packages/
cat > ~/.local/share/mime/packages/text-markdown.xml << 'EOF'
<?xml version="1.0" encoding="utf-8"?>
<mime-info xmlns="http://www.freedesktop.org/standards/shared-mime-info">
  <mime-type type="text/markdown">
    <comment>Markdown document</comment>
    <glob pattern="*.md"/>
    <glob pattern="*.markdown"/>
  </mime-type>
</mime-info>
EOF
update-mime-database ~/.local/share/mime/
```

**2. Create the desktop entry:**

The `.desktop` file tells the system how to launch O'Mark from a file manager
or via `xdg-open`.

`%u` (not `%f`) so a link between O'Mark documents can carry the target
heading as a URL fragment — `%f` is a local-path field code with no defined
fragment handling, `%u` is the one that passes the full URI through.
`x-scheme-handler/o-mark` is a scheme only O'Mark ever claims, used only to
reach a fresh instance from a link in another O'Mark document — it still only
ever opens `.md` files, the same as `text/markdown`.

```bash
cat > ~/.local/share/applications/o-mark.desktop << 'EOF'
[Desktop Entry]
Name=O'Mark
Comment=Minimal Markdown Viewer
Exec=o-mark %u
Icon=text-x-markdown
Type=Application
MimeType=text/markdown;text/x-markdown;x-scheme-handler/o-mark;
Categories=Utility;
EOF
```

**3. Register as default and update the application database:**

```bash
xdg-mime default o-mark.desktop text/markdown
xdg-mime default o-mark.desktop text/x-markdown
xdg-mime default o-mark.desktop x-scheme-handler/o-mark
update-desktop-database ~/.local/share/applications/
```

**4. Verify:**

```bash
xdg-mime query filetype /path/to/file.md  # must print text/markdown, not text/plain
xdg-mime query default text/markdown      # must print o-mark.desktop
xdg-mime query default x-scheme-handler/o-mark  # must print o-mark.desktop
```

If step 1 was skipped and the first command returns `text/plain`, the handler
registered in step 3 will never be used — `xdg-open` resolves by detected type,
not by file extension.

`%u` applies to every launch, not just links between documents: opening a
file from a file manager or via plain `o-mark file.md` still works, since
`o-mark` accepts either a bare local path or a URI on its command line.

To revert to a previous default:

```bash
xdg-mime default <previous-app>.desktop text/markdown
xdg-mime default <previous-app>.desktop text/x-markdown
```

## 4. Verify

```bash
o-mark /path/to/file.md
```

## Uninstall

```bash
rm -f ~/.local/bin/o-mark       # user install
sudo rm -f /usr/local/bin/o-mark  # system install
```
