# O'Mark

A minimalist markdown viewer. Go + Qt6 / QML.

```
o-mark document.md
```

## Features

- Headings, paragraphs, emphasis, lists, blockquotes, code blocks, tables
- Task lists (`- [x]`), definition lists, footnotes, subscript/superscript, `==highlight==`
- Math (`$inline$` / `$$block$$`, via embedded KaTeX) and Mermaid diagrams
- GitHub-style admonitions (`> [!NOTE]`, `[!TIP]`, `[!WARNING]`, `[!IMPORTANT]`, `[!CAUTION]`)
- Local images inlined as `data:` URLs; collapsible YAML frontmatter
- Links to other `.md` files open in a new O'Mark instance (see [INSTALL.md § 3](INSTALL.md))
- 7 viewer themes (System, following the active Omarchy palette, plus 6 built-in: GitHub, Writer, Night, Sepia, Mono, Minimal) — extensible by dropping CSS into `~/.config/o-mark/themes/`
- Live reload: the open document, `config.toml`, and theme files are watched and re-rendered on change
- Keyboard-first: arrow/PgUp/PgDn/Home/End scroll, `Ctrl+±`/`Ctrl+0` zoom, `Esc` toggles the toolbar, `T`/`Z` jump to its controls, `Ctrl+Q` to quit
- Window title shows the current file path

## Configuration

`~/.config/o-mark/config.toml` (created on first run) controls the starting
theme, document max-width and font size, scrollbar visibility, and toolbar
position/visibility. See the comments in the generated file for details.

## Requirements

- Qt6, WebEngine, and `pkg-config` — see [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) for versions and supported distros
- Go and CGO — see [docs/BUILDING.md](docs/BUILDING.md)

## Quick start

```bash
cd src
go build -o target/o-mark .
./target/o-mark ../examples/general.md
```

More examples (math, mermaid, tables, admonitions, images) in [examples/](examples/).

## Install

See [INSTALL.md](INSTALL.md), or run `scripts/install.sh` for an automated
build + install with a dependency check.

## Docs

| Document | Content |
|----------|---------|
| [docs/BUILDING.md](docs/BUILDING.md) | Build instructions |
| [docs/DEPENDENCIES.md](docs/DEPENDENCIES.md) | System and Go dependencies |
| [docs/FONTS.md](docs/FONTS.md) | Recommended fonts |
| [INSTALL.md](INSTALL.md) | Install to a system prefix |
