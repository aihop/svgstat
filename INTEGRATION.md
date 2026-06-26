# INTEGRATION.md

# SVGStat Cross-System Integration

> This document defines how SVGStat integrates with APayShop and Shoply.
>
> The goal is to keep billing, tenant lifecycle, and runtime SVG serving
> strictly separated while still forming one coherent SaaS product.

---

# 1. System Roles

The full product is split across three repositories:

## APayShop

Responsible for:

* official website
* pricing pages
* marketing entry
* user-facing account center
* purchase and renewal entry

APayShop is the customer-facing commercial portal.

It should not serve production SVG rendering traffic.

---

## Shoply

Responsible for:

* tenant lifecycle
* project provisioning
* subscription state
* billing truth
* entitlement checks
* API and admin control-plane workflows

Shoply is the control-plane source of truth.

It should not join the hot SVG request path.

---

## SVGStat

Responsible for:

* final SVG URLs
* runtime rendering
* runtime analytics
* Redis-first counters
* historical analytics aggregation
* runtime-oriented project config cache

SVGStat is the execution plane for developer-facing SVG embeds.

---

# 2. Why This Split Exists

Each system optimizes for a different type of work:

* APayShop optimizes for conversion and account experience.
* Shoply optimizes for SaaS lifecycle and operational control.
* SVGStat optimizes for low-latency stateless rendering.

Trying to merge these into one runtime path would create:

* slower renders
* weaker boundaries
* billing logic leaking into embed traffic
* harder scalability

---

# 3. High-Level Data Ownership

## APayShop Owns

Examples:

* pricing display state
* purchase UI state
* account-center interaction state
* website-facing session and presentation context

---

## Shoply Owns

Examples:

* tenant identity
* project lifecycle state
* billing and entitlement truth
* subscription activation
* team and user ownership relationships
* control-plane API credentials and provisioning workflow

---

## SVGStat Owns

Examples:

* runtime counters
* daily historical aggregates
* rendering configuration snapshots needed for runtime
* widget rendering settings
* SVG-facing project cache and access metadata

SVGStat may keep local copies of control-plane fields required for runtime, but those copies are derived state, not authoritative truth.

---

# 4. Core Integration Principle

The most important system rule is:

```text
Billing and lifecycle move downstream into SVGStat through explicit sync.
They are never resolved synchronously during hot SVG rendering.
```

This means:

* APayShop does not call SVGStat during checkout to render entitlement decisions.
* SVGStat does not call Shoply on every badge or counter request.
* Shoply pushes or synchronizes state into SVGStat ahead of runtime traffic.

---

# 5. Primary Business Flow

Recommended purchase and provisioning flow:

```text
User visits APayShop
    │
    ▼
User purchases or renews plan
    │
    ▼
APayShop confirms payment
    │
    ▼
APayShop notifies Shoply
    │
    ▼
Shoply creates or updates tenant/project entitlement
    │
    ▼
Shoply calls SVGStat internal sync API
    │
    ▼
SVGStat refreshes runtime config and cache
    │
    ▼
User receives or continues using SVG URLs
```

This keeps the commercial process and the rendering process decoupled.

---

# 6. Runtime Render Flow

The runtime SVG flow should look like this:

```text
Embed client requests SVG URL
    │
    ▼
Cloudflare or edge cache
    │
    ▼
SVGStat resolves project from memory or Redis
    │
    ▼
SVGStat reads hot metrics
    │
    ▼
SVGStat renders template
    │
    ▼
SVGStat collects analytics through Redis pipeline
    │
    ▼
SVG response returned
```

Forbidden in this flow:

* synchronous APayShop calls
* synchronous Shoply calls
* PostgreSQL writes
* provisioning logic
* payment logic

---

# 7. Sync Directions

## APayShop -> Shoply

Use when commercial state changes.

Examples:

* initial purchase success
* renewal success
* upgrade or downgrade intent
* cancellation or expiration notifications

APayShop should not directly become the lifecycle source of truth for SVGStat projects.

---

## Shoply -> SVGStat

Use when runtime-serving state must change.

Examples:

* create project
* activate project
* disable project
* rotate public key or token
* update plan-derived render limits
* update widget or theme config snapshot

This is the most important integration direction for runtime correctness.

---

# 8. Local Shadow State in SVGStat

SVGStat may persist local shadow fields for runtime use, such as:

* project status
* public slug
* render eligibility flags
* limits relevant to widget or badge output
* public-facing configuration snapshot

Rules:

* local copies must be refreshable
* local copies must not become independent sources of truth
* conflicts should resolve in favor of Shoply control-plane data

---

# 9. Failure Handling Across Systems

Expected degraded behaviors:

* APayShop unavailable -> purchases or portal flows affected, existing SVG embeds continue where cached/runtime state allows
* Shoply unavailable -> provisioning and lifecycle updates delayed, existing hot render traffic should continue using synced state
* SVGStat unavailable -> embed traffic fails even if billing and portal are healthy

This is intentional.

The runtime system should be able to survive temporary upstream control-plane outages after sync.

---

# 10. Security Boundaries

Cross-system communication must be explicit and authenticated.

Recommended controls:

* service-to-service signatures
* hashed API credentials
* narrow internal endpoints
* allowlists or gateway protection
* audit logging for privileged sync actions

Never trust front-end claims to mutate runtime project state directly inside SVGStat.

---

# 11. Schema Ownership Rule

A critical rule for future implementation:

* user, billing, and team truth belong to Shoply
* website-facing account experience belongs to APayShop
* runtime analytics and render persistence belong to SVGStat

If SVGStat stores user, billing, or team data, it should only do so as a minimal derived reference required for runtime or dashboard projection, never as primary ownership.

---

# 12. Integration Invariants

The following rules are mandatory:

1. APayShop is the commercial portal, not the runtime engine.
2. Shoply is the lifecycle source of truth.
3. SVGStat is the runtime execution plane for SVG embeds.
4. Billing and provisioning never execute inside hot render requests.
5. SVGStat consumes upstream state through explicit sync, not per-request dependency.
6. Derived local state in SVGStat must remain overrideable by Shoply truth.
7. Failure in APayShop or Shoply should degrade control-plane freshness before it breaks hot render traffic.

These rules keep the overall architecture scalable, understandable, and safe to evolve.
