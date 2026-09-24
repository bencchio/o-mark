# Keyboard Shortcuts

O'Mark is built to be used from the keyboard. This page lists every shortcut.

## Reading

| Action | Keyboard | Mouse |
|---|---|---|
| Scroll one page | `PgUp` / `PgDn` | Wheel / trackpad |
| Move to the next / previous link | `Tab` / `Shift+Tab` | — |
| Open the focused link | `Enter` | Click the link |

## Navigating text

The active word is highlighted as you read. The arrow keys move that highlight
through the document instead of scrolling (scrolling is `PgUp` / `PgDn`).

| Action | Keyboard |
|---|---|
| Next / previous word | `→` / `←` |
| Nearest word on the line above / below | `↑` / `↓` |
| Next / previous sentence | `Ctrl+→` / `Ctrl+←` |
| Next / previous paragraph | `Ctrl+↑` / `Ctrl+↓` |
| First / last word of the current line | `Home` / `End` |
| First / last word of the document | `Ctrl+Home` / `Ctrl+End` |
| Start / stop selecting (marks the current word as the anchor) | `Space` |
| Copy the selection, or the active word if there is none | `Super+C` |

While selecting, the same movement keys expand or shrink the marked range from
the anchor. `Space` stops selecting, clears the range, and returns the cursor
to the anchor word. `Super+C` copies and stops selecting too, but keeps the
range highlighted. Clicking the document moves the highlight to that word and
clears any selection.

Links to other `.md` files open in a new O'Mark window. This needs O'Mark
registered as your markdown handler — see [INSTALL.md](../INSTALL.md).

## Tables

| Action | Keyboard | Mouse |
|---|---|---|
| Sort by a column | Hover the column header, press `s` | Click the column header |
| Advance the sort (ascending → descending → off) | Press `s` again on the same header | Click the same header again |

A badge confirms each step ("SORTED ↓ Column", "SORT CLEARED"), and an arrow
on the header shows the active sort column and direction while it's on.
Sorting a different column starts a fresh cycle on that column — only one
column is sorted at a time. The third step returns the table to its original
order. The sort is for the current session only: it isn't saved, and it never
changes the source file.

## Search

| Action | Keyboard |
|---|---|
| Open search | `/` |
| Next / previous match | `Enter` / `Shift+Enter` |
| Toggle case-sensitive matching (`Aa`) | Click it, or `Tab` to it from the input and press `Enter`/`Space` |
| Move focus between the input and `Aa` | `Tab` / `Shift+Tab` |
| Close search | `Esc` |

Opening search shows the toolbar if it was hidden. It prefills the input
from the current word-marked range or mouse selection, if there is one.
Matches are highlighted as you type; the counter next to the input shows
your position ("3 / 12") or that there are no matches.

`/` is search. `Ctrl+F` does nothing — the web engine's native find bar is
disabled.

Closing search with `Esc` only closes search — it returns focus to the
document but leaves the toolbar exactly as it was. A second `Esc` (with
search already closed) is the regular toolbar toggle described below.

## Zoom

| Action | Keyboard | Mouse |
|---|---|---|
| Zoom in | `Ctrl++` or `Ctrl+=` | `Ctrl` + wheel up |
| Zoom out | `Ctrl+-` | `Ctrl` + wheel down |
| Reset zoom to the default (100%, or `zoom_default`) | `Ctrl+0` | Click `↺` in the toolbar |

The current zoom level is shown in the toolbar.

## Page

| Action | Keyboard | Mouse |
|---|---|---|
| Flip between portrait and landscape | `Ctrl+R` | Click the orientation label in the toolbar |

The orientation lasts for the session only, also on a document that declares its
own, and is not saved. The PDF export uses the same sheet.

## Export

| Action | Keyboard | Mouse |
|---|---|---|
| Export PDF next to the open file | `Ctrl+P` | Click `PDF` in the toolbar |

The PDF is written beside the markdown (`foo.md` → `foo.pdf`). If that file
already exists, O'Mark does not overwrite it.

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

The toolbar holds the search bar (see [Search](#search)), the zoom reset
button (`↺`), the PDF export control, the page orientation control, and the
treatment selector.

### Moving around the toolbar

| Action | Keyboard |
|---|---|
| Go straight to the treatment selector | `T` |
| Go straight to the zoom reset `↺` | `Z` |
| Move between the controls | `←` / `→` |
| Reset zoom (when `↺` is selected) | `Enter` or `Space` |
| Flip the orientation (when it is selected) | `Enter` or `Space` |

### Changing the treatment

With the treatment selector focused:

| Action | Keyboard |
|---|---|
| Switch treatment without opening the list | `↑` / `↓` |
| Open the treatment list | `Space` or `Enter` |
| Move through the list | `↑` / `↓` |
| Pick the highlighted treatment | `Enter` |
| Close the list, keeping the current treatment | `Esc` |

Closing the list with `Esc` leaves the toolbar open — you need a second `Esc` to
go back to the document.

The treatment you are using when you quit becomes the one O'Mark starts with next
time. See [the configuration section of the README](../README.md#configuration)
for the rest of the settings.
