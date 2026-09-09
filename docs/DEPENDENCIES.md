# Dependencies

## System

| Dependency | Version | Why it's needed |
|---|---|---|
| Qt (Widgets, Quick, QuickControls2, QuickLayouts) | ≥ 6.5 | Builds the window chrome and hosts the WebEngine view |
| Qt WebEngine (Quick) | ≥ 6.5 | Renders the document HTML (`MarkdownViewer.qml`) — plain Qt Quick is not enough. Ships Chromium ≥ 108, which the text navigation needs for the CSS Custom Highlight API (Chromium ≥ 105) that paints the marked range |
| pkg-config | — | miqt uses it to locate the Qt6 modules (`.pc` files); o-mark also uses it for `omarchy-theme` |
| omarchy-theme | 1.0.0-beta.2 (`libomarchy_theme.so.1`) | Resolves the active Omarchy color palette at runtime. Install from [bencchio/omarchy-theme](https://github.com/bencchio/omarchy-theme); the GitHub release tag `0.3.0` is how the library is installed, not the ABI number |
| CGO | — | Required by miqt |

Supported distributions: Arch Linux (rolling — keep the system up to date with
`pacman -Syu`) and Ubuntu 24.04 LTS. On Arch, see [README.md § Install](../README.md)
for the 0.7.0 package or `makepkg -si` from that tag. `scripts/install.sh`
checks these dependencies for you.

## Go modules

| Module | Version | Why it's needed |
|---|---|---|
| `github.com/mappu/miqt` | v0.14.0 | Qt6 bindings for Go |
| `github.com/yuin/goldmark` | v1.8.6 | Parses markdown and renders it to HTML |
| `github.com/BurntSushi/toml` | v1.6.0 | Reads `config.toml` |

## Bundled JavaScript

These are embedded into the binary at build time, so O'Mark renders everything
offline and never contacts a CDN.

| Library | Version | Why it's needed |
|---|---|---|
| KaTeX | 0.18.4 | Renders `$inline$` and `$$block$$` math |
| Mermaid | 11.17.0 | Renders ` ```mermaid ` diagrams |
| highlight.js | 11.12.0 | Syntax highlighting in code blocks (common bundle) |
| highlightjs-line-numbers.js | 2.9.1 | Line numbers in code blocks |
