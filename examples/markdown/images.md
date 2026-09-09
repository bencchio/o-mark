# Local Images

Tests for `resolveImages`: local paths get inlined as base64; external URLs
and missing paths fall back to their alt text.

---

## Case 1 — Local relative path ✓ should display

A relative path to a file that exists in `assets/`. `resolveImages` should
find it, read it from disk, and inline it as a base64 `data:` URI:

![Local gradient](../assets/gradient.png)

---

## Case 2 — External HTTP URL ✗ should show alt text: *[External placeholder]*

An `http://`-style external URL. o-mark does not fetch remote images, so the
`<img>` should fall back to rendering its alt text instead of an empty box:

![External placeholder](https://via.placeholder.com/400x200)

---

## Case 3 — External HTTPS URL ✗ should show alt text: *[HTTPS image]*

Same fallback behaviour as Case 2, over `https://` — confirms the scheme
itself isn't what triggers the fallback, remoteness is:

![HTTPS image](https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png)

---

## Case 4 — Missing path ✗ should show alt text: *[Missing image]*

A relative path to a file that does not exist. `resolveImages` should fail
to read it and fall back to the alt text instead of erroring out:

![Missing image](does-not-exist.png)

---

## Case 5 — Relative path into a missing subdirectory ✗ should show alt text: *[Image in subdir]*

Same as Case 4, but the missing path also traverses a subdirectory that
doesn't exist — checks that directory resolution fails gracefully too:

![Image in subdir](subdir/photo.png)

---

## Case 6 — Image with no alt text ✗ should show *[image]*

A missing image with an empty alt attribute. Confirms the fallback still
renders something sensible (a generic `[image]` placeholder) instead of
blank text when there's no alt to fall back to:

![](also-missing.png)

---

## Reference — Text flow around an image

Paragraph before the image.

![Local gradient](../assets/gradient.png)

Paragraph after the image. Text flow should continue without odd gaps.
