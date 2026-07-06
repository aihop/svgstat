# SCHEMA.md

# SVGStat V1 Schema Draft

> This document defines the first concrete database schema draft for SVGStat.
>
> It is intentionally narrower than a full SaaS platform schema because billing,
> tenant ownership, and account truth live upstream in Shoply and APayShop.
>
> For migration sequencing and SQL examples, see `MIGRATIONS.md` and `SQL_DRAFTS.md`.

---

# 1. Scope

This schema draft covers the minimum durable tables needed for:

* runtime-serving project state
* runtime-facing API credentials
* historical analytics
* widget configuration
* synchronized plan and render limits
* control-plane sync auditability

It does not define:

* users
* teams
* billing ledgers
* payment orders
* website-facing sessions

Those belong to upstream systems.

---

# 2. Table Set

V1 recommended tables:

* `projects`
* `api_keys`
* `daily_statistics`
* `widget_settings`
* `project_limits`
* `project_sync_events`

These six tables are enough to support the first production loop:

* Shoply syncs project state
* SVGStat serves runtime URLs
* Redis collects hot analytics
* workers persist daily aggregates

---

# 3. Design Notes

Design goals:

* keep runtime-serving fields close to `projects`
* isolate optional or frequently changing policy into dedicated tables
* keep historical analytics append-oriented
* make upstream sync auditable and retry-safe

Recommended ID style:

* string primary keys for public-facing entities when cross-system references matter
* `external_project_id` to link back to Shoply
* `tenant_id` as upstream foreign identity, not a local ownership truth

---

# 4. projects

Purpose:

* local durable runtime representation of a Shoply-managed project

Suggested columns:

```text
id                    text primary key
external_project_id   text not null unique
tenant_id             text not null
slug                  text not null unique
name                  text not null
description           text null
status                text not null
visibility            text not null default 'public'
public_token_hash     text null
default_theme         text null
default_locale        text null
render_enabled        boolean not null default true
badge_enabled         boolean not null default true
widget_enabled        boolean not null default true
chart_enabled         boolean not null default false
last_synced_at        timestamptz null
deleted_at            timestamptz null
created_at            timestamptz not null
updated_at            timestamptz not null
```

Suggested indexes:

```text
unique(slug)
unique(external_project_id)
index(tenant_id)
index(status)
index(last_synced_at)
```

Suggested checks:

```text
status in ('pending', 'active', 'grace', 'disabled', 'archived')
visibility in ('public', 'private')
```

Notes:

* `public_token_hash` stores the hashed runtime token, not plaintext
* `render_enabled` and sibling flags are runtime switches derived from control-plane sync
* `deleted_at` supports soft deletion without breaking historical analytics references

---

# 5. api_keys

Purpose:

* store runtime-facing API credentials used by SVGStat endpoints or internal consumers

Suggested columns:

```text
id                text primary key
project_id        text not null
name              text not null
key_hash          text not null unique
scope             text not null default 'public_runtime'
status            text not null default 'active'
expires_at        timestamptz null
last_used_at      timestamptz null
rotated_from_id   text null
created_at        timestamptz not null
updated_at        timestamptz not null
```

Suggested indexes:

```text
unique(key_hash)
index(project_id)
index(status)
index(expires_at)
```

Suggested checks:

```text
status in ('active', 'revoked', 'expired')
```

Foreign keys:

```text
project_id -> projects.id
rotated_from_id -> api_keys.id
```

Notes:

* plaintext keys should never be stored
* `scope` leaves room for future separation between public runtime, dashboard, and internal credentials

---

# 6. daily_statistics

Purpose:

* persist daily aggregated analytics produced by workers

Suggested columns:

```text
id                text primary key
project_id        text not null
date              date not null
pv                bigint not null default 0
uv                bigint not null default 0
requests          bigint not null default 0
bots              bigint not null default 0
referrers         jsonb not null default '{}'
countries         jsonb not null default '{}'
devices           jsonb not null default '{}'
browsers          jsonb not null default '{}'
created_at        timestamptz not null
updated_at        timestamptz not null
```

Suggested indexes:

```text
unique(project_id, date)
index(date)
```

Foreign keys:

```text
project_id -> projects.id
```

Notes:

* distributions are stored as aggregated JSON objects, not per-request rows
* workers should upsert on `(project_id, date)` for idempotent aggregation

---

# 7. widget_settings

Purpose:

* persist widget-specific presentation configuration per project

Suggested columns:

```text
id                text primary key
project_id        text not null
widget_key        text not null
status            text not null default 'active'
settings          jsonb not null default '{}'
version           integer not null default 1
created_at        timestamptz not null
updated_at        timestamptz not null
```

Suggested indexes:

```text
unique(project_id, widget_key)
index(status)
```

Suggested checks:

```text
status in ('active', 'disabled')
version >= 1
```

Foreign keys:

```text
project_id -> projects.id
```

Notes:

* `widget_key` identifies a logical widget such as `overview`, `timeline`, or `top-referrers`
* the actual settings payload should remain presentation-focused

---

# 8. project_limits

Purpose:

* store plan-derived runtime limits and policy knobs synchronized from Shoply

Suggested columns:

```text
id                        text primary key
project_id                text not null unique
plan_code                 text not null
max_requests_per_minute   integer null
cache_ttl_seconds         integer null
badge_enabled             boolean not null default true
widget_enabled            boolean not null default true
chart_enabled             boolean not null default false
render_enabled            boolean not null default true
effective_from            timestamptz null
effective_until           timestamptz null
created_at                timestamptz not null
updated_at                timestamptz not null
```

Suggested indexes:

```text
unique(project_id)
index(plan_code)
index(effective_until)
```

Foreign keys:

```text
project_id -> projects.id
```

Notes:

* this table keeps fast-changing commercial entitlements out of the core `projects` identity record
* if the product later prefers fewer tables, some boolean flags can be folded back into `projects`, but V1 keeps the policy surface explicit

---

# 9. project_sync_events

Purpose:

* record upstream sync attempts from Shoply for auditability, replay, and idempotency

Suggested columns:

```text
id                  text primary key
event_id            text not null unique
project_id          text null
external_project_id text null
tenant_id           text null
sync_type           text not null
payload             jsonb not null
status              text not null
error_code          text null
error_message       text null
processed_at        timestamptz null
created_at          timestamptz not null
updated_at          timestamptz not null
```

Suggested indexes:

```text
unique(event_id)
index(sync_type)
index(status)
index(external_project_id)
index(created_at)
```

Suggested checks:

```text
status in ('received', 'processed', 'rejected', 'failed')
sync_type in ('project.upsert', 'project.disable', 'project.rotateKey', 'project.refreshCache')
```

Foreign keys:

```text
project_id -> projects.id
```

Notes:

* `payload` is the raw synchronized body used for replay and debugging
* `event_id` is the idempotency anchor

---

# 10. Recommended Creation Order

Suggested migration order:

```text
000001_create_projects.sql
000002_create_api_keys.sql
000003_create_daily_statistics.sql
000004_create_widget_settings.sql
000005_create_project_limits.sql
000006_create_project_sync_events.sql
```

This order keeps foreign-key dependencies straightforward.

---

# 11. Suggested Runtime Mapping

Recommended mapping from contracts to tables:

* Shoply `project.upsert` -> `projects`, `project_limits`
* Shoply `project.disable` -> `projects.status`, `project_limits.render_enabled`
* Shoply `project.rotateKey` -> `api_keys`
* Widget config sync -> `widget_settings`
* Daily worker aggregation -> `daily_statistics`
* Every sync request -> `project_sync_events`

This gives every upstream operation a clear persistence destination.

---

# 12. Non-Goals

Do not add these tables to SVGStat V1 unless runtime truly requires them:

* `users`
* `teams`
* `subscriptions`
* `orders`
* `invoices`
* `payments`

Those models belong to upstream systems and would blur ownership boundaries.

---

# 13. Schema Invariants

The following rules are mandatory:

1. `projects` represent runtime-serving identity, not SaaS ownership truth.
2. `api_keys` store hashes only.
3. `daily_statistics` store aggregated daily facts only.
4. `project_limits` remain synchronized from Shoply-derived policy.
5. `project_sync_events` preserve idempotency and replayability.
6. Cross-system identifiers such as `external_project_id` remain stable once assigned.
7. Schema growth should favor runtime needs, not upstream duplication.

These rules keep the first SVGStat schema minimal, durable, and compatible with the three-system architecture.
