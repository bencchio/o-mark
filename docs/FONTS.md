# Fonts

These fonts are recommended for the best visual experience with O'Mark themes.
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

## Per-theme font requirements

| Theme   | Primary font     | Package (Arch)        | Package (Ubuntu)       |
| ------- | ---------------- | --------------------- | ---------------------- |
| github  | Inter            | `inter-font`          | `fonts-inter`          |
| night   | Inter            | — (same as above)     | —                      |
| minimal | Inter            | — (same as above)     | —                      |
| writer  | Liberation Serif | (bundled with system) | (bundled with system)  |
| sepia   | Source Serif 4   | `ttf-source-serif-4`  | `fonts-source-serif-4` |
| mono    | JetBrains Mono   | `ttf-jetbrains-mono`  | `fonts-jetbrains-mono` |

**Note:** Liberation Serif is included in `libreoffice` or `liberation-fonts`.
It is typically already installed on most Linux systems.
