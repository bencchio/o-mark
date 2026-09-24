# Configuration

O'Mark reads its options from `~/.config/o-mark/config.toml`, created on first
run by copying the embedded default. You can edit it at any time; the two keys
that a session changes (`viewer_theme` and `toolbar_visible`) are written back on
quit, and the rest of the file — including your comments — is preserved.

| Key | Default | Values | Notes |
|-----|---------|--------|-------|
| `viewer_theme` | `"github"` | `"system"`, `"github"`, `"writer"`, `"night"`, `"sepia"`, `"mono"`, `"minimal"` | Overwritten on quit with the theme active at the end of the session. |
| `page_format` | `"a4"` | `"a3"`, `"a4"`, `"a5"` | Size of the page sheet, on screen and in the exported PDF. A document can declare its own in its front matter. |
| `page_orientation` | `"portrait"` | `"portrait"` / `"landscape"` | Default orientation of the sheet. A document can declare its own; `Ctrl+R` flips it for the session only. Changes only when you edit it here. |
| `pdf_margin_vertical` | `"25mm"` | a length in `mm`, up to `50mm` | Margin above and below the content of every page of the exported PDF. |
| `pdf_margin_horizontal` | `"10mm"` | a length in `mm`, up to `50mm` | Margin left and right of the content of every page of the exported PDF. It does not change the screen. |
| `zoom_default` | `1.0` | `0.5` to `2.0` | Zoom at launch, and what `Ctrl+0` and the toolbar reset return to. It is not saved between sessions. |
| `font_size_base` | `"default"` | `"default"`, or a size in `px` (`"16"` or `"17px"`) | `"default"` uses the size each theme defines. |
| `show_scrollbars` | `false` | `true` / `false` | Show the native browser scrollbars in the `WebEngineView`. |
| `code_line_numbers` | `true` | `true` / `false` | Show line numbers on code blocks. Single-line blocks never show them. |
| `toolbar_position` | `"bottom"` | `"top"` / `"bottom"` | Position of the toolbar. |
| `toolbar_visible` | `false` | `true` / `false` | Toolbar visibility at launch. Overwritten on quit with its state at the end of the session (`Esc` toggles it). |

## Behavior on quit

When you close the app, O'Mark rewrites only `viewer_theme` and `toolbar_visible` in
`config.toml`, leaving every other line — comments included — untouched. This is
what lets you keep documentation notes in the file itself. The page format and
orientation are defaults: they change only when you edit them, never because of
what you did in a session.

An older config that lacks some keys gets them on quit, each one with its
comment and its default value, so the file documents every option. Keys you
already have, and your own lines, are never touched.

## Upgrading from `document_max_width`

The old `document_max_width` key is gone: the page size belongs to the format,
not to a reading width. A config that still has it opens as landscape when the
value is `"297mm"` and as portrait for anything else; the next quit removes the
key and writes `page_orientation` (and `page_format`) in its place, once.

## Reset

To restore the factory defaults, delete the file and relaunch:

```bash
rm ~/.config/o-mark/config.toml
```

It is regenerated from the embedded default on the next run.
