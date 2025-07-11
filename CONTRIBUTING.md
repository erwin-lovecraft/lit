# Contributing to Lit

Welcome! 🎉 Thank you for your interest in contributing to **Lit** – a modular, production-grade collection of Go libraries that simplifies backend development.

This document provides guidelines to help you get started and ensure consistent, high-quality contributions.

---

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Ways to Contribute](#ways-to-contribute)
- [Contribution Requirements](#contribution-requirements)
- [Development Workflow](#development-workflow)
- [License](#license)

---

## Code of Conduct

Please be respectful and inclusive in all interactions. By participating in this project, you agree to abide by our [Code of Conduct](./CODE_OF_CONDUCT.md).

---

## Ways to Contribute

We welcome all kinds of contributions:

### 🐞 Report Bugs

If you find a bug, [open an issue](https://github.com/viebiz/lit/issues/new/choose) with:
- A clear title and description
- Steps to reproduce
- Expected and actual behavior
- Environment details (Go version, OS)

### 📚 Improve Documentation

Help clarify usage, improve readability, or correct mistakes in:
- README files
- Module documentation
- Examples

### ✨ Suggest Features

Propose improvements or new modules via [GitHub Discussions](https://github.com/viebiz/lit/discussions) or open an issue with the `enhancement` label.

### 🧑‍💻 Submit Code

Contribute to existing modules (like `ioutil`, `iam`, `guard`, `broker`, `monitoring`, etc.) or propose new reusable components.

---

## Contribution Requirements

To maintain code quality and consistency, please follow these guidelines:

### 🧹 Code Style

- Use `gofmt` before committing (`go fmt ./...`)
- Use clear, descriptive naming (e.g., `NewHandler`, `WithConfig`)
- Keep functions short and focused
- Prefer composition over inheritance

### ✅ Testing

- All logic **must** be covered by **unit tests**
- Use table-driven tests where applicable
- Mock external dependencies to keep tests isolated
- Run `make test` and ensure all tests pass locally

---

### 🤖 CI & Automation

- This project uses **automated CI** to run tests on each **Pull Request (PR) and Merge Request (MR)**
- **Only PRs with all tests passing will be eligible for approval**
- The CI pipeline includes:
    - Code formatting/linting
    - Unit test execution
    - Build validation

---

## Development Workflow

### 1. Fork and Clone

```bash
git clone https://github.com/YOUR_USERNAME/lit.git
cd lit
```
### 2. Create a Branch

```
git checkout -b feature/my-feature
```

Use descriptive branch names (fix/config-reload, feat/grpc-middleware).

### 3. Make Changes

- Keep commits focused and small
- Add tests for new logic
- Update documentation if needed

### 4. Format & Test

```
make test
```

### 5. Push and Open a PR

```
git push origin feature/my-feature
```

Then, open a pull request on GitHub. Please fill out the PR template and describe the change clearly.

## License

By contributing to this project, you agree that your contributions will be licensed under the project's [Apache-2.0 License](./LICENSE).
