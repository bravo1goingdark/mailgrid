## 🚀 Feature: `mailgrid serve` (API Server Mode)

Add a lightweight HTTP server that allows users to trigger email campaigns via a RESTful JSON API.

### ✅ Tasks
- [ ] Add `mailgrid serve` command to launch an HTTP server
- [ ] Expose POST `/send` endpoint for triggering campaigns
- [ ] Validate and parse JSON payload
- [ ] Run campaign using existing pipeline (CSV, templating, sending)
- [ ] Return batch ID and status as JSON

### 🛠️ Implementation Notes (Production Grade)
- Use `net/http` or `chi` router for simplicity and performance
- Run actual campaign logic in a separate goroutine and return immediately
- Use job queue with concurrency limit (`--max-workers`)
- Use UUID for campaign/batch ID
- Add basic bearer token or API key auth
- Log all requests + internal errors in structured format (JSONL)
- Add graceful shutdown and signal handling
---

## 📊 Feature: Campaign Analytics via Tracking Pixel

Track email opens and link clicks by embedding a tracking pixel and redirect URLs.

### ✅ Tasks
- [ ] Generate unique campaign ID per send
- [ ] Embed `<img src="/open/{{.email}}?cid=xxx">` in email body
- [ ] Add `/open/{email}` endpoint in `mailgrid serve`
- [ ] Log event to SQLite or flat JSONL file
- [ ] Optional: track link clicks via `/click?url=...&cid=...`

### 🛠️ Implementation Notes (Production Grade)
- Pixel route must respond with `1x1 transparent gif` and `Content-Type: image/gif`
- Store open events in SQLite table `open_events(email, campaign_id, ts, ua, ip)`
- Use UUIDs for campaign and user identification
- Do not block pixel request with DB writes — use buffered channel + batch flush
- Use user-agent + IP for basic device/location stats (opt-in)
- Track only the first open per user to prevent false inflation

---

## 🧠 Feature: Bounce Detection and Suppression List

Detect bounce events from SMTP and maintain a suppression list to avoid retrying.

### ✅ Tasks
- [ ] Parse SMTP response codes after each send
- [ ] Identify hard bounces (5XX codes, especially 550/554/5.1.1)
- [ ] Add email to local suppression list on permanent failure
- [ ] Skip suppressed emails in future sends
- [ ] Add `--no-suppression` flag to override

### 🛠️ Implementation Notes (Production Grade)
- Maintain `suppressed_emails.db` with schema `(email, reason, ts)`
- Use prepared statements for SQLite inserts (avoid IO bottleneck)
- Suppression list should be read into memory on start (hash set)
- Bounce check logic should run in sending thread to avoid race conditions
- Add TTL (e.g., 90 days) support to auto-prune old entries
- Export suppressed list as CSV via optional `mailgrid suppressions --export`

---

## 🧵 Feature: Per-Recipient State Tracking

Track the delivery state of each recipient for observability and retry capabilities.

### ✅ Tasks
- [ ] Record status per recipient (e.g., SENT, FAILED, BOUNCED, OPENED)
- [ ] Save state as JSONL or SQLite (`send_log.db`)
- [ ] Allow filtering by campaign ID or email

### 🛠️ Implementation Notes (Production Grade)
- Use campaign UUID and recipient email as composite key
- Record timestamp, retries, and SMTP response
- Batch write logs using a buffered writer or INSERT transaction
- Add CLI command to inspect past campaigns (`mailgrid history --campaign <id>`)
- Ensure thread-safe updates if sending is concurrent

--- 

## 📦 Feature: SMTP Pooling and Rotation

Allow configuring multiple SMTP credentials and rotate between them per email.

### ✅ Tasks
- [ ] Extend config to support SMTP pools:
```json
{
  "smtp_pool": [
    { "host": "...", "port": 587, "user": "...", "pass": "..." },
    { "host": "...", "port": 587, "user": "...", "pass": "..." }
  ]
}
```
- Randomly assign SMTP per message or use round-robin
- Handle failover if a server is unreachable
- Add `--pool-strategy git = random|round-robin`