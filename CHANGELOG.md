# Changelog

High-level history of O'Mark releases, summarized for the end user.

## 0.7.0 — First Arch package on GitHub

- An Arch package (`o-mark-0.7.0-1-x86_64.pkg.tar.zst`) ships with the
  GitHub release. Install [omarchy-theme](https://github.com/bencchio/omarchy-theme)
  first, then `pacman -U` the package or build tag `0.7.0` with `makepkg`.

## 0.6.x — Omarchy theme as a library

- The System theme follows the active Omarchy palette through
  `libomarchy_theme` instead of parsing `colors.toml`.
- Palette changes apply live via the library's signal, without a file
  watcher on the theme files. Font still follows the Omarchy CSS file.
- Requires the `omarchy-theme` system library (`1.0.0-beta.2`, SONAME
  `libomarchy_theme.so.1`).

## 0.5.x — Code blocks, navigation, reading

- Syntax-highlighted code blocks with line numbers and long-line wrapping,
  colored per theme.
- Block `$$...$$` math that no longer corrupts LaTeX via inline parsing.
- Updated bundled libraries: Mermaid 11, KaTeX 0.18, highlight.js 11.12.
- Helix-style navigation: word cursor, sentence/paragraph/line movement,
  range marking from an anchor and `Super+C` copy.
- Live reload and theme persistence: a hand-edited theme or the config
  comments survive a relaunch.
- A first automated test suite for the renderer and config.
- User documentation (keyboard shortcuts, dependencies, fonts) and a
  reorganized set of examples by dialect.

## 0.4.x — Stabilization and hardening

- A dedicated navigation layer and per-theme accent colors.
- Security and path-handling fixes around image resolution and link
  filtering.

## 0.3.x — Themes and features

- The six built-in viewer themes plus the System theme from the Omarchy
  palette, extensible with custom CSS.
- Local images inlined as `data:` URLs; collapsible YAML frontmatter.

## 0.2.x — Themes, toolbar, navigation

- Theme support, a toolbar, and fuller keyboard navigation.

## 0.1.x — The basics

- A window that renders markdown to styled HTML with scrolling, keyboard
  control, and support for CommonMark plus GitHub and Obsidian dialects.
