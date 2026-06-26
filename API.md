# API.md

# SVGStat API Design

> This document defines the API surface of SVGStat.
>
> It covers runtime SVG endpoints, analytics collection, dashboard queries, and
> control-plane integration with APayShop and Shoply.

---

# 1. API Philosophy

SVGStat exposes APIs for two very different traffic classes:

1. Runtime SVG traffic
2. Control-plane and dashboard traffic

These classes must remain separated.

Runtime SVG endpoints are optimized for:

* low latency
* cacheability
* stateless execution
* high request volume

Control-plane endpoints are optimized for:

* project lifecycle management
* configuration synchronization
* historical queries
* secure administrative access

Never design a single endpoint that tries to serve both roles.

---

# 2. System Context

SVGStat is the runtime engine in a three-repository system:

* `Shoply` is the SaaS base for billing, tenancy, subscription state, and project provisioning.
* `APayShop` is the public website, pricing portal, and user account center.
* `SVGStat` serves the actual SVG URLs and collects runtime analytics.

API consequences:

* APayShop may initiate purchase and user-facing entry flows.
* Shoply is the source of truth for tenant/project lifecycle.
* SVGStat should expose stable APIs for rendering, analytics, and project sync.
* Neither APayShop nor Shoply should sit in the hot rendering path.

---

# 3. API Categories

## Runtime SVG

Used by browsers, README renderers, Markdown renderers, and static websites.

Examples:

* counter SVG
* badge SVG
* widget SVG
* chart SVG

Expected properties:

* GET only when possible
* cache-friendly
* no authentication round-trip to upstream systems
* no PostgreSQL writes

---

## Analytics Ingestion

Used to record request metadata associated with SVG access.

Expected properties:

* lightweight
* Redis-first
* non-blocking
* safe to call on every render request

---

## Dashboard Query

Used by APayShop user center or Shoply dashboard surfaces to query historical data.

Expected properties:

* authenticated
* project-scoped
* JSON responses
* can read Redis and PostgreSQL

---

## Control Plane

Used by Shoply to provision or update projects inside SVGStat.

Examples:

* create project
* rotate keys
* update plan
* disable project
* sync theme/widget configuration

These endpoints are not public embed endpoints.

---

# 4. Endpoint Shape

Suggested public surface:

```text
GET  /svg/:projectSlug/counter/:name.svg
GET  /svg/:projectSlug/badge/:name.svg
GET  /svg/:projectSlug/widget/:name.svg
GET  /svg/:projectSlug/chart/:name.svg

POST /v1/track

GET  /v1/projects/:projectId/overview
GET  /v1/projects/:projectId/trends
GET  /v1/projects/:projectId/referrers
GET  /v1/projects/:projectId/countries
GET  /v1/projects/:projectId/browsers
GET  /v1/projects/:projectId/devices

POST /internal/v1/projects/sync
POST /internal/v1/projects/disable
POST /internal/v1/projects/rotate-key
POST /internal/v1/projects/refresh-cache
```

The exact path layout may evolve, but the separation between:

* `/svg/...`
* `/v1/...`
* `/internal/v1/...`

should remain stable.

---

# 5. Authentication Rules

## Public Runtime Endpoints

SVG embed endpoints should avoid interactive authentication.

Allowed identification methods:

* signed project token
* public project slug with project-level access rules
* cache-resolved project configuration

Never call APayShop or Shoply synchronously during a hot SVG request to resolve identity.

---

## Dashboard Endpoints

Dashboard and management endpoints require authenticated callers.

Allowed callers:

* Shoply backend
* APayShop backend acting on behalf of the signed-in user
* internal services with machine credentials

Rules:

* every request must resolve a project
* every request must enforce tenant isolation
* no caller may access another tenant's project data

---

## Internal Endpoints

Internal endpoints must never rely on client-side trust.

Use:

* service-to-service signatures
* API keys stored as hashes
* allowlist or gateway protection when applicable

---

# 6. Response Rules

## SVG Endpoints

Return:

* `image/svg+xml`
* deterministic SVG body
* cache headers suitable for the endpoint type

Do not return JSON from SVG endpoints except explicit debug-only or internal tooling endpoints.

---

## JSON Endpoints

All JSON APIs should use a consistent envelope.

Suggested structure:

```json
{
  "success": true,
  "data": {},
  "error": null,
  "requestId": "..."
}
```

Error responses should remain machine-readable and stable across endpoints.

Suggested error structure:

```json
{
  "success": false,
  "data": null,
  "error": {
    "code": "project_not_found",
    "message": "Project does not exist or is not available"
  },
  "requestId": "..."
}
```

---

# 7. Project Isolation

Every endpoint must resolve project scope first.

Isolation applies to:

* runtime counters
* Redis keys
* dashboard queries
* cache entries
* rate limits
* API credentials

Never derive authorization from user input alone without validating project ownership or key scope.

---

# 8. Runtime Request Rules

For `/svg/...` requests:

* validate request parameters
* resolve project from memory cache first
* read counters from Redis or memory
* render through templates
* collect analytics asynchronously or through a lightweight Redis pipeline
* return SVG

Forbidden in hot path:

* PostgreSQL writes
* direct Shoply calls
* direct APayShop calls
* heavy joins
* blocking background work

---

# 9. Control-Plane Integration

Shoply is the control-plane source of truth.

Typical flow:

```text
User purchases plan in APayShop
        │
        ▼
APayShop notifies Shoply
        │
        ▼
Shoply provisions or updates project
        │
        ▼
Shoply calls SVGStat internal API
        │
        ▼
SVGStat refreshes cached project config
```

This keeps:

* billing in APayShop
* lifecycle and tenancy in Shoply
* runtime rendering in SVGStat

---

# 10. Versioning

Public JSON APIs should be versioned when contract stability matters.

Recommended:

* `/v1/...` for public and dashboard APIs
* `/internal/v1/...` for service-to-service APIs

SVG URL formats should change only with strong backward-compatibility reasons, because embed links are hard to migrate once published.

---

# 11. Rate Limiting

Rate limiting must be project-aware and endpoint-aware.

Examples:

* public SVG endpoints: high volume, low per-request cost
* dashboard APIs: lower volume, heavier reads
* internal sync APIs: low volume, privileged access

A single global limit is not sufficient for a multi-tenant SVG platform.

---

# 12. Observability

Every API should contribute structured telemetry.

Recommended labels:

* request type
* endpoint category
* project id
* cache hit or miss
* render latency
* Redis latency
* response size
* status code

Do not log secrets, raw API keys, or personally sensitive values into request logs.

---

# 13. API Invariants

The following rules are mandatory:

1. SVG endpoints stay stateless.
2. Hot render paths never write PostgreSQL.
3. Control-plane APIs stay separate from public embed APIs.
4. Project isolation is enforced before data access.
5. Shoply is the lifecycle source of truth.
6. APayShop is a user-facing portal, not the SVG runtime.
7. Error formats remain consistent.
8. Backward compatibility matters for published SVG URLs.

These rules protect SVGStat from coupling business control-plane logic into runtime rendering.
