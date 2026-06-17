# AI Instructions

This file defines the rules and conventions that AI coding assistants should follow when working in this repository. It is tool-agnostic and serves as the single source of truth for all AI-generated code and commit messages.

## Commit Messages: Conventional Commits

Always use Conventional Commits for every commit message.

**Format:** `<type>(optional scope): <description>`

| Type                                                                | Effect        | When to use                      |
|---------------------------------------------------------------------|---------------|----------------------------------|
| `fix`                                                               | PATCH release | Patches a bug                    |
| `feat`                                                              | MINOR release | Introduces a new feature         |
| `BREAKING CHANGE` footer                                            | MAJOR release | Introduces a breaking API change |
| `build`, `chore`, `ci`, `docs`, `style`, `refactor`, `perf`, `test` | No release    | All other changes                |

**Rules:**

- A scope may be added in parentheses for extra context: `feat(parser): add ability to parse arrays`. A scope may **NOT** contain a slash (`/`).
- Breaking changes must include `BREAKING CHANGE:` in the footer: `feat: drop support for Node 6`
- Commit message titles must also match the project pattern: `^(fix|feat|build|chore|ci|docs|style|refactor|perf|test)/[a-z0-9._-]+$`

Write commit messages using the Conventional Commits format, ensuring the header (`type(scope): summary`) is clear and descriptive, as it will be displayed on GitHub release pages and used for changelogs. Focus the header on user-visible, meaningful change descriptions and avoid vague wording. Always document breaking changes explicitly in the footer using `BREAKING CHANGE:` (do not use the `!` notation).
