# RecSys Demo UI

A lightweight React + Vite UI for exploring the RecSys API. Provides seeding,
recommendations, bandit playground, rules, segments, and data management.

## Quick start

```bash
pnpm install
pnpm dev
```

Open http://localhost:3000.

## Configuration

Config is centralized in `src/config.ts`. It reads environment variables from
Vite at build/runtime:

- `VITE_API_BASE_URL` (default: `/api`)
- `VITE_SWAGGER_UI_URL` (default: `http://localhost:8081`)
- `VITE_CUSTOM_CHATGPT_URL` (optional)

Do not read envs directly in feature code; inject via `config`.

## Project layout (UI)

- `src/ui/` shared UI (design tokens, globals, AppShell, Toast)
- `src/components/` views and sections
  - `rules/` split modal components for rules
- `src/hooks/` reusable hooks (`useQuerySync`, `useAsync`)
- `src/services/` API adapters (wrap generated client)
- `src/lib/api-client/` generated OpenAPI client (do not edit)

## Conventions

- Use design tokens from `src/ui/tokens.ts` for spacing, colors, and typography
- Keep long-running or external calls in `services/`; views should stay thin
- Use `useAsync` for async flows and surface toasts via `useToast`
- Sync shareable state with `useQuerySync`; persist with `storageKey`

## Scripts

```bash
pnpm lint           # eslint
pnpm lint:fix       # eslint --fix
pnpm typecheck      # tsc --noEmit
pnpm build          # vite build
pnpm preview        # vite preview
pnpm codegen:api    # regenerate OpenAPI client from swagger.json
```

## Notes

- This UI strives for consistency and clarity; avoid ad-hoc inline styles.
- Prefer generated client services over manual fetch.

# RecSys Demo UI

The demo UI includes several powerful features for testing and exploring the recommendation system.

## User Traits Editor

The demo now includes a comprehensive **User Traits Editor** integrated as an accordion within the "Seed Data" section that allows you to:

1. **Quick Preview**: See a summary of configured traits and their probabilities without opening the editor
2. **Configure Dynamic Traits**: Define custom trait keys (e.g., `plan`, `age_group`, `interests`) with:
   - **Include Probability**: Control how often each trait appears in generated users (0-1)
   - **Value Options**: Define multiple possible values for each trait
   - **Value Probabilities**: Set weighted probabilities for each value (e.g., 60% "free", 30% "plus", 10% "pro")

3. **Edit User Traits in Browser**: 
   - Select any generated user from a dropdown
   - View and edit their current traits
   - Add new traits or modify existing ones
   - Update user traits directly in the browser

4. **Accordion Interface**: 
   - Collapsible section to save space
   - Always-visible preview of current configuration
   - Easy toggle between collapsed and expanded states

5. **Example Trait Configurations**:

```json
{
    "plan": {
    "probability": 1.0,
    "values": [
        {"value": "free", "probability": 0.6},
        {"value": "plus", "probability": 0.3},
        {"value": "pro", "probability": 0.1}
    ]
    },
    "age_group": {
    "probability": 0.8,
    "values": [
        {"value": "18-24", "probability": 0.2},
        {"value": "25-34", "probability": 0.3},
        {"value": "35-44", "probability": 0.25},
        {"value": "45-54", "probability": 0.15},
        {"value": "55+", "probability": 0.1}
    ]
    }
}
```

### Dynamic User Generation

When seeding data, users are now generated with traits based on your configuration:

- Each trait has a chance to be included based on its probability
- Values are selected using weighted random selection
- Fallback to default `plan` trait if no configurations are provided

### Events per User Configuration

The seeding system now supports realistic event generation:

- **Min/Max Events per User**: Set a range of events each user will generate
- **Randomized Distribution**: Each user gets a random number of events between min and max
- **Consistent Events**: Set min=max for the same number of events per user
- **Realistic Patterns**: More closely mimics real user behavior with varied activity levels

### Usage

1. **Configure Data Counts**: Set the number of users, items, and events per user (min/max range)
2. **Configure Traits**: Use the "User Traits Configuration" accordion in the "Seed Data" section to set up your desired trait configurations
3. **Preview Configuration**: See a quick summary of your trait setup without opening the accordion
4. **Seed Data**: Click "Seed Data" to generate users with your configured traits and event patterns
5. **Edit Users**: Select generated users from the dropdown to view and edit their traits
6. **Test Recommendations**: Use the updated user data to test how traits and event patterns affect recommendations

## Deploying to Vercel

- Create new project.
- Root directory `/demo-ui/`.
- Paste env variables.
- Set branch tracking as `production`.
- Generate custom domain.
