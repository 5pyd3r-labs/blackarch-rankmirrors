# Contributing

Thanks for taking a look at this project. Contributions, bug reports, and suggestions are welcome.

## Getting Started

1. Fork the repository
2. Clone your fork locally
3. Create a new branch for your change

   ```
   git checkout -b feature/short-description
   ```

## Commit Messages

This project follows conventional commit format:

```
<type>(<scope>): <subject>
```

Common types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`

Examples:
```
feat(auth): add rate limiting to login endpoint
fix(parser): handle empty input without crashing
docs(readme): update installation steps
```

Keep the subject line under 50 characters, imperative mood ("add" not
"added"), no trailing period. Add a body if the change needs more context —
explain what and why, not how.

## Submitting a Pull Request

1. Make sure your branch is up to date with `main`
2. Push your branch and open a PR against `main`
3. Fill out the PR template — describe what changed and why
4. Link any related issues

PRs are reviewed by the maintainers as time allows. This is a
solo-maintained project, so response times may vary.

## Code Style

- Match the existing style/formatting already in the file you're editing
- Keep changes focused — one logical change per PR
- Add tests for new functionality where the project has an existing test
  suite

## Reporting Bugs

Use the issue templates provided. Include steps to reproduce, expected vs
actual behavior, and environment details where relevant.

## Reporting Security Issues

Do not open a public issue for security vulnerabilities. See
[SECURITY.md](SECURITY.md) for responsible disclosure instructions.
