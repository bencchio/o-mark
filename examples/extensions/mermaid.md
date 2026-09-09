# Mermaid Diagrams

Validates that o-mark's `AssetPostLoad`-injected mermaid.js renders each of
the 9 diagram types the library supports: flowchart, sequence, class, state,
entity-relationship, gantt, git graph, mindmap, timeline. Each section is a
minimal but non-trivial example of that diagram's syntax.

---

## Flowchart

A decision tree with branching edges — checks node shapes (`[]`, `{}`) and
labelled edges (`-->|Yes|`):

```mermaid
graph TD
    A[Start] --> B{Condition?}
    B -->|Yes| C[Step A]
    B -->|No| D[Step B]
    C --> E[End]
    D --> E
```

---

## Sequence Diagram

Message passing between named participants, including a synchronous call, a
self-call, and a dashed return arrow:

```mermaid
sequenceDiagram
    participant U as User
    participant A as App
    participant S as Server

    U->>A: Opens document
    A->>S: GET /file.md
    S-->>A: 200 OK (markdown)
    A->>A: Renders HTML
    A-->>U: Shows document
```

---

## Class Diagram

Two classes with typed fields and a relationship arrow between them:

```mermaid
classDiagram
    class ThemePalette {
        +string Background
        +string Foreground
        +string Accent
        +string Surface
        +string Border
    }
    class ViewerTheme {
        +string ID
        +string Label
        +string Bg
        +string HTML
    }
    ThemePalette --> ViewerTheme : generates
```

---

## State Diagram

A state machine with entry/exit points (`[*]`) and labelled transitions:

```mermaid
stateDiagram-v2
    [*] --> Idle
    Idle --> Loading : open file
    Loading --> Rendered : parse ok
    Loading --> Error : parse fails
    Rendered --> Idle : close file
    Error --> Idle : dismiss
    Rendered --> [*]
```

---

## Entity Relationship Diagram

Entities with cardinality-annotated relationships and one entity's attribute
list expanded:

```mermaid
erDiagram
    DOCUMENT ||--o{ IMAGE : embeds
    DOCUMENT ||--o| FRONTMATTER : has
    THEME ||--o{ DOCUMENT : styles
    DOCUMENT {
        string path
        string title
    }
    THEME {
        string id
        string label
    }
```

---

## Gantt Chart

A project timeline with dated sections and task states (`done`, `active`,
pending):

```mermaid
gantt
    title Cycle v0.5
    dateFormat  YYYY-MM-DD
    section Features
    Code blocks & examples   :done, 2026-08-01, 5d
    Syntax highlighting      :active, 2026-08-06, 4d
    Reading improvements     :2026-08-10, 3d
    section Stabilization
    Known issues cleanup     :2026-08-13, 3d
```

---

## Git Graph

A branch, feature commits, and a merge back into `main` — checks branching
and merge rendering:

```mermaid
gitGraph
    commit id: "init"
    branch feature
    checkout feature
    commit id: "add examples"
    commit id: "add math known issues"
    checkout main
    merge feature
    commit id: "release v0.5.1"
```

---

## Mindmap

A root node branching into multiple levels of nested children:

```mermaid
mindmap
  root((o-mark))
    Markdown
      CommonMark
      GFM
      Obsidian
    Extensions
      Math
      Mermaid
      Admonitions
    Reading
      Themes
      Navigation
```

---

## Timeline

A chronological sequence of periods, one of them with multiple stacked
events:

```mermaid
timeline
    title o-mark release history
    v0.3 : Footnotes, subscript/superscript
         : Highlight, local images
         : Admonitions, math (KaTeX)
    v0.4 : Themes and Omarchy integration
    v0.5 : Code blocks and navigation
```
