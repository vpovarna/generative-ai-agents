# Payload examples

Use these with the producer CLI.

**From file (bash):**
```bash
go run cmd/producer/main.go -d "$(cat resources/payloads/pass.json)"
```

| File | Description | Expected verdict |
|------|-------------|------------------|
| `pass.json` | Good answer, on-topic and grounded in context | `pass` |
| `fail-early.json` | Very short answer; triggers precheck early exit | `fail` (no LLM calls) |
| `review.json` | Adequate but incomplete answer | `review` or `pass` depending on judges |
