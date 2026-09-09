# Installing O'Mark

On Arch, prefer the package in [README.md § Install](README.md): install
[omarchy-theme](https://github.com/bencchio/omarchy-theme), then
`cd packaging/arch && makepkg -si`. This page is the copy-to-prefix path.

## 1. Build

```bash
cd src
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -o target/o-mark .
```

Requirements: Go 1.26+, Qt6 (`qt6-base`, `qt6-declarative`, `qt6-webengine`), CGO enabled.

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

```bash
cat > ~/.local/share/applications/o-mark.desktop << 'EOF'
[Desktop Entry]
Name=O'Mark
Comment=Minimal Markdown Viewer
Exec=o-mark %f
Icon=text-x-markdown
Type=Application
MimeType=text/markdown;text/x-markdown;
Categories=Utility;
EOF
```

**3. Register as default and update the application database:**

```bash
xdg-mime default o-mark.desktop text/markdown
xdg-mime default o-mark.desktop text/x-markdown
update-desktop-database ~/.local/share/applications/
```

**4. Verify:**

```bash
xdg-mime query filetype /path/to/file.md  # must print text/markdown, not text/plain
xdg-mime query default text/markdown      # must print o-mark.desktop
```

If step 1 was skipped and the first command returns `text/plain`, the handler
registered in step 3 will never be used — `xdg-open` resolves by detected type,
not by file extension.

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
