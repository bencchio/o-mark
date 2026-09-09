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
- Keyboard-first: arrow/PgUp/PgDn/Home/End scroll, `Ctrl+±`/`Ctrl+0` zoom, `Esc` toggles the toolbar, `T`/`Z` jump to its controls, `Ctrl+Q` to quit — full list in [docs/KEYBOARD-SHORTCUTS.md](docs/KEYBOARD-SHORTCUTS.md)
- Window title shows the current file path

## Configuration

`~/.config/o-mark/config.toml` (created on first run) controls the starting
theme, document max-width and font size, scrollbar visibility, code block line
numbers, and toolbar position/visibility. See the comments in the generated
file for details.

The theme you are using when you quit is saved as the one O'Mark starts with
next time.

## Requirements

- Qt6, WebEngine, and `pkg-config` — see [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) for versions and supported distros
- Go and CGO — see [docs/BUILDING.md](docs/BUILDING.md)

## Quick start

```bash
cd src
go build -o target/o-mark .
./target/o-mark ../examples/markdown/general.md
```

More examples in [examples/](examples/), grouped by what they cover:
[markdown/](examples/markdown/) for the CommonMark, GitHub and Obsidian
dialects, [extensions/](examples/extensions/) for math, Mermaid and the rest,
and [code/](examples/code/) for syntax highlighting across 10 languages.

## Install

See [INSTALL.md](INSTALL.md), or run `scripts/install.sh` for an automated
build + install with a dependency check.

## Docs

| Document | Content |
|----------|---------|
| [docs/KEYBOARD-SHORTCUTS.md](docs/KEYBOARD-SHORTCUTS.md) | Every keyboard shortcut and mouse action |
| [docs/BUILDING.md](docs/BUILDING.md) | Build instructions |
| [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) | System and Go dependencies |
| [docs/FONTS.md](docs/FONTS.md) | Recommended fonts |
| [INSTALL.md](INSTALL.md) | Install to a system prefix |
