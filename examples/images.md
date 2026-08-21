# Local Images

Paths locales incrustados como base64. URLs externas y rutas inexistentes muestran el texto alternativo como respaldo.

---

## Caso 1 — Path relativo (mismo directorio) ✓ debe mostrarse

![Local gradient](gradient.png)

---

## Caso 2 — URL externa HTTP ✗ debe mostrar el alt: *[External placeholder]*

![External placeholder](https://via.placeholder.com/400x200)

---

## Caso 3 — URL externa HTTPS ✗ debe mostrar el alt: *[HTTPS image]*

![HTTPS image](https://upload.wikimedia.org/wikipedia/commons/4/47/PNG_transparency_demonstration_1.png)

---

## Caso 4 — Path inexistente ✗ debe mostrar el alt: *[Missing image]*

![Missing image](no-existe.png)

---

## Caso 5 — Path relativo a subdirectorio inexistente ✗ debe mostrar el alt: *[Image in subdir]*

![Image in subdir](subdir/foto.png)

---

## Caso 6 — Imagen sin alt ✗ debe mostrar *[image]*

![](no-existe-tampoco.png)

---

## Referencia — Flujo de texto alrededor de imagen

Párrafo antes de la imagen.

![Local gradient](gradient.png)

Párrafo después de la imagen. El flujo de texto debe continuar sin saltos extraños.
