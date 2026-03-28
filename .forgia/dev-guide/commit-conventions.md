# Commit Conventions

## Format

```
type(scope): description

[optional body]

[optional footer]
```

## Types

| Type | When |
|------|------|
| `feat` | New feature or capability |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `test` | Adding or updating tests |
| `chore` | Build process, CI, tooling |

## Scope

- FD-linked: `feat(FD-001): add user authentication`
- SDD-linked: `feat(FD-001/SDD-002): implement login endpoint`
- General: `fix: resolve race condition in cache`

## Rules

1. Subject line under 72 characters
2. Use imperative mood: "add", not "added" or "adds"
3. No period at the end of the subject
4. Body explains WHY, not WHAT (the diff shows what)
5. AI-generated commits include footer:
   ```
   Co-Authored-By: Claude <noreply@anthropic.com>
   ```
6. Reference issues when applicable:
   ```
   Closes #123
   ```

## Examples

```
feat(FD-001): add JWT authentication middleware

Implements HMAC-based token validation for the gateway proxy.
Chose HMAC over RSA because this is a single-issuer system.

Co-Authored-By: Claude <noreply@anthropic.com>
```

```
fix(FD-002/SDD-001): handle timeout in venue bridge

The paper bridge was silently swallowing connection timeouts,
causing the agent loop to hang. Now raises BridgeTimeoutError.
```
