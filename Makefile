.PHONY: up-dev up-prod down logs backend-run backend-build backend-test frontend-dev frontend-build fmt help

help:
	@echo "Targets:"
	@echo "  up-dev         — docker compose up with dev profile (backend+frontend+mongo+minio)"
	@echo "  up-prod        — docker compose up without dev-only services"
	@echo "  down           — docker compose down"
	@echo "  logs           — tail docker compose logs"
	@echo "  backend-run    — run Go backend against local env"
	@echo "  backend-build  — go build ./..."
	@echo "  backend-test   — go test ./..."
	@echo "  frontend-dev   — npm run dev"
	@echo "  frontend-build — npm run build"
	@echo "  fmt            — gofmt backend + prettier frontend (best effort)"

up-dev:
	docker compose --profile dev up -d --build

up-prod:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

backend-run:
	cd backend && go run ./cmd/api

backend-build:
	cd backend && go build ./...

backend-test:
	cd backend && go test ./...

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

fmt:
	cd backend && gofmt -w .
	cd frontend && npx --yes prettier -w "app/**/*.{ts,tsx}" "components/**/*.{ts,tsx}" "lib/**/*.{ts,tsx}" || true
