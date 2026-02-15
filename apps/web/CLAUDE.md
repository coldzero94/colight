# Web

Next.js 15 frontend. See root CLAUDE.md for project rules.

## Structure

```
src/
  app/(auth)/        # Login, signup, callback
  app/(main)/        # Dashboard, experiences, analysis, coaching
  app/(admin)/       # Admin pages
  api/generated/     # @hey-api/openapi-ts (do not edit)
  components/        # auth/, layout/, common/, admin/
  hooks/             # React Query hooks
  stores/            # Zustand (auth-store)
  lib/               # api-client, schemas, utilities
```

## Rules

- Tailwind only, no CSS Modules
- Server Components for data, Client Components (`'use client'`) for interactivity
- Auth: JWT in Zustand (localStorage) → Axios interceptor

## Commands

```bash
moon run web:dev              # Dev (port 4000)
moon run web:build            # Build
moon run web:lint             # ESLint
moon run web:typecheck        # tsc --noEmit
moon run web:test             # vitest
moon run web:generate-client  # OpenAPI → TS client
```

## TDD Rules

### Workflow

1. **Red**: Write `__tests__/*.test.ts(x)` ONLY → `moon run web:test` → confirm FAIL
2. **Green**: Write minimal component/hook → `moon run web:test` → confirm PASS
3. Never write test + implementation together

### Test Structure

```
components/foo/           → components/foo/__tests__/foo.test.tsx
hooks/use-foo.ts          → hooks/__tests__/use-foo.test.ts
lib/utils.ts              → lib/__tests__/utils.test.ts
lib/validations/foo.ts    → lib/validations/__tests__/foo.test.ts
```

### Patterns

**Component test** (render + assert + interact):
```tsx
it('renders title and calls handler', () => {
  const onSubmit = vi.fn();
  render(<FooCard title="테스트" onSubmit={onSubmit} />);
  expect(screen.getByText('테스트')).toBeInTheDocument();
  fireEvent.click(screen.getByRole('button', { name: '제출' }));
  expect(onSubmit).toHaveBeenCalledOnce();
});
```

**Hook test** (renderHook + act):
```tsx
it('fetches data on mount', async () => {
  vi.mocked(apiFn).mockResolvedValue({ data: mockData });
  const { result } = renderHook(() => useFoo());
  await waitFor(() => expect(result.current.data).toEqual(mockData));
});
```

**Validation test** (schema parse):
```ts
it('rejects empty title', () => {
  const result = fooSchema.safeParse({ title: '' });
  expect(result.success).toBe(false);
});
```

### Conventions

- API mocking: `vi.mock('@/lib/api/foo')` or MSW for integration
- Use `screen.getByRole` > `getByText` > `getByTestId` (accessibility order)
- `userEvent` for realistic interactions, `fireEvent` for simple clicks
- Don't test: `api/generated/`, layouts, config files, types-only files
