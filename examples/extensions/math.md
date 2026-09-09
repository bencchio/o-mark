# Math (KaTeX)

Mathematical analysis: series, integrals, probability, linear algebra.
Inline and block, including multi-line `\begin{aligned}`, plus documented
known issues below.

---

## Series and Limits

The **Taylor series** of a function $f(x)$ around $a = 0$ is:

$$
f(x) = \sum_{n=0}^{\infty} \frac{f^{(n)}(0)}{n!} x^n
$$

For the exponential function, every derivative equals itself, so $f^{(n)}(0) = 1$ for all $n$:

$$e^x = \sum_{n=0}^{\infty} \frac{x^n}{n!} = 1 + x + \frac{x^2}{2!} + \frac{x^3}{3!} + \cdots$$

The **geometric series** converges for $|r| < 1$:

$$\sum_{n=0}^{\infty} r^n = \frac{1}{1-r}$$

---

## Euler's Formula and Complex Analysis

**Euler's formula** connects the exponential function to trigonometry for any real $\theta$:

$$e^{i\theta} = \cos\theta + i\sin\theta$$

Setting $\theta = \pi$ yields the celebrated identity $e^{i\pi} + 1 = 0$, which unites the five
fundamental constants $e$, $i$, $\pi$, $1$, and $0$.

The **modulus** of a complex number $z = a + bi$ is $|z| = \sqrt{a^2 + b^2}$, and its argument
is $\arg(z) = \arctan\!\left(\frac{b}{a}\right)$.

---

## Calculus

**Integration by parts**: for differentiable functions $u$ and $v$,

$$\int u \, dv = uv - \int v \, du$$

The **Gaussian integral** is a cornerstone of probability and physics:

$$\int_{-\infty}^{\infty} e^{-x^2} dx = \sqrt{\pi}$$

More generally, for $\alpha > 0$:

$$\int_{-\infty}^{\infty} e^{-\alpha x^2} dx = \sqrt{\frac{\pi}{\alpha}}$$

The **Fundamental Theorem of Calculus** states that if $F'(x) = f(x)$, then:

$$\int_a^b f(x)\, dx = F(b) - F(a)$$

---

## Probability and Statistics

The **normal distribution** $\mathcal{N}(\mu, \sigma^2)$ has probability density:

$$p(x) = \frac{1}{\sigma\sqrt{2\pi}} \exp\!\left(-\frac{(x-\mu)^2}{2\sigma^2}\right)$$

Its mean is $\mathbb{E}[X] = \mu$ and variance $\text{Var}(X) = \sigma^2$.

**Bayes' theorem** relates conditional probabilities $P(A \mid B)$ and $P(B \mid A)$:

$$P(A \mid B) = \frac{P(B \mid A)\, P(A)}{P(B)}$$

---

## Fourier Transform

The **Fourier transform** of an integrable function $f : \mathbb{R} \to \mathbb{C}$ is:

$$\hat{f}(\xi) = \int_{-\infty}^{\infty} f(x)\, e^{-2\pi i \xi x}\, dx$$

and the **inverse transform** reconstructs $f$ from its frequency content:

$$f(x) = \int_{-\infty}^{\infty} \hat{f}(\xi)\, e^{2\pi i \xi x}\, d\xi$$

**Parseval's theorem** expresses conservation of energy between domains:

$$\int_{-\infty}^{\infty} |f(x)|^2\, dx = \int_{-\infty}^{\infty} |\hat{f}(\xi)|^2\, d\xi$$

---

## Multi-line Block Test

> **Known issue:** KaTeX ignores literal line breaks inside `$$...$$`.
> Multi-line expressions need explicit `\begin{aligned}...\end{aligned}` with
> `\\` row separators — plain multi-line content still renders, but collapses
> onto one line instead of aligning. This is a KaTeX limitation, not a parser
> one: the block below reaches KaTeX exactly as written.
>
> Three former parser issues are fixed and covered by the regression sections
> below: stray `$` in prose no longer opens inline math, multi-line `$$`
> blocks convert inside any wrapper tag, and block content is now taken
> verbatim, so `\\` survives and `^`/`~`/`==` inside a formula are no longer
> mistaken for superscript, subscript or highlight markup.

**Maxwell's equations** — four lines using `\begin{aligned}` with `\\` breaks:

$$
\begin{aligned}
\nabla \cdot \mathbf{E} &= \frac{\rho}{\varepsilon_0} \\
\nabla \cdot \mathbf{B} &= 0 \\
\nabla \times \mathbf{E} &= -\frac{\partial \mathbf{B}}{\partial t} \\
\nabla \times \mathbf{B} &= \mu_0\mathbf{J} + \mu_0\varepsilon_0\frac{\partial \mathbf{E}}{\partial t}
\end{aligned}
$$

**Derivation steps** — chained equalities with `\\`:

$$
\begin{aligned}
e^{i\pi} + 1 &= \cos\pi + i\sin\pi + 1 \\
             &= -1 + i \cdot 0 + 1 \\
             &= 0
\end{aligned}
$$

---

## Regression — Stray dollars in prose

Every dollar sign in this section must stay literal text — none of these
lines should render as math:

The service costs $5 and $10 for the premium tier.

Our prices are $1, $2 and $3 depending on volume.

In bash, `$HOME` expands, and prose mentions of $PATH or $variable names
should not open a formula either.

Spaced delimiters are rejected too, so $ x $ stays plain text, while proper
inline math like $x^2 + 1$ still renders.

---

## Regression — Multi-line block inside wrappers

A multi-line `$$` block must convert to a KaTeX block wherever it appears,
not only as a bare paragraph.

Plain paragraph, no `\begin{aligned}` (renders as one line — KaTeX
limitation noted above):

$$
E = mc^2
\quad
p = mv
$$

Inside a blockquote:

> The quadratic formula:
>
> $$
> x = \frac{-b \pm \sqrt{b^2 - 4ac}}{2a}
> $$

Inside a list item:

- First item with a block:

  $$
  \sum_{k=1}^{n} k = \frac{n(n+1)}{2}
  $$

- Second item, plain text.

## Regression — Block content is verbatim

A block whose formula contains `^`, `~`, `_` or `==` must reach KaTeX
untouched — these used to be eaten by the superscript, subscript and
highlight extensions before the block was parsed:

$$
f(x) = \sum_{n=0}^{\infty} \frac{f^{(n)}(0)}{n!} x^n
$$

And `\\` must survive as a row separator, not collapse into a single
backslash:

$$
\begin{aligned}
a_1 &= b^2 \\
c_1 &= d^2 \\
e_1 &= f^2
\end{aligned}
$$

---

And a fenced code block containing `$$` lines must stay code, untouched:

```
$$
this is code, not math
$$
```

---

## Linear Algebra

For a square matrix $A$, the **characteristic polynomial** is $\det(A - \lambda I) = 0$.
If $A\mathbf{v} = \lambda\mathbf{v}$ for a nonzero vector $\mathbf{v}$, then $\lambda$ is an
**eigenvalue** and $\mathbf{v}$ an **eigenvector**.

The **trace** equals the sum of eigenvalues $\lambda_1, \ldots, \lambda_n$:

$$\text{tr}(A) = \sum_{i=1}^{n} \lambda_i$$

and the **determinant** equals their product:

$$\det(A) = \prod_{i=1}^{n} \lambda_i$$
