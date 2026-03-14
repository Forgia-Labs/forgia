# Python Conventions

> Auto-loaded when `pyproject.toml`, `setup.py`, or `requirements.txt` is detected.

## Project Structure

- Package manager: `uv` (preferred) or `hatch`
- Project config: `pyproject.toml` (not `setup.py`)
- Source layout: `src/<package>/` for libraries, flat for applications
- Tests: `tests/` directory, mirroring source structure

## Type Hints

- Required on all public functions and methods
- Use `from __future__ import annotations` for modern syntax
- Use `typing` module types: `Optional`, `Union`, `list[str]` (3.10+)
- Complex types: define `TypeAlias` or `TypedDict`

```python
from __future__ import annotations

def calculate_risk(
    positions: list[Position],
    max_drawdown: float = 0.05,
) -> RiskAssessment:
    """Calculate portfolio risk given open positions."""
    ...
```

## Async

- Runtime: `asyncio` (stdlib)
- Use `async def` / `await` consistently
- Use `asyncio.gather()` for concurrent tasks
- Use `asyncio.TaskGroup` (3.11+) for structured concurrency
- Never mix sync and async — use `asyncio.to_thread()` for blocking calls

## Error Handling

- Custom exceptions: inherit from domain-specific base exception
- Use `try/except` with specific exception types (never bare `except:`)
- Log exceptions with `logger.exception()` (includes traceback)
- Raise early, handle late

```python
class EngineError(Exception):
    """Base exception for trading engine."""

class BridgeTimeoutError(EngineError):
    """Venue bridge connection timed out."""

class InsufficientMarginError(EngineError):
    """Not enough margin to open position."""
```

## Dependencies

- HTTP server: `FastAPI` + `uvicorn`
- HTTP client: `httpx` (async) or `requests` (sync)
- Database: `sqlalchemy` (async with `asyncpg`) or `psycopg`
- Validation: `pydantic` v2
- Task queue: `celery` or `arq`
- ML: `scikit-learn`, `pandas`, `numpy`

## Style

- Formatter: `ruff format` (replaces black)
- Linter: `ruff check` (replaces flake8, isort, etc.)
- Line length: 88 (ruff default)
- Imports: sorted by `ruff` (stdlib → third-party → local, separated by blank lines)
- Naming: `snake_case` functions/variables, `PascalCase` classes, `UPPER_CASE` constants
- Docstrings: Google style for public functions

## Testing

- Framework: `pytest`
- Fixtures: use `conftest.py` for shared fixtures
- Async tests: `pytest-asyncio` with `@pytest.mark.asyncio`
- Database tests: use real database (not mocks) — see constitution
- Test names: `test_<behavior>_<scenario>`
- Coverage: `pytest-cov`, report to stdout

```python
@pytest.mark.asyncio
async def test_engine_rejects_position_without_approval(
    engine: TradingEngine,
    mock_venue: PaperBridge,
):
    with pytest.raises(ApprovalRequiredError):
        await engine.open_position(symbol="EURUSD", side="buy", size=1.0)
```

## Config

- Use `pydantic-settings` for environment-based config
- TOML files for application config (`tomllib` stdlib in 3.11+)
- Never hardcode secrets — use env vars or secret managers
- Use `.env` files for local development only (never commit)

## What NOT to Do

- Don't use `print()` for logging — use `logging` module
- Don't use mutable default arguments (`def f(x=[])`)
- Don't use `global` or `nonlocal` unless absolutely necessary
- Don't ignore type checker warnings — fix them
- Don't use `*` imports (`from module import *`)
