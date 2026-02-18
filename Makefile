.PHONY: proto
proto:
	@echo "Generating gRPC code..."
	protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		proto/product/v1/*.proto

.PHONY: migrate
migrate:
	@echo "Running Spanner migrations..."
	go run cmd/migrate/main.go

.PHONY: test
test:
	@echo "Running E2E and Unit tests..."
	go test ./... -v

.PHONY: run
run:
	@echo "Starting gRPC server..."
	go run cmd/server/main.go

.PHONY: emulator
emulator:
	@echo "Starting Spanner emulator container..."
	docker-compose up -d

.PHONY: emulator-stop
emulator-stop:
	docker-compose down

.PHONY: init-emulator
init-emulator:
	@echo "Waiting for emulator to start..."
	@sleep 5
	@# We use your logic here but add --project to avoid global config issues
	export SPANNER_EMULATOR_HOST=localhost:9010 && \
	gcloud config configurations create emulator || true && \
	gcloud config set auth/disable_credentials true && \
	gcloud config set project test-project && \
	gcloud config set api_endpoint_overrides/spanner "http://localhost:9020/" && \
	gcloud spanner instances create test-instance \
		--config=emulator-config \
		--description="Test Instance" \
		--nodes=1 || true && \
	gcloud spanner databases create test-database --instance=test-instance || true
	@echo "Spanner Emulator is ready."