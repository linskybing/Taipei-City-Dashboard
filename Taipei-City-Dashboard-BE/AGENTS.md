## Backend Agent Rules

These rules inherit the root hackathon redlines. If this file conflicts with the root `AGENTS.md`, the root hackathon rules win.

## Scope

- Work in Go with Gin, GORM, Redis, Cobra, PostgreSQL/PostGIS, Qdrant, and the existing Gateway/ToolRouter architecture.
- Read `taipei-city-dashboard-back-end` before editing controllers, routes, services, models, AI gateway, cache, or CLI code.

## Compliance Boundaries

- AI must stay behind the backend gateway. The official chat endpoint is `POST /api/v1/ai/chat/twai`.
- TWCC config must come from ENV-backed config such as `TWCC_API_URL`, `TWCC_API_KEY`, and `TWCC_MODEL`.
- The only authorized hackathon model is `llama3.3-ffm-70b-16k-chat`. Do not introduce or preserve `llama3.3-ffm-70b-32k-chat`.
- Tool Calling must execute approved server-side tools only, enforce bounded loops/timeouts/argument sizes, and persist request/action/token usage to `ai_chatlog` or the official audit table.
- Component APIs must keep existing `query_type` parsing and SQL aliases compatible with frontend formats.

## Implementation Standards

- Keep handlers thin: controllers validate request/response, services own business logic, models own persistence.
- Use `DBDashboard` for analytical data and `DBManager` for management/configuration data.
- Keep new Go files under 200 lines; split providers, request types, tool logic, and tests by responsibility.
- Use idiomatic Go names, `gofmt`, structured errors, request context propagation, and no hardcoded secrets.
- Run targeted `go test` for changed packages and `go test ./test/aiassistant` for AI compliance-sensitive changes.
