# Domain Model & Invariants

This document defines **business invariants** that must hold regardless of
storage, transport, or implementation language.

Violations are domain errors — not HTTP errors.

---

## 🧩 Tenant

### Entity
```text
Tenant
- id (UUID)
- name (string)
- slug (string, unique)
- created_at (timestamp)
```

### Invariants
- slug:
    - lowercase
    - URL-safe
    - globally unique
- name:
    - trimmed
    - non-empty
- tenant identity is immutable

### Errors
- ErrTenantNotFound
- ErrTenantSlugTaken
- ErrInvalidSlug

## Template

### Entity
```text
Template
- id (UUID)
- tenant_id (UUID)
- name (string)
- slug (string)
- created_at (timestamp)
```

### Invariants
- belongs to exactly one tenant
- slug unique per tenant
- cannot be deleted once published
- templates are containers for versions

### Errors
- ErrTemplateNotFound
- ErrTemplateSlugTaken
- ErrTemplateImmutable

## Template Version

### Entity
```text
TemplateVersion
- id (UUID)
- template_id (UUID)
- version (int)
- status (draft | published)
- content (json)
- created_at (timestamp)
```

### Invariants
- exactly one published version per template
- versions are append-only
- published versions are immutable
- drafts may be replaced or deleted

### Transitions
```text
draft → published (allowed)
published → draft (forbidden)
```

### Errors
- ErrVersionNotFound
- ErrVersionAlreadyPublished
- ErrPublishedVersionImmutable

## Cross-Entity Invariants
- Tenants own templates
- Templates own versions
- Deleting a tenant:
    - forbidden if templates exist
- Deleting a template:
    - forbidden if published version exists

## Enforcement Layers
| Layer   | Responsibility      |
| ------- | ------------------- |
| Domain  | Define invariants   |
| Service | Enforce rules       |
| Repo    | Enforce constraints |
| DB      | Final safety net    |

## Testing Implications
- Every invariant must be covered by:
    - service-level unit tests
    - repo-level integration tests
- No invariant may rely solely on HTTP validation.

## Next Domain Expansion
- Sessions
- Conversations
- Messages
- Transcript artifacts
