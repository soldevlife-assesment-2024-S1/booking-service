.PHONY: build clean development and deploy

run:
	@echo "Running application..."
	go run cmd/main.go

clean-tools:
	@echo "Cleaning tools..."
	docker compose -f infrastructure-devops/docker-compose.yml down --rmi all

lint:
	@echo "Running lint..."
	golangci-lint run ./internal/...

unit-test:
	@echo "Running tests"
	go test -v -covermode=count ./... -coverprofile=coverage.cov
	go tool cover -func=coverage.cov 

coverage:
	@echo "Running tests with coverage"
	go tool cover -html=./test/coverage/coverage.out


scan:
	@echo "Running scann..."
	gosec ./internal/...

migrate-up:
	@echo "Migrating up..."
	migrate -path database/migration -database $(DB_URL) -verbose up

migrate-down:
	@echo "Migrating down..."
	migrate -path database/migration -database $(DB_URL) -verbose down

migrate-force:
	@echo "Migrating force..."
	migrate -path database/migration -database $(DB_URL) -verbose force $(version)

## postgres://postgres:postgres@100.83.50.92:5432/postgres?sslmode=disable