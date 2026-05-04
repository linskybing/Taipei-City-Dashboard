## Frontend Agent Rules

These rules inherit the root hackathon redlines. If this file conflicts with the root `AGENTS.md`, the root hackathon rules win.

## Scope

- Work in Vue 3, Vite, Pinia, Vue Router, Axios, Mapbox, deck.gl, Three.js, Sass, Apexcharts, and the existing dashboard component system.
- Read `taipei-city-dashboard-fe-core` first for structure and style. Add `fe-components`, `fe-maps`, `fe-dashboards`, or `fe-ui-design` when the task touches those areas.

## Compliance Boundaries

- Do not add chart libraries. Use only `apexcharts`/`vue3-apexcharts` and existing chart component patterns.
- Do not call TWCC, OpenAI, Anthropic, browser AI, or any model provider directly from frontend code. AI UI must call backend APIs, especially `POST /api/v1/ai/chat/twai`.
- Do not expose `TWCC_API_KEY`, `TWCC_MODEL`, provider URLs, or model names in frontend source, bundles, static files, or env examples.
- Render only supported dashboard data formats: `two_d`, `percent`, `three_d`, `map_legend`, and `time`.
- New competition components must include a Taipei/New Taipei selector or equivalent region switch when applicable.

## Implementation Standards

- Follow local Vue structure: imports, constants, props, local state, computed, methods, lifecycle, template, scoped SCSS.
- Keep new component files under 200 lines. Split helpers, options, or config builders when a component grows.
- Reuse existing stores, router patterns, utility functions, CSS variables, and `src/assets/configs`.
- Do not hardcode data payloads for competition dashboards. Use backend/database-driven component data.
- Prefer read-only validation before formatter commands because `npm run build` and `npm run lint` run `eslint --fix`.
