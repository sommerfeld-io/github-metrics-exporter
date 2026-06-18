# AI-Assisted Development

This project is developed with the help of AI coding assistants. We use [Claude Code](https://code.claude.com) and [GitHub Copilot](https://github.com/features/copilot) in parallel.

AI assistants are guided by instruction files that define coding conventions, commit message rules, and project-specific guidelines. We maintain a single source of truth for these instructions and use symlinks so that each tool reads from the same file:

| Symlink         | Points to (master file)                   |
|-----------------|-------------------------------------------|
| `CLAUDE.md`     | `.github/copilot-instructions.md`         |
| `src/CLAUDE.md` | `.github/instructions/go.instructions.md` |

The master instruction files live under `.github/` and are committed to the repository:

- `.github/copilot-instructions.md` - project-wide conventions (commit messages, general style)
- `.github/instructions/go.instructions.md` - Go-specific coding guidelines for the `src/` subtree

The symlinks are also committed to git. Still, they are created locally by running `task symlinks` (see `taskfile.yml` for the exact `ln` commands used).

> [!NOTE]
> Because this repository uses symlinks, it is intended to be worked on in a Linux environment. The devcontainer setup makes it possible to work on a Windows host since the dev container is Linux, and the actual development environment runs inside that container.
