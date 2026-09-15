<h1 align="center">KalaSetu</h1>

<p align="center">Helping traditional artists and craftspeople find work, funding, and an audience.</p>

<p align="center">
  <img src="https://img.shields.io/badge/Backend-Go-00ADD8?logo=go&logoColor=white" alt="Go" />
  <img src="https://img.shields.io/badge/Frontend-Kotlin_Multiplatform-7F52FF?logo=kotlin&logoColor=white" alt="Kotlin Multiplatform" />
  <img src="https://img.shields.io/badge/API-Gin-00ADD8" alt="Gin" />
  <img src="https://img.shields.io/badge/Database-PostgreSQL-336791?logo=postgresql&logoColor=white" alt="PostgreSQL" />
  <img src="https://img.shields.io/badge/GraphQL-gqlgen-E10098?logo=graphql&logoColor=white" alt="GraphQL" />
  <img src="https://img.shields.io/badge/License-AGPL--3.0-blue.svg" alt="License" />
</p>

---

# Overview

### Why we're building this

Traditional artists and craftspeople are often deeply skilled, but invisible outside their immediate community. Organizers looking to hire talent for an event, and sponsors looking to fund a local creator, rarely have a reliable place to look — and the artists themselves have no real equivalent of the platforms other industries take for granted, built around how their work actually gets discovered, hired, and paid.

Generic social media wasn't built for this. Getting "likes" on a post doesn't pay rent, and a photo of a finished piece scrolling past in a feed rarely turns into an actual sale or booking.

KalaSetu is our attempt to close that gap: a platform built specifically for the cultural sector, where artists and craftspeople can be discovered, hired, funded, and paid — not just seen.

### Who it's for

- **Artists & craftspeople** — potters, weavers, painters, musicians, and other traditional practitioners looking for paid work and recognition
- **Organizers** — people running cultural events, craft fairs, or exhibitions who need to find and hire talent
- **Sponsors** — individuals or organizations who want to fund creators or cultural events directly
- **Audiences** — people who want to discover cultural events near them and support the artists behind them

---

# Repository Structure

```
kalasetu/
├── kalasetu-backend/     # Go backend
└── kalasetu-frontend/    # Kotlin Multiplatform frontend
```

Each component has its own detailed README: [`kalasetu-backend/`](kalasetu-backend/README.md) and [`kalasetu-frontend/`](kalasetu-frontend/README.md).

---

# Architecture

### How it's built

KalaSetu is a mobile app backed by a Go server, all shared behind one API. If you're curious about the technical side — the architecture, the database, how everything connects — that detail lives in the two component repositories:

- [`kalasetu-backend/`](kalasetu-backend/README.md) — the Go server, database, and API
- [`kalasetu-frontend/`](kalasetu-frontend/README.md) — the Android/iOS app

---

# Current Features

### What you can do on KalaSetu today

**Build a profile.** Artists and craftspeople set up a profile describing their craft, so organizers and sponsors can actually find them — instead of relying on word of mouth or a scattered Instagram page.

**Post your work.** Share photos or videos of a finished piece, an event, or work in progress, and get direct engagement from people who care about the craft.

**Create and discover events.** An organizer can list a craft fair or cultural event with a banner image, and audiences can browse what's happening near them.

**Apply for opportunities.** Artists can apply directly to listed opportunities and track where their application stands, rather than sending a message into the void.

---

# Technology Stack

Go + Gin for the server, GraphQL (gqlgen) for the API, PostgreSQL for data, S3 object storage for media, and a Kotlin Multiplatform app. Full details in [`kalasetu-backend/README.md`](kalasetu-backend/README.md).

---

# Getting Started

Backend setup (Docker, environment variables, the GraphQL playground at <http://localhost:8080/>) is covered in [`kalasetu-backend/README.md`](kalasetu-backend/README.md). Frontend build and run steps are in [`kalasetu-frontend/README.md`](kalasetu-frontend/README.md).

---

# Development

For a developer tour of the backend directory layout, database migrations, and regenerating the GraphQL code, see [`kalasetu-backend/README.md`](kalasetu-backend/README.md).

---

# Contributing

Contributions are welcome.

1. Fork the repository.
2. Create a feature branch.
3. Commit changes with clear, descriptive messages (e.g. `feat: add onboarding endpoint`).
4. Open a pull request describing the change and its motivation.

---

# License

KalaSetu is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**. See the `LICENSE` file for details.