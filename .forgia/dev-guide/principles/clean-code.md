# Clean Code Principles

> Always loaded. Language-agnostic rules that apply to every codebase.

## Naming

- Names reveal intent — if you need a comment to explain a variable name, the name is wrong
- Avoid abbreviations unless universally understood (`ctx`, `err`, `db` are OK; `proc_mgr_svc` is not)
- Functions: verb + noun (`calculateTotal`, `parse_config`, `send_notification`)
- Booleans: `is_`, `has_`, `can_`, `should_` prefix (`is_active`, `has_permission`)
- Collections: plural (`users`, `pending_orders`), single item: singular (`user`, `order`)
- Constants: describe the meaning, not the value (`MAX_RETRY_ATTEMPTS = 3`, not `THREE = 3`)

## Functions

- Do ONE thing — if you can describe it with "and", split it
- Max 3 parameters (prefer structs/objects for more)
- No side effects hidden in the name — `getUser()` should not modify state
- Prefer pure functions where possible (same input → same output, no mutation)
- Early return > deeply nested if/else

```
# Bad
def process(data):
    if data:
        if data.is_valid():
            if data.has_permission():
                return do_work(data)
            else:
                raise PermissionError()
        else:
            raise ValidationError()
    else:
        raise ValueError()

# Good
def process(data):
    if not data:
        raise ValueError()
    if not data.is_valid():
        raise ValidationError()
    if not data.has_permission():
        raise PermissionError()
    return do_work(data)
```

## Comments

- Code tells HOW, comments tell WHY
- Don't comment WHAT the code does — make the code readable instead
- Acceptable comments: business logic reasoning, regulatory requirements, non-obvious performance decisions, workarounds with links to issues
- Never commit commented-out code — that's what git is for

## Error Handling

- Fail fast, fail loud — don't silently swallow errors
- Handle errors at the right level (not too deep, not too shallow)
- Error messages must include: what failed, why, and what the user/dev can do about it
- Never use exceptions for control flow

## DRY vs WET

- Rule of Three: duplicate once is fine, duplicate twice means extract
- Don't DRY prematurely — two similar things might diverge later
- Prefer clear duplication over wrong abstraction
- When you DO abstract, the abstraction must be easier to understand than the duplication

## Boy Scout Rule

- Leave code better than you found it — but only if you're already touching it
- Don't refactor code you're not working on (scope creep)
- Small improvements compound: better names, simplified conditions, removed dead code

## Testing (Universal)

- Test behavior, not implementation
- One assert per test (logical assert, not physical)
- Test names describe the scenario: `test_<what>_<when>_<then>`
- Tests are documentation — reading them should explain what the code does
- Don't test the framework, test YOUR code
