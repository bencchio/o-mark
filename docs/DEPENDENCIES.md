# Dependencies — O'Mark

Dependencias de Go (módulos externos).

## Sistema

| dependencia | versión | propósito |
|-------------|---------|-----------|
| Qt (Widgets, Quick, QuickControls2, QuickLayouts) | ≥ 6.5 | Chromium ≥ 111 para `color-mix()` en CSS del tema System |
| Qt WebEngine (Quick) | ≥ 6.5 | Render del HTML del documento (`MarkdownViewer.qml`) — no basta con Qt Quick a secas |
| pkg-config | — | miqt usa `pkg-config` para localizar los módulos Qt6 (`.pc` files) |
| CGO | — | Requerido por miqt |

Distribuciones soportadas: Arch Linux (rolling — mantener el sistema actualizado con `pacman -Syu`) y Ubuntu 24.04 LTS. Ver `scripts/install.sh` para el chequeo automático de estas dependencias.

## Go modules

| módulo | versión | propósito |
|--------|---------|-----------|
| `github.com/mappu/miqt` | v0.14.0 | Bindings Qt6 para Go |
| `github.com/yuin/goldmark` | v1.8.2 | Parseo y renderizado de markdown a HTML |
| `github.com/BurntSushi/toml` | v1.6.0 | Parseo de colors.toml de Omarchy (temas) |
