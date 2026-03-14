# TypeScript Conventions

> Auto-loaded when `tsconfig.json` or `package.json` with TypeScript dependency is detected.

## Project Structure

- Package manager: `pnpm` (preferred) or `npm`
- Config: `tsconfig.json` with `strict: true`
- Source: `src/` directory
- Tests: `tests/` or `__tests__/` or colocated `*.test.ts`

## Strict Mode

- Always enable `strict: true` in tsconfig
- Never use `any` — use `unknown` and narrow with type guards
- Prefer `const` over `let`, never use `var`
- Use explicit return types on exported functions

```typescript
// Good
export function calculateTotal(items: CartItem[]): number {
  return items.reduce((sum, item) => sum + item.price * item.qty, 0);
}

// Bad — implicit return type, uses any
export function calculateTotal(items: any[]) {
  return items.reduce((sum: any, item: any) => sum + item.price * item.qty, 0);
}
```

## React (if applicable)

- Functional components only (no class components)
- Props: define with `interface`, not `type` (for better error messages and extensibility)
- State: `useState` for local, `useReducer` for complex, context for shared
- Effects: minimize `useEffect` — prefer derived state and event handlers
- Styling: Tailwind CSS (utility-first) or CSS Modules

```typescript
interface DashboardProps {
  positions: Position[];
  onRefresh: () => void;
}

export function Dashboard({ positions, onRefresh }: DashboardProps) {
  const totalPnL = positions.reduce((sum, p) => sum + p.pnl, 0);
  // ...
}
```

## Tauri (if applicable)

- IPC: use `@tauri-apps/api/core` for `invoke()`
- Commands: typed with shared types between Rust and TypeScript
- State: Tauri state management for cross-process data
- Events: `listen()` / `emit()` for Rust ↔ frontend communication

## Error Handling

- Use discriminated unions for error types (not exceptions for control flow)
- Wrap external API calls in try/catch with specific error handling
- Use `Result<T, E>` pattern for functions that can fail predictably

```typescript
type Result<T, E = Error> =
  | { ok: true; value: T }
  | { ok: false; error: E };
```

## Dependencies

- UI: `React` + `Vite`
- Components: `shadcn/ui` (copy-paste, not dependency)
- Styling: `Tailwind CSS` v4
- State: `zustand` or `jotai` (not Redux unless needed)
- Forms: `react-hook-form` + `zod` validation
- HTTP: `fetch` API (native) or `ky`
- Testing: `vitest` + `@testing-library/react`

## Style

- Formatter: `prettier` or `biome`
- Linter: `eslint` with strict TypeScript rules, or `biome`
- Imports: absolute paths with `@/` alias, sorted automatically
- Naming: `camelCase` functions/variables, `PascalCase` components/types, `UPPER_CASE` constants
- Files: `kebab-case.ts` for utilities, `PascalCase.tsx` for components

## Testing

- Framework: `vitest`
- Component tests: `@testing-library/react` (test behavior, not implementation)
- E2E: `playwright` (if needed)
- Test names: `it('should <behavior> when <scenario>')`

```typescript
describe('Dashboard', () => {
  it('should display total PnL from all positions', () => {
    render(<Dashboard positions={mockPositions} onRefresh={vi.fn()} />);
    expect(screen.getByText('$1,234.56')).toBeInTheDocument();
  });
});
```

## What NOT to Do

- Don't use `any` — ever
- Don't use `enum` — use `as const` objects or union types
- Don't use `namespace` — use ES modules
- Don't mutate props or state directly
- Don't use `index` as React key (unless list is static and never reordered)
