# Configuration

O'Mark reads its options from `~/.config/o-mark/config.toml`, created on first
run by copying the embedded default. You can edit it at any time; the two keys
that a session changes (`viewer_theme` and `toolbar_visible`) are written back
on quit, and the rest of the file — including your comments — is preserved.

| Key | Default | Values | Notes |
|-----|---------|--------|-------|
| `viewer_theme` | `"github"` | `"system"`, `"github"`, `"writer"`, `"night"`, `"sepia"`, `"mono"`, `"minimal"` | Overwritten on quit with the theme active at the end of the session. |
| `document_max_width` | `"default"` | `"default"`, or a CSS length (`210mm`, `50ch`, `80%`, …) | `"default"` uses the width each theme defines. |
| `font_size_base` | `"default"` | `"default"`, or a size in `px` (`"16"` or `"17px"`) | `"default"` uses the size each theme defines. |
| `show_scrollbars` | `false` | `true` / `false` | Show the native browser scrollbars in the `WebEngineView`. |
| `code_line_numbers` | `true` | `true` / `false` | Show line numbers on code blocks. Single-line blocks never show them. |
| `toolbar_position` | `"bottom"` | `"top"` / `"bottom"` | Position of the toolbar. |
| `toolbar_visible` | `false` | `true` / `false` | Toolbar visibility at launch. Overwritten on quit with its state at the end of the session (`Esc` toggles it). |

## Behavior on quit

When you close the app, O'Mark rewrites only `viewer_theme` and
`toolbar_visible` in `config.toml`, leaving every other line — comments
included — untouched. This is what lets you keep documentation notes in the
file itself.

## Reset

To restore the factory defaults, delete the file and relaunch:

```bash
rm ~/.config/o-mark/config.toml
```

It is regenerated from the embedded default on the next run.
