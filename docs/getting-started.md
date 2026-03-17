# Getting Started

## Prerequisites

- [mise](https://mise.jdx.dev/) — task runner and tool manager
- [Docker](https://docs.docker.com/get-docker/) — for OpenHands agent runtime
- [Claude Code](https://claude.ai/claude-code) — for slash commands (optional but recommended)

## Install Forgia

```bash
git clone git@github.com:Deepzima/forgia.git ~/forgia
cd ~/forgia

# Install Claude Code slash commands
mise run claude:install

# Pull OpenHands (optional, only if you want autonomous agents)
mise run openhands:install
```

## Initialize a Project

```bash
cd /path/to/your/project

# Option 1: via CLI
~/forgia/bin/forgia init

# Option 2: via Claude Code
/project-init
```

This creates `.forgia/` in your project with:
- `constitution.md` — edit with your project rules
- `fd/` — Feature Designs
- `sdd/` — Execution Specs
- `ops/` — Task tracking
- `dev-guide/` — Conventions

## Your First Feature

### 1. Create a Feature Design

`/fd-new` accepts three types of input:

**Free-text description**
```
/fd-new "Add user authentication"
```

**From a GitHub issue**

Requires [GitHub CLI](https://cli.github.com/) (`gh`) to be installed and authenticated.

```
# Full URL
/fd-new https://github.com/owner/repo/issues/42

# Short reference (repo inferred from git remote)
/fd-new #42

# Short reference with explicit repo
/fd-new #42 --repo owner/repo
```

**From a GitLab issue**

Requires either a `GITLAB_TOKEN` environment variable or [glab CLI](https://gitlab.com/gitlab-org/cli) installed.

```bash
export GITLAB_TOKEN=glpat-xxxxxxxxxxxxxxxxxxxx  # read_api scope
```

```
# Full URL (host, owner, repo and issue number extracted automatically)
/fd-new https://gitlab.com/owner/repo/-/issues/42

# Self-hosted GitLab
/fd-new https://git.example.com/owner/repo/-/issues/42

# Short reference (gl# prefix)
/fd-new gl#42 --repo owner/repo

# Short reference targeting a self-hosted instance
/fd-new gl#42 --repo owner/repo --gitlab-host git.example.com
```

In all cases, this creates `.forgia/fd/FD-NNN-kebab-title.md` with the issue title, labels, assignee, and description pre-filled.

### 2. Fill in the FD

Edit the FD to describe:
- The problem (what and why)
- Solutions considered (at least 2)
- Chosen architecture
- Component interfaces
- Expected SDDs

### 3. Review

```
/fd-review FD-001
```

The review checks problem clarity, architecture diagram, interfaces, and constitution compliance.

### 4. Generate SDDs

```
/fd-sdd FD-001
```

This creates N SDDs in `.forgia/sdd/FD-001/`, one per component.

### 5. Execute

```bash
# Option A: OpenHands (autonomous)
mise run sdd .forgia/sdd/FD-001/SDD-001.md

# Option B: All SDDs in parallel
mise run sdd:batch FD-001

# Option C: Claude Code (interactive)
/sdd-assign SDD-001 claude-code

# Option D: Manual
# Read the SDD and implement yourself
```

### 6. Verify

```
/fd-verify FD-001
```

Checks all SDDs are done, Work Logs are filled, tests pass.

### 7. Close

```
/fd-close FD-001
```

Archives the FD, updates changelog, aggregates retrospectives.

## Health Check

```bash
~/forgia/bin/forgia doctor
```

Checks: vault structure, Docker, OpenHands, mise, Claude commands, LLM API key.

## Using with Obsidian

The `.forgia/` directory is an Obsidian-compatible vault. Open it in Obsidian to get:
- Dataview dashboard (`_dashboard.md`)
- Kanban boards for task tracking
- Linked notes between FDs, SDDs, and tasks

Recommended Obsidian plugins: Dataview, Kanban, Templater, Calendar.
