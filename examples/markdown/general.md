# General Markdown

End-to-end integration example mixing basic and advanced markdown to see
several things at once. For per-format validation see `commonmark.md`,
`github.md`, `obsidian.md`, `images.md` and `../extensions/extensions.md`,
`../extensions/math.md`, `../extensions/mermaid.md`, `../code/code.md`.

Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor
incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis
nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.

## Heading 2 — Emphasis

Normal text, **bold text**, *italic text*, ***bold and italic***, ~~strikethrough~~.
Inline `code snippet` within a sentence. And a [link to example](https://example.com).

### Heading 3 — Lists

Unordered list:

- Item one with some lorem ipsum text to show wrapping behavior
- Item two
  - Nested item A
  - Nested item B
- Item three

Ordered list:

1. First item
2. Second item
   1. Nested ordered A
   2. Nested ordered B
3. Third item

#### Heading 4 — Blockquote

> Lorem ipsum dolor sit amet, consectetur adipiscing elit. Pellentesque habitant
> morbi tristique senectus et netus et malesuada fames ac turpis egestas.
>
> — Author Name

##### Heading 5 — Code Blocks

Inline code: `fmt.Println("hello, world")`

Fenced code block:

```go
package main

import "fmt"

func main() {
    for i := range 5 {
        fmt.Printf("Lorem ipsum %d\n", i)
    }
}
```

Shell example:

```bash
./src/target/o-mark examples/markdown/general.md
```

###### Heading 6 — Tables

| Column A       | Column B       | Column C       |
|----------------|----------------|----------------|
| Lorem ipsum    | Dolor sit amet | Consectetur    |
| Adipiscing     | Elit sed do    | Eiusmod tempor |
| Incididunt ut  | Labore et      | Dolore magna   |

Table with alignment:

| Left           | Center         | Right          |
|:---------------|:--------------:|---------------:|
| Lorem          | Ipsum          | Dolor          |
| Sit amet       | Consectetur    | Adipiscing     |

---

## Links and Images

External link: [OpenStreetMap](https://www.openstreetmap.org)

Local image (see `images.md` for the full image test matrix). External URLs
show the alt text as fallback:

![Alt text for image](https://via.placeholder.com/400x200)
