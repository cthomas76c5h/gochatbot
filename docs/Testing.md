# GoChatbot Testing Strategy

## 🎯 Goals
- Make refactoring safe
- Enable aggressive iteration
- Catch regressions early
- Keep feedback fast

---

## 🧪 Test Categories

### 1. Unit Tests (Fast, Cached)

Covers:
- Validation logic
- Pagination encode/decode
- HTTP handlers
- Service business rules

Characteristics:
- No DB
- No network
- Use fakes/mocks
- Cached by Go automatically

Run:
```sh
go test ./...
```

### 2. Integration Tests (Postgres)

Covers:
- SQL correctness
- Constraints
- Pagination queries
- Error mapping from DB

Characteristics:
- Uses testcontainers-go
- Spins up real Postgres
- Slower but deterministic
- Isolated to repo layer

Run:
```sh
go test ./internal/repo
```

## 🧰 Test Infrastructure
Postgres Test Harness
- Located in internal/testdb
- Uses Docker + testcontainers

Includes:
- startup retry loop (Windows-safe)
- migration application
- per-test isolation

## 🧠 Error Testing Philosophy
Assert behavior, not messages
- Use errors.Is
- HTTP tests check status codes, not internals
- Repo tests verify real DB behavior

## 🧼 What Is Not Tested
- Standard library behavior
- pgx internals
- chi routing mechanics
- Tests focus on our logic, not dependencies.

## Test Workflow
### Fast Inner Loop
```bash
go test ./...
```

### Force Rerun (No Cache)
```bash
go test ./... -count=1
```

### Single Package
```bash
go test ./internal/httpapi
```

### Single Test
```bash
go test ./internal/service -run TestCreateTenant_Conflict
```

## 🧯 Common Pitfalls Avoided
- DB tests mixed with unit tests
- Error string comparisons
- Over-mocking
- Non-deterministic ordering

## ➡️ Next Testing Steps
- Table-driven tests for Templates
- Property tests for pagination
- Build tags for DB tests (optional)
- CI parallelization
