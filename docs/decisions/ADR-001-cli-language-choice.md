# ADR-001: CLI language choice for Forgia v1.0.0

## Status

Accepted

## Context

La CLI di Forgia e' attualmente uno script Bash di ~600 righe (`bin/forgia`) che ha raggiunto il suo limite di complessita'. Il progetto si sta muovendo verso la v1.0.0 con requisiti che Bash non puo' soddisfare:

- **MCP server**: JSON-RPC over stdio e' impraticabile in Bash
- **Parsing strutturato**: TOML/YAML gestiti con `grep` + `sed` si rompono su edge case (commenti, valori multiline)
- **Error handling**: `set -e` + return code non distingue errori recuperabili da fatali
- **Cross-platform**: macOS e Linux divergono su `sed`, `find`, `stat`, `wc`
- **Testing**: ogni test deve lanciare un sottoprocesso, rendendo la suite lenta e fragile

Serve un linguaggio unico per CLI, MCP server, slash command engine, guardrails e runner abstraction, tutto da un unico codebase.

Tre candidati sono stati valutati tramite spike prototipi (SDD-001): **Go**, **Rust**, **TypeScript**.

## Evaluation Criteria

I criteri derivano da FD-008:

1. **MCP feasibility** — capacita' di implementare un server MCP (JSON-RPC over stdio)
2. **Parsing** — supporto per TOML, YAML e frontmatter markdown
3. **Cross-platform** — macOS + Linux senza workaround
4. **Distribution** — semplicita' di distribuzione (binario singolo preferito)
5. **Contributor accessibility** — facilita' di onboarding per nuovi contributori
6. **Iteration speed** — velocita' di sviluppo e ciclo edit-compile-test
7. **Library maturity** — maturita' dell'ecosistema per le esigenze del progetto

## Benchmark Data (SDD-001)

Dati raccolti su macOS Darwin 25.2.0 (Apple Silicon arm64). Media di 5 run (escluso cold run).

### MCP Server

| Metrica | Go | Rust | TypeScript |
|---|---|---|---|
| Dimensione binario | 5.8 MB | 590 KB | N/A (73 MB node_modules) |
| Tempo avvio (warm) | ~10 ms | ~10 ms | ~330 ms |
| RSS a idle | ~10 MB | ~3.3 MB | ~83 MB |

### Frontmatter + TOML CLI

| Metrica | Go | Rust | TypeScript |
|---|---|---|---|
| Dimensione binario | 4.0 MB | 1.1 MB | N/A (51 MB node_modules) |
| Tempo avvio (warm) | ~5 ms | ~10 ms | ~330 ms |
| RSS a idle | ~5 MB | ~3.8 MB | ~83 MB |

## Evaluation Matrix

| Criterio | Go | Rust | TypeScript |
|---|---|---|---|
| MCP feasibility | `mcp-go` SDK maturo, integrazione semplice | `rmcp` richiede rustc >= 1.85, fallback a JSON-RPC manuale | SDK ufficiale (`@modelcontextprotocol/sdk`), best-in-class |
| Parsing (TOML/YAML) | `pelletier/go-toml/v2`, `gopkg.in/yaml.v3` — stabili e ben documentati | `serde` + derive macro — ergonomico ma ecosistema frontmatter meno maturo | `gray-matter`, `@iarna/toml` — funzionali, ecosistema ampio |
| Cross-platform | Binario statico, `GOOS`/`GOARCH` per cross-compile triviale | Binario statico, cross-compile possibile ma piu' complesso (linker) | Richiede Node.js runtime su ogni piattaforma |
| Distribution | Binario singolo ~5 MB, zero dipendenze runtime | Binario singolo ~1 MB, zero dipendenze runtime | `npx` o `node_modules` — richiede Node.js installato |
| Contributor accessibility | Curva di apprendimento bassa, linguaggio semplice | Borrow checker e lifetime aggiungono frizione significativa | Molto accessibile, ma runtime dependency e' un trade-off |
| Iteration speed | Compilazione veloce (~1s), `go test` immediato | Compilazione lenta (10-30s per build incrementali), friction del borrow checker | Nessuna compilazione, ma startup ~330 ms per ogni run |
| Library maturity | Ecosistema CLI maturo (`cobra`, `viper`), MCP SDK stabile | `clap` eccellente, ma MCP SDK (`rmcp`) richiede toolchain recente | Ecosistema vastissimo, MCP SDK ufficiale |

## Decision

**Go e' il linguaggio scelto per la riscrittura della CLI di Forgia.**

### Justification

1. **Miglior bilanciamento performance/ergonomia**: startup ~5-10 ms e RSS ~5-10 MB sono eccellenti per un tool CLI. Rust e' piu' performante in assoluto (RSS ~3 MB, binario ~1 MB) ma la differenza non e' significativa per un CLI tool — nessun utente percepisce la differenza tra 5 ms e 10 ms di startup.

2. **MCP SDK maturo**: `mcp-go` funziona out-of-the-box. Il prototipo SDD-001 ha integrato il server MCP senza problemi. L'SDK Rust (`rmcp`) invece richiede rustc >= 1.85 (edition 2024), incompatibile con la toolchain attuale (1.81), ed ha richiesto un fallback a JSON-RPC manuale nello spike.

3. **Distribution ottimale**: un singolo binario statico di ~5 MB senza dipendenze runtime. TypeScript richiede Node.js (73 MB di node_modules solo per il server MCP), inaccettabile per un tool che deve essere semplice da installare.

4. **Iteration speed superiore**: compilazione in ~1 secondo, `go test` immediato. Rust ha tempi di compilazione 10-30x piu' lunghi. TypeScript non compila ma ha startup 30x piu' lento per ogni esecuzione.

5. **Contributor accessibility**: Go ha la curva di apprendimento piu' bassa tra i tre candidati. Il linguaggio e' volutamente semplice — meno costrutti da imparare, meno sorprese. Questo abbassa la barriera di ingresso per nuovi contributori.

6. **Ecosistema CLI completo**: `cobra` + `viper` sono lo standard de facto per CLI in Go. Auto-generazione di help, completions, e configurazione da file/env/flag. Equivalenti in Rust (`clap`) e TypeScript (`commander`) esistono ma l'integrazione con l'ecosistema Go e' piu' coesa.

### TypeScript e' stato scartato perche':

- Startup ~330 ms (30x piu' lento di Go/Rust) — percepibile dall'utente su ogni invocazione CLI
- RSS ~83 MB (25x piu' di Go) — eccessivo per un tool da riga di comando
- Richiede Node.js runtime — aggiunge frizione all'installazione per utenti non-JS
- Nonostante l'SDK MCP ufficiale sia un vantaggio, non compensa i trade-off su performance e distribution

### Rust e' stato scartato perche':

- L'SDK MCP (`rmcp`) non e' utilizzabile con la toolchain attuale (richiede edition 2024, rustc >= 1.85)
- Tempi di compilazione significativamente piu' lunghi rallentano il ciclo di sviluppo
- Il borrow checker aggiunge frizione su un codebase che e' prevalentemente I/O e manipolazione stringhe
- La curva di apprendimento limita il pool di contributori
- I vantaggi in performance (binario piu' piccolo, RSS minimo) non sono rilevanti per questo use case

## Consequences

### Cosa cambia

- **Nuovo codebase Go** in `cmd/forgia/` (o struttura equivalente) che sostituira' `bin/forgia`
- **MCP server** integrato nel binario Go, esposto come sottocomando (`forgia mcp serve`)
- **Guardrails engine** in Go, con parsing nativo di `deny.toml`
- **Runner abstraction** in Go, per orchestrare agent (Claude Code, OpenHands, manuale)
- **Slash commands** potranno essere implementati come tool MCP nativi

### Trade-off accettati

- **Verbose error handling**: `if err != nil` aggiunge boilerplate, ma rende gli errori espliciti (in linea con la constitution: "no silent fallbacks")
- **Agent Orchestrator (Composio) e' TypeScript**: la comunicazione avverra' via IPC (subprocess o HTTP) invece che per chiamata diretta di libreria. Questo e' accettabile perche' l'orchestrator e' un componente separato.
- **Generics limitati**: Go ha generics dal 1.18 ma l'ecosistema li usa con moderazione. Per un CLI tool questo non e' un problema.

### Backward compatibility

- `bin/forgia` (Bash) rimane funzionante durante il periodo di migrazione
- I test e2e esistenti (`tests/e2e.sh`) devono restare verdi
- La migrazione avverra' comando per comando: ogni sottocomando Go sostituisce il corrispondente Bash
- Una volta che tutti i comandi sono migrati, `bin/forgia` diventa un wrapper che delega al binario Go, poi viene rimosso

### Follow-up work

1. **SDD per scaffolding Go** — struttura del progetto Go (`cmd/`, `internal/`, `go.mod`), primo comando (`forgia status`)
2. **SDD per config parser** — parsing nativo di `config.toml`, `deny.toml`, frontmatter SDD/FD
3. **SDD per MCP server** — implementazione completa del server MCP con tool `forgia_status`, `fd_list`, `sdd_list`
4. **SDD per guardrails engine** — enforce di `deny.toml` pre/post esecuzione agent
5. **SDD per runner abstraction** — supporto Claude Code, OpenHands, esecuzione manuale
6. **Aggiornamento CLAUDE.md** — riflettere il nuovo stack Go nella documentazione del progetto

### Timeline stimata

La migrazione e' incrementale. Non serve una riscrittura big-bang: ogni comando migrato e' immediatamente utilizzabile mentre il resto continua a funzionare in Bash.
