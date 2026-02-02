# =========================
# Config
# =========================
APP_NAME := wunderlist-api
DOCKER_COMPOSE := docker compose
GO_MAIN := ./cmd/api/main.go

# =========================
# Local Go
# =========================

.PHONY: run
run:
	@echo "▶️ Running API locally"
	go run $(GO_MAIN)

.PHONY: build
build:
	@echo "🔨 Building Go binary"
	go build -o bin/$(APP_NAME) $(GO_MAIN)

.PHONY: test
test:
	@echo "🧪 Running tests"
	go test ./...

.PHONY: test-integration
test-integration:
	@echo "🧪 Running integration tests"
	go test ./tests/integration/...

# =========================
# Docker
# =========================

.PHONY: docker-build
docker-build:
	@echo "🐳 Building Docker images (no cache)"
	$(DOCKER_COMPOSE) build --no-cache

.PHONY: docker-up
docker-up:
	@echo "🚀 Starting containers"
	$(DOCKER_COMPOSE) up -d

.PHONY: docker-up-build
docker-up-build:
	@echo "🚀 Rebuilding + starting containers"
	$(DOCKER_COMPOSE) up -d --build

.PHONY: docker-down
docker-down:
	@echo "🛑 Stopping containers"
	$(DOCKER_COMPOSE) down

.PHONY: docker-logs
docker-logs:
	@echo "📜 Tailing API logs"
	$(DOCKER_COMPOSE) logs -f api

.PHONY: docker-restart
docker-restart:
	@echo "♻️ Restarting API container"
	$(DOCKER_COMPOSE) restart api

# =========================
# Cleanup
# =========================

.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts"
	rm -rf bin

.PHONY: docker-nuke
docker-nuke:
	@echo "💣 Removing containers, images, volumes"
	$(DOCKER_COMPOSE) down -v --rmi all --remove-orphans
