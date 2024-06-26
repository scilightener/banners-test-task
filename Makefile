migrate-up:
	@echo "Migrating database up..."
	@CONFIG_PATH=./configs/local.json go run ./cmd/migrator -migrations-path ./migrations -direction up

migrate-down:
	@echo "Migrating database down..."
	@CONFIG_PATH=./configs/local.json go run ./cmd/migrator -migrations-path ./migrations -direction down

run:
	@echo "Running server..."
	@CONFIG_PATH=./configs/local.json go run ./cmd/app

jwt-user:
	@echo "Generating JWT token for user..."
	@CONFIG_PATH=./configs/local.json go run ./cmd/jwt-generator -role user

jwt-admin:
	@echo "Generating JWT token for admin..."
	@CONFIG_PATH=./configs/local.json go run ./cmd/jwt-generator -role admin

lint:
	@echo "Running linter..."
	@golangci-lint run ./... -c .golangci.yml

docker:
	@echo "Running docker-compose..."
	@docker-compose up -d banners-postgres
	@docker-compose up -d banners-api

docker-down:
	@echo "Stopping docker-compose..."
	@docker-compose down --remove-orphans

docker-deps:
	@echo "Running dependencies in docker..."
	@docker-compose -f local.docker-compose.yaml up -d
	@CONFIG_PATH=./configs/local.docker.deps.json go run ./cmd/migrator -migrations-path ./migrations -direction up
	@make run

docker-deps-down:
	@echo "Stopping dependencies in docker..."
	@docker-compose -f local.docker-compose.yaml down --remove-orphans

test: docker-deps
	@echo "Running tests..."
	@go test ./tests
	@make docker-deps-down
