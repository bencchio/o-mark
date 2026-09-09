# Building O'Mark

## Requirements

- Go 1.26+ (see the `go` directive in `src/go.mod`)
- Qt ≥ 6.5 development libraries (`qt6-base-devel` / `libqt6-dev`)
- Qt6 Quick + Controls (`qt6-declarative-devel` / `libqt6quick6`)
- Qt6 WebEngine (`qt6-webengine-devel` / `libqt6webenginequick6`)
- CGO enabled (default)

## Build

```bash
cd src
go build -o target/o-mark .
```

QML files, the default config, and the bundled JavaScript libraries (KaTeX,
Mermaid, highlight.js) are all embedded at compile time, so the resulting
binary is self-contained and works offline.

## Run

```bash
./target/o-mark <file.md>
./target/o-mark --version
```

## Version

The version string is defined in `src/main.go` (`var version`), kept in
sync with the `VERSION` file. To override at build time:

```bash
go build -ldflags "-X main.version=0.3.0" -o target/o-mark .
```

## Clean

```bash
rm -rf src/target/
```
