# Backend Architecture Overview

## Purpose
This project implements the backend of a Facebook-like social network using a Clean / Hexagonal Architecture-inspired approach. The goal is to balance clarity, maintainability, and simplicity while respecting the project constraints (SQLite, migrations, Docker, WebSockets, sessions).

The backend is designed as a modular monolith: one backend service, one database, but with strong internal boundaries.

## High-Level Architecture
Transport Layer (HTTP / WebSocket)
↓
Application Layer (Use Cases)
↓
Domain Layer (Entities & Rules)
↓
Infrastructure Layer (SQLite, FS)

Dependencies always point inward. Inner layers never depend on outer layers.

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

### 3) Infrastructure Layer
The infrastructure layer provides technical implementations for external systems.

Includes:
- SQLite repositories
- Database connection and migrations
- Filesystem access (images/GIFs)
- Session and cookie handling

Characteristics:
- Implements domain repository interfaces
- Replaceable without changing business logic
- Isolated from application and domain layers

SQLite is used as required by the project.

### 4) Transport Layer
The transport layer exposes the backend to clients.

Includes:
- HTTP handlers (REST endpoints)
- WebSocket handlers (real-time chat)

Responsibilities:
- Parse requests
- Validate input
- Call use cases
- Format responses

This layer contains no business logic and no SQL.

## Dependency Injection
Dependencies are wired manually in the application entry point (`main.go`):
- Database initialization
- Repository construction
- Use case services
- HTTP and WebSocket handlers

This keeps control explicit and avoids hidden framework behavior.

## Database Strategy
- SQLite is the single database
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
- SQLite
- Migrations
- Docker
- Sessions and cookies
- WebSockets
- Real-time chat
- Notifications

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
│   │   ├── sqlite/
│   │   │   ├── sqlite.go        # DB init + migrations
│   │   │   └── repositories/
│   │   │       ├── user.go
│   │   │       ├── post.go
│   │   │       ├── group.go
│   │   │       ├── chat.go
│   │   │       └── notification.go
│   │   │
│   │   └── migrations/
│   │       └── sqlite/
│   │           ├── 000001_create_users.up.sql
│   │           ├── 000001_create_users.down.sql
│   │           ├── ...
│   │
│   ├── auth/
│   │   └── session.go           # Sessions & cookies
│   │
│   └── utils/
│       └── env.go
│
├── data/
│   └── social.db
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
