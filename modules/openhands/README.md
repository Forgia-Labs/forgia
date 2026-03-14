# OpenHands Module

Agent runtime per Forgia. Esegue gli SDD in container Docker isolati.

## Setup

```bash
# Install (pull image)
mise run openhands:install

# Configure LLM backend
cp modules/openhands/config.toml.template modules/openhands/config.toml
# Edit config.toml with your API key

# Or use environment variables (recommended)
export ANTHROPIC_API_KEY="sk-ant-..."
```

## Usage

```bash
# Start UI (http://localhost:3000)
mise run openhands:up

# Execute a single SDD
mise run sdd .forgia/sdd/FD-001/SDD-001.md

# Execute all SDDs for an FD in parallel
mise run sdd:batch FD-001

# View logs
mise run openhands:logs

# Stop
mise run openhands:down
```

## Local LLM

Per usare Ollama/Exo invece di Claude API:

```bash
export FORGIA_LLM_MODEL="ollama/qwen2.5:32b"
# Ollama deve essere raggiungibile da dentro il container
# Su macOS: host.docker.internal:11434
```
