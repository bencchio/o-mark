# Code Blocks

Prueba de bloques de código con distintos lenguajes, longitud y contenido especial.

---

## Go

```go
package main

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct {
	Theme    string
	FontSize int
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("cannot read %s: %w", filepath.Base(path), err)
	}
	_ = data
	return &Config{Theme: "github", FontSize: 14}, nil
}

func main() {
	cfg, err := LoadConfig(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Printf("theme=%s size=%d\n", cfg.Theme, cfg.FontSize)
}
```

---

## Python

```python
from dataclasses import dataclass
from pathlib import Path
import tomllib

@dataclass
class Config:
    theme: str = "github"
    font_size: int = 14

def load_config(path: Path) -> Config:
    with path.open("rb") as f:
        data = tomllib.load(f)
    return Config(**data)

if __name__ == "__main__":
    cfg = load_config(Path("~/.config/o-mark/config.toml").expanduser())
    print(f"theme={cfg.theme} size={cfg.font_size}")
```

---

## Bash

```bash
#!/usr/bin/env bash
set -euo pipefail

OMARK_BIN="./src/target/o-mark"
EXAMPLES_DIR="./docs/examples"

for f in "$EXAMPLES_DIR"/*.md; do
    echo "Opening: $f"
    "$OMARK_BIN" "$f" &
    sleep 0.5
done

wait
echo "All done."
```

---

## JSON

```json
{
  "id": "github",
  "label": "GitHub",
  "bg": "#ffffff",
  "palette": {
    "background": "#ffffff",
    "foreground": "#24292e",
    "accent": "#0366d6",
    "surface": "#f6f8fa",
    "border": "#dfe2e5"
  }
}
```

---

## TOML

```toml
# ~/.config/o-mark/config.toml
viewer_theme = "github"
document_max_width = "210mm"
font_size_base = 14
show_scrollbars = false
```

---

## CSS

```css
:root {
  --o-mark-bg: #ffffff;
  --o-mark-fg: #24292e;
  --o-mark-accent: #0366d6;
  --o-mark-font: Inter, "Helvetica Neue", sans-serif;
}

body {
  font-family: var(--o-mark-font);
  font-size: 16px;
  line-height: 1.6;
  color: var(--o-mark-fg);
  background-color: var(--o-mark-bg);
  max-width: 210mm;
  margin: 8mm auto;
}

code {
  font-family: "JetBrains Mono", monospace;
  background-color: var(--o-mark-surface);
  padding: 2px 5px;
  border-radius: 4px;
}
```

---

## Código inline en contexto

Usar `os.ReadFile` en lugar de `ioutil.ReadFile`. El tipo de retorno es `([]byte, error)`.

Ejecutar con `go build -o target/o-mark .` desde el directorio `src/`.

La función `filepath.Join(dir, name)` no escapa `../` — usar `filepath.EvalSymlinks` + `filepath.Rel` para validar.

---

## Bloque vacío y sin lenguaje

Sin lenguaje especificado:

```
Este bloque no tiene resaltado de lenguaje.
Debe mostrarse en fuente monoespaciada.
```

Vacío:

```go
```

---

## Caracteres especiales

```
<html> & "quotes" & 'apostrophes'
a < b > c && d || e
{{ template }} {% block %} {# comment #}
```

```diff
- línea eliminada
+ línea añadida
  línea sin cambios
```
