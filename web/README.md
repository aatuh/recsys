# API Boilerplate web (Next.js)

Next.js app router with Tailwind, TypeScript, and generated API client types.

## Quickstart

```bash
pnpm install
cp .env.example .env
pnpm dev
```

## Scripts

- `pnpm dev` - local dev server
- `pnpm lint` - lint workspace
- `pnpm typecheck` - TypeScript type check
- `pnpm test` - run unit tests
- `pnpm generate:api` - generate typed API models from `../api/swagger/swagger.json`

## Env

Set in `.env` (see `.env.example`):

- `NEXT_PUBLIC_API_BASE_URL` - API base, e.g. `http://localhost:8000/api/v1`
- `NEXT_PUBLIC_APP_URL` - public site URL
- `NEXT_PUBLIC_APP_ORG_ID` - demo org ID for the Foo list
- `NEXT_PUBLIC_APP_NAMESPACE` - demo namespace for the Foo list
- `NEXT_PUBLIC_DEFAULT_LOCALE` - default locale (e.g. `en`)
- `NEXT_PUBLIC_SUPPORTED_LOCALES` - comma-separated locales
- `NEXT_PUBLIC_LOCALE_AUTO_DETECT` - toggle accept-language detection

## Notes

- Replace the Foo domain packages under `packages/services/foo/` with your own domain.
- Run `pnpm generate:api` after API changes to refresh types.
