# AI Chat Integration Tests

## Objective

Validate and commit the backend AI chat changes, including live `POST /api/v1/ai/chat/twai` integration coverage, response body shape checks, audit-log verification, and isolated token-consuming test execution.

## Affected Files And Modules

- `Taipei-City-Dashboard-BE/app/controllers/ai.go`
- `Taipei-City-Dashboard-BE/app/models/ai.go`
- `Taipei-City-Dashboard-BE/app/services/ai/*`
- `Taipei-City-Dashboard-BE/app/services/ai/assistant/*`
- `Taipei-City-Dashboard-BE/test/aiassistant/*`
- `Taipei-City-Dashboard-BE/test/ai_token_consuming/*`

## Implementation Notes

- Added live AI chat integration tests under `test/ai_token_consuming/` so tests that consume TWCC token are clearly separated from normal CI-style tests.
- Live tests are gated by `RUN_AI_TOKEN_TESTS=1`.
- Live tests validate the expected non-stream JSON response body fields, positive token usage, `provider=twcc`, the authorized model, and the `audit_ref` to `ai_chatlog` mapping.
- Live tests create temporary PostgreSQL databases with the `tcd_ai_*_test_*` prefix for manager and dashboard connections, run only test schema setup against those databases, then close connections and drop the temporary databases.
- No package manifest changes were made.

## Hackathon Compliance Checks

- AI provider remains behind the backend gateway endpoint `POST /api/v1/ai/chat/twai`.
- TWCC config is read from environment variables; no secrets or API keys were hardcoded.
- Authorized model check enforces `llama3.3-ffm-70b-16k-chat`.
- The non-compliant 32k TWCC model identifier was not introduced in runtime code.
- No OpenAI, Anthropic, local model, browser AI, or alternate model provider usage was introduced.
- No new dependencies were added to `go.mod`, `go.sum`, `package.json`, or lockfiles.
- No database initialization or destructive operation was run against existing application databases; destructive cleanup is restricted to temporary test DB names with the `tcd_ai_` prefix.

## Validation

- `RUN_AI_TOKEN_TESTS=0 go test ./app/controllers ./app/services/ai/... ./test/aiassistant ./test/ai_token_consuming -count=1` passed.
- `go test ./...` from `Taipei-City-Dashboard-BE` passed.
- `.agents/scripts/agent_quality_gate.py` passed before and after the live test.
- `git diff --check` passed.
- Compliance scan for non-TWCC AI provider strings and the non-compliant 32k model found no runtime violations.
- Live token-consuming test passed in a temporary Go container on the `br_dashboard` network:
  - `docker run --rm --network br_dashboard --env-file docker/.env -e GIN_MODE=test -e RUN_AI_TOKEN_TESTS=1 ... go test ./test/ai_token_consuming -run TestLiveAIChatEndpointWritesAuditLog -count=1 -v`
  - Latest live run used the authorized `llama3.3-ffm-70b-16k-chat` model and returned positive token usage.
  - Temporary DBs `tcd_ai_manager_test_*` and `tcd_ai_dashboard_test_*` were created and dropped during test cleanup.
- Verified no `tcd_ai_%` temporary databases remained in `postgres-manager` or `postgres-data` after the live test.

## Skipped Validation

- Frontend build and lint were not run because this change did not modify frontend source.
- Data-End tests were not run because this change did not modify ETL/DAG code.
- Streaming endpoint body shape was not validated here because the frontend currently calls AI chat with `stream: false`; existing provider tests cover streaming tool-markup suppression.

## Remaining Risks

- Live AI tests consume TWCC team token and should remain opt-in.
- The live test validates a concise non-streaming assistant response and audit write path; broader prompt behavior still depends on the model response and should be covered by focused scenario tests as features are added.
