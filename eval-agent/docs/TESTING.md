# Testing Guide

Comprehensive test cases for eval-agent API endpoints.

## Prerequisites

Ensure the API server is running:

```bash
cd eval-agent
go run cmd/api/main.go
```

Server should be listening on `http://localhost:18082`.

---

## Health Check

Verify the service is running:

```bash
curl http://localhost:18082/api/v1/health
```

**Expected response:**
```json
{
  "status": "ok",
  "version": "1.0.0"
}
```

---

## Full Pipeline Evaluation

### Test 1: Happy Path (High Quality Answer)

```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-001",
    "event_type": "agent_response",
    "agent": {
      "name": "my-agent",
      "type": "rag",
      "version": "1.0.0"
    },
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Its capital city is Paris, which is also the largest city in the country.",
      "answer": "The capital of France is Paris."
    }
  }'
```

**Expected:**
- All prechecks pass (scores close to 1.0)
- All LLM judges pass (scores > 0.9)
- `"verdict": "pass"`
- `"confidence": > 0.9`

---

### Test 2: Early Exit (Very Poor Answer)

```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-002",
    "event_type": "agent_response",
    "agent": {"name": "my-agent", "type": "rag", "version": "1.0.0"},
    "interaction": {
      "user_query": "Explain the theory of relativity in detail",
      "context": "Einstein developed the theory of relativity.",
      "answer": "ok"
    }
  }'
```

**Expected:**
- Precheck scores very low (< 0.2 average)
- Early exit triggered
- `"verdict": "fail"`
- `stages` array contains only prechecks (no LLM judges)

---

### Test 3: Review Verdict (Medium Quality)

```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "evt-003",
    "event_type": "agent_response",
    "agent": {"name": "my-agent", "type": "rag", "version": "1.0.0"},
    "interaction": {
      "user_query": "What is encryption?",
      "context": "Encryption is a method to secure data by converting it into unreadable format.",
      "answer": "Encryption protects data."
    }
  }'
```

**Expected:**
- `"verdict": "pass"` (confidence higher then 0.8)
- Answer is relevant but incomplete. Score 0.6
- Completeness judge should score lower. Score 0.5

---

## Single Judge Tests

### Relevance Judge

#### Test 1: Highly Relevant Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/relevance \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-relevance-pass",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe. Paris is its capital.",
      "answer": "The capital of France is Paris."
    }
  }'
```

**Expected:**
- `"score": > 0.9`
- `"verdict": "pass"`
- Reason mentions direct relevance to query

#### Test 2: Irrelevant Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/relevance \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-relevance-fail",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is the capital of France?",
      "context": "France is a country in Western Europe.",
      "answer": "Pizza is a popular Italian food."
    }
  }'
```

**Expected:**
- `"score": < 0.3`
- `"verdict": "fail"`
- Reason mentions answer does not address query

---

### Faithfulness Judge

#### Test 1: Faithful Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/faithfulness \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-faithfulness-pass",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What encryption does the product support?",
      "context": "Our product supports AES-256 encryption for data at rest and TLS 1.3 for data in transit.",
      "answer": "The product supports AES-256 encryption for data at rest."
    }
  }'
```

**Expected:**
- `"score": > 0.9`
- `"verdict": "pass"`
- Reason confirms answer is grounded in context

#### Test 2: Hallucination Detected

```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/faithfulness?threshold=0.9" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-faithfulness-fail",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What encryption does the product support?",
      "context": "Our product supports AES-256 encryption for data at rest.",
      "answer": "The product supports AES-256 encryption and quantum-resistant algorithms."
    }
  }'
```

**Expected:**
- `"score": < 0.5` (hallucination: quantum-resistant not in context)
- `"verdict": "fail"` (with threshold 0.9)
- Reason mentions unsupported claim

---

### Coherence Judge

#### Test 1: Coherent Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/coherence \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-coherence-pass",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "How does encryption work?",
      "context": "Encryption uses algorithms to scramble data.",
      "answer": "Encryption works by using mathematical algorithms to transform readable data into scrambled format. This ensures security by making data unreadable without the key."
    }
  }'
```

**Expected:**
- `"score": > 0.9`
- `"verdict": "pass"`
- Reason mentions logical flow and consistency

#### Test 2: Incoherent Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/coherence \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-coherence-fail",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "How does encryption work?",
      "context": "Encryption uses mathematical algorithms to scramble data.",
      "answer": "Encryption scrambles data. But also, pizza is delicious. The algorithm ensures security. Cats are better than dogs."
    }
  }'
```

**Expected:**
- `"score": < 0.4`
- `"verdict": "fail"`
- Reason mentions incoherence and unrelated statements

---

### Completeness Judge

#### Test 1: Complete Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/completeness \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-completeness-pass",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Explain both encryption and decryption",
      "context": "Encryption converts plaintext to ciphertext. Decryption reverses this process.",
      "answer": "Encryption converts readable plaintext into unreadable ciphertext using a key. Decryption is the reverse process that converts ciphertext back to plaintext using the decryption key."
    }
  }'
```

**Expected:**
- `"score": > 0.9`
- `"verdict": "pass"`
- Reason confirms both parts addressed

#### Test 2: Incomplete Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/completeness \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-completeness-fail",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Explain both encryption and decryption, and provide examples of each",
      "context": "Encryption converts plaintext to ciphertext. Decryption reverses this. Example: AES-256.",
      "answer": "Encryption converts plaintext to ciphertext using algorithms like AES-256."
    }
  }'
```

**Expected:**
- `"score": < 0.5` (missed decryption and examples)
- `"verdict": "fail"`
- Reason mentions incomplete coverage of query parts

---

### Instruction Judge

#### Test 1: Instructions Followed

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/instruction \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-instruction-pass",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "List exactly 3 encryption algorithms in bullet points",
      "context": "Common algorithms: AES, RSA, ChaCha20, Blowfish, Twofish",
      "answer": "• AES\n• RSA\n• ChaCha20"
    }
  }'
```

**Expected:**
- `"score": > 0.9`
- `"verdict": "pass"`
- Reason confirms count and format match instructions

#### Test 2: Count Violation

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/instruction \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-instruction-fail",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "List exactly 3 encryption algorithms",
      "context": "Common algorithms: AES, RSA, ChaCha20, Blowfish, Twofish",
      "answer": "1. AES\n2. RSA"
    }
  }'
```

**Expected:**
- `"score": < 0.6` (only provided 2 instead of 3)
- `"verdict": "fail"`
- Reason mentions instruction violation

#### Test 3: Minor Overshoot (Still Passes)

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/instruction \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-instruction-overshoot",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "Give me 3 examples of encryption algorithms",
      "context": "Common algorithms: AES, RSA, ChaCha20, Blowfish",
      "answer": "Examples of encryption algorithms are AES, RSA, ChaCha20, and Blowfish."
    }
  }'
```

**Expected:**
- `"score": ~ 0.8` (asked for 3, gave 4 - minor overshoot)
- `"verdict": "pass"`
- Reason mentions slight excess but acceptable

---

## Threshold Testing

### Test with Lenient Threshold (0.5)

```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/relevance?threshold=0.5" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-lenient",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is AI?",
      "context": "AI is artificial intelligence.",
      "answer": "AI is related to computers and technology."
    }
  }'
```

**Expected:**
- Score might be ~ 0.6-0.7 (somewhat relevant but vague)
- `"verdict": "pass"` (exceeds 0.5 threshold)

---

### Test with Strict Threshold (0.95)

```bash
curl -X POST "http://localhost:18082/api/v1/evaluate/judge/relevance?threshold=0.95" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-strict",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is AI?",
      "context": "AI stands for Artificial Intelligence.",
      "answer": "AI refers to artificial intelligence, which is technology that mimics human cognition."
    }
  }'
```

**Expected:**
- Score ~ 0.9-0.95
- `"verdict"` depends on whether score exceeds 0.95
- May be "fail" if score is exactly 0.9

---

## Edge Cases

### Empty Answer

```bash
curl -X POST http://localhost:18082/api/v1/evaluate \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-empty",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is encryption?",
      "context": "Encryption secures data.",
      "answer": ""
    }
  }'
```

**Expected:**
- Early exit triggered (format checker fails)
- `"verdict": "fail"`

---

### No Context (Faithfulness Judge)

```bash
curl -X POST http://localhost:18082/api/v1/evaluate/judge/faithfulness \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "test-no-context",
    "event_type": "agent_response",
    "agent": {"name": "test", "type": "rag", "version": "1.0"},
    "interaction": {
      "user_query": "What is AI?",
      "context": "",
      "answer": "AI is artificial intelligence."
    }
  }'
```

**Expected:**
- Faithfulness judge should handle gracefully
- Score may default to high (no context to contradict)

---

## Using jq for Filtering

Extract specific judge scores:

```bash
curl -X POST http://localhost:18082/api/v1/evaluate ... | jq '.stages[] | select(.name=="relevance-judge")'
```

Show only verdict and confidence:

```bash
curl -X POST http://localhost:18082/api/v1/evaluate ... | jq '{verdict, confidence}'
```
