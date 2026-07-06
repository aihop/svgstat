# CONTRACTS.md

# SVGStat Cross-System Contracts

> This document defines the concrete integration contracts between APayShop,
> Shoply, and SVGStat.
>
> It focuses on request/response shapes, field ownership, idempotency rules, and
> compatibility expectations.

---

# 1. Contract Purpose

Architecture documents explain who owns which responsibility.

This document explains:

* who calls whom
* which fields must be transmitted
* which system owns each field
* how to keep requests idempotent
* how to preserve backward compatibility

These contracts are intended to guide the first implementation and future API evolution.

---

# 2. System Roles

## APayShop

Acts as:

* official website
* pricing and purchase entry
* user-facing account center
* upstream commercial event initiator

---

## Shoply

Acts as:

* tenant and project lifecycle source of truth
* billing and entitlement source of truth
* control-plane caller of SVGStat

---

## SVGStat

Acts as:

* runtime SVG execution plane
* runtime analytics collector
* holder of runtime-oriented synchronized project state

---

# 3. Contract Rules

All cross-system contracts should follow these rules:

* use JSON for service-to-service payloads
* use `camelCase` field naming
* include explicit versioning
* include idempotency information for mutation requests
* keep field ownership unambiguous
* never send billing truth from SVGStat upstream as if it were authoritative

---

# 4. Envelope Format

Recommended JSON response envelope:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "requestId": "req_01J..."
}
```

Recommended JSON error envelope:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "project_not_found",
    "message": "Project does not exist"
  },
  "requestId": "req_01J..."
}
```

Rules:

* `requestId` should be generated per request
* `error.code` should be stable and machine-readable
* `error.message` may evolve for clarity

---

# 5. APayShop -> Shoply Contract

Purpose:

* communicate commercial events that may affect entitlement or lifecycle

Recommended endpoint:

```text
POST /internal/v1/svgstat/orders/activated
```

Suggested request body:

```json
{
  "version": "2026-06-26",
  "eventId": "evt_01J123456789",
  "eventType": "subscription.activated",
  "occurredAt": "2026-06-26T10:30:00Z",
  "source": "apayshop",
  "user": {
    "externalUserId": "usr_1001",
    "email": "user@example.com"
  },
  "subscription": {
    "planCode": "svgstat_pro",
    "billingCycle": "monthly",
    "status": "active",
    "startedAt": "2026-06-26T10:30:00Z",
    "renewAt": "2026-07-26T10:30:00Z"
  },
  "order": {
    "orderId": "ord_202606260001",
    "paymentProvider": "wechat",
    "providerReference": "wx_abc123"
  }
}
```

Field ownership:

* `planCode`, `billingCycle`, `status`, `renewAt` belong to Shoply after ingestion
* `orderId` and payment metadata are event evidence, not long-term SVGStat runtime state

Idempotency:

* `eventId` must be unique for the commercial event
* Shoply should treat repeated `eventId` submissions as safe retries

---

# 6. Shoply -> SVGStat Project Sync Contract

Purpose:

* synchronize runtime-serving project state into SVGStat

Recommended endpoint:

```text
POST /internal/v1/projects/sync
```

Suggested request body:

```json
{
  "version": "2026-06-26",
  "eventId": "evt_sync_01J123",
  "syncType": "project.upsert",
  "occurredAt": "2026-06-26T10:35:00Z",
  "tenant": {
    "tenantId": "tenant_001"
  },
  "project": {
    "externalProjectId": "proj_shoply_001",
    "slug": "acme-readme",
    "name": "Acme README Analytics",
    "status": "active",
    "visibility": "public",
    "publicToken": "pst_live_xxx",
    "planCode": "svgstat_pro"
  },
  "runtimePolicy": {
    "renderEnabled": true,
    "badgeEnabled": true,
    "widgetEnabled": true,
    "chartEnabled": true,
    "maxRequestsPerMinute": 600,
    "cacheTtlSeconds": 60
  },
  "presentation": {
    "defaultTheme": "dark",
    "defaultLocale": "en"
  }
}
```

Expected response:

```json
{
  "success": true,
  "data": {
    "accepted": true,
    "projectId": "svgstat_proj_001",
    "syncState": "updated"
  },
  "error": null,
  "requestId": "req_01J..."
}
```

Rules:

* `externalProjectId` is the upstream project identifier from Shoply
* `tenantId` is owned by Shoply
* `publicToken` is runtime-facing synchronized state, not billing truth
* SVGStat may transform this into local cache and local durable runtime records

Idempotency:

* `eventId` should deduplicate retries
* repeated `project.upsert` with same payload should be safe

---

# 7. Shoply -> SVGStat Project Disable Contract

Purpose:

* stop or restrict runtime-serving behavior for a project

Recommended endpoint:

```text
POST /internal/v1/projects/disable
```

Suggested request body:

```json
{
  "version": "2026-06-26",
  "eventId": "evt_disable_01J123",
  "syncType": "project.disable",
  "occurredAt": "2026-06-26T10:40:00Z",
  "tenant": {
    "tenantId": "tenant_001"
  },
  "project": {
    "externalProjectId": "proj_shoply_001",
    "slug": "acme-readme",
    "status": "disabled"
  },
  "reason": {
    "code": "subscription_expired",
    "message": "Subscription expired and grace period ended"
  }
}
```

Expected behavior:

* mark local runtime state as disabled
* evict or refresh runtime cache
* future public render requests respond according to policy

Possible render outcomes:

* deny rendering
* render fallback badge
* render disabled state SVG

That policy should be consistent per plan and product decision.

---

# 8. Shoply -> SVGStat Key Rotation Contract

Purpose:

* rotate runtime-facing access credentials without forcing manual database edits

Recommended endpoint:

```text
POST /internal/v1/projects/rotate-key
```

Suggested request body:

```json
{
  "version": "2026-06-26",
  "eventId": "evt_rotate_01J123",
  "occurredAt": "2026-06-26T10:45:00Z",
  "project": {
    "externalProjectId": "proj_shoply_001"
  },
  "key": {
    "name": "publicRuntimeKey",
    "keyId": "key_001",
    "keyHash": "sha256:abcdef123456",
    "expiresAt": "2026-12-31T23:59:59Z"
  }
}
```

Rules:

* plaintext keys should not be stored durably in SVGStat
* `keyHash` should be the stored value
* retries with same `eventId` must remain safe

---

# 9. SVGStat Dashboard Query Contract

Purpose:

* provide historical analytics to APayShop or Shoply surfaces

Recommended endpoint:

```text
GET /v1/projects/:projectId/overview
```

Example response:

```json
{
  "success": true,
  "data": {
    "projectId": "svgstat_proj_001",
    "today": {
      "pv": 1203,
      "uv": 488,
      "requests": 1401,
      "bots": 91
    },
    "last7Days": {
      "pv": 8412,
      "uv": 3204
    },
    "topReferrers": [
      {
        "name": "github.com",
        "count": 502
      }
    ],
    "topCountries": [
      {
        "name": "US",
        "count": 430
      }
    ]
  },
  "error": null,
  "requestId": "req_01J..."
}
```

Rules:

* caller auth is required
* query must be project-scoped
* near-real-time fields may read Redis
* historical fields may read PostgreSQL

---

# 10. Public SVG URL Contract

Purpose:

* provide a stable embed URL that developers can place into README files and websites

Recommended shape:

```text
GET /svg/:projectSlug/counter/:name.svg?label=Visitors&style=flat&theme=dark
```

Supported parameter classes:

* `label`
* `style`
* `theme`
* `locale`
* `format`
* `token`

Rules:

* URLs should remain backward compatible once public
* query parameters must be validated and sanitized
* `token` should be optional only for explicitly public projects
* SVG endpoints return `image/svg+xml`, not JSON

Example:

```text
/svg/acme-readme/counter/visitors.svg?label=Visitors&theme=dark
```

---

# 11. Tracking Contract

Purpose:

* optionally accept explicit analytics ingestion outside automatic render-path collection

Recommended endpoint:

```text
POST /v1/track
```

Suggested request body:

```json
{
  "version": "2026-06-26",
  "project": {
    "projectId": "svgstat_proj_001"
  },
  "event": {
    "type": "render",
    "resourceType": "counter",
    "resourceName": "visitors"
  },
  "request": {
    "referrer": "https://github.com/acme/repo",
    "userAgent": "Mozilla/5.0 ...",
    "country": "US",
    "device": "desktop",
    "browser": "Chrome"
  }
}
```

Rules:

* avoid duplicating automatic render-path counting unless explicitly intended
* normalize labels before storage
* do not accept unbounded arbitrary payloads

---

# 12. Status Vocabulary

Recommended project statuses:

* `pending`
* `active`
* `grace`
* `disabled`
* `archived`

Recommended sync types:

* `project.upsert`
* `project.disable`
* `project.rotateKey`
* `project.refreshCache`

Recommended commercial event types:

* `subscription.activated`
* `subscription.renewed`
* `subscription.upgraded`
* `subscription.cancelled`
* `subscription.expired`

These values should be treated as contract enums once implementation begins.

---

# 13. Compatibility Rules

Contracts should evolve with discipline.

Allowed:

* add optional fields
* add new event types
* add new response fields

Use caution:

* changing status semantics
* renaming public fields
* changing SVG URL shape

Avoid without version bump:

* removing required fields
* changing field meaning
* changing auth requirements incompatibly

---

# 14. Contract Invariants

The following rules are mandatory:

1. APayShop emits commercial events; it is not the lifecycle source of truth.
2. Shoply syncs runtime-serving project state into SVGStat.
3. SVGStat stores derived runtime state and historical analytics, not billing truth.
4. Every mutation request includes a version and idempotency event identifier.
5. Public SVG URLs are treated as long-lived public contracts.
6. Service-to-service contracts use stable machine-readable error codes.
7. Field ownership stays explicit across all payloads.

These rules keep the three-system platform interoperable and safe to evolve.
