# Keyboard Shortcuts

O'Mark is built to be used from the keyboard. This page lists every shortcut.

## Reading

| Action | Keyboard | Mouse |
|---|---|---|
| Scroll one line | `↑` / `↓` | Wheel / trackpad |
| Scroll one page | `PgUp` / `PgDn` | — |
| Jump to the top | `Home` | — |
| Jump to the bottom | `End` | — |
| Move to the next / previous link | `Tab` / `Shift+Tab` | — |
| Open the focused link | `Enter` | Click the link |

Links to other `.md` files open in a new O'Mark window. This needs O'Mark
registered as your markdown handler — see [INSTALL.md](../INSTALL.md).

## Zoom

| Action | Keyboard | Mouse |
|---|---|---|
| Zoom in | `Ctrl++` or `Ctrl+=` | `Ctrl` + wheel up |
| Zoom out | `Ctrl+-` | `Ctrl` + wheel down |
| Reset zoom to 100% | `Ctrl+0` | Click `↺` in the toolbar |

The current zoom level is shown in the toolbar.

## Quitting

| Action | Keyboard |
|---|---|
| Quit O'Mark | `Ctrl+Q` |

## The toolbar

The toolbar is hidden by default. `Esc` is how you reach it, and it works as a
simple on/off switch:

- Toolbar hidden → `Esc` → toolbar appears and takes keyboard focus
- Toolbar visible → `Esc` → toolbar hides and focus returns to the document

While the toolbar has focus, a coloured line between it and the document tells
you that keys go to the toolbar rather than to the page. O'Mark remembers which
control you were on the last time, so `Esc` brings you back to it.

The toolbar holds two controls: the zoom reset button (`↺`) and the theme
selector.

### Moving around the toolbar

| Action | Keyboard |
|---|---|
| Go straight to the theme selector | `T` |
| Go straight to the zoom reset `↺` | `Z` |
| Move between the two controls | `←` / `→` |
| Reset zoom (when `↺` is selected) | `Enter` or `Space` |

### Changing the theme

With the theme selector focused:

| Action | Keyboard |
|---|---|
| Switch theme without opening the list | `↑` / `↓` |
| Open the theme list | `Space` or `Enter` |
| Move through the list | `↑` / `↓` |
| Pick the highlighted theme | `Enter` |
| Close the list, keeping the current theme | `Esc` |

Closing the list with `Esc` leaves the toolbar open — you need a second `Esc` to
go back to the document.

The theme you are using when you quit becomes the one O'Mark starts with next
time. See [the configuration section of the README](../README.md#configuration)
for the rest of the settings.
