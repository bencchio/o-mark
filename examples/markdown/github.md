# GitHub Flavored Markdown — Basic and Advanced

Validates GFM coverage per the [markdown.org compatibility table](https://markdown.org/compatibility/):
basic GFM plus strikethrough (`✓` in GFM), and advanced GFM **without** the
colour/underline/comments/accordion passthrough-HTML tricks (`✕` in GFM —
those are CommonMark-only via raw HTML, see `commonmark.md`). No other
extensions here — tables, task lists, footnotes belong to
`../extensions/extensions.md`.

---

## Basic — Emphasis and strikethrough

**Bold**, *italic*, ***bold and italic***, ~~strikethrough~~ (GFM adds this
on top of CommonMark).

Strikethrough combined: ~~**bold strikethrough**~~ and ~~*italic
strikethrough*~~.

---

## Basic — Autolinks

GFM extended autolinks (bare URLs, no angle brackets required):

www.example.com

https://example.com/path?query=1

user@example.com

> **Gap:** o-mark does not enable goldmark's `extension.Linkify` — bare URLs
> above render as plain text, not as clickable links. Angle-bracket autolinks
> (`<https://example.com>`) still work since those are CommonMark, not GFM.

---

## Basic — Line breaks

GFM treats a single trailing newline plus two spaces as a hard break, same
as CommonMark:
Line one.  
Line two.

---

## Advanced — Nested lists with mixed markers

- Unordered item
  1. Ordered nested
  2. Ordered nested
     - Unordered nested again
- Back to top level

---

## Advanced — Blockquote with lazy continuation

> GFM keeps CommonMark's lazy continuation rule
this line without `>` should still be part of the blockquote.

---

## Advanced — What GFM does NOT add

The following CommonMark-adjacent tricks are **not** part of GFM and should
render as literal text or be ignored, not styled:

Colour syntax: `` `#ff0000` `` stays a code span, not a coloured swatch.

Underline: GFM has no native underline syntax (`__text__` renders as bold,
not underline).

HTML comments `<!-- like this -->` are stripped by the HTML sanitizer/pass,
not rendered as an accordion or hidden note.
