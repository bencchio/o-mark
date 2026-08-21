# Inline Formatting

Prueba de todos los elementos de formato inline: énfasis, tachado, subscript, superscript, highlight y combinaciones.

---

## Énfasis básico

Normal, **negrita**, *cursiva*, ***negrita y cursiva***.

Con guiones bajos: __negrita__, _cursiva_, ___ambos___.

Mixto: **negrita con *cursiva* dentro** y *cursiva con **negrita** dentro*.

---

## Tachado

~~texto tachado~~ al inicio. Texto normal ~~tachado~~ en medio. Al final ~~tachado~~.

~~Frase entera tachada con **negrita** y *cursiva* dentro.~~

---

## Subscript y Superscript

Subscript: H~2~O, CO~2~, C~6~H~12~O~6~, log~2~(n).

Superscript: E=mc^2^, x^n^, a^2^ + b^2^ = c^2^, (a+b)^n^.

Combinados: H~2~^+^, x~i~^2^.

Adjacent: a~1~a~2~a~3~, x^1^x^2^x^3^.

---

## Highlight

==texto resaltado==. Normal ==resaltado== normal.

==Highlight con **negrita** dentro== y ==con *cursiva* dentro==.

Varias marcas: ==uno== dos ==tres== cuatro ==cinco==.

---

## Código inline

Variable: `count`, función: `fmt.Println()`, tipo: `[]byte`.

Con símbolos: `a + b`, `x > 0`, `s := "hello"`, `i++`.

En contexto: usa `os.ReadFile` en lugar de `ioutil.ReadFile` (deprecated).

---

## Combinaciones

**negrita** + ==highlight==: **==texto negrita y resaltado==**.

*cursiva* + ~~tachado~~: *~~cursiva tachada~~*.

`código` con ~~tachado alrededor~~: ~~`codigo`~~.

Subscript en fórmula: La concentración de H~2~O en *agua pura* es **100%**.

Superscript en referencia: Ver nota^1^ y también^2^.

---

## Casos límite

Tilde sin cerrar: a~b (no debe renderizar como subscript).

Caret sin cerrar: a^b (no debe renderizar como superscript).

Iguales solos: == (no debe renderizar como highlight).

Doble tilde simple: ~~ (no debe renderizar como tachado).

Escapes: \~\~no tachado\~\~, \*\*no negrita\*\*, \==no highlight\==.

---

## Inline en distintos contextos

### En heading

Título con **negrita**, *cursiva* y `código`

#### En lista

- Item con **negrita** y H~2~O
- Item con ==highlight== y E=mc^2^
- Item con ~~tachado~~ y `código`

#### En blockquote

> Cita con **negrita**, *cursiva*, H~2~O y ==highlight==.

#### En tabla

| Elemento | Ejemplo |
|---|---|
| Negrita | **texto** |
| Cursiva | *texto* |
| Subscript | H~2~O |
| Superscript | mc^2^ |
| Highlight | ==texto== |
| Tachado | ~~texto~~ |
| Código | `texto` |
