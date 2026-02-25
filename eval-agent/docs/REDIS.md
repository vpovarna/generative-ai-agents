# Redis Stream Consumer Mode

Advanced usage guide for running eval-agent as a Redis Stream consumer for asynchronous evaluation.

## Overview

In Redis Stream consumer mode, eval-agent:
- Connects to a Redis instance
- Joins a consumer group (`eval-group`)
- Consumes messages from the `eval-events` stream
- Processes evaluation requests asynchronously
- Supports graceful shutdown and acknowledgment

## Use Cases

- **High-throughput evaluation**: Process evaluation requests asynchronously
- **Decoupled architecture**: Separate evaluation from agent response generation
- **Multiple consumers**: Scale horizontally with multiple eval-agent instances
- **Fault tolerance**: Redis Streams provide message persistence and redelivery

---

## Configuration

Add Redis configuration to your `.env`:

```env
# Redis Stream Configuration
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_STREAM_NAME=eval-events
REDIS_CONSUMER_GROUP=eval-group
REDIS_CONSUMER_NAME=eval-consumer-1
```

---

## Running the Consumer

Start the Redis Stream consumer:

```bash
cd eval-agent
go run cmd/main.go
```

**Expected output:**
```
{"level":"info","message":"Starting eval-agent Redis Stream consumer"}
{"level":"info","stream":"eval-events","group":"eval-group","message":"Consumer started"}
```

The consumer will block and wait for messages on the stream.

---

## Sending Evaluation Requests

### Option 1: CLI Producer (Recommended)

Use the built-in CLI producer:

```bash
go run cmd/producer/main.go -d '{
  "event_id": "evt-001",
  "event_type": "agent_response",
  "agent": {
    "name": "my-agent",
    "type": "rag",
    "version": "1.0.0"
  },
  "interaction": {
    "user_query": "What is the capital of France?",
    "context": "France is a country in Western Europe. Its capital city is Paris.",
    "answer": "The capital of France is Paris."
  }
}'
```

**Flags:**
- `-d <json>`: Inline JSON payload
- `-f <file>`: Read payload from file
- `--redis-addr <addr>`: Redis address (default: localhost:6379)
- `--stream <name>`: Stream name (default: eval-events)

**Example with file:**
```bash
echo '{"event_id":"evt-002",...}' > payload.json
go run cmd/producer/main.go -f payload.json
```

### Option 2: redis-cli

Send messages directly via redis-cli:

```bash
redis-cli XADD eval-events '*' payload '{
  "event_id": "evt-001",
  "event_type": "agent_response",
  "agent": {"name": "my-agent", "type": "rag", "version": "1.0.0"},
  "interaction": {
    "user_query": "What is the capital of France?",
    "context": "France is a country in Western Europe. Its capital city is Paris.",
    "answer": "The capital of France is Paris."
  }
}'
```

### Option 3: Go Redis Client

```go
import "github.com/redis/go-redis/v9"

rdb := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

payload := map[string]interface{}{
    "event_id": "evt-001",
    "event_type": "agent_response",
    "agent": map[string]string{
        "name": "my-agent",
        "type": "rag",
        "version": "1.0.0",
    },
    "interaction": map[string]string{
        "user_query": "What is AI?",
        "context": "AI is artificial intelligence.",
        "answer": "AI is technology that mimics human intelligence.",
    },
}

payloadJSON, _ := json.Marshal(payload)
rdb.XAdd(ctx, &redis.XAddArgs{
    Stream: "eval-events",
    Values: map[string]interface{}{
        "payload": string(payloadJSON),
    },
})
```

---

## Consumer Behavior

### Message Processing

1. **Receive message** from stream
2. **Parse payload** from `payload` field
3. **Execute evaluation** (full pipeline)
4. **Log result** (confidence, verdict, stage scores)
5. **Acknowledge (ACK)** message to Redis
6. **Repeat**

### Consumer Groups

Redis Streams support consumer groups, allowing multiple eval-agent instances to process messages in parallel:

```bash
# Terminal 1
REDIS_CONSUMER_NAME=eval-consumer-1 go run cmd/main.go

# Terminal 2
REDIS_CONSUMER_NAME=eval-consumer-2 go run cmd/main.go
```

Each consumer in the group receives different messages, enabling horizontal scaling.

### Graceful Shutdown

Press `Ctrl+C` to trigger graceful shutdown:

```
^C
{"level":"info","message":"Received interrupt signal, shutting down..."}
{"level":"info","message":"Consumer shutdown complete"}
```

---

## Monitoring

### View Stream Contents

```bash
# List all messages in stream
redis-cli XRANGE eval-events - +

# Get stream length
redis-cli XLEN eval-events

# Get latest 10 messages
redis-cli XREVRANGE eval-events + - COUNT 10
```

### View Consumer Group Status

```bash
# Get consumer group info
redis-cli XINFO GROUPS eval-events

# Get consumer status
redis-cli XINFO CONSUMERS eval-events eval-group

# Get pending messages
redis-cli XPENDING eval-events eval-group
```

### Monitor in Real-Time

```bash
# Watch stream for new messages
redis-cli --scan --pattern eval-events

# Monitor all Redis commands
redis-cli MONITOR
```

---

## Message Schema

The `payload` field must contain a JSON object matching the HTTP API schema:

```json
{
  "event_id": "evt-001",
  "event_type": "agent_response",
  "agent": {
    "name": "my-agent",
    "type": "rag",
    "version": "1.0.0"
  },
  "interaction": {
    "user_query": "What is the capital of France?",
    "context": "France is a country in Western Europe. Its capital city is Paris.",
    "answer": "The capital of France is Paris."
  }
}
```

**Required fields:**
- `event_id`: Unique identifier
- `event_type`: Typically `"agent_response"`
- `agent.name`: Agent identifier
- `interaction.user_query`: User's query
- `interaction.answer`: Agent's response

**Optional fields:**
- `interaction.context`: Retrieved documents or context
- `agent.type`, `agent.version`: Metadata

---

## Error Handling

### Invalid Payload

If payload is invalid JSON or missing required fields:
- **Logged as error**
- **Message is ACKed** (removed from stream to prevent reprocessing)

Example log:
```json
{"level":"error","error":"missing required field: user_query","message":"Invalid payload"}
```

### Processing Errors

If evaluation fails (e.g., AWS Bedrock error):
- **Logged as error**
- **Message is ACKed** (not retried by default)
- Consider implementing dead-letter queue for production

---

## Production Considerations

### Redis Configuration

For production deployments:

```env
# Use Redis Cluster or Sentinel for HA
REDIS_ADDR=redis-cluster:6379
REDIS_PASSWORD=<strong-password>

# Enable TLS if needed
REDIS_TLS_ENABLED=true

# Connection pool settings
REDIS_MAX_RETRIES=3
REDIS_POOL_SIZE=10
```

### Consumer Scaling

Run multiple consumers with unique names:

```bash
# Deploy as systemd services or Kubernetes pods
eval-consumer-1:
  REDIS_CONSUMER_NAME=eval-consumer-1 go run cmd/main.go

eval-consumer-2:
  REDIS_CONSUMER_NAME=eval-consumer-2 go run cmd/main.go
```

### Stream Trimming

Prevent unbounded stream growth:

```bash
# Trim to max 10,000 messages (cron job or Redis config)
redis-cli XTRIM eval-events MAXLEN ~ 10000
```

Or configure automatic trimming in producer:

```bash
redis-cli XADD eval-events MAXLEN ~ 10000 * payload '...'
```

### Monitoring & Alerting

Monitor key metrics:
- **Stream length**: `XLEN eval-events`
- **Consumer lag**: `XPENDING eval-events eval-group`
- **Processing rate**: Messages ACKed per second
- **Error rate**: Failed evaluations per minute

Set up alerts for:
- Stream length > 10,000 (backlog)
- Consumer lag > 100 messages
- Error rate > 5%

---

## Troubleshooting

### Consumer Not Receiving Messages

1. **Check consumer is running:**
   ```bash
   ps aux | grep "cmd/main.go"
   ```

2. **Verify consumer group exists:**
   ```bash
   redis-cli XINFO GROUPS eval-events
   ```

3. **Check for pending messages:**
   ```bash
   redis-cli XPENDING eval-events eval-group
   ```

4. **Claim pending messages** (if consumer crashed):
   ```bash
   redis-cli XCLAIM eval-events eval-group eval-consumer-1 3600000 <message-id>
   ```

### Messages Not Being ACKed

Check consumer logs for errors. If messages remain pending:

```bash
# View pending messages
redis-cli XPENDING eval-events eval-group - + 10

# Manually ACK a message (caution!)
redis-cli XACK eval-events eval-group <message-id>
```

### Redis Connection Errors

```bash
# Test Redis connectivity
redis-cli -h localhost -p 6379 ping

# Check Redis logs
redis-cli INFO server
```

---

## Comparison: HTTP API vs Redis Stream

| Feature | HTTP API | Redis Stream |
|---------|----------|--------------|
| **Latency** | Low (sync) | Higher (async) |
| **Throughput** | Medium | High |
| **Scalability** | Vertical | Horizontal |
| **Fault tolerance** | None | Message persistence |
| **Use case** | Real-time UI | Batch processing |

**Recommendation:**
- Use **HTTP API** for real-time evaluation in user-facing applications
- Use **Redis Stream** for high-volume batch evaluation or decoupled architectures

---

## Next Steps

- Set up monitoring dashboards (Grafana + Prometheus)
- Implement dead-letter queue for failed messages
- Add result publishing to output stream or database
- Configure automatic stream trimming
