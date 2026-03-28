# Design Patterns Reference

> Always loaded. Use the RIGHT pattern for the problem — don't force patterns where they don't fit.

## When to Use Patterns

Patterns are **solutions to recurring problems**, not goals. Apply them when:
- You recognize the problem they solve
- The alternative is demonstrably worse
- The pattern simplifies, not complicates

**Never use a pattern just because you can.** Simple code > clever code.

---

## Creational

### Builder
**When**: Object has many optional parameters, or construction is multi-step.
```
config = ConfigBuilder().port(8080).tls(true).timeout(30).build()
```
**Not when**: Object has 1-3 required fields — use a constructor.

### Factory
**When**: The concrete type depends on runtime data (config, feature flags, user input).
```
def create_bridge(venue: str) -> Bridge:
    match venue:
        case "binance": return BinanceBridge()
        case "paper":   return PaperBridge()
```
**Not when**: You always know which type you need at compile time.

### Singleton
**Almost never.** Use dependency injection instead. Singletons hide dependencies and make testing hard.
**Exception**: Logger, connection pool — things with true process-wide identity.

---

## Structural

### Adapter
**When**: You need to make an existing interface work with code expecting a different interface.
```
// Your code expects Authenticator, but LDAP library has its own API
struct LdapAdapter { client: LdapClient }
impl Authenticator for LdapAdapter { ... }
```

### Facade
**When**: A subsystem has many moving parts and callers don't need the complexity.
```
// Instead of: connect(), authenticate(), query(), parse(), close()
result = database.execute("SELECT ...")
```

### Decorator / Middleware
**When**: Adding behavior (logging, auth, caching, retry) without modifying existing code.
```
app.use(auth_middleware)
app.use(logging_middleware)
app.use(rate_limit_middleware)
```

---

## Behavioral

### Strategy
**When**: You need to swap algorithms at runtime.
```
class TradingEngine:
    def __init__(self, strategy: Strategy):
        self.strategy = strategy

    def evaluate(self, candles):
        return self.strategy.signal(candles)
```

### Observer / Event
**When**: Multiple components need to react to state changes without coupling.
```
event_bus.emit("position_opened", position)
// Multiple listeners: logger, risk_manager, notifier
```
**Not when**: There's only one listener — direct call is simpler.

### State Machine
**When**: An entity has distinct states with defined transitions.
```
Draft → Submitted → Approved → InProgress → Done
                  ↘ Rejected → Draft
```
**Use for**: Order lifecycle, FD status, connection state, deployment pipeline.

### Command
**When**: You need to queue, undo, or replay operations.
```
commands = [OpenPosition(...), SetStopLoss(...), ClosePosition(...)]
for cmd in commands:
    cmd.execute()
```

---

## Architecture Patterns

### Repository
**When**: Separating data access from business logic.
```
trait UserRepository {
    fn find_by_id(&self, id: &str) -> Result<User>;
    fn save(&self, user: &User) -> Result<()>;
}
// Implementations: PostgresUserRepo, InMemoryUserRepo (for tests)
```

### CQRS (Command Query Responsibility Segregation)
**When**: Read and write models have different shapes or performance needs.
- **Command**: write, validate, enforce rules
- **Query**: read, optimize for display, denormalize if needed
**Not when**: Simple CRUD — CQRS adds complexity.

### Event Sourcing
**When**: You need complete audit trail, or the "how we got here" matters as much as "where we are".
**Not when**: Simple state management. Event sourcing is powerful but complex.

---

## Anti-Patterns to Avoid

| Anti-Pattern | Problem | Instead |
|---|---|---|
| God Object | One class does everything | Single Responsibility |
| Premature Abstraction | Abstract before you understand the problem | Wait for Rule of Three |
| Golden Hammer | Using one pattern/tool for everything | Choose the right tool |
| Cargo Cult | Using patterns without understanding why | Understand the problem first |
| Lava Flow | Dead code left "just in case" | Delete it, git remembers |
| Spaghetti | No structure, everything calls everything | Clear dependency direction |
