# SVGStat

> Developer-first SVG Analytics Platform.

SVGStat is a high-performance analytics platform built around **dynamic SVG rendering**.

It enables developers to embed real-time counters, badges, widgets, charts and analytics into GitHub README files, Markdown documents and static websites—without requiring JavaScript.

---

## Why SVGStat?

Most analytics solutions are designed for websites.

SVGStat is designed for developers.

Instead of injecting JavaScript, SVGStat delivers analytics through SVG images that can be embedded almost anywhere.

Supported scenarios include:

* GitHub README
* Markdown
* Documentation
* Static websites
* Developer dashboards
* Open source projects

SVGStat focuses on making analytics lightweight, fast and developer-friendly.

---

# Features

## Dynamic SVG Counter

Display real-time counters.

Examples:

* Visitors
* Downloads
* Stars
* Followers
* Custom counters

---

## README Analytics

Track README traffic without JavaScript.

Metrics include:

* Page Views
* Unique Visitors
* Referrer
* Country
* Device
* Browser

---

## SVG Badge

Generate dynamic badges.

Examples:

* Visitors
* Downloads
* Version
* Build Status
* Custom Metrics

---

## SVG Widgets

Render rich widgets inside Markdown.

Examples:

* Statistics
* Charts
* Heatmaps
* Timelines
* Project summaries

---

## Comment System (Planned)

GitHub-native comments rendered as SVG.

Designed to work inside Markdown.

---

## Dashboard

Modern analytics dashboard including:

* Today
* Yesterday
* Last 7 Days
* Last 30 Days
* Referrer
* Country
* Browser
* Device
* Trends

---

# Architecture

SVGStat operates inside a three-repository SaaS architecture.

```text
Developers / Browsers / README Embeds
                 │
                 ▼
           Cloudflare CDN
                 │
                 ▼
           SVGStat Go Engine
        ┌────────┼────────┐
        ▼        ▼        ▼
   Renderer  Analytics   API
        │        │        │
        └────┬───┴────────┘
             ▼
           Redis
             │
             ▼
           Worker
             │
             ▼
        PostgreSQL

APayShop Official Site / User Portal
             │
             ▼
 Shoply SaaS Base / Billing / Project Lifecycle
```

Responsibilities are strictly separated.

## Go Service

Responsible for:

* SVG Rendering
* Analytics
* Counters
* Widgets
* Badge Generation
* Event Collection
* API

---

## APayShop

Responsible for:

* Official Website
* Pricing Portal
* User Account Center
* Public Product Entry
* Purchase Entry Flow

APayShop does not render production SVG URLs.

---

## Shoply

Responsible for:

* Authentication
* OAuth
* Billing
* Subscription
* Dashboard
* User Management
* Project Management
* API Keys

Shoply manages lifecycle and control-plane state.

Shoply never renders SVG in the hot path.

---

# Design Philosophy

SVGStat follows several principles.

## Performance First

High-frequency requests should avoid database access.

Redis is always the first write target.

---

## API First

Every feature must be exposed through APIs.

The dashboard is only a consumer of APIs.

---

## Stateless Rendering

Every SVG request should be stateless.

Horizontal scaling should be effortless.

---

## Renderer & Analytics Separation

Rendering and analytics are independent modules.

Neither should depend on the other.

---

## Cache First

Cache hierarchy:

```text
Memory Cache

↓

Redis

↓

PostgreSQL
```

---

# Performance Goals

SVG Rendering

* P95 < 20 ms
* P99 < 50 ms

Analytics

* Redis-first
* Worker aggregation
* Zero database writes on hot paths

---

# Project Structure

```text
cmd/
    api/

internal/
    analytics/
    renderer/
    counter/
    badge/
    widget/
    cache/
    middleware/
    worker/
    auth/
    config/
    metrics/
    project/

pkg/

templates/

configs/

scripts/

deploy/

docs/
```

---

# Tech Stack

Backend

* Go
* PostgreSQL
* Redis

Platform

* Shoply SaaS

Infrastructure

* Cloudflare
* Docker
* Docker Compose

Future

* Kubernetes
* Multi-region Deployment

---

# Roadmap

## V1

* Dynamic Counter
* Badge
* Project Dashboard
* Analytics
* README Statistics

---

## V2

* Widgets
* Heatmaps
* Charts
* Timeline
* Team Support

---

## V3

* SVG Comment
* Public Dashboard
* Team Collaboration
* Marketplace

---

# Self Hosting

Coming Soon.

---

# Documentation

* AGENT.md
* ARCHITECTURE.md
* API.md
* CONTRACTS.md
* INTEGRATION.md
* CONTRIBUTING.md
* DATABASE.md
* SCHEMA.md
* MIGRATIONS.md
* SQL_DRAFTS.md
* REDIS.md
* WORKER.md
* RENDERER.md
* ANALYTICS.md

---

# Contributing

Contributions are welcome.

Before contributing, please read:

* CONTRIBUTING.md
* AGENT.md

---

# License

MIT License.

---

# Vision

SVGStat is not simply a visitor counter.

It aims to become the infrastructure layer for developer-facing SVG analytics.

Just as Prometheus became the standard for metrics collection and Grafana became the standard for visualization, SVGStat aims to become the standard platform for analytics embedded in SVG.

Our mission is simple:

> Make analytics as easy to embed as an image.
