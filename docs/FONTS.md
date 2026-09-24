# Fonts

These fonts are recommended for the best visual experience with O'Mark.
Without them, the viewer falls back to Liberation, DejaVu, and Noto fonts,
which are available on most Linux distributions.

## Arch Linux

```
sudo pacman -S inter-font ttf-jetbrains-mono tex-gyre-fonts adobe-source-serif-fonts
```

## Ubuntu / Debian

```
sudo apt install fonts-inter fonts-source-serif-4 fonts-jetbrains-mono tex-gyre-fonts
```

## Choosing the font

The `font` key of `config.toml` sets the font of the document: `"mono"` (the
default) is the Omarchy mono font, and any CSS font stack works.

| `font` | Package (Arch) | Package (Ubuntu) |
| ------ | -------------- | ---------------- |
| `"mono"` | whatever font your Omarchy theme uses | — |
| `"Inter, sans-serif"` | `inter-font` | `fonts-inter` |
| `"JetBrains Mono, monospace"` | `ttf-jetbrains-mono` | `fonts-jetbrains-mono` |
| `"Liberation Serif, serif"` | `liberation-fonts` | bundled with the system |

Liberation Serif is included in `libreoffice` or `liberation-fonts`, and it is
typically already installed on most Linux systems. If none of the fonts in your
stack is installed, O'Mark logs a warning and the browser falls back to its own
default.
