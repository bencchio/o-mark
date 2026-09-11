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
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -o target/o-mark .
```

`CGO_CXXFLAGS` silences Qt/g++ warnings from miqt (`-Wsfinae-incomplete`). `scripts/install.sh` sets the same flag.

QML files, the default config, and the bundled JavaScript libraries (KaTeX,
Mermaid, highlight.js) are all embedded at compile time, so the resulting
binary is self-contained and works offline.

### Release build

Add `-ldflags="-s -w"` to strip the symbol table and DWARF debug info —
roughly a third smaller, with no behavior change:

```bash
CGO_CXXFLAGS="-Wno-sfinae-incomplete" go build -ldflags="-s -w" -o target/o-mark .
```

Keep the default (unstripped) build for local development — stripping
loses the ability to symbolize a crash. The Arch package
(`packaging/arch/PKGBUILD`) doesn't use this flag directly: it builds with
full debug info and relies on `options=('debug')` so `makepkg` splits it
into a separate `o-mark-debug` package, stripping only the installed
binary — see `diagnose-crash` when you need the symbols back.

## Run

```bash
./target/o-mark <file.md>
./target/o-mark --version
```

## Testing

```bash
cd src
go test ./...
go vet ./...
```

The suite in `src/internal/*_test.go` covers math, `$` flanking, code-fence
detection, post-load injection, highlight CSS order, config decoding and theme
invariants. The files under `examples/` double as manual rendering checks —
open each with the app to validate output the automated tests cannot assert.

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
