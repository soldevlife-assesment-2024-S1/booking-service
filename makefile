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

DB_HOST=soldevlife-booking-service2024053106135688160000000a.cgyygeag3oal.ap-southeast-1.rds.amazonaws.com
DB_PORT=5432
DB_USER=soldevlife
DB_PASSWORD=soldevlife
DB_NAME=postgres
DB_SSL=disable

DB_URL=postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSL)

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