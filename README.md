# Product Catalog Service

A clean architecture implementation of a product catalog service using **Go**, **gRPC**, and **Google Cloud Spanner**. This service is designed for high-precision pricing and reliable state transitions.

## 🏗 Architecture & Design Patterns

This service is built using **Domain-Driven Design (DDD)** and **Clean Architecture** to ensure maintainability and testability.

### Core Patterns

- **Domain Purity (CRITICAL)**: The core domain layer has zero external dependencies. It contains pure Go business logic, protecting it from infrastructure changes.
- **Golden Mutation Pattern**: Every write operation follows a strict atomic flow: `Load Aggregate` -> `Mutate State` -> `Collect Mutations` -> `Apply Plan`.
- **CQRS**: Commands (writes) go through the domain aggregate and repositories, while Queries (reads) use optimized read models to bypass aggregate overhead.
- **Transactional Outbox**: Business state changes and event intents are persisted in a single Spanner transaction, ensuring 100% reliable event publishing.
- **High-Precision Pricing**: All monetary calculations use `math/big.Rat` (stored as numerator/denominator) to ensure zero rounding errors.

---

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- `protoc` (Protocol Buffers compiler)

### 1. Environment Setup

The service is designed to run against the Google Cloud Spanner emulator for local development.

```bash
# Start the Spanner emulator
docker-compose up -d

# Initialize the emulator (create instance/database)
make init-emulator

```

### 2. Database Migrations

Run the schema DDL against the emulator to create the `products` and `outbox_events` tables.

```bash
make migrate

```

### 3. Build & Run

```bash
# Generate gRPC code from proto
make proto

# Run the gRPC server
go run cmd/server/main.go

```

---

## 🧪 Testing

The service includes a comprehensive E2E test suite that verifies the entire flow from usecase to the Spanner emulator, including business rule validation (e.g., "Cannot discount an inactive product").

```bash
# Run all tests (Unit + E2E)
go test -v ./...

```

---

## 📁 Project Structure

```text
product-catalog-service/
├── cmd/server/             # Main entry point
├── internal/
│   ├── app/product/
│   │   ├── domain/         # Pure domain aggregates & business logic
│   │   ├── usecases/       # Interactors (Golden Mutation Pattern)
│   │   ├── queries/        # Optimized read models (CQRS)
│   │   └── contracts/      # Interfaces for repositories & committer
│   ├── pkg/
│   │   ├── committer/      # Atomic CommitPlan wrapper
│   │   └── clock/          # Time abstractions for testability
│   ├── repo/               # Spanner repository implementations
│   └── transport/grpc/     # Thin gRPC handlers & error mapping
├── proto/                  # Protocol Buffer definitions
├── migrations/             # Spanner DDL scripts
└── tests/e2e/              # Integration tests against emulator

```

---

## 📡 API Endpoints (gRPC)

### Commands

- `CreateProduct`: Creates a new product (Default: Inactive).
- `UpdateProduct`: Updates basic product details.
- `ActivateProduct`: Enables the product for sale and discounts.
- `ApplyDiscount`: Attaches a percentage-based discount with start/end dates.

### Queries

- `GetProduct`: Retrieves a product with its **Effective Price** (Base - Active Discount).
- `ListProducts`: Paginated listing with category filtering.

---

## ⚖️ Trade-offs & Decisions

- **No Background Processor**: As per requirements, outbox events are stored but not dispatched. In a production system, a separate worker would poll or stream the `outbox_events` table.
- **Change Tracking**: The aggregate uses a `ChangeTracker` to ensure `UpdateMut` only touches modified columns, reducing Spanner write load.
- **Graceful Shutdown**: The server handles `SIGTERM` to allow active Spanner sessions and gRPC calls to finish cleanly.

```

```
