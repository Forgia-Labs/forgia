---
title: "Creating an FD"
description: "Use /fd-new to create a Feature Design from a GitHub issue, free-text description, or scratch."
navigation:
  title: "Creating"
---

# Creating an FD

## From a GitHub issue

Requires [GitHub CLI](https://cli.github.com/) installed and authenticated.

```
# Full URL
/fd-new https://github.com/owner/repo/issues/42

# Short reference (repo inferred from git remote)
/fd-new #42

# Short reference with explicit repo
/fd-new #42 --repo owner/repo
```

Claude fetches the issue title, body, labels, and comments, then rewrites them into a clean FD — problem statement, solutions, architecture diagrams.

## From a GitLab issue

Requires `GITLAB_TOKEN` env var or [glab CLI](https://gitlab.com/gitlab-org/cli).

```bash
export GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx

/fd-new https://gitlab.com/owner/repo/-/issues/42
# Self-hosted:
/fd-new https://git.example.com/owner/repo/-/issues/42
```

## From free text

```
/fd-new "Add rate limiting to the API gateway"
```

Claude creates the FD structure from your description and asks clarifying questions if needed.

## What gets created

A file at `.forgia/fd/FD-NNN-kebab-title.md` with all sections pre-filled:

- Problem derived from the issue or description
- Two solutions with pros/cons (one marked as chosen)
- Architecture diagrams in Mermaid
- Interface table
- Placeholder SDDs

## After creation

Edit the FD to:
1. Refine the problem statement
2. Add your chosen architecture details
3. Define component interfaces
4. List the SDDs you want generated

Then run `/fd-review FD-NNN` to advance to approved.
