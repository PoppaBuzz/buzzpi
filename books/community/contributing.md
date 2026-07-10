# Contributing to BuzzPi

Thank you for considering contributing to BuzzPi. This project thrives on community involvement, and every contribution is valued.

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md) in all interactions.

---

## Ways to Contribute

- **Code** — submit pull requests for bug fixes, features, or improvements
- **Documentation** — improve or translate documentation
- **Design** — contribute to the design system, icons, or UI mockups
- **Testing** — test on different devices and report issues
- **Ideas** — open an issue or start a discussion
- **Translations** — help localize the Android app and documentation
- **Bug reports** — file detailed bug reports so we can fix issues

---

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/buzzpi/buzzpi.git`
3. Set up the development environment for the component you want to work on (see component-specific guides)
4. Create a branch from `main`: `git checkout -b your-feature-name`
5. Make your changes

---

## Development Workflow

1. Ensure your branch is based on the latest `main`
2. Make focused, atomic commits with clear messages
3. Write or update tests as needed
4. Run linters and checks before opening a PR
5. Open a pull request against `main`

### Commit Message Style

We use conventional commit format:

- `feat:` — a new feature
- `fix:` — a bug fix
- `docs:` — documentation changes
- `refactor:` — code refactoring
- `test:` — adding or updating tests
- `chore:` — maintenance tasks

---

## RFC Process

Significant changes to BuzzPi start as an RFC. Before implementing a large feature, open an RFC pull request in the `rfcs/` directory.

See [GOVERNANCE.md](GOVERNANCE.md) for the full RFC lifecycle.

---

## Style Guides

### Go

- Standard `gofmt` formatting
- Follow idiomatic Go conventions
- Use meaningful names and avoid premature abstraction

### Kotlin

- Follow the [Kotlin style guide](https://kotlinlang.org/docs/coding-conventions.html)
- Use Jetpack Compose conventions for UI code
- Keep functions focused and composable

### Documentation

- Write in clear, plain English
- Use Markdown with consistent formatting
- Include code examples where helpful

---

## Pull Request Process

1. Ensure your PR description clearly describes the problem and solution
2. Reference the related issue or RFC if applicable
3. Verify that all CI checks pass
4. At least one maintainer review is required before merging
5. Maintainers may request changes or ask questions

---

## Questions

If you have questions, open a discussion on GitHub or reach out through the community channels.
