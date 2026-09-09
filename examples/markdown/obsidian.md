# Obsidian Flavored Markdown — Basic and Advanced

Validates Obsidian-dialect coverage per the [markdown.org compatibility
table](https://markdown.org/compatibility/): basic Obsidian plus
strikethrough, and advanced Obsidian **with** colour support (`✓` in
Obsidian, unlike GFM). Obsidian-specific extensions (wiki-links, embeds,
tags, callouts) live in `../extensions/extensions.md` since o-mark doesn't
implement Obsidian's app-level features — only what goldmark actually
parses.

---

## Basic — Emphasis and strikethrough

**Bold**, *italic*, ***bold and italic***, ~~strikethrough~~ (Obsidian
supports this, same as GFM).

---

## Advanced — Colour (via inline HTML passthrough)

Obsidian supports colour highlighting through its own syntax
(`==text==` plus a colour picker) and via inline HTML/CSS. o-mark exposes
colour only through the same raw-HTML passthrough as `commonmark.md`:

<span style="color: #d73a49;">Red text via inline HTML span.</span>

<span style="color: #22863a;">Green text via inline HTML span.</span>

> **Gap:** o-mark's `==highlight==` extension (see
> `../extensions/extensions.md`) renders a highlight background, not
> Obsidian's per-instance colour picker output — colour here only works via
> raw HTML passthrough, same mechanism as `commonmark.md`.

---

## Advanced — Nested blockquotes

> Level one.
> > Level two, nested inside level one.
> > > Level three.

---

## Advanced — Line breaks

Obsidian, like GFM, treats a trailing double-space as a hard break:
Line one.  
Line two.
