.PHONY: proto
proto:
	@echo "Cleaning old gen files..."
	rm -rf gen/go/product/v1
	@echo "Creating directory..."
	mkdir -p gen/go/product/v1
	@echo "Generating gRPC code..."
	protoc --proto_path=proto/product/v1 \
		--go_out=gen/go/product/v1 --go_opt=paths=source_relative \
		--go-grpc_out=gen/go/product/v1 --go-grpc_opt=paths=source_relative \
		proto/product/v1/*.proto

.PHONY: migrate
migrate:
	@echo "Running Spanner migrations..."
	# Ensure the emulator host is set for the migration tool
	export SPANNER_EMULATOR_HOST=localhost:9010 && go run cmd/migrate/main.go

.PHONY: test
test:
	@echo "Running E2E and Unit tests..."
	# The tests need to know where the emulator is
	export SPANNER_EMULATOR_HOST=localhost:9010 && go test ./... -v

.PHONY: run
run:
	@echo "Starting gRPC server..."
	export SPANNER_EMULATOR_HOST=localhost:9010 && go run cmd/server/main.go

.PHONY: emulator
emulator:
	@echo "Starting Spanner emulator container..."
	docker compose up -d

.PHONY: emulator-stop
emulator-stop:
	docker compose down

.PHONY: init-emulator
init-emulator:
	@echo "Initializing Spanner instance and database via Docker..."
	@# This uses the official SDK image to configure the emulator running on your host
	docker run --rm --network host google/cloud-sdk:slim /bin/bash -c "\
		gcloud config configurations create emulator || true && \
		gcloud config set auth/disable_credentials true && \
		gcloud config set project test-project && \
		gcloud config set api_endpoint_overrides/spanner 'http://localhost:9020/' && \
		gcloud spanner instances create test-instance --config=emulator-config --description='Test Instance' --nodes=1 || true && \
		gcloud spanner databases create test-database --instance=test-instance || true"
	@echo "Spanner Emulator is ready."