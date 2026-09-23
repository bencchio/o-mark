# Manual

A quick guide to using O'Mark. It is a keyboard-first, CLI-driven markdown
viewer: you open a file, and the document renders in the window.

## Opening a document

```bash
o-mark document.md
```

Pass no argument to open an empty viewer. Links to other `.md` files in a
document open a new O'Mark instance (requires the delegation registered by
`scripts/install.sh --register-mime`, see [INSTALL.md § 3](../INSTALL.md)).

## Reading

`PgUp` / `PgDn` scroll the document. The arrow keys and `Home` / `End` move a
word cursor through the text:

- `←` / `→` move by word; `↑` / `↓` by visual line.
- `Ctrl+←` / `Ctrl+→` move by sentence; `Ctrl+↑` / `Ctrl+↓` by paragraph.
- `Home` / `End` jump to the first/last word of the line; `Ctrl+Home` /
  `Ctrl+End` to the start/end of the document.

Press `Space` at a word to set an anchor and extend a range; `Super+C` copies
the marked text. A mouse drag selects normally. See
[docs/KEYBOARD-SHORTCUTS.md](KEYBOARD-SHORTCUTS.md) for the full list.

## Page separators

A line with only `---`, `***` or `___` starts a new page instead of drawing a
horizontal rule: the document renders as separate pages, marked by a gap on
screen, and each page prints (or exports to PDF) on its own physical page. A
separator inside a fenced code block or a Mermaid diagram is left alone — it
never splits the page. A document with no such line keeps rendering as one
continuous document, unchanged. There is no keyboard navigation between
pages: it stays a continuous document with marked pages, not a slide viewer.

## Export to PDF

`Ctrl+P` (or the `PDF` control in the toolbar) writes a print-colored PDF
next to the open file (`document.md` → `document.pdf`). If that PDF already
exists, it is left alone. The toolbar is hidden for the capture.

## Zoom

`Ctrl++` / `Ctrl+-` zoom in and out, `Ctrl+0` resets to the default. The
toolbar shows the current zoom percentage.

## Toolbar and themes

`Esc` toggles the toolbar; `T` / `Z` jump to its controls. The toolbar holds
the theme selector, the zoom indicator, PDF export and the current file name.
Themes and their customization live in [docs/THEMES.md](THEMES.md).

## Live reload

The open document, `config.toml`, and every theme file are watched. Editing
and saving any of them re-renders the viewer in place — you see a "Reloaded"
badge when the document itself changes.

## Configuration

`~/.config/o-mark/config.toml` controls the starting theme, document width,
font size, scrollbars, line numbers and toolbar position/visibility. Full
reference in [docs/CONFIG.md](CONFIG.md).

## FAQ

**Remote images don't show.**
Only local images are inlined as `data:` URLs. Remote (`http`/`https`) URLs
are rendered as a placeholder with the alt text — O'Mark never makes network
requests.

**A `.md` link does not open a new instance.**
That behavior needs the URL handler registered with
`scripts/install.sh --register-mime`.

**Bare URLs in text are not clickable.**
The Markdown parser has no `Linkify` extension, so a URL needs link syntax
(`[text](url)`) to become a link.

**Arrow keys don't scroll.**
Arrows move the word cursor; `PgUp` / `PgDn` scroll the document.
