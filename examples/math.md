# Math (KaTeX)

Análisis matemático: series, integrales, probabilidad, álgebra lineal. Inline y block, incluye `\begin{aligned}` multi-línea.

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

> **Nota:** KaTeX ignora los saltos de línea literales dentro de `$$...$$`. Para expresiones
> multi-línea se requiere `\begin{aligned}...\end{aligned}` con `\\` explícitos.

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

## Linear Algebra

For a square matrix $A$, the **characteristic polynomial** is $\det(A - \lambda I) = 0$.
If $A\mathbf{v} = \lambda\mathbf{v}$ for a nonzero vector $\mathbf{v}$, then $\lambda$ is an
**eigenvalue** and $\mathbf{v}$ an **eigenvector**.

The **trace** equals the sum of eigenvalues $\lambda_1, \ldots, \lambda_n$:

$$\text{tr}(A) = \sum_{i=1}^{n} \lambda_i$$

and the **determinant** equals their product:

$$\det(A) = \prod_{i=1}^{n} \lambda_i$$
