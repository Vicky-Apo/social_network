# Backend Architecture Overview

## Purpose
This project implements the backend of a Facebook-like social network using a Clean / Hexagonal Architecture-inspired approach. The goal is to balance clarity, maintainability, and simplicity while respecting the project constraints (Postgres, migrations, Docker, WebSockets, sessions).

The backend is designed as a modular monolith: one backend service, one database, but with strong internal boundaries.

## High-Level Architecture
Transport Layer (REST / GraphQL / WebSocket)
↓
Application Layer (Use Cases)
↓
Domain Layer (Entities & Rules)
↓
Infrastructure Layer (Postgres, FS)

Dependencies always point inward. Inner layers never depend on outer layers.
Strict rule: Domain and Application must not import DB, HTTP, WebSocket, or framework packages.

## Architectural Style
- Modular monolith
- Clean Architecture principles
- Hexagonal (Ports & Adapters) concepts
- Manual dependency injection
- No frameworks forcing structure

This avoids unnecessary microservices while keeping the codebase professional and scalable.

## Layer Responsibilities

### 1) Domain Layer
The domain layer contains the core business model of the social network.

Includes:
- Entities (users, posts, groups, messages, notifications)
- Value objects
- Business rules
- Repository interfaces (ports)

Characteristics:
- Pure Go code
- No SQL
- No HTTP
- No WebSocket logic
- No external dependencies

This layer represents the business truth of the application.

### 2) Application Layer (Use Cases)
The application layer implements use cases, such as:
- User registration and authentication
- Follow / unfollow logic
- Post creation with privacy rules
- Group management
- Chat and messaging
- Notification generation

Responsibilities:
- Orchestrate domain entities
- Enforce business rules
- Call repository interfaces
- Remain independent of transport and storage

All business logic lives here.

Use-case boundaries and DTO flow:
- Each use case exposes a single entry point per action (e.g., `CreatePost`, `SendMessage`).
- Inputs are request DTOs defined in the application layer (no HTTP/GraphQL types).
- Outputs are response DTOs defined in the application layer (no DB models).
- Transport maps incoming payloads → request DTOs, calls use case, maps response DTOs → transport response.

### 3) Infrastructure Layer
The infrastructure layer provides technical implementations for external systems.

Includes:
- Postgres repositories
- Database connection and migrations
- Filesystem access (images/GIFs)
- Session and cookie handling

Characteristics:
- Implements domain repository interfaces
- Replaceable without changing business logic
- Isolated from application and domain layers

Postgres is used as required by the project.

### 4) Transport Layer
The transport layer exposes the backend to clients.

Includes:
- REST handlers
- GraphQL handlers
- WebSocket handlers (real-time chat)

Responsibilities:
- Parse requests
- Validate input
- Call use cases
- Format responses

This layer contains no business logic and no SQL.

Transport responsibilities and when to use each:
- REST: commands and simple reads (auth, create/update/delete, small lookups).
- GraphQL: complex reads (feeds, profiles, dashboards) with client-defined shapes.
- WebSocket: real-time events only (chat, notifications, presence, typing).

Strict rule: Transport must never call the database directly. It only calls use cases.

## Dependency Injection
Dependencies are wired manually in the application entry point (`main.go`):
- Database initialization
- Repository construction
- Use case services
- HTTP and WebSocket handlers

This keeps control explicit and avoids hidden framework behavior.

## Database Strategy
- Postgres is the single database
- Schema is designed first (ERD)
- Versioned migrations manage schema evolution
- Migrations run automatically on server startup
- Foreign keys and indexes are explicitly defined

Database logic is fully isolated from business logic.

## Deployment Model
- One backend container
- One frontend container
- No microservices
- No API gateway

This matches the project requirements and avoids unnecessary complexity.

## Benefits of This Architecture
### Clear Separation of Concerns
Each layer has a single responsibility, preventing tight coupling.

### Maintainability
Features can be added or modified without breaking unrelated parts.

### Testability
Business logic can be tested independently of the database and transport layers.

### Scalability of Features
The codebase grows in features without becoming unmanageable.

### Simplicity in Deployment
Despite internal modularity, the system remains easy to build and run.

### Alignment with Project Requirements
The architecture fully supports:
- Postgres
- Migrations
- Docker
- Sessions and cookies
- WebSockets
- Real-time chat
- Notifications

## Implementation Guidance (practical rules)
- Domain never imports `net/http`, GraphQL packages, WebSocket packages, or database drivers.
- Application never imports transport or database packages. It depends on interfaces only.
- Infrastructure implements interfaces defined by the domain/application layers.
- Transport is thin: map input → DTO, call use case, map output → response.
- Keep DTOs in the application layer; do not reuse DB models or transport payloads.
- If a rule is broken, fix the layering instead of adding shortcuts.

## REST vs GraphQL vs WebSocket (usage guide)
- REST endpoints should map 1:1 to use cases (commands and simple reads).
- GraphQL resolvers should call use cases; resolvers must not execute raw SQL.
- WebSocket handlers decode messages and call use cases; broadcasting is infrastructure.

## Example Flow (Create Post)
1) REST handler receives JSON.
2) Maps JSON to `CreatePostRequest`.
3) Calls `CreatePost` use case.
4) Use case validates rules, calls repository interface.
5) Repository implementation writes to Postgres.
6) Use case returns `CreatePostResponse`.
7) Handler maps response to JSON.

## Final Statement
This architecture provides a professional-grade structure while remaining appropriate for the scope of the project. It avoids premature optimization and unnecessary microservices, focusing instead on correctness, clarity, and long-term maintainability.

The backend remains:
- Easy to reason about
- Easy to extend
- Easy to test
- Easy to deploy




backend/
├── cmd/
│   ├── server/
│   │   └── main.go              # Application entry point
│   │
│   └── migrateversion/
│       └── main.go              # Optional migration inspection tool
│
├── internal/                    # Application core (not importable outside)
│   ├── domain/                  # Business entities + ports
│   │   ├── user/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   ├── post/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   ├── group/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   ├── chat/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   ├── notification/
│   │   │   ├── entity.go
│   │   │   └── repository.go
│   │   └── auth/
│   │       └── entity.go
│   │
│   ├── usecase/                 # Application logic (services)
│   │   ├── auth/
│   │   │   └── service.go
│   │   ├── user/
│   │   │   └── service.go
│   │   ├── post/
│   │   │   └── service.go
│   │   ├── group/
│   │   │   └── service.go
│   │   ├── chat/
│   │   │   └── service.go
│   │   └── notification/
│   │       └── service.go
│   │
│   ├── transport/               # Delivery layer
│   │   ├── http/
│   │   │   ├── handler/
│   │   │   │   ├── auth.go
│   │   │   │   ├── user.go
│   │   │   │   ├── post.go
│   │   │   │   ├── group.go
│   │   │   │   └── notification.go
│   │   │   ├── middleware/
│   │   │   │   ├── auth.go
│   │   │   │   └── logging.go
│   │   │   └── router.go
│   │   │
│   │   └── websocket/
│   │       ├── hub.go
│   │       └── handler.go
│   │
│   └── config/
│       └── config.go            # Config & env loading
│
├── pkg/                         # Infrastructure & shared adapters
│   ├── db/
│   │   ├── postgres/
│   │   │   ├── postgres.go      # DB init + migrations
│   │   │   └── repositories/
│   │   │       ├── user.go
│   │   │       ├── post.go
│   │   │       ├── group.go
│   │   │       ├── chat.go
│   │   │       └── notification.go
│   │   │
│   │   └── migrations/
│   │       └── postgres/
│   │           ├── 000001_create_users_table.up.sql
│   │           ├── 000001_create_users_table.down.sql
│   │           ├── ...
│   │
│   ├── auth/
│   │   └── session.go           # Sessions & cookies
│   │
│   └── utils/
│       └── env.go
│
├── docs/
│   ├── migrations.md
│   └── architecture.md
│
├── Dockerfile
├── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
