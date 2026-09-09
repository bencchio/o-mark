# Dependencies

## System

| Dependency | Version | Why it's needed |
|---|---|---|
| Qt (Widgets, Quick, QuickControls2, QuickLayouts) | ≥ 6.5 | Needs Chromium ≥ 111 for the `color-mix()` CSS used by the System theme |
| Qt WebEngine (Quick) | ≥ 6.5 | Renders the document HTML (`MarkdownViewer.qml`) — plain Qt Quick is not enough |
| pkg-config | — | miqt uses it to locate the Qt6 modules (`.pc` files) |
| CGO | — | Required by miqt |

Supported distributions: Arch Linux (rolling — keep the system up to date with
`pacman -Syu`) and Ubuntu 24.04 LTS. `scripts/install.sh` checks these
dependencies for you.

## Go modules

| Module | Version | Why it's needed |
|---|---|---|
| `github.com/mappu/miqt` | v0.14.0 | Qt6 bindings for Go |
| `github.com/yuin/goldmark` | v1.8.2 | Parses markdown and renders it to HTML |
| `github.com/BurntSushi/toml` | v1.6.0 | Reads `config.toml` and the Omarchy theme palette |

## Bundled JavaScript

These are embedded into the binary at build time, so O'Mark renders everything
offline and never contacts a CDN.

| Library | Version | Why it's needed |
|---|---|---|
| KaTeX | 0.18.4 | Renders `$inline$` and `$$block$$` math |
| Mermaid | 11.17.0 | Renders ` ```mermaid ` diagrams |
| highlight.js | 11.12.0 | Syntax highlighting in code blocks (common bundle) |
| highlightjs-line-numbers.js | 2.9.1 | Line numbers in code blocks |
