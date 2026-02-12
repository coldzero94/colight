# Async Job Processing Options for Colight Go Backend

> Research Date: 2026-02-11
> Context: Evaluating async job queue solutions for Playwright browser automation (Phase 10)

## Executive Summary

This research evaluates async job processing options for the Colight Go backend, specifically to handle heavy Playwright browser automation for wanted.co.kr crawling. The primary goal is to determine if we can avoid deploying a separate worker service while using PostgreSQL (already available via Supabase) without adding Redis or Python dependencies.

**Recommendation**: Use **River (PostgreSQL-based, Go-native)** in **embedded mode** initially, with option to scale to separate worker process later if needed.

---

## Current Architecture Context

| Component | Details |
|-----------|---------|
| Backend | Go (Gin + Ent) on Koyeb free tier |
| Resources | 512MB RAM, 0.1 vCPU, 2GB SSD |
| Database | Supabase PostgreSQL (pgvector enabled) |
| Heavy Jobs | Playwright browser automation for wanted.co.kr (Phase 10) |
| Reference | mindhit uses Redis + Asynq with separate worker service |

### Constraints
- Koyeb free tier: 512MB RAM, 0.1 vCPU (low resources)
- No timeout per request (long-running jobs are supported)
- Already have PostgreSQL (Supabase)
- Prefer avoiding additional infrastructure (Redis, Python worker)

---

## Option 1: Go Backend Built-in Async

### 1.1 Goroutines + Channels (Simplest)

**Description**: Use Go's native concurrency primitives for background job processing.

**Architecture**:
```go
// Simple goroutine approach
go func() {
    result := crawlWithPlaywright(url)
    saveToDatabase(result)
}()
```

**Pros**:
- Zero external dependencies
- Minimal complexity
- No additional deployment
- Perfect for low-volume jobs (1-10 per hour)

**Cons**:
- No built-in retry logic
- No job persistence (lost on restart)
- No job queue management
- Poor visibility into job status
- Memory leaks if not managed properly (goroutines don't exit)
- Risk of OOM on Koyeb free tier (512MB RAM)

**Resource Impact**:
- Goroutine overhead: ~2KB per goroutine
- Playwright process: ~700MB (headless), ~1GB (standard)
- **Critical**: Playwright alone exceeds Koyeb free tier RAM limit

**Best For**: Very simple background tasks (email sending, logging), NOT browser automation

**Verdict**: ❌ **Not Recommended** - Playwright memory requirements exceed free tier capacity, no retry/persistence

---

### 1.2 River (PostgreSQL-based Job Queue for Go)

**Description**: Fast, reliable background job processing using PostgreSQL as the backend. Go-native, transactional, with built-in web UI.

**GitHub**: [riverqueue/river](https://github.com/riverqueue/river)
**Documentation**: [riverqueue.com](https://riverqueue.com/)

**Key Features**:
- Transactional job enqueueing (ACID guarantees)
- LISTEN/NOTIFY for sub-millisecond job pickup
- Multiple queues with priority support
- Scheduled and periodic jobs (cron-like)
- Job cancellation, retries, and unique jobs
- Web UI for monitoring
- Batch job insertion (PostgreSQL COPY FROM)
- Multi-language support (Python/Ruby clients can enqueue)

**Architecture Options**:

**Option A: Embedded Mode (Same Process)**
```go
// Run workers in the API server process
client, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        river.QueueDefault: {MaxWorkers: 10},
        "crawling":         {MaxWorkers: 2}, // Limit concurrent Playwright jobs
    },
})
client.Start(ctx) // Starts workers inline
```

**Option B: Separate Worker (Different Process)**
```go
// API server - insert-only client (no Queues config, no Start())
apiClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})

// Worker process - worker-only client
workerClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        "crawling": {MaxWorkers: 5},
    },
})
workerClient.Start(ctx)
```

**Performance**:
- ~10,000 trivial jobs/second on commodity hardware
- Uses PostgreSQL binary protocol (pgx) for efficiency
- LISTEN/NOTIFY for real-time job pickup (vs polling)
- Batch operations minimize database round trips

**Pros**:
- ✅ Uses existing PostgreSQL (no Redis needed)
- ✅ Transactional job safety (job enqueued only if transaction commits)
- ✅ Go-native (same language as API)
- ✅ Can start embedded, scale to separate worker later
- ✅ Built-in retry logic, job persistence
- ✅ Web UI for monitoring
- ✅ Active development, modern codebase

**Cons**:
- ⚠️ Playwright memory usage still a concern (700MB+ per browser)
- ⚠️ Embedded mode: API server + worker in same 512MB RAM (likely OOM)
- ⚠️ Separate worker: requires second Koyeb instance (paid tier)

**Resource Estimation (Embedded Mode)**:
```
API Server:       ~50MB
River Client:     ~20MB
Playwright (1):   ~700MB (headless)
Total:            ~770MB -> EXCEEDS 512MB free tier
```

**Resource Estimation (Separate Worker)**:
```
API Server:       ~50MB (insert-only client minimal overhead)
Worker Instance:  ~50MB + 700MB Playwright = ~750MB
Total:            ~800MB across 2 instances (1 free + 1 paid)
```

**Best For**:
- Initial MVP: Embedded mode with **strict concurrency limits** (1 Playwright job at a time)
- Production: Separate worker process on dedicated instance

**Deployment Strategy**:
1. **Phase 10 (MVP)**: Embedded mode, MaxWorkers: 1 for "crawling" queue
2. **Post-MVP**: Deploy separate worker on paid Koyeb instance (Nano: $5.5/mo, 1GB RAM)

**Verdict**: ✅ **RECOMMENDED** - Best balance of functionality and simplicity, scales gracefully

---

### 1.3 Asynq (Redis-based, Go-native)

**Description**: Redis-backed distributed task queue for Go. Used by mindhit reference project.

**GitHub**: [hibiken/asynq](https://github.com/hibiken/asynq)

**Key Features**:
- Redis as backend (fast, in-memory)
- Retry with exponential backoff
- Task priority, scheduled tasks
- Unique tasks, rate limiting
- Web UI (asynqmon)

**PostgreSQL Support**: ❌ **Redis-only** (as of 2026, no native PostgreSQL backend)

**Alternative**: Asynq is part of **Neoq** library which provides PostgreSQL backend option, but Neoq is less mature than River.

**Pros**:
- ✅ Proven, mature library
- ✅ Fast (in-memory Redis)
- ✅ Used by mindhit (reference project)

**Cons**:
- ❌ Requires Redis (additional infrastructure)
- ❌ No PostgreSQL backend support
- ❌ Need to manage Redis instance (cost, maintenance)
- ⚠️ Still requires separate worker process for heavy jobs

**Cost Impact**:
- Supabase doesn't offer Redis
- Need external Redis provider (Upstash free tier: 10K commands/day, 256MB)
- Or self-host Redis on Koyeb (requires paid instance)

**Verdict**: ❌ **Not Recommended** - Adds Redis dependency, no advantage over River for PostgreSQL users

---

### 1.4 Database Polling (Simple Job Table)

**Description**: Create a `jobs` table, poll for pending jobs, process in background goroutine.

**Architecture**:
```sql
CREATE TABLE jobs (
    id UUID PRIMARY KEY,
    type VARCHAR(50),
    payload JSONB,
    status VARCHAR(20), -- pending, processing, completed, failed
    attempts INT DEFAULT 0,
    created_at TIMESTAMP,
    scheduled_at TIMESTAMP
);
```

```go
// Background poller
go func() {
    ticker := time.NewTicker(5 * time.Second)
    for range ticker.C {
        jobs := fetchPendingJobs(db)
        for _, job := range jobs {
            go processJob(job)
        }
    }
}()
```

**Pros**:
- ✅ Very simple implementation
- ✅ No external dependencies
- ✅ Full control over logic

**Cons**:
- ❌ Inefficient polling (vs LISTEN/NOTIFY)
- ❌ Manual retry logic implementation
- ❌ No built-in concurrency control
- ❌ Race conditions without proper locking
- ❌ No monitoring/UI
- ❌ Reinventing the wheel (River does this better)

**Performance**: Polling every 5 seconds means 5-second latency. LISTEN/NOTIFY (River) has sub-millisecond latency.

**Verdict**: ❌ **Not Recommended** - River provides all this functionality out-of-the-box

---

## Option 2: PostgreSQL + Procrastinate (Python)

**Description**: Python async job queue using PostgreSQL, requiring separate Python worker service.

**GitHub**: [procrastinate-org/procrastinate](https://github.com/procrastinate-org/procrastinate)
**Documentation**: [procrastinate.readthedocs.io](https://procrastinate.readthedocs.io/)

**Architecture**:
```
Go API Server (Koyeb Free)
    |
    | INSERT INTO procrastinate_jobs
    v
PostgreSQL (Supabase)
    ^
    | SELECT ... FOR UPDATE SKIP LOCKED
    |
Python Worker (Koyeb Paid)
    |
    | Playwright automation
    v
Save results to PostgreSQL
```

**Key Features**:
- Python 3.10+ async/await support
- PostgreSQL-based (13+)
- Retry logic, periodic tasks
- Django integration
- LISTEN/NOTIFY support

**Go Integration**:
Go API would **directly insert** into Procrastinate's PostgreSQL jobs table:
```go
// Go API enqueues job by inserting into procrastinate_jobs table
_, err := db.Exec(`
    INSERT INTO procrastinate_jobs (queue_name, task_name, args, scheduled_at)
    VALUES ($1, $2, $3, NOW())
`, "crawling", "crawl_wanted", jsonPayload)
```

**Pros**:
- ✅ Uses existing PostgreSQL
- ✅ Python has mature Playwright support
- ✅ Separate worker isolates heavy jobs from API

**Cons**:
- ❌ Requires Python runtime (different language)
- ❌ Need to deploy separate Python worker service
- ❌ Added complexity: maintain two codebases (Go + Python)
- ❌ No type safety between Go enqueue and Python worker
- ❌ Deployment overhead (manage Python dependencies, Docker image)
- ❌ Go API directly writes to Procrastinate's internal schema (fragile)

**Deployment**:
- API: Go on Koyeb free tier
- Worker: Python on Koyeb paid tier (Nano: $5.5/mo)

**Playwright Support**: Python Playwright is official, well-maintained (playwright-python)

**Verdict**: ⚠️ **Possible but NOT Recommended** - Adds language complexity, no advantage over Go-native River

---

## Option 3: Hybrid (Separate Worker, Same Language)

**Description**: Use River with Go API (insert-only) + separate Go worker process, similar to mindhit's architecture but with PostgreSQL instead of Redis.

**Architecture**:
```
apps/backend/
├── cmd/
│   ├── api/          # Main API server (insert-only River client)
│   └── worker/       # Background worker (River workers enabled)
├── internal/
│   ├── jobs/         # Shared job definitions
│   │   ├── crawl_wanted.go
│   │   └── crawl_jobkorea.go
│   └── ...
```

**API Server** (apps/backend/cmd/api/main.go):
```go
// Insert-only client (no workers)
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    // No Queues config = insert-only mode
})

// Enqueue job
_, err = riverClient.Insert(ctx, CrawlWantedArgs{URL: url}, nil)
```

**Worker Server** (apps/backend/cmd/worker/main.go):
```go
// Worker-only client
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        "crawling": {MaxWorkers: 3},
    },
    Workers: river.NewWorkers(),
})

// Register workers
river.AddWorker(riverClient.Config().Workers, &CrawlWantedWorker{})

// Start processing
riverClient.Start(ctx)
```

**Pros**:
- ✅ Unified language (Go for both API and worker)
- ✅ Shared codebase (job definitions in internal/jobs)
- ✅ Type-safe job enqueue/dequeue
- ✅ Isolates heavy jobs from API server
- ✅ Can scale workers independently
- ✅ Uses existing PostgreSQL

**Cons**:
- ⚠️ Requires separate Koyeb instance (paid tier: $5.5/mo for Nano)
- ⚠️ Deployment complexity (two services to manage)

**Cost Breakdown**:
- API Server: Koyeb Free tier (512MB RAM)
- Worker: Koyeb Nano tier (1GB RAM) - $5.5/mo
- Total: $5.5/mo (vs $0 for embedded mode)

**When to Use**:
- Production workload (multiple crawls per hour)
- Need horizontal scaling of workers
- API server stability is critical

**Verdict**: ✅ **RECOMMENDED for Production** - Best architecture for scale, same as mindhit pattern

---

## Comparison Table

| Option | Complexity | Cost | Pros | Cons | Recommendation |
|--------|------------|------|------|------|----------------|
| **Goroutines + Channels** | Low | Free | Zero dependencies, simple | No persistence/retry, memory risk | ❌ Not for Playwright |
| **River (Embedded)** | Low | Free | Go-native, PostgreSQL, easy start | RAM limit (512MB vs 700MB Playwright) | ✅ MVP only (1 job at a time) |
| **River (Separate Worker)** | Medium | $5.5/mo | Scalable, type-safe, isolated | Requires paid tier | ✅ **Production** |
| **Asynq (Redis)** | Medium | Free-$10/mo | Mature, fast | Requires Redis | ❌ Unnecessary complexity |
| **Database Polling** | Low | Free | Simple, full control | Poor performance, manual work | ❌ Reinventing wheel |
| **Procrastinate (Python)** | High | $5.5/mo | Separate worker, Python Playwright | Two languages, fragile integration | ❌ Language mismatch |

---

## Memory and CPU Considerations

### Goroutine Memory Overhead
- Single goroutine: ~2KB initial stack
- 1000 goroutines: ~2MB (negligible)
- Context switching overhead: Minimal (Go scheduler is efficient)

**Best Practice**: Use worker pool pattern (limit concurrent goroutines to 10-20 per CPU core)

Source: [Optimizing Goroutine Usage](https://vatsalchauhan.medium.com/optimizing-goroutine-usage-for-high-concurrency-0b7448e55ac8)

### Playwright Memory Usage
- **Headless Chromium**: ~700MB peak memory
- **Standard Chromium**: ~1GB peak memory
- **playwright-go**: ~50MB Node.js bridge process

**Koyeb Free Tier**: 512MB RAM, 0.1 vCPU
- **Verdict**: Cannot run Playwright in embedded mode without OOM risk

Source: [Playwright Browser Footprint](https://datawookie.dev/blog/2025/06/playwright-browser-footprint/)

### PostgreSQL Job Queue Performance

**Polling vs LISTEN/NOTIFY**:
- Polling (every 5s): 5-second latency, constant CPU load
- LISTEN/NOTIFY: Sub-millisecond latency, event-driven (efficient)

**River's Approach**: Hybrid (LISTEN/NOTIFY + periodic poll fallback)

Source: [PostgreSQL LISTEN/NOTIFY for Job Queues](https://aminediro.com/posts/pg_job_queue/)

---

## Koyeb Free Tier Limitations

| Resource | Limit | Impact on Async Jobs |
|----------|-------|----------------------|
| RAM | 512MB | Cannot run Playwright + API in same process |
| vCPU | 0.1 vCPU | CPU-bound tasks are slow |
| Storage | 2GB SSD | Adequate for logs |
| Timeout | None | ✅ Long-running jobs supported |
| Instances | 1 per org | Cannot scale horizontally on free tier |
| Region | Frankfurt or DC | Single region only |

**Key Insight**: Koyeb free tier supports long-running jobs (no timeout), but RAM is the bottleneck for Playwright.

Source: [Koyeb Free Tier Pricing](https://www.koyeb.com/pricing)

---

## Recommended Architecture

### Phase 1: MVP (Free Tier Only)

**Use River in Embedded Mode with Strict Limits**

```go
// apps/backend/cmd/api/main.go
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        river.QueueDefault: {MaxWorkers: 10},   // Light jobs
        "crawling":         {MaxWorkers: 1},     // CRITICAL: Only 1 concurrent Playwright
    },
    Workers: river.NewWorkers(),
})

// Register crawler worker
river.AddWorker(riverClient.Config().Workers, &CrawlWantedWorker{
    playwright: playwright, // Reuse single browser instance
})

riverClient.Start(ctx)
```

**Key Constraints**:
- MaxWorkers: 1 for "crawling" queue (prevent OOM)
- Reuse single browser instance (avoid spawning multiple Chromium processes)
- Monitor memory usage closely

**Tradeoffs**:
- ✅ Zero cost
- ✅ Simple deployment
- ⚠️ Sequential processing (only 1 crawl at a time)
- ⚠️ API server restart = job interruption

---

### Phase 2: Production (Paid Tier)

**Use River with Separate Worker Process**

**Deployment**:
- API Server: Koyeb Free tier (512MB RAM) - insert-only client
- Worker: Koyeb Nano tier (1GB RAM, $5.5/mo) - worker client

**API Server** (apps/backend/cmd/api/):
```go
// Insert-only client (no workers)
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{})

// Enqueue job
_, err = riverClient.Insert(ctx, CrawlWantedArgs{URL: url}, nil)
```

**Worker** (apps/backend/cmd/worker/):
```go
// Worker client with concurrency
riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
    Queues: map[string]river.QueueConfig{
        "crawling": {MaxWorkers: 3}, // Safe on 1GB RAM
    },
    Workers: river.NewWorkers(),
})

river.AddWorker(riverClient.Config().Workers, &CrawlWantedWorker{})
riverClient.Start(ctx)
```

**Benefits**:
- ✅ API server stability (isolated from heavy jobs)
- ✅ Horizontal scaling (add more worker instances)
- ✅ Type-safe, single codebase (Go)
- ✅ PostgreSQL-native (no Redis)

**Cost**: $5.5/mo (Nano tier)

---

## Implementation Checklist

### Phase 10.1 (MVP - Embedded Mode)

- [ ] Install River: `go get github.com/riverqueue/river`
- [ ] Run River migrations: `river migrate-up --database-url $DATABASE_URL`
- [ ] Create `internal/jobs/crawl_wanted.go` worker
- [ ] Configure River client in `cmd/api/main.go` (embedded mode)
- [ ] Set `MaxWorkers: 1` for "crawling" queue
- [ ] Implement Playwright worker with browser instance reuse
- [ ] Add job enqueue endpoint: `POST /api/crawl` -> `riverClient.Insert()`
- [ ] Monitor memory usage in Koyeb dashboard
- [ ] Test: Enqueue 3 jobs sequentially, verify single-threaded processing

### Phase 10.2 (Production - Separate Worker)

- [ ] Create `cmd/worker/main.go` entrypoint
- [ ] Move job definitions to `internal/jobs/` (shared package)
- [ ] Configure API server as insert-only client (no `Queues` config)
- [ ] Configure worker server with `Queues` and `Workers`
- [ ] Create Dockerfile for worker service
- [ ] Deploy worker to Koyeb Nano tier
- [ ] Set `MaxWorkers: 3` for "crawling" queue
- [ ] Test: Enqueue 5 jobs, verify concurrent processing (up to 3)
- [ ] Set up River Web UI for monitoring

---

## Resources and References

### River (PostgreSQL Job Queue)
- [GitHub: riverqueue/river](https://github.com/riverqueue/river)
- [Official Documentation](https://riverqueue.com/docs)
- [River: Fast, Robust Job Queue - brandur.org](https://brandur.org/river)
- [Hacker News Discussion](https://news.ycombinator.com/item?id=38349716)

### PostgreSQL as Job Queue
- [Rethinking the Queue: PostgreSQL vs Redis](https://medium.com/@soaebhasan04/rethinking-the-queue-is-postgresql-a-viable-alternative-to-redis-for-job-processing-45f580fd4236)
- [PostgreSQL LISTEN/NOTIFY Performance](https://aminediro.com/posts/pg_job_queue/)
- [Choose Postgres Queue Technology](https://adriano.fyi/posts/2023-09-24-choose-postgres-queue-technology/)

### Playwright Performance
- [Playwright Browser Footprint](https://datawookie.dev/blog/2025/06/playwright-browser-footprint/)
- [Playwright Performance Improvements 2026](https://www.skyvern.com/blog/puppeteer-vs-playwright-complete-performance-comparison-2025/)
- [Python vs Node.js for Web Scraping](https://pixeljets.com/blog/web-scraping-playwright-python-nodejs/)

### Go Concurrency Best Practices
- [Optimizing Goroutine Usage](https://vatsalchauhan.medium.com/optimizing-goroutine-usage-for-high-concurrency-0b7448e55ac8)
- [Goroutine Worker Pools](https://goperf.dev/01-common-patterns/worker-pool/)
- [Job Queues in Go](https://www.opsdash.com/blog/job-queues-in-go.html)

### Koyeb Platform
- [Koyeb Free Tier Pricing](https://www.koyeb.com/pricing)
- [Koyeb Instances Reference](https://www.koyeb.com/docs/reference/instances)
- [Koyeb vs Competitors](https://www.freetiers.com/directory/koyeb)

### Procrastinate (Python)
- [Procrastinate Documentation](https://procrastinate.readthedocs.io/)
- [GitHub: procrastinate-org/procrastinate](https://github.com/procrastinate-org/procrastinate)

---

## Conclusion

**Final Recommendation**: Use **River (PostgreSQL-based, Go-native)** with a two-phase approach:

1. **MVP (Phase 10)**: Embedded mode with strict concurrency (MaxWorkers: 1)
   - Zero cost, simple deployment
   - Acceptable for low-volume testing (1-5 crawls/hour)
   - Monitor memory closely, accept sequential processing

2. **Production (Post-MVP)**: Separate worker process on Koyeb Nano tier
   - $5.5/mo cost
   - Scalable, reliable, isolated from API server
   - Same architecture as mindhit (Redis + Asynq), but PostgreSQL-native

**Why River over alternatives**:
- ✅ Uses existing PostgreSQL (no Redis)
- ✅ Go-native (type-safe, single language)
- ✅ Flexible deployment (embedded → separate worker)
- ✅ Production-ready (transactional safety, retries, monitoring)
- ✅ Active development, modern codebase

**Avoid**:
- ❌ Goroutines + Channels (no persistence, memory risk)
- ❌ Asynq (requires Redis)
- ❌ Procrastinate (adds Python complexity)
- ❌ Database polling (River does this better)
