# ARCHITECTURE.md

# SVGStat System Architecture

> This document defines the architecture of SVGStat.
>
> It describes **why** the system is designed this way, how each component interacts, and which architectural boundaries must never be violated.
>
> If implementation details conflict with this document, the implementation should be reconsidered.

---

# 1. Design Goals

SVGStat is designed around one primary requirement:

> Render dynamic SVGs with extremely low latency while collecting analytics at scale.

The system should satisfy the following goals:

* Low latency
* Horizontal scalability
* Stateless rendering
* Multi-tenant SaaS support
* Clear module boundaries
* Independent deployment
* Long-term maintainability

---

# 2. Architecture Overview

SVGStat is the runtime engine inside a three-repository SaaS architecture.

```text
Developers / Browsers / README Embeds
                    │
                    ▼
             Cloudflare CDN
                    │
                    ▼
              Go SVG Engine
                    │
      ┌─────────────┼─────────────┐
      ▼             ▼             ▼
  Renderer      Analytics      REST API
      │             │             │
      └──────┬──────┘             │
             ▼                    │
           Redis                  │
             │                    │
             ▼                    │
           Worker                 │
             │                    │
             ▼                    │
        PostgreSQL                │
                                  │
    APayShop Official Site / User Portal
                                  │
                                  ▼
                  Shoply SaaS Base / Billing / Lifecycle
```

The Go service handles all runtime traffic.

APayShop manages public website entry and user-center experience.

Shoply manages SaaS lifecycle, billing, and project control-plane operations.

---

# 3. Service Responsibilities

## Go SVG Engine

Responsible for:

* SVG rendering
* Analytics collection
* Badge generation
* Widget rendering
* REST API
* Worker execution
* Cache management

The Go service must remain stateless.

---

## Shoply

Responsible for:

* Authentication
* User management
* OAuth
* Billing
* Subscription
* Dashboard UI
* Teams
* API keys
* Project management

Shoply never renders SVG.

Shoply never participates in hot rendering paths.

---

## APayShop

Responsible for:

* official website
* pricing and plan discovery
* user account center
* purchase entry flow
* upstream notification into Shoply after payment

APayShop should never sit in the hot SVG render path.

---

# 4. Request Flows

## SVG Request

```text
Client
    │
Cloudflare
    │
Go HTTP
    │
Memory Cache
    │
Redis Pipeline
    │
Renderer
    │
SVG
    │
Response
```

Characteristics:

* No PostgreSQL writes
* No Shoply calls
* No blocking tasks

---

## Dashboard Request

```text
Browser
    │
Shoply
    │
Go API
    │
Redis
    │
PostgreSQL
    │
JSON Response
```

Dashboard traffic is separated from rendering traffic.

---

## Control-Plane Provisioning Flow

```text
User purchases in APayShop
    │
    ▼
APayShop notifies Shoply
    │
    ▼
Shoply provisions or updates project state
    │
    ▼
Shoply syncs SVGStat config
    │
    ▼
SVGStat refreshes runtime cache
```

Provisioning is separate from rendering.

---

# 5. Core Modules

## Renderer

The renderer is responsible only for presentation.

Responsibilities:

* Load SVG templates
* Apply data
* Produce SVG output

The renderer must never:

* Count visitors
* Access PostgreSQL
* Authenticate users

---

## Analytics

Analytics is responsible only for data collection.

Responsibilities:

* Page views
* Unique visitors
* Referrers
* Countries
* Browsers
* Devices

Analytics never generates SVG.

---

## Worker

Workers process asynchronous tasks.

Examples:

* Daily aggregation
* Cache refresh
* Cleanup
* Historical statistics

Workers should never affect request latency.

---

# 6. Storage Strategy

SVGStat uses layered storage.

```text
Application Memory
        │
        ▼
      Redis
        │
        ▼
   PostgreSQL
```

Each layer has a different responsibility.

## Memory

* Project configuration
* Frequently accessed objects
* Small hot datasets

---

## Redis

Runtime data.

Examples:

* Counters
* Today statistics
* Sessions
* Referrers
* Browser distribution

Redis is optimized for write throughput.

---

## PostgreSQL

Persistent storage.

Examples:

* Projects
* Historical reports
* Daily aggregates
* Billing metadata

PostgreSQL is not a real-time analytics database.

---

# 7. Cache Strategy

Cache priority:

```text
Memory
   │
Redis
   │
Database
```

The application should always attempt the highest cache layer first.

---

# 8. Analytics Pipeline

Every request follows the same pipeline.

```text
HTTP Request
      │
Normalize
      │
Bot Detection
      │
Resolve Project
      │
Memory Cache
      │
Redis Pipeline
      │
Response
      │
Worker Aggregation
      │
PostgreSQL
```

Collection should never block rendering.

---

# 9. Rendering Pipeline

Rendering should remain deterministic.

```text
Resolve Project
      │
Load Theme
      │
Load Template
      │
Apply Data
      │
Generate SVG
      │
Compress
      │
Return
```

Rendering must not contain business logic.

---

# 10. Multi-Tenancy

SVGStat is a SaaS platform.

Every request belongs to a project.

Isolation rules:

* Data isolation
* Cache isolation
* Rate limit isolation
* API key isolation

No project should be able to access another project's data.

---

# 11. Horizontal Scaling

The Go service is stateless.

Any instance should handle any request.

```text
Cloudflare
      │
Load Balancer
      │
 ┌────┴────┐
 │         │
Go #1   Go #2
 │         │
 └────┬────┘
      ▼
Redis Cluster
      ▼
PostgreSQL
```

Scaling should require adding instances, not changing code.

---

# 12. Failure Handling

Failures should degrade gracefully.

Examples:

* Redis unavailable → return SVG with cached/default values if possible.
* Worker unavailable → analytics aggregation delayed, rendering unaffected.
* PostgreSQL unavailable → rendering continues using cached runtime data where feasible.

The rendering path should be resilient.

---

# 13. Security

Never trust client input.

All requests should be:

* validated
* sanitized
* rate-limited

Secrets must never be embedded in SVG output.

---

# 14. Future Expansion

New features should extend existing modules rather than rewrite them.

Potential additions:

* Comment SVG
* Timeline widgets
* Heatmaps
* Public dashboards
* Team analytics
* Plugin system

Architecture should remain stable as capabilities grow.

---

# 15. Architectural Invariants

The following rules are non-negotiable:

1. Renderer never imports Analytics.
2. Analytics never imports Renderer.
3. SVG requests never write to PostgreSQL.
4. Runtime analytics always go through Redis first.
5. Business logic never lives in HTTP handlers.
6. Workers never block user requests.
7. Templates generate SVG; code should not manually concatenate SVG strings.
8. Every package owns a single responsibility.
9. APayShop and Shoply never join the hot SVG path.
10. Services communicate through stable APIs, not shared implementation details.
11. Maintainability is more important than cleverness.

These invariants define the long-term architecture of SVGStat and should guide every future design decision.
