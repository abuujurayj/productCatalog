# Product Catalog Service

A high-performance, domain-driven microservice for managing product catalogs and pricing, built with **Go**, **gRPC**, and **Google Cloud Spanner**.

## 🏗 Architectural Overview

This service follows **Clean Architecture** and **Domain-Driven Design (DDD)** principles. It implements the **Golden Mutation Pattern** and the **Transactional Outbox Pattern** to ensure strict data consistency and reliable event publishing.

### Key Design Decisions

- **Domain Purity**: The `internal/app/product/domain` layer has zero external dependencies (no Spanner, no Proto, no Context).
- **Precision Arithmetic**: All financial calculations use `math/big.Rat` to prevent floating-point errors.
- **Atomic Mutations**: Using `CommitPlan` to batch business logic changes and outbox events into a single Spanner transaction.
- **CQRS**: Command logic is encapsulated in the domain aggregate, while Queries (Read Models) are optimized for performance.

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- `protoc` (for generating gRPC code)

### 1. Start Infrastructure

Run the Google Cloud Spanner emulator:

```bash
docker-compose up -d

```

### 2. Run Migrations

(Ensure you have the Spanner CLI or a migration tool configured to point to `localhost:9010`)

```bash
make migrate

```

### 3. Run Tests (E2E)

The tests verify the full flow: Aggregate Logic -> Mutation Generation -> Spanner Persistence -> Outbox Event Creation.

```bash
go test ./tests/e2e/... -v

```

### 4. Start the gRPC Server

```bash
go run cmd/server/main.go

```

---

## 🛠 Project Structure

- `internal/app/product/domain`: Pure business logic, Aggregates, and Value Objects.
- `internal/app/product/usecases`: Application logic (Interactors) that orchestrate the Golden Mutation Pattern.
- `internal/app/product/repo`: Spanner-specific implementations that return Mutations.
- `pkg/committer`: Wrapper for the atomic `CommitPlan` execution.
- `proto/`: gRPC API definitions.

---

## 📡 API Endpoints (gRPC)

| Method          | Description                                               |
| --------------- | --------------------------------------------------------- |
| `CreateProduct` | Initializes a new product in the catalog.                 |
| `ApplyDiscount` | Applies a percentage-based discount with date validation. |
| `ListProducts`  | Paginated query for active products.                      |
| `GetProduct`    | Fetches a product with its current effective price.       |

---

## 📝 License

Proprietary / Test Task

```

```
