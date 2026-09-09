# Code Blocks

Fenced code blocks across 10 languages, plus edge cases (empty block, no
language, special characters). Syntax highlighting colors are out of scope
here — see v0.5.2.

---

## Bash

Shell-style comments, variables, quoting, and a glob loop — checks that `#`
inside a fenced block isn't mistaken for a heading:

```bash
#!/usr/bin/env bash
set -euo pipefail

OMARK_BIN="./src/target/o-mark"
EXAMPLES_DIR="./examples"

for f in "$EXAMPLES_DIR"/**/*.md; do
    echo "Opening: $f"
    "$OMARK_BIN" "$f" &
    sleep 0.5
done

wait
echo "All done."
```

---

## C

Preprocessor directives, a struct typedef, and pointer syntax — checks that
`#include`/`*` don't get parsed as markdown syntax inside the block:

```c
#include <stdio.h>
#include <stdlib.h>

typedef struct {
    char *theme;
    int font_size;
} Config;

Config *load_config(const char *path) {
    Config *cfg = malloc(sizeof(Config));
    cfg->theme = "github";
    cfg->font_size = 14;
    return cfg;
}

int main(int argc, char **argv) {
    if (argc < 2) {
        fprintf(stderr, "usage: %s <path>\n", argv[0]);
        return 1;
    }
    Config *cfg = load_config(argv[1]);
    printf("theme=%s size=%d\n", cfg->theme, cfg->font_size);
    free(cfg);
    return 0;
}
```

---

## C++

Templates-adjacent syntax (`std::optional`, `<>`) and `&`/`&&` references —
checks that angle brackets don't get treated as raw HTML:

```cpp
#include <iostream>
#include <optional>
#include <string>

struct Config {
    std::string theme = "github";
    int fontSize = 14;
};

std::optional<Config> loadConfig(const std::string &path) {
    if (path.empty()) {
        return std::nullopt;
    }
    return Config{};
}

int main(int argc, char **argv) {
    if (argc < 2) {
        std::cerr << "usage: " << argv[0] << " <path>\n";
        return 1;
    }
    auto cfg = loadConfig(argv[1]);
    if (!cfg) {
        std::cerr << "cannot load config\n";
        return 1;
    }
    std::cout << "theme=" << cfg->theme << " size=" << cfg->fontSize << "\n";
    return 0;
}
```

---

## Go

Tab-indented code (Go's `gofmt` convention) — checks that literal tab
characters inside a fenced block render as fixed indentation, not as
CommonMark's 4-space indented-code trigger:

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

## Rust

Lifetimes-adjacent syntax, `impl` blocks, and `?`/`match` — checks operators
that could clash with markdown's own emphasis/list syntax:

```rust
use std::fs;
use std::path::Path;

struct Config {
    theme: String,
    font_size: u32,
}

impl Default for Config {
    fn default() -> Self {
        Config { theme: "github".to_string(), font_size: 14 }
    }
}

fn load_config(path: &Path) -> Result<Config, std::io::Error> {
    let _data = fs::read_to_string(path)?;
    Ok(Config::default())
}

fn main() {
    let path = std::env::args().nth(1).expect("missing path");
    match load_config(Path::new(&path)) {
        Ok(cfg) => println!("theme={} size={}", cfg.theme, cfg.font_size),
        Err(e) => eprintln!("error: {e}"),
    }
}
```

---

## Python

Significant indentation and decorators (`@dataclass`) — checks that Python's
whitespace-sensitive syntax survives untouched inside the fence:

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

## JavaScript

Template literals with `${}` interpolation and arrow functions — checks that
backticks and braces inside the block don't get parsed as markdown code
spans:

```js
class Config {
  constructor(theme = "github", fontSize = 14) {
    this.theme = theme;
    this.fontSize = fontSize;
  }
}

async function loadConfig(path) {
  const raw = await fetch(path).then((r) => r.json());
  return new Config(raw.theme, raw.fontSize);
}

loadConfig("./config.json")
  .then((cfg) => console.log(`theme=${cfg.theme} size=${cfg.fontSize}`))
  .catch((err) => console.error("cannot load config:", err));
```

---

## TypeScript

Interfaces and type annotations (`: string`, `as Config`) — checks that the
`:` used by type syntax doesn't collide with definition-list syntax:

```ts
interface Config {
  theme: string;
  fontSize: number;
}

async function loadConfig(path: string): Promise<Config> {
  const response = await fetch(path);
  const raw: unknown = await response.json();
  return raw as Config;
}

loadConfig("./config.json").then((cfg: Config) => {
  console.log(`theme=${cfg.theme} size=${cfg.fontSize}`);
});
```

---

## Ruby

Symbols (`:theme`), blocks (`do...end`), and string interpolation (`#{}`) —
checks that `#{` isn't misread as a heading or comment:

```ruby
require "toml"

Config = Struct.new(:theme, :font_size) do
  def self.default
    new("github", 14)
  end
end

def load_config(path)
  data = TOML.load_file(path)
  Config.new(data["theme"], data["font_size"])
rescue Errno::ENOENT
  Config.default
end

cfg = load_config(ARGV[0])
puts "theme=#{cfg.theme} size=#{cfg.font_size}"
```

---

## Java

Verbose type declarations and checked exceptions (`throws IOException`) —
the longest lines in this file, useful for checking wrap/overflow behavior:

```java
import java.nio.file.Files;
import java.nio.file.Path;
import java.io.IOException;

public class Config {
    String theme = "github";
    int fontSize = 14;

    static Config load(String path) throws IOException {
        byte[] data = Files.readAllBytes(Path.of(path));
        Config cfg = new Config();
        return cfg;
    }

    public static void main(String[] args) throws IOException {
        Config cfg = Config.load(args[0]);
        System.out.printf("theme=%s size=%d%n", cfg.theme, cfg.fontSize);
    }
}
```

---

## Inline code in context

Use `os.ReadFile` instead of `ioutil.ReadFile`. The return type is
`([]byte, error)`.

Build with `go build -o target/o-mark .` from the `src/` directory.

`filepath.Join(dir, name)` does not escape `../` — use
`filepath.EvalSymlinks` + `filepath.Rel` to validate.

---

## Empty block and no language

A fence with no language tag after the backticks — should still render as
monospace code, just without a language annotation:

```
This block has no language for highlighting.
It should render in monospace font.
```

A fence with a language tag but no content between the backticks — should
render as an empty code block, not collapse or error:

```go
```

---

## Special characters

HTML-significant characters (`<`, `>`, `&`, quotes) and template-engine
delimiters (`{{ }}`, `{% %}`, `{# #}`) inside an untagged fence — should
render as literal text, not get interpreted as HTML or escaped twice:

```
<html> & "quotes" & 'apostrophes'
a < b > c && d || e
{{ template }} {% block %} {# comment #}
```

A `diff`-tagged block with `+`/`-` prefixes — checks that diff markers don't
get parsed as a markdown list inside the fence:

```diff
- removed line
+ added line
  unchanged line
```
