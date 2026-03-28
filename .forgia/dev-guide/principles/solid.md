# SOLID Principles

> Always loaded. Apply pragmatically — SOLID is a compass, not a GPS.

## S — Single Responsibility

> A module should have one, and only one, reason to change.

- One class/module = one job
- If you can't describe what it does without "and", split it
- **Practical test**: if a change in business logic forces you to change a module that handles persistence, they're coupled

```
# Bad — handles validation AND persistence AND notification
class OrderService:
    def create_order(self, data):
        self.validate(data)
        self.save_to_db(data)
        self.send_email(data)

# Good — each concern is separate
class OrderValidator: ...
class OrderRepository: ...
class OrderNotifier: ...
class CreateOrderUseCase:
    def __init__(self, validator, repo, notifier): ...
```

## O — Open/Closed

> Open for extension, closed for modification.

- Add new behavior by adding new code, not changing existing code
- Use interfaces/traits + implementations, not if/else chains
- **Practical test**: can you add a new variant without touching existing code?

```
# Bad — must modify every time a new venue is added
def connect(venue):
    if venue == "binance": ...
    elif venue == "kraken": ...

# Good — add new venue by implementing trait
trait VenueBridge { fn connect(&self); }
struct BinanceBridge;
struct KrakenBridge;
// Adding CoinbaseBridge doesn't touch existing code
```

## L — Liskov Substitution

> Subtypes must be substitutable for their base types.

- If `Dog` extends `Animal`, anywhere you use `Animal` you must be able to use `Dog`
- Don't override methods to throw "not supported" — that violates LSP
- **Practical test**: does the subtype honor ALL contracts of the parent?

```
# Bad — violates LSP
class ReadOnlyRepo(Repository):
    def save(self, item):
        raise NotImplementedError()  # Surprise!

# Good — separate interfaces
trait Readable { fn find(&self, id: &str) -> Item; }
trait Writable { fn save(&self, item: &Item); }
```

## I — Interface Segregation

> No client should be forced to depend on methods it doesn't use.

- Small, focused interfaces > one large interface
- If a class implements an interface but leaves methods empty, the interface is too fat

```
# Bad — forces implementors to handle everything
trait DataStore {
    fn read();
    fn write();
    fn watch();      // Not all stores support watching
    fn replicate();  // Not all stores replicate
}

# Good — compose what you need
trait Reader { fn read(); }
trait Writer { fn write(); }
trait Watcher { fn watch(); }
// impl Reader + Writer for Postgres
// impl Reader + Writer + Watcher for Redis
```

## D — Dependency Inversion

> Depend on abstractions, not concretions.

- High-level modules should not depend on low-level modules
- Both should depend on abstractions (interfaces/traits)
- **This is the most impactful SOLID principle** — it enables testing, swapping implementations, and clean architecture

```
# Bad — business logic depends directly on infrastructure
class TradingEngine:
    def __init__(self):
        self.db = PostgresConnection("localhost:5432")

# Good — depends on abstraction, injected from outside
class TradingEngine:
    def __init__(self, repo: PositionRepository):
        self.repo = repo

# In production: TradingEngine(PostgresRepo(...))
# In tests:      TradingEngine(InMemoryRepo())
```

## When to Break SOLID

SOLID is guidance, not law. Break it when:
- **Prototyping**: get it working first, refactor later
- **Scripts**: a 50-line script doesn't need interfaces
- **Performance**: sometimes coupling is faster and the bottleneck matters
- **Simplicity**: if the abstraction is harder to understand than the concrete code, skip it

The goal is **maintainable code**, not **pattern-compliant code**.
