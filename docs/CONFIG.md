# Configuration

O'Mark reads its options from `~/.config/o-mark/config.toml`, created on first
run by copying the embedded default. You can edit it at any time; the keys
that a session changes (`viewer_theme`, `toolbar_visible` and `page_orientation`) are written back
on quit, and the rest of the file — including your comments — is preserved.

| Key | Default | Values | Notes |
|-----|---------|--------|-------|
| `viewer_theme` | `"github"` | `"system"`, `"github"`, `"writer"`, `"night"`, `"sepia"`, `"mono"`, `"minimal"` | Overwritten on quit with the theme active at the end of the session. |
| `page_format` | `"a4"` | `"a4"` | Size of the page sheet, on screen and in the exported PDF. A4 is the only format for now. |
| `page_orientation` | `"portrait"` | `"portrait"` / `"landscape"` | Sheet orientation at launch. Overwritten on quit with the orientation active at the end of the session (`Ctrl+R` flips it). |
| `font_size_base` | `"default"` | `"default"`, or a size in `px` (`"16"` or `"17px"`) | `"default"` uses the size each theme defines. |
| `show_scrollbars` | `false` | `true` / `false` | Show the native browser scrollbars in the `WebEngineView`. |
| `code_line_numbers` | `true` | `true` / `false` | Show line numbers on code blocks. Single-line blocks never show them. |
| `toolbar_position` | `"bottom"` | `"top"` / `"bottom"` | Position of the toolbar. |
| `toolbar_visible` | `false` | `true` / `false` | Toolbar visibility at launch. Overwritten on quit with its state at the end of the session (`Esc` toggles it). |

## Behavior on quit

When you close the app, O'Mark rewrites only `viewer_theme`, `toolbar_visible`,
`page_format` and `page_orientation` in `config.toml`, leaving every other line —
comments included — untouched. This is what lets you keep documentation notes in
the file itself.

## Upgrading from `document_max_width`

The old `document_max_width` key is gone: the page size belongs to the format,
not to a reading width. A config that still has it opens as landscape when the
value is `"297mm"` and as portrait for anything else; the next quit removes the
key and writes `page_format` and `page_orientation` in its place.

## Reset

To restore the factory defaults, delete the file and relaunch:

```bash
rm ~/.config/o-mark/config.toml
```

It is regenerated from the embedded default on the next run.
