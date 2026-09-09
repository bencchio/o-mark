# CommonMark — Basic and Advanced

Validates pure CommonMark compliance (not GFM/Obsidian extensions — those are
covered in `github.md` and `obsidian.md`; general extensions live in
`../extensions/extensions.md`). Each section notes what's expected; use it to
spot differences between the spec and what goldmark actually renders. No
strikethrough here (`✕` in CommonMark) — see `github.md`/`obsidian.md`.

---

## Thematic breaks

Three or more `-`, `*`, or `_` (with or without spaces between them) form a
thematic break (`<hr>`):

---

***

___

- - -

* * *

Fewer than three is NOT a thematic break (stays a list or plain text):

- -

---

## ATX headings

# H1
## H2
### H3
#### H4
##### H5
###### H6

With optional closing `#` (must not appear in the text):

## H2 with closing sequence ##

More than 6 `#` is NOT a heading — should stay a paragraph with literal `#`:

####### This is not a heading, it's 7 hashes

Empty heading (valid, no content):

##

---

## Setext headings

Setext heading, level 1
========================

Setext heading, level 2
------------------------

---

## Backslash escapes

Escaping special characters: \*not italic\*, \# not a heading, \[not a
link\](url), literal backslash: \\

---

## Entity references

Ampersand: `&amp;` → &amp; — Quote: `&quot;` → &quot; — Decimal numeric:
`&#35;` → &#35; — Hex numeric: `&#x22;` → &#x22;

---

## Code spans

Single backtick: `code`

Double backticks to include a literal backtick inside: `` `backtick` ``

Leading/trailing space gets trimmed (should not show as extra padding):
` code `

---

## Emphasis and strong emphasis

*Italic with asterisk* vs _italic with underscore_

**Bold with asterisk** vs __bold with underscore__

***Bold and italic combined***

Nested emphasis: *italic with **bold** inside*

Intraword — CommonMark rule: `foo*bar*baz` should trigger emphasis (asterisk
works intraword), but `foo_bar_baz` should NOT (underscore intraword doesn't
trigger, per the flanking rule):

foo*bar*baz

foo_bar_baz

---

## Links

Inline: [inline link](https://example.com "optional title")

Full reference: [reference text][ref1]

Collapsed reference: [ref2][]

Shortcut reference: [ref3]

Autolink: <https://example.com>

Email autolink: <test@example.com>

[ref1]: https://example.com/ref1 "ref1 title"
[ref2]: https://example.com/ref2
[ref3]: https://example.com/ref3

---

## Hard vs soft line breaks

Line with two trailing spaces (hard break, should cut with `<br>`):
This line should start fresh.

Line with no trailing spaces (soft break, should join into the same
paragraph as one text block):
This line should stay glued to the previous one.

---

## Blockquote with lazy continuation

> First line of the blockquote with explicit `>`
this line without `>` should still be part of the same blockquote (lazy
continuation per CommonMark).

---

## Lists — tight vs loose

Tight list (no blank lines between items — items should NOT be wrapped in
`<p>`):

- Item one
- Item two
- Item three

Loose list (blank lines between items — each item SHOULD be wrapped in
`<p>`):

- Item one

- Item two

- Item three

---

## Marker change — should split into two lists

- Item with dash (list A)

+ Item with plus (list B, new — marker change splits the list)

* Item with asterisk (list C, new)

---

## Ordered list with a start number other than 1

3. Starts at 3
4. Continues at 4
5. Continues at 5

---

## Indented code (4 spaces, no fences)

    this is a code block via indentation
    no need for ```fences```

---

## Raw HTML block (passthrough)

<div style="border: 1px solid; padding: 4px;">
This HTML block should pass through as-is, not get escaped as text.
</div>

---

## Paragraph with multiple lines (soft-wrapped)

This is one line.
This is another line in the same paragraph.
And a third line, still the same paragraph.

This is a separate paragraph, split by a blank line.
