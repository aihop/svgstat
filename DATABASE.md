# DATABASE.md

# SVGStat Database Design

> This document defines the persistent data model for SVGStat.
>
> PostgreSQL is used as the durable store inside SVGStat.
>
> It is **not** responsible for real-time analytics.

---

# 1. Design Principles

PostgreSQL is designed for:

* durability
* consistency
* historical storage
* SVGStat-owned persistent data

PostgreSQL should never become the hot path for SVG rendering.

---

# 2. System Context

SVGStat is only one part of the broader product:

* `APayShop` owns the public website, pricing portal, and user-facing account center.
* `Shoply` owns tenancy, billing, subscription truth, and project lifecycle.
* `SVGStat` owns runtime rendering, runtime analytics, and historical analytics persistence.

This means the SVGStat database should store SVGStat-owned durable state, not become a duplicate SaaS master database.

---

# 3. Responsibilities

Database stores:

* Projects
* API Keys
* Historical Statistics
* Aggregated Reports
* Project Configuration
* Runtime-facing shadow state synchronized from the control plane when needed

Database does **not** store:

* Real-time PV
* Runtime Counters
* Active Sessions
* Temporary Cache
* Billing source-of-truth records
* Team ownership source-of-truth records
* Full account-center source-of-truth records

Runtime hot data belongs in Redis.

Billing, team ownership, and account-center truth belong in upstream systems such as Shoply or APayShop.

---

# 4. Ownership Boundaries

## Owned by Shoply

Examples:

* tenant identity
* subscription truth
* billing state
* team relationships
* project lifecycle truth

SVGStat may cache or shadow selected fields required for runtime, but Shoply remains authoritative.

---

## Owned by SVGStat

Examples:

* render-facing project records
* hashed API keys for SVG access
* historical daily aggregates
* widget and render configuration needed by the runtime

This keeps the runtime database focused and compact.

---

# 5. Storage Layers

```text
Memory
    │
    ▼
Redis
    │
    ▼
PostgreSQL
```

Each layer serves a different purpose.

Never bypass the hierarchy without a strong reason.

---

# 6. Entity Relationship

```text
Shoply Tenant / Project
          │
          ▼
SVGStat Projects
    │      │
    │      ├── API Keys
    │      ├── Daily Stats
    │      ├── Widgets
    │      └── Settings
```

Each project is isolated.

No shared analytics data.

---

# 7. projects

Purpose:

Represents an analytics project.

Typical fields:

```text
id

external_project_id

tenant_id

slug

name

description

status

created_at

updated_at
```

Rules:

* slug must be globally unique.
* Project IDs never change.
* Deleting a project should use soft delete unless explicitly purged.
* Upstream lifecycle state should be synchronized from Shoply rather than invented independently inside SVGStat.

---

# 8. api_keys

Purpose:

Authentication between clients and SVGStat.

Fields:

```text
id

project_id

name

key_hash

last_used_at

expires_at

created_at
```

Rules:

* Never store plaintext API keys.
* Store hashes only.
* Keys should be revocable.
* If keys are rotated upstream, SVGStat must update its local runtime copy through explicit sync.

---

# 9. daily_statistics

Purpose:

Historical analytics.

Typical fields:

```text
id

project_id

date

pv

uv

requests

bots

countries

devices

browsers

created_at
```

This table stores aggregated values only.

Never insert one row per request.

---

# 10. widget_settings

Purpose:

Widget configuration.

Examples:

* Theme
* Colors
* Layout
* Counter Style
* Badge Style

Store configuration as JSONB when appropriate.

---

# 11. control_plane_shadow_state

Purpose:

Optional local snapshot of upstream control-plane fields required for runtime-serving decisions.

Examples:

* upstream project status
* render eligibility flags
* plan-derived limits relevant to runtime output
* synchronized public configuration

Rules:

* This state is derived.
* Shoply remains the source of truth.
* The schema should remain narrow and runtime-oriented.

---

# 12. Soft Delete

Preferred for:

* Projects

Use:

```text
deleted_at
```

instead of immediate deletion.

Hard delete only after retention policies are satisfied.

---

# 13. Index Strategy

Every table should define indexes intentionally.

Examples:

Projects

```text
slug
```

Daily Statistics

```text
(project_id, date)
```

API Keys

```text
key_hash
```

Avoid unnecessary indexes.

Every index increases write cost.

---

# 14. JSONB Usage

Use JSONB only for flexible configuration.

Suitable:

* Widget Config
* Theme Config
* Custom Metadata

Avoid JSONB for:

* Primary relationships
* Frequently filtered fields
* Analytics counters

---

# 15. Migrations

All schema changes must be versioned.

Example:

```text
000001_create_users.sql

000002_create_projects.sql

000003_create_api_keys.sql
```

Never edit historical migrations.

Create new migrations instead.

---

# 16. Transactions

Use transactions only when necessary.

Examples:

* Project creation
* control-plane shadow updates
* aggregate persistence

Do not wrap long-running operations in a transaction.

---

# 17. Constraints

Prefer database constraints.

Examples:

* UNIQUE
* FOREIGN KEY
* CHECK

Business rules should be enforced in both:

* Application
* Database

---

# 18. Naming Convention

Tables:

plural

```text
users

projects

daily_statistics

api_keys
```

Columns:

snake_case

Primary Key:

```text
id
```

Foreign Keys:

```text
project_id

user_id
```

---

# 19. Archiving

Historical statistics older than retention policies may be archived.

Runtime analytics should never depend on archived data.

---

# 20. Future Expansion

Potential future tables:

```text
badges

widgets

timelines

heatmaps

public_dashboards

audit_logs

plugins

project_sync_events

project_limits
```

The existing schema should accommodate these additions without major redesign.

---

# 21. Database Invariants

The following rules are mandatory:

1. PostgreSQL is never the hot path.
2. Runtime analytics always flow through Redis first.
3. Historical tables store aggregated data only.
4. IDs are immutable.
5. API keys are stored as hashes.
6. Migrations are append-only.
7. Shoply remains the source of truth for billing, tenancy, and lifecycle.
8. SVGStat stores only the durable data it needs for runtime and history.
9. Schema evolution must preserve backward compatibility where possible.

These principles ensure the database remains reliable, scalable, and maintainable as SVGStat evolves.
