# Themes

O'Mark ships seven viewer themes that control the look of the document area:
**System** plus six built-in CSS themes. You can also add your own.

## The seven themes

| Theme | Background | Notes |
|-------|-----------|-------|
| System | dynamic | Follows the active Omarchy palette (background, foreground, accent, code colors) and font. Requires a running Omarchy install; falls back to a light palette otherwise. |
| GitHub | `#ffffff` | GitHub-flavored: Inter font, explicit table header contrast. |
| Writer | `#e8e0d0` | Warm paper background, Liberation Serif body. |
| Night | dark | Dark theme with its own navigation/accent colors. |
| Sepia | sepia | Sepia paper, Liberation Serif body. |
| Mono | mono | JetBrains Mono throughout. |
| Minimal | minimal | Minimalist Inter-based theme. |

## Where themes live

On first run O'Mark exports its six built-in CSS files to
`~/.config/o-mark/themes/`. From then on, the file on disk is the source of
truth for that theme: a hand edit wins over the embedded version and survives
a relaunch. Deleting a file restores the built-in on the next launch.

The `System` theme is not a file — it is rendered on the fly from the Omarchy
palette and the current system font.

## Adding a custom theme

Drop a `.css` file into `~/.config/o-mark/themes/` and restart. The theme
appears in the selector with a label derived from the file name (e.g.
`mytheme.css` becomes "Mytheme"). To make it fit the viewer chrome, set the
variables other themes use (see an existing file for the full set), most
importantly:

```css
:root {
  --o-mark-bg: #ffffff;
  --o-mark-fg: #1f2328;
  --o-mark-accent: #0969da;
  --o-mark-surface: #f6f8fa;
  --o-mark-border: #d8dee4;
  --o-mark-code-bg: #f6f8fa;
  --o-mark-code-fg: #1f2328;
  --o-mark-link: #0969da;
  --o-mark-heading: #1f2328;
  --o-mark-selection-bg: #b6e3ff;
  --o-mark-selection-fg: #1f2328;
  --o-mark-font: ...;
  --o-mark-admonition-note: ...;
  --o-mark-admonition-tip: ...;
  --o-mark-admonition-warning: ...;
  --o-mark-admonition-important: ...;
  --o-mark-admonition-caution: ...;
}
```

A custom theme is picked up alongside the built-ins and does not need any
other registration.

## Fonts

Each theme names its preferred fonts with CSS fallbacks. Missing fonts fall
back to Liberation, DejaVu, and Noto. See [docs/FONTS.md](FONTS.md) for the
recommended packages per distribution.
