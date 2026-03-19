# Functional Architecture

> What Forgia does, how the components interact, and how data flows through the system.
> Companion to [go-integration-guide.md](go-integration-guide.md) which covers _how_ the code is structured.

## 1. System Overview

```mermaid
flowchart TD
    subgraph human ["Human Layer"]
        Dev["Developer"]
        Team["Team (PR review, board)"]
    end

    subgraph forgia ["Forgia (single Go binary)"]
        CLI["CLI Commands\ninit, status, doctor, validate,\nexec, watch, batch, render"]
        MCP["MCP Server\n(tool provider for agents)"]
        Skills["Skills Engine\n(/fd-new, /fd-review, /fd-sdd,\n/arch-init, /arch-review)"]
    end

    subgraph agents ["Agent Layer"]
        Claude["Claude Code\n(host — architect)"]
        Sandbox["Claude Code\n(sandbox — builder)"]
    end

    subgraph external ["External Services"]
        Board["Project Board\n(GitHub/GitLab)"]
        CodeMem["codebase-memory-mcp\n(Tier 3 knowledge)"]
        Docker["Docker Engine\n(sandbox containers)"]
        Beads["Beads\n(local cache)"]
        RTK["RTK\n(token compression)"]
    end

    subgraph storage ["Storage"]
        Vault[".forgia/\n(source of truth)"]
        Git["Git\n(versioning)"]
    end

    Dev -->|"interactive"| CLI
    Dev -->|"interactive"| Claude
    Claude -->|"MCP"| MCP
    Claude -->|"slash commands"| Skills
    CLI -->|"spawns"| Sandbox
    MCP -->|"ToolProvider"| CodeMem
    MCP -->|"ProjectBoard"| Board
    MCP -->|"BeadsClient"| Beads
    Sandbox -->|"runs in"| Docker
    Docker -->|"includes"| RTK
    Team -->|"review"| Board
    Board <-->|"sync"| Vault
    CLI --> Vault
    MCP --> Vault
    Skills --> Vault
    Vault --> Git

    style forgia fill:#fff3cd,stroke:#ffc107
    style storage fill:#d4edda,stroke:#28a745
    style external fill:#cce5ff,stroke:#0d6efd
```

## 2. Artifact Lifecycle

Every artifact in Forgia follows a lifecycle from creation to archival.

> **Note**: The Architecture Phase (`/arch-init`, `/arch-review`, `/arch-update`) is defined in #27 and not yet implemented. The FD and SDD phases are implemented and tested.

```mermaid
stateDiagram-v2
    [*] --> Architecture: /arch-init

    state "Architecture Phase" as Architecture {
        [*] --> SystemContext: system-context.yaml
        [*] --> Containers: containers.yaml
        [*] --> Contexts: contexts/*.yaml
        SystemContext --> Reviewed: /arch-review
        Containers --> Reviewed
        Contexts --> Reviewed
    }

    Architecture --> FD: /fd-new (from context)

    state "Feature Design Phase" as FD {
        [*] --> Planned
        Planned --> InReview: /fd-review
        InReview --> Approved: all checks pass
        InReview --> Planned: revision needed
        Planned --> Abandoned: team decides to drop
        InReview --> Rejected: competitive FD loses
        Approved --> Rejected: /fd-close --reject
    }

    FD --> SDD: /fd-sdd

    state "SDD Phase" as SDD {
        [*] --> SDDPlanned
        SDDPlanned --> Validated: forgia validate
        Validated --> Executing: forgia exec
        Executing --> Done: agent completes
        Executing --> Failed: agent fails
        Failed --> SDDPlanned: rework
    }

    SDD --> Verify: /fd-verify
    Verify --> Closed: /fd-close
    Closed --> ArchUpdate: /arch-update
    ArchUpdate --> Architecture: feedback loop

    Rejected --> [*]: archived as negative ADR
    Abandoned --> [*]: archived with reason
    Closed --> [*]: archived with retrospective
```

## 3. Data Model

### Artifact hierarchy

```
.forgia/ (vault — source of truth)
│
├── architecture/                    ← C4 Level 1-2 (YAML + derived MD)
│   ├── system-context.yaml
│   ├── containers.yaml
│   ├── technology-decisions.yaml
│   ├── quality-attributes.yaml
│   ├── constraints.yaml
│   └── glossary.yaml
│
├── contexts/                        ← DDD Bounded Contexts (YAML + derived MD)
│   ├── <context-name>.yaml          ← interfaces, dependencies, language
│   └── ...
│
├── fd/                              ← Feature Designs (YAML source, MD derived)
│   ├── FD-<hash>-<slug>.yaml       ← agent reads/writes this
│   ├── FD-<hash>-<slug>.md         ← auto-generated for human review
│   └── _templates/
│
├── sdd/                             ← Execution Specs
│   ├── FD-<hash>/
│   │   ├── SDD-001-<slug>.yaml     ← agent reads/writes this
│   │   ├── SDD-001-<slug>.md       ← auto-generated for human review
│   │   └── ...
│   └── _templates/
│
├── constitution.md                  ← immutable rules
├── config.toml                      ← team config (committed)
├── config.local.toml                ← personal overrides (gitignored)
│
├── guardrails/
│   ├── deny.toml                    ← read/execute/write deny patterns
│   └── ignore                       ← context exclusion patterns
│
├── dev-guide/
│   ├── principles/                  ← clean-code, SOLID, design-patterns
│   ├── lang/                        ← per-language conventions (auto-detected)
│   └── *.md                         ← coding, commit, review conventions
│
├── learnings/                       ← feedback from closed FDs (committed)
│   ├── FD-<hash>.yaml               ← failure modes, successful patterns, suggestions
│   └── ...
│
├── logs/                            ← gitignored
│   ├── exec-<sdd>-<timestamp>.json  ← execution reports
│   └── exec-<sdd>-<timestamp>.log   ← execution logs
│
└── .gitignore                       ← excludes logs/, run/, .beads/
```

### ID System

Hash-based, collision-proof:

```
FD-a3f2   ← 4-char hex from sha256(title + timestamp + author)
SDD-001   ← sequential within FD folder (no cross-FD collision)
```

Same hash everywhere:

```
GitHub Project card ID = Beads task ID = FD file ID = SDD folder name
```

### ID System — migration note

> **Current state**: sequential IDs (FD-001, FD-002) in Markdown with YAML frontmatter.
> **Target state**: hash-based IDs (FD-a3f2) in YAML source with derived MD.
> Migration will happen as part of #35 (GitHub Projects sync). Existing sequential FDs will coexist — no forced migration.

### Format duality — migration note

> **Current state**: Markdown files with YAML frontmatter (`fd/FD-001.md`).
> **Target state**: YAML source of truth with derived MD (`fd/FD-a3f2.yaml` + `fd/FD-a3f2.md`).
> This is a breaking change that happens with the Go rewrite. Migration plan TBD in #35.

### Format duality

```
*.yaml    ← source of truth (agent reads/writes, CLI parses, DB syncs)
*.md      ← derived view (auto-generated for human review, PR diff, Obsidian)
```

The MD is never edited directly. Changes go through YAML, MD is regenerated by `forgia render` (a CLI subcommand that converts YAML → Markdown for any artifact).

**Exceptions** (Markdown-only, human-authored):
- `constitution.md` — prose rules, not structured data
- `dev-guide/*.md` — conventions, not parseable metadata
- `guardrails/ignore` — gitignore-style patterns

## 4. Memory Tiers

```mermaid
flowchart TD
    subgraph tier1 ["Tier 1: Hot Memory (always loaded, ~800 lines)"]
        Arch["architecture/*.yaml\nsystem-context, containers,\ntechnology-decisions"]
        Const["constitution.md"]
        Guard["guardrails/deny.toml"]
        Principles["principles/\nclean-code, SOLID, patterns"]
    end

    subgraph tier2 ["Tier 2: Domain (loaded per bounded context, ~2000 lines)"]
        Context["contexts/<domain>.yaml\ninterfaces, deps, language"]
        Lang["lang/<detected>.md\nGo, Rust, TS, K8s, Shell"]
        FDCurrent["Current FD"]
        SDDCurrent["Current SDD"]
    end

    subgraph tier3 ["Tier 3: Cold (on-demand via MCP, unlimited)"]
        CodeGraph["codebase-memory-mcp\nsearch_graph, trace_call_path,\ndetect_changes, get_architecture"]
        ADR["Closed FDs\n(retrospectives, decisions)"]
        Code["Source code\n(only referenced files)"]
    end

    tier1 -->|"every session"| Agent["Agent"]
    tier2 -->|"per domain"| Agent
    tier3 -->|"on-demand query"| Agent

    style tier1 fill:#f8d7da,stroke:#dc3545
    style tier2 fill:#fff3cd,stroke:#ffc107
    style tier3 fill:#cce5ff,stroke:#0d6efd
```

| Tier | What | Loaded when | Size |
|------|------|-------------|------|
| 1 (Hot) | Constitution, guardrails, architecture overview, principles | Every session | ~800 lines |
| 2 (Domain) | Bounded context, language conventions, current FD/SDD | Per domain | ~2000 lines |
| 3 (Cold) | Code graph, closed FDs, source code | On-demand MCP query | Unlimited |

## 5. Security Layers

```mermaid
flowchart TD
    subgraph L1 ["L1: Prompt (deny.toml in agent context)"]
        DenyPrompt["Agent told:\nNEVER read .env, .ssh, .gnupg\nNEVER execute pass show, env grep\nNEVER write constitution, config"]
    end

    subgraph L2 ["L2: Permissions (settings.json / permission-mode)"]
        Host["Host mode:\nsettings.json (286 allow, 45 deny)\npermission-mode: default"]
        SandboxPerm["Sandbox mode:\n--dangerously-skip-permissions\n(safe because L3 isolates)"]
    end

    subgraph L3 ["L3: OS Isolation (Docker Engine API)"]
        Mount["Only project dir mounted\nNo ~/.ssh, ~/.gnupg, ~/.aws"]
        Net["Network allowlist\napi.anthropic.com, github.com"]
        Resources["CPU/memory limits"]
    end

    subgraph L4 ["L4: Post-Exec Scan"]
        FileCheck["Scan created files\nvs deny.toml write patterns"]
        SecretCheck["Scan file contents\nfor API key patterns"]
        Rollback["Auto-rollback\non violation"]
    end

    L1 --> L2 --> L3 --> L4

    style L1 fill:#fff3cd,stroke:#ffc107
    style L3 fill:#f8d7da,stroke:#dc3545
    style L4 fill:#f8d7da,stroke:#dc3545
```

| Layer | Host (architect) | Sandbox (builder) |
|-------|-----------------|-------------------|
| L1 Prompt | deny.toml injected | deny.toml injected |
| L2 Permissions | settings.json + default mode | --dangerously-skip-permissions |
| L3 OS | None (dev is in the loop) | Docker: mount + network + resources |
| L4 Post-Exec | Optional | Mandatory |

> **Note**: RTK (token compression) lives in the sandbox as a **cost optimization**, not a security layer. See §6 Dual-Profile Runner.

## 6. Dual-Profile Runner

```mermaid
flowchart TD
    Command["forgia exec / watch / batch"]

    Command --> Config{"config.toml\nsandbox setting?"}

    Config -->|"sandbox = none"| Host["Host Profile\n(architect)"]
    Config -->|"sandbox = docker"| Sandbox["Sandbox Profile\n(builder)"]

    Host --> Claude_Host["claude\n--permission-mode auto\non the Mac directly"]

    Sandbox --> Docker_API["Docker Engine API\n(Unix socket)"]
    Docker_API --> Container["Sandbox Container"]
    Container --> RTK_Hook["RTK auto-rewrite hook"]
    Container --> Claude_Sandbox["claude\n--dangerously-skip-permissions"]

    Claude_Host --> PostExec["Post-Exec\nUpdate status\nWrite report"]
    Claude_Sandbox --> PostExec

    style Host fill:#fff3cd,stroke:#ffc107
    style Sandbox fill:#d4edda,stroke:#28a745
    style Container fill:#cce5ff,stroke:#0d6efd
```

### OpenHands Runner

When configured as runner, OpenHands provides full container isolation with its own agent loop:

```mermaid
flowchart TD
    Forgia["forgia exec\n--runner=openhands"] --> Prompt["Build prompt\n(constitution + guardrails + SDD)"]
    Prompt --> Container["OpenHands Container\n(Docker, API key)"]

    subgraph oh ["OpenHands Agent Loop"]
        Read["Read SDD"] --> Plan["Plan"]
        Plan --> Code["Write Code"]
        Code --> Test["Run Tests"]
        Test -->|fail| Fix["Fix"]
        Fix --> Test
        Test -->|pass| Commit["Commit + PR"]
    end

    Container --> oh
    oh --> Report["Extract Work Log\nUpdate SDD status"]

    style oh fill:#cce5ff,stroke:#0d6efd
```

| Aspect | Claude Code (sandbox) | OpenHands |
|--------|----------------------|-----------|
| Auth | Max subscription (flat) | API key (per-token) |
| Agent loop | Claude native | OpenHands own loop |
| Model | Claude only | Any (Claude, GPT, Ollama) |
| Isolation | Docker via Forgia | Docker via OpenHands |
| UI | None (headless) | Web UI on :3000 (optional) |

## 7. External Service Integration

```mermaid
flowchart LR
    subgraph forgia ["Forgia Go Binary"]
        Core["Core\n(vault, config, guardrails)"]
        TP["ToolProvider\nRegistry"]
        PB["ProjectBoard\nInterface"]
        RM["Runner\nManager"]
        BC["Beads\nClient"]
    end

    subgraph services ["External Services (all optional)"]
        CM["codebase-memory-mcp\nToolProvider impl"]
        GH["GitHub Projects\nProjectBoard impl"]
        GL["GitLab Board\nProjectBoard impl"]
        LB["LocalBoard\nProjectBoard fallback"]
        DockerE["Docker Engine\nRunner impl"]
        BD["Beads (bd CLI)\nlocal cache"]
    end

    TP --> CM
    PB --> GH
    PB --> GL
    PB --> LB
    RM --> DockerE
    BC --> BD

    style forgia fill:#fff3cd,stroke:#ffc107
    style services fill:#cce5ff,stroke:#0d6efd
```

**Design principle: everything optional with graceful fallback.**

| Service | Interface | Required? | Fallback |
|---------|-----------|-----------|----------|
| codebase-memory-mcp | `ToolProvider` | No | No Tier 3, skills work with less context |
| GitHub Projects | `ProjectBoard` | No | `LocalBoard` (filesystem only) |
| GitLab Board | `ProjectBoard` | No | `LocalBoard` |
| Docker Engine | `DockerSandbox` | No | Host mode (permission-mode auto) |
| RTK | In-container hook | No | No token compression |
| Beads (bd) | `BeadsClient` | No | Vault files only (no dep graph) |
| Claude Code | `Runner` | Yes | Primary runner (only required service) |

## 8. The Autonomous Cycle (Future Vision)

```mermaid
flowchart TD
    Arch["Architecture\n+ DDD Contexts"] --> Propose

    subgraph cycle ["Autonomous Cycle"]
        Propose["Agent proposes FD\n(based on completed FDs\n+ architecture gaps\n+ context files)"]

        Propose --> SelfReview["Self-Review\n(coherence with\nprevious FDs)"]

        SelfReview --> Gate{"Intelligent Gate\nCoherent?"}

        Gate -->|"Conflict found"| Stop["STOP\nNotify human\nExplain conflict"]

        Gate -->|"Coherent"| Generate["/fd-sdd → SDDs\nforgia watch → execute\n/fd-verify → check"]

        Generate --> Feedback["Feedback Loop\nUpdate architecture\nUpdate contexts\nClose FD"]

        Feedback --> Propose
        Stop -->|"Human resolves"| Propose
    end

    style Gate fill:#f8d7da,stroke:#dc3545
    style Stop fill:#f8d7da,stroke:#dc3545
    style Propose fill:#d4edda,stroke:#28a745
    style Feedback fill:#cce5ff,stroke:#0d6efd
```

### Gate checks (uses all 3 memory tiers)

| Check | Tier | Source |
|-------|------|--------|
| Contradicts previous FD decisions? | 3 (Cold) | Closed FDs retrospectives |
| Breaks bounded context contracts? | 2 (Domain) | contexts/*.yaml interfaces |
| Quality attributes still achievable? | 1 (Hot) | architecture/quality-attributes.yaml |
| Uses rejected patterns? | 3 (Cold) | Work Log failure modes |
| Dependencies satisfied? | 2 (Domain) | contexts/*.yaml dependencies |
| Blast radius acceptable? | 3 (Cold) | codebase-memory-mcp detect_changes |

### Human gates

The human intervenes at **3 moments only**:

1. **Approve FD** — agent proposed and self-reviewed, human confirms
2. **Resolve conflict** — gate blocked, human decides how to proceed
3. **Review PR** — code is ready, human merges

Everything else is autonomous.

### Who triggers the cycle

| Trigger | How | When |
|---------|-----|------|
| Manual | Developer runs `/fd-new` or `forgia status` → decides next FD | Default, always available |
| Cron | `claude -p "check forgia status, propose next FD"` via crontab | Nightly/weekly autonomous check |
| GitHub Action | On schedule or on issue label `forgia:fd` | CI-driven |
| MCP tool | Another agent calls `forgia_next_fd()` | Agent-to-agent orchestration |

## 9. Feedback Loop — From Code Back to Architecture

The critical missing piece: after code is produced, lessons learned must flow **back** to DDD documents, influencing future FDs. Without this, the architecture diverges from reality.

### The Gap (today)

```
DDD → FD → SDD → Code → Work Log → STOP
                                     ↑ feedback dies here
```

### The Complete Loop

```mermaid
flowchart TD
    SDD_Done["SDD Done\n(Work Log filled)"] --> Verify["/fd-verify\n(all SDDs pass)"]

    Verify --> Close["/fd-close FD-a3f2"]

    Close --> Feedback["/fd-feedback FD-a3f2\n(NEW SKILL)"]

    Feedback --> C1["Update contexts/\n(actual interfaces\nvs planned)"]
    Feedback --> C2["Update architecture/\n(new containers,\nrevised quality attrs)"]
    Feedback --> C3["Update glossary\n(new terms introduced)"]
    Feedback --> C4["Write learning record\n(.forgia/learnings/\nfailure modes, patterns)"]

    C1 & C2 & C3 & C4 --> Review["/arch-review\n(coherence check\nafter updates)"]

    Review --> NextFD["Next FD proposal\n(gate reads updated\narchitecture + learnings)"]

    style Feedback fill:#fff3cd,stroke:#ffc107
    style C4 fill:#f3e5f5,stroke:#9c27b0
    style Review fill:#f8d7da,stroke:#dc3545
```

### What `/fd-feedback` does

New skill that runs after `/fd-close`. Reads all Work Logs of the closed FD and routes feedback to the right documents:

| Work Log content | Routes to | Example |
|-----------------|-----------|---------|
| Interface was different than planned | `contexts/<name>.yaml` interfaces section | "CheckRisk takes position + account, not just position" |
| New container/service introduced | `architecture/containers.yaml` | "Added Redis cache for session state" |
| Technology decision changed | `architecture/technology-decisions.yaml` | "Switched from gRPC to HTTP for internal — simpler debugging" |
| Latency/performance measured | `architecture/quality-attributes.yaml` | "Actual p99: 85ms (planned: 100ms) — margin OK" |
| New term introduced | `architecture/glossary.yaml` | "BridgeTimeout: max wait for MT5 bridge response" |
| Failure mode discovered | `.forgia/learnings/<fd-id>.yaml` (NEW) | "gRPC streaming caused memory leak under load" |
| Pattern that worked well | `.forgia/learnings/<fd-id>.yaml` | "Builder pattern for config was clean and testable" |
| Suggestion for future FDs | `.forgia/learnings/<fd-id>.yaml` | "Add integration test template to SDD for DB-dependent features" |

### Learnings directory (new)

```
.forgia/
  learnings/                        ← NEW: persisted knowledge from execution
    FD-a3f2.yaml                    ← learnings from FD-a3f2
    FD-b7c1.yaml                    ← learnings from FD-b7c1
```

```yaml
# .forgia/learnings/FD-a3f2.yaml
fd: FD-a3f2
title: "Scaffold ZeroClaw + Tauri"
closed: "2026-03-20"

failure_modes:
  - pattern: "gRPC streaming"
    context: "Internal service communication"
    outcome: "Memory leak under sustained load"
    recommendation: "Use HTTP for internal, gRPC only for external"

successful_patterns:
  - pattern: "Builder pattern for config"
    context: "Multi-field struct initialization"
    outcome: "Clean, testable, extensible"

suggestions:
  - "Add DB integration test template to SDD for features with persistence"
  - "Include load test acceptance criterion for any streaming interface"

interface_corrections:
  - context: "trading-engine"
    interface: "CheckRisk"
    planned: "check_risk(position) → approved/rejected"
    actual: "check_risk(position, account) → RiskResult{approved, reason, limits}"

new_terms:
  - term: "BridgeTimeout"
    meaning: "Max wait for MT5 bridge response before circuit breaker trips"
```

### How the Gate Intelligente uses learnings

When a new FD is proposed, the gate reads `.forgia/learnings/` (Tier 3 Cold Memory):

```mermaid
flowchart LR
    NewFD["Proposed FD-c4d5\nuses gRPC streaming"] --> Gate["Gate Intelligente"]

    Gate --> ReadLearnings["Read learnings/FD-a3f2.yaml"]

    ReadLearnings --> Match["MATCH: failure_mode\n'gRPC streaming → memory leak'"]

    Match --> Block["STOP: FD-a3f2 found that gRPC streaming\ncauses memory leaks. Consider HTTP instead.\nSee: .forgia/learnings/FD-a3f2.yaml"]

    style Block fill:#f8d7da,stroke:#dc3545
    style Match fill:#fff3cd,stroke:#ffc107
```

### Complete lifecycle with feedback

```mermaid
stateDiagram-v2
    [*] --> Architecture: /arch-init

    Architecture --> FD: /fd-new
    FD --> SDD: /fd-sdd (after /fd-review)
    SDD --> Code: forgia exec/watch/batch
    Code --> Verify: /fd-verify
    Verify --> Close: /fd-close

    Close --> FeedbackSkill: /fd-feedback (NEW)

    state "Feedback Routing" as FeedbackSkill {
        [*] --> ReadWorkLogs: Read all SDD Work Logs
        ReadWorkLogs --> RouteContexts: Update contexts/ (interfaces)
        ReadWorkLogs --> RouteArch: Update architecture/ (containers, QA)
        ReadWorkLogs --> RouteLearnings: Write learnings/ (failure modes, patterns)
        RouteContexts --> ArchReview
        RouteArch --> ArchReview
        RouteLearnings --> ArchReview
        ArchReview: /arch-review (coherence)
    }

    FeedbackSkill --> Architecture: updated DDD docs
    Architecture --> FD: next FD (informed by learnings)
```

### Skill definition

```
/fd-feedback FD-a3f2
```

| Attribute | Value |
|-----------|-------|
| Category | `architecture` |
| Mode | `both` (slash command + MCP tool `forgia_fd_feedback`) |
| Triggers | After `/fd-close` (can be called manually or auto-triggered) |
| Reads | All SDD Work Logs for the closed FD |
| Writes | `contexts/`, `architecture/`, `learnings/` |
| Then runs | `/arch-review` to verify coherence |

### Auto-trigger on `/fd-close`

`/fd-close` should call `/fd-feedback` automatically:

```
/fd-close FD-a3f2
  1. Archive FD with retrospective
  2. → /fd-feedback FD-a3f2 (auto-triggered)
       a. Read Work Logs
       b. Route to contexts/architecture/learnings
       c. Run /arch-review
  3. Update board card → "Closed"
  4. Done
```

## 10. Project Board Sync

```mermaid
sequenceDiagram
    participant Vault as .forgia/ (YAML)
    participant Forgia as Forgia MCP
    participant Board as ProjectBoard
    participant Beads as Beads (cache)

    Note over Vault: FD created locally

    Vault->>Forgia: FD-a3f2 created
    Forgia->>Board: CreateCard(FD-a3f2)
    Board-->>Forgia: card confirmed
    Forgia->>Beads: cache card + ID

    Note over Vault: FD status changes

    Vault->>Forgia: FD-a3f2 → approved
    Forgia->>Board: MoveCard(FD-a3f2, "Approved")
    Forgia->>Beads: update cache

    Note over Board: Another engineer moves card

    Board->>Forgia: card FD-b7c1 → "In Progress"
    Forgia->>Vault: update FD-b7c1.yaml status
    Forgia->>Beads: update cache
```

### Competitive FDs on the board

| FD Proposed          | FD Approved    | SDD In Progress    | Done     |
|---------------------|----------------|-------------------|----------|
| FD-a3f2 (approach A) | FD-c4d5        | SDD-001 of FD-c4d5 | FD-x1y2  |
| FD-b7c1 (approach B) |                | SDD-002 of FD-c4d5 |          |
|   └ competes with a  |                |                    |          |

When FD-a3f2 is approved, FD-b7c1 automatically moves to "Rejected" column with reason.

### Card created from web (reverse sync)

A PM or stakeholder can create a card directly on the board (web UI). Forgia syncs it back:

```mermaid
sequenceDiagram
    participant PM as PM (web)
    participant Board as ProjectBoard
    participant Forgia as Forgia MCP
    participant Vault as .forgia/

    PM->>Board: Create card "New feature idea"
    Note over Board: Card has no .forgia/ file yet

    Forgia->>Board: SyncToVault() (poll or webhook)
    Board-->>Forgia: new card detected
    Forgia->>Forgia: generate hash ID (FD-c4d5)
    Forgia->>Vault: create FD-c4d5-new-feature.yaml (status: planned, from board)
    Forgia->>Board: update card with FD-c4d5 ID
```

## 10. Skill System

A skill is any capability Forgia exposes — to humans (slash commands), to agents (MCP tools), or both. Three types:

```mermaid
flowchart TD
    subgraph types ["Skill Types"]
        SC["SlashCommand\n(/fd-new, /fd-review)\nUser invokes explicitly"]
        MT["MCPTool\n(forgia_blast_radius)\nAgent invokes directly"]
        Both["Both\n(/arch-review + forgia_arch_review)\nSame Go logic, two entry points"]
    end

    subgraph impl ["Implementation"]
        Native["Native Skill\nGo logic in internal/"]
        Composite["Composite Skill\nWraps external MCP provider\n+ adds Forgia logic"]
    end

    SC --> Native
    MT --> Native
    MT --> Composite
    Both --> Native

    subgraph providers ["Composite sources"]
        CM["codebase-memory-mcp\nsearch_graph, detect_changes,\ntrace_call_path, get_architecture"]
    end

    Composite --> providers

    style SC fill:#fff3cd,stroke:#ffc107
    style MT fill:#cce5ff,stroke:#0d6efd
    style Both fill:#d4edda,stroke:#28a745
    style Composite fill:#f3e5f5,stroke:#9c27b0
```

### Skill Registry

All skills register in a central `skill.Registry`. The MCP server and slash command installer both read from it:

```
skill.Registry
  ├── SlashCommands() → list of .md files to install
  ├── MCPTools()      → list of MCP tool definitions to advertise
  └── Get(name)       → execute skill logic
```

### Slash Command Skills (user-facing)

| Skill | Category | Mode | Issue |
|-------|----------|------|-------|
| `/fd-new` | fd | both | #29, #30 |
| `/fd-deep` | fd | slash_command | — |
| `/fd-review` | fd | both | #18 |
| `/fd-sdd` | fd | both | #35 |
| `/fd-verify` | fd | both | #17 |
| `/fd-close` | fd | both | #29 |
| `/fd-feedback` | architecture | both | #27 (feedback loop) |
| `/fd-status` | fd | both | — |
| `/arch-init` | architecture | both | #27 |
| `/arch-review` | architecture | both | #27 |
| `/arch-update` | architecture | both | #27 |
| `/fd-threat-model` | design | both | #21 |
| `/fd-arch-review` | design | both | #22 |
| `/sdd-dry-run` | design | both | #23 |

### Composite Skills (codebase-memory-mcp derived)

These wrap external MCP provider tools with Forgia logic (guardrails check, context enrichment):

| Skill | Provider Tool | Forgia adds | Used by |
|-------|--------------|-------------|---------|
| `forgia_search_code` | `search_graph` | Guardrails filter on results | Agent search |
| `forgia_blast_radius` | `detect_changes` | Cross-check with bounded contexts | Gate intelligente |
| `forgia_trace_calls` | `trace_call_path` | Validate context interface contracts | `/arch-review` |
| `forgia_architecture` | `get_architecture` | Merge with existing architecture/ | `/arch-init` |
| `forgia_dead_code` | `search_graph(dead_code)` | Filter by SDD scope | Cleanup |
| `forgia_reindex` | `index_repository` | Trigger after exec | Post-exec |

### Skill evolution

```
Today (bash):       .md file → Claude interprets → Read/Write/Bash
Transition:         .md file → Claude calls forgia CLI → Go logic
Future (MCP):       .md thin wrapper → forgia_tool() MCP call → Go logic
                    Agent can also call MCP tool directly (no .md needed)
```

Slash commands stay for **explicit UX** and **context savings**. MCP tools exist for **agent-to-agent** calls.

## 11. Configuration Hierarchy

```mermaid
flowchart TD
    subgraph committed ["Committed to git (shared)"]
        Constitution["constitution.md\n(immutable principles)"]
        Config["config.toml\n(team settings)"]
        Guardrails["guardrails/deny.toml\n(security rules)"]
        DevGuide["dev-guide/\n(conventions)"]
        Arch["architecture/\n(system design)"]
        Contexts["contexts/\n(bounded contexts)"]
    end

    subgraph personal ["Personal (gitignored)"]
        LocalConfig["config.local.toml\n(runner prefs, max_turns)"]
        Logs["logs/\n(exec reports)"]
        BeadsDB[".beads/\n(local cache)"]
    end

    subgraph protected ["Protected (CODEOWNERS)"]
        Constitution
        Guardrails
        Config
        Arch
    end

    subgraph open ["Open to all"]
        FDs["fd/\n(anyone can propose)"]
        SDDs["sdd/\n(generated)"]
    end

    style committed fill:#d4edda,stroke:#28a745
    style personal fill:#f0f0f0,stroke:#999
    style protected fill:#f8d7da,stroke:#dc3545
```

### Loading priority

```
CLI flags > Environment vars > config.local.toml > config.toml > Go defaults
```

## 12. Feature Dependency Graph

```mermaid
flowchart TD
    Scaffold["#28 Go Scaffold ✅"]

    Scaffold --> Interfaces["#41 Go Interfaces"]

    Interfaces --> Board["#35 Project Board\n+ Hash ID\n+ Beads Cache"]
    Interfaces --> CodeMem["#36 codebase-memory-mcp\n(Tier 3)"]

    Board --> Competitive["#29 Competitive FD\n+ Reject"]
    Board --> MultiEng["#30 pt2 Multi-Eng CI"]

    CodeMem --> DDD["#27 DDD Skills\n(arch-init, review, update)"]

    DDD --> Threat["#21 /fd-threat-model"]
    DDD --> ArchReview["#22 /fd-arch-review"]
    CodeMem --> DryRun["#23 /sdd-dry-run"]

    DDD --> DesignParent["#20 Design Skills (parent)"]
    Threat --> DesignParent
    ArchReview --> DesignParent
    DryRun --> DesignParent

    subgraph phase3 ["Phase 3 — v1.1.0"]
        Sandbox["#32 Dual-Profile\nSandbox + RTK"]
        Tokens["#13 Token/Cost\nTracking"]
        PrePost["#17 Pre/Post\nExec Scan"]
        Scoring["#18 Compliance\nScoring"]
        Bounds["#19 Behavior\nConstraints"]
        Parallel["#15 Parallel\n(Agent Orchestrator)"]
    end

    DesignParent -.-> phase3

    style Scaffold fill:#d4edda,stroke:#28a745
    style Board fill:#fff3cd,stroke:#ffc107
    style CodeMem fill:#cce5ff,stroke:#0d6efd
    style phase3 fill:#f0f0f0,stroke:#999
```

## References

- [go-integration-guide.md](go-integration-guide.md) — Go code patterns, process management
- [Forgia README](../README.md) — user-facing documentation
- [Doc 09 — Architecture-Driven Development](../../docs/spec-driven-development/09-architecture-driven-development.md) — conceptual model (in ai-bots knowledge base)
- [C4 Model](https://c4model.com/)
- [DDD Bounded Context](https://martinfowler.com/bliki/BoundedContext.html)
- [Codified Context paper](https://arxiv.org/html/2602.20478v1) — 3-tier memory
