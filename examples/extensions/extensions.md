# Extensions

All extensions o-mark enables beyond CommonMark/dialect basics, except LaTeX
(`math.md`) and Mermaid (`mermaid.md`), which keep their own files. Each
section states whether goldmark actually parses the syntax (o-mark enables
`extension.Table`, `extension.DefinitionList`, `extension.TaskList`,
`extension.Strikethrough`, `extension.Footnote`, plus custom
subscript/superscript, highlight and math extensions — see
`src/internal/markdown.go`) or whether it's a documented gap that renders as
plain text.

---

## Tables

`extension.Table` (GFM tables): a basic grid, a column-alignment variant
(`:---`, `:--:`, `---:`), and a single-column edge case.

| Column A | Column B | Column C |
|----------|----------|----------|
| Cell 1   | Cell 2   | Cell 3   |
| Cell 4   | Cell 5   | Cell 6   |

With alignment:

| Left     |  Center  |    Right |
|:---------|:--------:|---------:|
| Lorem    |  Ipsum   |    Dolor |
| Sit amet | Consect. | Adipisc. |

Single column:

| Solo   |
|--------|
| Row 1  |
| Row 2  |

---

## Task lists

`extension.TaskList` (GFM task lists): checked and unchecked items should
render as disabled checkboxes, not literal `[x]`/`[ ]` text.

- [x] Completed task
- [ ] Pending task
- [x] Another completed task
- [ ] Final pending task

---

## Footnotes

`extension.Footnote`: reference markers should render as superscript links
that jump to a generated footnote list at the end of the document, including
one with inline code inside its definition.

Here is a sentence with a footnote.[^1] And another one.[^2]

[^1]: This is the first footnote content.
[^2]: This is the second footnote, with `inline code` inside.

---

## Autolinks

> **Gap:** o-mark does not enable `extension.Linkify` — bare URLs like
> www.example.com or https://example.com stay plain text. Only CommonMark
> angle-bracket autolinks work: <https://example.com>, <test@example.com>.

---

## Heading IDs

> **Gap:** o-mark does not enable automatic heading-ID generation (no
> `parser.WithAutoHeadingID` / GFM auto-id extension) — headings render
> without an `id` attribute, so in-page anchor links (`#heading-2-emphasis`)
> don't resolve to anything in the current DOM.

---

## Emoji shortcodes

> **Gap:** o-mark does not enable an emoji-shortcode extension — `:smile:`
> renders as the literal text `:smile:`, not 😄.

---

## Callouts / Admonitions

The 5 supported types: NOTE, TIP, WARNING, IMPORTANT, CAUTION.

> [!NOTE]
> This is a **note**. Use it for general information that is useful but not critical.

> [!TIP]
> This is a **tip**. Use it for helpful suggestions or best practices.

> [!WARNING]
> This is a **warning**. Use it to alert the user to potential issues.

> [!IMPORTANT]
> This is **important** information. Use it to highlight key points the user must not miss.

> [!CAUTION]
> This is a **caution**. Use it to warn about risks or dangerous actions.

Multi-paragraph admonition:

> [!NOTE]
> First paragraph of the note. Lorem ipsum dolor sit amet.
>
> Second paragraph. Consectetur adipiscing elit, sed do eiusmod tempor.

Mixed with a regular blockquote (no admonition style, keeps the grey left
border):

> This is just a regular blockquote.
> It should keep the grey left border.

> [!TIP]
> Admonition right after a regular blockquote.

Inline formatting inside admonitions:

> [!WARNING]
> You can use `inline code`, **bold**, *italic*, and [links](https://example.com) inside admonitions.

> [!CAUTION]
> - List item one
> - List item two
> - List item three

---

## Definition lists

`extension.DefinitionList`: a `Term` / `: Definition` pair, plus a term with
two stacked definitions, should render as `<dl>`/`<dt>`/`<dd>`.

Term
: Definition of the term with lorem ipsum dolor sit amet.

Another Term
: First definition line.
: Second definition line for the same term.

---

## Subscript and superscript

o-mark's custom `~text~`/`^text^` extension. Should render `<sub>`/`<sup>`
only when both delimiters of a pair are present; an unclosed marker must
stay literal text instead of swallowing the rest of the line.

Basic: Water molecule: H~2~O. Energy: E=mc^2^. Carbon-14: C~14~. Footnote-style: x^n+1^.

Wider coverage: H~2~O, CO~2~, C~6~H~12~O~6~, log~2~(n).

Superscript: E=mc^2^, x^n^, a^2^ + b^2^ = c^2^, (a+b)^n^.

Combined: H~2~^+^, x~i~^2^.

Adjacent: a~1~a~2~a~3~, x^1^x^2^x^3^.

Edge cases — unclosed markers should NOT render as sub/superscript:

Unclosed tilde: a~b.

Unclosed caret: a^b.

Escapes: \~\~not strikethrough\~\~.

In different contexts:

- List item with **bold** and H~2~O
- List item with ==highlight== and E=mc^2^

> Blockquote with **bold**, *italic*, H~2~O and ==highlight==.

| Element     | Example  |
|-------------|----------|
| Subscript   | H~2~O    |
| Superscript | mc^2^    |

---

## Highlight

o-mark's custom `==text==` extension, rendering a `<mark>`-style background.
A lone unmatched `==` at the end must stay literal, not open an unclosed
highlight that swallows the rest of the document.

==Highlighted text== with ==lorem ipsum== and normal text in between.

==Highlight with **bold** inside== and ==with *italic* inside==.

Several marks: ==one== two ==three== four ==five==.

Edge case — a lone `==` with nothing to close should NOT render as highlight: ==.

---

## Combinations

Extensions layered together — checks that the custom subscript/superscript
and highlight parsers compose correctly with CommonMark's own emphasis and
code-span parsing instead of conflicting with them.

**bold** + ==highlight==: **==bold and highlighted text==**.

*italic* + ~~strikethrough~~: *~~italic strikethrough~~*.

`code` + ~~strikethrough~~ around it: ~~`code`~~.

---

## Wiki-links, embeds, tags — Obsidian app-level syntax

> **Gap:** o-mark parses only what goldmark understands. Obsidian's
> app-level syntax below is not registered as an extension, so it renders as
> literal text, not as an interactive link/embed/tag.

Wiki-link: `[[Some Page]]` stays literal text.

Embed: `![[gradient.png]]` stays literal text (use standard `![alt](path)`
syntax instead — see `../markdown/images.md`).

Tag: `#project-tag` stays literal text (not a clickable tag).

---

## Frontmatter (YAML)

> **Known issue:** o-mark's frontmatter handling (`extractFrontmatter` in
> `src/internal/postprocess.go`) does not use a YAML parser — it extracts the
> raw block between `---` fences and displays it as-is inside a collapsible
> `<details>` element, without validating or interpreting YAML syntax.

```
---
title: Example Document
tags: [markdown, test]
date: 2026-08-21
---
```
