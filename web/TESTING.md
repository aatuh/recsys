# Testing guide

API Boilerplate web is wired for unit/component testing and prepared for e2e.

## Unit tests (Vitest)

- Run once: `pnpm test`
- Watch: `pnpm test:watch`
- Config: `vitest.config.ts` (jsdom env, V8 coverage, path alias `@ -> src`)
- Setup: `vitest.setup.ts` loads `@testing-library/jest-dom`
- Example: `tests/http/client.test.ts` shows how to inject a mock `fetch` and assert retries/timeouts.
- Mocking HTTP: use `createMockHttpClient` from `@api-boilerplate-core/http/testing` or pass `fetchImpl` to `createHttpClient`.

## Component testing

- Use `@testing-library/react` with Vitest (jsdom) to render components and simulate user actions via `@testing-library/user-event`.
- Import helpers from `vitest.setup.ts` (`jest-dom` matchers) for assertions.

## API mocking

- Prefer `createMockHttpClient` or `fetchImpl` override for units; for higher-level tests, wire MSW or a similar interceptor (not added by default).

## E2E readiness

- Add Playwright (`pnpm dlx playwright install`) and configure a `test:e2e` script when pages stabilize. The app runs with `pnpm dev`; point Playwright to `http://localhost:3000`.

## Notes

- Abort/timeout-safe requests live in `@api-boilerplate-core/http` (`src/client.ts`) to keep tests deterministic.
- Keep new tests colocated under `tests/` or alongside components (`*.test.tsx`).
