# GoChatbot Architecture

## 🎯 Purpose
This project is a refactor of a legacy Node.js chatbot backend into a **modern Go + Postgres** system with:

- explicit domain boundaries
- strict error semantics
- test-first, incremental migration
- long-term maintainability

Go module name: **`gochatbot`**

---

## 🧱 Architectural Principles

### 1. Layered, Dependency-Directional Design
```
HTTP (chi)
↓
Service (business rules)
↓
Repository (Postgres)
↓
Database
```


Dependencies flow **downward only**:
- HTTP knows about Service
- Service knows about Repo
- Repo knows about Postgres
- No layer reaches upward

---

### 2. Domain Errors Are the Contract
All meaningful errors are defined in `internal/domain` and propagated upward.

Examples:
- `ErrTenantNotFound`
- `ErrTenantSlugTaken`
- `ErrInvalidSlug`
- `ErrInvalidCursor`

Rules:
- Use `errors.Is`, never `==`
- HTTP layer maps domain errors → status codes
- Repo layer never returns HTTP errors

---

## 📁 Directory Structure
```
gochatbot/
├─ cmd/api/main.go # HTTP server entrypoint
├─ internal/
│ ├─ domain/ # Domain errors & invariants
│ ├─ validate/ # Pure validation (slug, phone, color)
│ ├─ pagination/ # Cursor encode/decode
│ ├─ httpapi/ # HTTP handlers (chi)
│ ├─ service/ # Business logic
│ ├─ repo/ # Postgres repositories
│ └─ testdb/ # Postgres test harness
```


---

## 🌐 HTTP Layer (`internal/httpapi`)

### Responsibilities
- Parse request input
- Validate parameters
- Call services
- Map domain errors → HTTP responses

### Characteristics
- No database logic
- No business rules
- Fully unit tested using fakes

### Example Status Mapping
| Condition | HTTP |
|---------|------|
| Invalid input | 400 |
| Not found | 404 |
| Conflict | 409 |
| Internal error | 500 |

---

## 🧠 Service Layer (`internal/service`)

### Responsibilities
- Enforce business rules
- Normalize inputs defensively
- Translate repo errors into domain errors
- Return HTTP-safe DTOs

### Characteristics
- No HTTP knowledge
- No SQL
- Testable with mocked repos

---

## 🗄 Repo Layer (`internal/repo`)

### Responsibilities
- Execute SQL queries
- Implement stable pagination
- Map DB errors → domain errors

### Details
- Uses **pgx v5**
- Cursor pagination based on:
- (created_at DESC, id DESC)
- Unique constraint → `ErrTenantSlugTaken`

---

## 🔁 Pagination Design

- Cursor structure:
```go
type Cursor struct {
  CreatedAt time.Time
  ID        string
}
```

- Encoded as base64
- Decoded and validated centrally
- Cursor errors return 400

## 🚀 Runtime
- cmd/api/main.go wires:
    - Postgres connection
    - Repo
    - Service
    - HTTP server
- Ready for:
    - pgxpool
    - graceful shutdown
    - request-scoped contexts

## ➡️ Next Architecture Steps
- Template + TemplateVersion domain
- Auth & tenant isolation
- Background jobs (email, transcripts)
- Replace Node endpoints incrementally
