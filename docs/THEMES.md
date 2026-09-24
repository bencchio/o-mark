# Theme and treatments

O'Mark shows the document with the colors of the active Omarchy theme. It
follows that theme live: switch theme in Omarchy and the viewer repaints, with
no restart. The viewer has no list of themes of its own.

## Treatments

On top of the theme, a **treatment** restates its colors. The toolbar has a
selector with every treatment the installed
[omarchy-lib-theme](https://github.com/bencchio/omarchy-lib-theme) names:

| Treatment | What it does |
|-----------|--------------|
| Original | The theme exactly as declared. |
| Inverted | The same identity with its polarity flipped: a dark theme becomes light and the other way round, each color keeping its hue. |
| HighContrast | Every color pushed further from the background, for readability. |
| Mono | The theme's own scale, from background to foreground, with no hue. |
| Print | White page, black text and greys. |

The names and their number come from the library, so a treatment it adds later
shows up in the selector with no change in O'Mark. Pick one with the mouse, or
press `T` to jump to the selector. Both the document and the toolbar take the
treatment's colors. The treatment active when you quit is saved as
`viewer_treatment` and used the next time.

The exported PDF always uses Print, whatever treatment is on screen.

Without a running Omarchy install the selector is hidden and the viewer uses a
light palette.

## Font

The document font is the `font` key of the config: `"mono"` (the default) is the
Omarchy mono font, and any CSS font stack works, such as
`"Liberation Serif, serif"`. `font_size` sets the base size. See
[docs/CONFIG.md](CONFIG.md) and [docs/FONTS.md](FONTS.md) for the recommended
packages.

## From the old themes

Earlier versions listed six CSS themes (GitHub, Writer, Night, Sepia, Mono and
Minimal) beside System. They are gone. The first launch moves the `.css` files
you had in `~/.config/o-mark/themes/` into `~/.config/o-mark/themes-old/`,
without deleting or overwriting anything, and a `viewer_theme` in your config is
dropped on the next quit. Custom CSS is not supported for now.
