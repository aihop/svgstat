# MIGRATIONS.md

# SVGStat Migration Strategy

> This document defines how schema changes should be introduced in SVGStat.
>
> Migrations are append-only, reviewable, and aligned with the three-system
> architecture in which Shoply owns control-plane truth and SVGStat owns only
> runtime-serving durable state.

---

# 1. Purpose

Migrations exist to:

* create the initial SVGStat durable schema
* evolve runtime-serving persistence safely
* preserve historical analytics correctness
* keep production upgrades repeatable

Migrations should not be used to smuggle upstream SaaS models into SVGStat.

---

# 2. Migration Principles

Every migration should follow these rules:

* append-only
* deterministic
* reviewable in code review
* safe to run once in production
* consistent with `DATABASE.md` and `SCHEMA.md`

Never edit an already-applied historical migration.

Create a new migration instead.

---

# 3. Naming Convention

Recommended file naming:

```text
000001_create_projects.sql
000002_create_api_keys.sql
000003_create_daily_statistics.sql
000004_create_widget_settings.sql
000005_create_project_limits.sql
000006_create_project_sync_events.sql
```

Rules:

* use zero-padded sequence numbers
* use lowercase snake_case names
* describe the primary intent in the filename

---

# 4. V1 Migration Order

Recommended initial creation order:

1. `projects`
2. `api_keys`
3. `daily_statistics`
4. `widget_settings`
5. `project_limits`
6. `project_sync_events`

Reasoning:

* foreign-key dependencies stay simple
* runtime identity lands first
* historical analytics tables can reference stable project IDs
* control-plane audit tables arrive after the core runtime tables exist

---

# 5. Safe Change Patterns

Prefer these patterns:

* add nullable column
* add column with safe default
* add new table
* add index concurrently when supported and justified
* backfill in controlled steps
* switch reads and writes after data is ready

Avoid risky one-step migrations that:

* rewrite hot large tables blindly
* drop important columns immediately
* merge upstream ownership truth into SVGStat runtime tables

---

# 6. Destructive Changes

Destructive schema changes require extra caution.

Examples:

* dropping a column
* changing data type incompatibly
* replacing a unique key
* renaming a public identifier field

Recommended sequence:

1. add the new shape
2. dual-write or backfill if necessary
3. switch reads
4. verify
5. remove the old shape in a later migration

Do not combine all of this in one irreversible migration.

---

# 7. Data Ownership Guardrail

A critical migration rule:

* do not add `users`, `teams`, `subscriptions`, `orders`, or similar upstream truth tables into SVGStat
* do not model Shoply billing logic inside SVGStat migrations
* only persist the runtime-derived state SVGStat truly needs

If a new field request appears, ask:

```text
Is this needed for runtime serving, historical analytics, or sync auditability?
```

If not, it probably belongs upstream.

---

# 8. Idempotency and Replay

Migration files themselves should be applied once by the migration runner.

Application-level replay concerns belong in tables such as `project_sync_events`, not in repeated mutation of schema history.

For data backfills introduced by migrations:

* make them restart-safe where possible
* log rows affected
* avoid hidden partial state

---

# 9. Index Discipline

Indexes should be added intentionally.

Add an index only when it supports:

* primary lookup paths
* uniqueness guarantees
* common dashboard or control-plane queries
* worker aggregation queries

Remember:

* every index increases write cost
* analytics hot paths still write to Redis first, not PostgreSQL

---

# 10. JSONB Discipline

Use JSONB only where flexibility is valuable.

Good candidates:

* `widget_settings.settings`
* `daily_statistics` distribution payloads
* `project_sync_events.payload`

Avoid JSONB for:

* primary identity
* frequently filtered status flags
* core join relationships

---

# 11. Migration Review Checklist

Before accepting a migration, verify:

* table ownership matches the three-system architecture
* required constraints are present
* unique keys are correct
* foreign keys are justified
* defaults are safe
* backfill plan exists when needed
* no upstream source-of-truth leakage is introduced

If a migration affects public identifiers such as `slug` or `external_project_id`, backward compatibility review is required.

---

# 12. Rollback Philosophy

Prefer roll-forward over ad hoc rollback.

Why:

* production rollback after partial writes is often unsafe
* data shape may already have changed
* forward corrective migrations are easier to reason about

For risky migrations:

* snapshot first when appropriate
* deploy during lower-risk windows
* keep remediation SQL prepared

---

# 13. V1 Migration Invariants

The following rules are mandatory:

1. Migrations are append-only.
2. Runtime tables must not duplicate upstream billing or account ownership truth.
3. Historical analytics remain aggregated, never per-request.
4. Public identifiers and sync identifiers remain stable once introduced.
5. Changes should preserve compatibility with `CONTRACTS.md` and `SCHEMA.md`.

These rules keep SVGStat schema evolution controlled and production-safe.
