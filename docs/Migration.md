# Node → Go Migration Plan

## 🎯 Objective
Incrementally replace the legacy Node.js chatbot backend with the GoChatbot
service **without downtime, data loss, or behavioral regressions**.

The strategy is **parallel-run + gradual cutover**, not a big-bang rewrite.

---

## 🧠 Core Principles

1. **One endpoint at a time**
2. **Go owns new writes first**
3. **Node remains the source of truth until proven otherwise**
4. **Every migrated endpoint is locked by tests**
5. **Rollback is always possible**

---

## 🧱 Migration Phases

---

## Phase 0 — Foundation (✅ DONE)

- Go module initialized
- Domain errors defined
- Validation ported
- Pagination standardized
- Tenant API implemented
- Postgres repo tested with real DB
- HTTP handlers tested
- Service layer isolated

At this point:
> Go is *structurally safer* than Node.

---

## Phase 1 — Shadow Read (Tenants)

**Goal:** Go reads from the same Postgres data as Node.

Steps:
1. Point Go service at production-replica DB
2. Disable writes (`POST /tenants`) in Go
3. Compare:
   - Node tenant responses
   - Go tenant responses
4. Log diffs only (no user impact)

Exit criteria:
- 100% response parity
- No unexplained mismatches

---

## Phase 2 — Write Dual-Path (Tenants)

**Goal:** Go becomes the write path.

Steps:
1. Enable `POST /tenants` in Go
2. Node proxies tenant creation to Go
3. Node continues reads
4. DB constraints enforce correctness

Safety:
- Node fallback remains available
- Unique slug conflicts detected centrally

Exit criteria:
- No production errors
- All writes flow through Go

---

## Phase 3 — Read Cutover (Tenants)

**Goal:** Go serves tenants fully.

Steps:
1. Route tenant reads to Go
2. Node no longer touches tenant tables
3. Archive Node tenant logic

Exit criteria:
- Stable traffic
- Monitoring clean
- Rollback unused

---

## Phase 4 — Templates + Versions

Repeat Phases 1–3 for:
- Templates
- Template Versions
- Publishing logic

---

## Phase 5 — Transcripts / Email

Final migrations:
- Transcript rendering
- PDF generation
- Email dispatch
- Background jobs

---

## 🧪 Migration Safety Nets

- Domain-level invariants enforced in Go
- Postgres constraints as last line of defense
- Dual-write detection logging
- HTTP error parity tests

---

## 🚨 Rollback Strategy

At any phase:
- Switch routing back to Node
- Keep DB unchanged
- No schema rollback required

---

## 📌 Final State

- Go is authoritative backend
- Node becomes thin proxy (or removed)
- All logic covered by tests
- Lower operational risk
