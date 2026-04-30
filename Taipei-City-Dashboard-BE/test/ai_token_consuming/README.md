# AI Token Consuming Tests

Tests in this folder call the live TWCC AFS API and consume team AI token.

They are skipped by default. Run them only when you intentionally want a live
AI chat integration check:

```sh
RUN_AI_TOKEN_TESTS=1 go test ./test/ai_token_consuming -run TestLiveAIChatEndpointWritesAuditLog -count=1 -v
```

Required runtime config:

- `TWCC_API_URL=https://api-ams.twcc.ai/api`
- `TWCC_API_KEY` set to an authorized TWCC key
- `TWCC_MODEL=llama3.3-ffm-70b-16k-chat`
- manager and dashboard PostgreSQL env vars pointing to an existing test-safe
  database

The test sends one non-streaming chat request. If the model decides to use
tools, the service may make additional TWCC requests inside the bounded tool
loop.

Database isolation:

- The test creates unique temporary PostgreSQL databases for manager and
  dashboard data, named with the `tcd_ai_*_test_*` prefix.
- The backend is pointed at those temporary databases for the test run.
- The `ai_chatlog` row is written only to the temporary manager database.
- Cleanup closes test connections, terminates sessions on the temporary
  databases, and drops the temporary databases.
- The test refuses to drop any database whose name does not use the temporary
  test prefix.
