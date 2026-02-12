# River Job Queue: Licensing & Pricing Research

**Research Date:** 2026-02-11

## Summary: Can We Use River for Free in Production?

**YES** - River is completely free to use in production without any paid plan. The core library and UI are open-source under MPL-2.0 license. River Pro is an optional paid add-on for advanced features.

---

## 1. Is River Open-Source?

**YES** - River is fully open-source under the **MPL-2.0 (Mozilla Public License 2.0)**.

### License History
- Originally released under LGPL license
- Changed to MPL-2.0 on November 22, 2023 (via [PR #60](https://github.com/riverqueue/river/issues/47))
- License change was made specifically to address Go's static linking concerns
- Both the core library (`riverqueue/river`) and UI (`riverqueue/riverui`) use MPL-2.0

### What MPL-2.0 Means
- **Commercial use allowed** - You can use River in commercial projects
- **Static linking friendly** - Unlike LGPL, MPL-2.0 doesn't require your entire app to be open-source
- **Modifications must be shared** - If you modify River's source code, those modifications must be shared
- **No patent retaliation** - Includes patent grant protections

**Sources:**
- [River GitHub Repository](https://github.com/riverqueue/river)
- [River UI GitHub Repository](https://github.com/riverqueue/riverui)
- [License Change Discussion (Issue #47)](https://github.com/riverqueue/river/issues/47)

---

## 2. Is River Free to Use?

**YES** - River is completely free to use with no cost.

### What's Included for Free:
- Core job queue library
- PostgreSQL-based job storage
- Reliable background job processing
- River UI (web interface) - self-hosted
- Transaction-safe job enqueueing
- Automatic retries and error handling
- Docker images and pre-built binaries
- Full production use without restrictions

**Sources:**
- [River Documentation](https://riverqueue.com/docs)
- [River Pro Pricing Page](https://riverqueue.com/pro)

---

## 3. Pricing Model

River operates on a **freemium model**:

### FREE Tier (Open Source)
- **Cost:** $0
- **License:** MPL-2.0 (open-source)
- **Includes:**
  - Core job queue functionality
  - River UI (web interface)
  - Full production capabilities
  - Self-hosted deployment
  - No developer limits
  - No revenue restrictions
  - No company size restrictions

### PRO Tier (Optional Paid Add-on)
- **Cost:** $125 USD/month (or annual plan with $300 savings)
- **Developer Limit:** Up to 20 developers
- **Additional Features:**
  - Workflows (complex task orchestration)
  - Batching
  - Concurrency limits
  - Dead letter queue
  - Email support
- **License:** Proprietary (distributed as private Go module)

### ENTERPRISE Tier (Optional)
- **Cost:** Custom pricing (contact sales)
- **Developer Limit:** Unlimited
- **Additional Features:**
  - All Pro features
  - Slack support
  - Invoice/PO billing
  - Site license

**Sources:**
- [River Pro Pricing Page](https://riverqueue.com/pro)
- [River Pro Documentation](https://riverqueue.com/docs/pro)

---

## 4. Usage Restrictions

### Open Source River (Free)
**NO RESTRICTIONS** on:
- Company size
- Revenue
- Number of developers
- Production use
- Commercial use
- Number of jobs processed

### River Pro (Paid)
**Restrictions based on developer count:**
- "Developer" defined as: "any individual who installs, runs, or develops against the Software"
- Excludes runtime-only environments (CI/CD, production servers)
- Must not exceed maximum developers for selected plan (20 for Pro, unlimited for Enterprise)

**Prohibited activities (River Pro only):**
- Reverse engineering or decompiling
- Redistributing or sublicensing
- Building competing products
- Publishing license keys
- Using unauthorized license keys

**NOTE:** These restrictions apply ONLY to River Pro features, NOT to the open-source River library.

**Sources:**
- [River Pro License Agreement](https://riverqueue.com/pro/license)

---

## 5. PostgreSQL + Go Alternatives

If River doesn't meet your needs, here are free, open-source alternatives:

### Gue
- **GitHub:** [vgarvardt/gue](https://github.com/vgarvardt/gue)
- **License:** MIT (fully permissive)
- **Status:** Actively maintained (v5 as of 2024)
- **Features:**
  - Transaction-level locks
  - Multiple PostgreSQL driver support (pgx v4, v5)
  - Adapter interface for different loggers
  - Originally a fork of que-go
- **Best for:** Simple, reliable job queuing with minimal dependencies

### pgq
- **GitHub:** [btubbs/pgq](https://github.com/btubbs/pgq)
- **License:** MIT (fully permissive)
- **Features:**
  - Easy-to-use API
  - Built-in retries
  - Exponential backoff support
- **Best for:** Lightweight job queuing without complex features

### que-go
- **GitHub:** [bgentry/que-go](https://github.com/bgentry/que-go)
- **Status:** UNMAINTAINED (author recommends using River instead)
- **License:** Not specified in search results
- **Note:** Interoperable with Ruby Que library

### Rickover
- **GitHub:** [Shyp/rickover](https://github.com/Shyp/rickover)
- **License:** Not specified in search results
- **Features:**
  - Job queue and scheduler
  - HTTP API
  - PostgreSQL-backed
- **Best for:** HTTP-based job queueing

**Sources:**
- [Gue GitHub Repository](https://github.com/vgarvardt/gue)
- [pgq GitHub Repository](https://github.com/btubbs/pgq)
- [que-go GitHub Repository](https://github.com/bgentry/que-go)
- [Rickover GitHub Repository](https://github.com/Shyp/rickover)

---

## Comparison Matrix

| Feature | River (Free) | River Pro | Gue | pgq |
|---------|-------------|-----------|-----|-----|
| License | MPL-2.0 | Proprietary | MIT | MIT |
| Cost | Free | $125/mo | Free | Free |
| PostgreSQL | Yes | Yes | Yes | Yes |
| Transaction-safe | Yes | Yes | Yes | ? |
| Web UI | Yes (free) | Yes | No | No |
| Workflows | No | Yes | No | No |
| Concurrency limits | No | Yes | No | No |
| Active maintenance | Yes | Yes | Yes | Less active |
| Commercial use | Yes | Yes | Yes | Yes |

---

## Recommendation

**For your project, River (free/open-source) is the best choice:**

1. **Fully free and open-source** - No hidden costs or limitations
2. **Production-ready** - Used by many companies
3. **MPL-2.0 license** - Commercial-friendly, no static linking issues
4. **Includes UI** - Built-in web interface for monitoring
5. **Active development** - Well-maintained with responsive maintainers
6. **PostgreSQL-native** - Built specifically for PostgreSQL
7. **Go-first design** - Designed by Go developers for Go ecosystem
8. **Transaction-safe** - Jobs never lost, atomic operations

**You only need River Pro if:**
- You need advanced workflows (complex task orchestration)
- You need concurrency limits at the job level
- You need commercial support
- You have a budget for premium features

For most projects, the free/open-source version is more than sufficient.

---

## Additional Resources

- [River GitHub](https://github.com/riverqueue/river)
- [River Documentation](https://riverqueue.com/docs)
- [River Blog (Technical Articles)](https://riverqueue.com/blog)
- [River: A Fast, Robust Job Queue for Go + Postgres (Blog Post)](https://brandur.org/river)
- [River Hacker News Discussion](https://news.ycombinator.com/item?id=38349716)

---

## Key Takeaways

1. River is **completely free** for production use
2. The core library is **open-source (MPL-2.0)** with no restrictions
3. River UI is **included for free** (also open-source)
4. River Pro is **optional** and only needed for advanced features
5. No company size, revenue, or usage limits on the free version
6. The license is **commercial-friendly** (MPL-2.0, not LGPL)
7. River is the **most feature-complete** PostgreSQL + Go job queue
8. Alternatives exist (Gue, pgq) but River offers the best balance of features and ease of use
