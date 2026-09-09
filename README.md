# O'Mark

A minimalist markdown viewer. Go + Qt6 / QML.

```
o-mark document.md
```

## Features

- Headings, paragraphs, emphasis, lists, blockquotes, code blocks, tables
- Task lists (`- [x]`), definition lists, footnotes, subscript/superscript, `==highlight==`
- Syntax-highlighted code blocks with line numbers; long lines wrap instead of overflowing
- Math (`$inline$` / `$$block$$`, via embedded KaTeX) and Mermaid diagrams
- GitHub-style admonitions (`> [!NOTE]`, `[!TIP]`, `[!WARNING]`, `[!IMPORTANT]`, `[!CAUTION]`)
- Local images inlined as `data:` URLs; collapsible YAML frontmatter
- Links to other `.md` files open in a new O'Mark instance (see [INSTALL.md § 3](INSTALL.md))
- 7 viewer themes (System, following the active Omarchy palette, plus 6 built-in: GitHub, Writer, Night, Sepia, Mono, Minimal) — extensible by dropping CSS into `~/.config/o-mark/themes/`
- Live reload: the open document, `config.toml`, and theme files are watched and re-rendered on change
- Keyboard-first: `PgUp`/`PgDn` scroll while arrow/`Home`/`End` move the word cursor, `Ctrl+±`/`Ctrl+0` zoom, `Esc` toggles the toolbar, `T`/`Z` jump to its controls, `Ctrl+Q` to quit — full list in [docs/KEYBOARD-SHORTCUTS.md](docs/KEYBOARD-SHORTCUTS.md)
- Window title shows the current file path

## Configuration

`~/.config/o-mark/config.toml` (created on first run) controls the starting
theme, document max-width and font size, scrollbar visibility, code block line
numbers, and toolbar position/visibility. See the comments in the generated
file for details.

The theme you are using when you quit is saved as the one O'Mark starts with
next time.

## Requirements

- Qt6, WebEngine, `pkg-config`, and `omarchy-theme` — see [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) for versions and supported distros
- Go and CGO — see [docs/BUILDING.md](docs/BUILDING.md)

## Quick start

```bash
cd src
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -o target/o-mark .
./target/o-mark ../examples/markdown/general.md
```

More examples in [examples/](examples/), grouped by what they cover:
[markdown/](examples/markdown/) for the CommonMark, GitHub and Obsidian
dialects, [extensions/](examples/extensions/) for math, Mermaid and the rest,
and [code/](examples/code/) for syntax highlighting across 11 languages.

## Install

On Arch, install [omarchy-theme](https://github.com/bencchio/omarchy-theme)
first.

From the [0.7.0 release](https://github.com/bencchio/o-mark/releases/tag/0.7.0):

```bash
sudo pacman -U o-mark-0.7.0-1-x86_64.pkg.tar.zst
```

Or build that tag:

```bash
git clone https://github.com/bencchio/o-mark.git
cd o-mark
git checkout 0.7.0
cd packaging/arch
makepkg -si
```

Without a package, see [INSTALL.md](INSTALL.md) or run `scripts/install.sh`.

## Docs

| Document | Content |
|----------|---------|
| [docs/MANUAL.md](docs/MANUAL.md) | How to use O'Mark, with a FAQ |
| [docs/CONFIG.md](docs/CONFIG.md) | Reference for `config.toml` |
| [docs/THEMES.md](docs/THEMES.md) | The 7 themes and custom CSS |
| [docs/KEYBOARD-SHORTCUTS.md](docs/KEYBOARD-SHORTCUTS.md) | Every keyboard shortcut and mouse action |
| [docs/BUILDING.md](docs/BUILDING.md) | Build and test instructions |
| [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) | System and Go dependencies |
| [docs/FONTS.md](docs/FONTS.md) | Recommended fonts |
| [INSTALL.md](INSTALL.md) | Install to a system prefix |
| [CHANGELOG.md](CHANGELOG.md) | Release history |
