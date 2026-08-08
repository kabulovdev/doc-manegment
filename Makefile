.PHONY: up-dev up-prod down logs backend-run backend-build backend-test backend-linux frontend-dev frontend-build frontend-dist fmt help

help:
	@echo "Targets:"
	@echo "  up-dev         — docker compose up with dev profile (backend+frontend+mongo+minio)"
	@echo "  up-prod        — docker compose up without dev-only services"
	@echo "  down           — docker compose down"
	@echo "  logs           — tail docker compose logs"
	@echo "  backend-run    — run Go backend against local env"
	@echo "  backend-build  — go build ./..."
	@echo "  backend-test   — go test ./..."
	@echo "  backend-linux  — cross-compile api+purge binaries for linux/amd64 (VM deploy)"
	@echo "  frontend-dev   — npm run dev"
	@echo "  frontend-build — npm run build"
	@echo "  frontend-dist  — standalone deploy build: make frontend-dist API_BASE=http://IP:8080/api/v1"
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

backend-linux:
	cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/api ./cmd/api
	cd backend && GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bin/purge ./cmd/purge

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-dist:
	@test -n "$(API_BASE)" || (echo "API_BASE berilmadi. Masalan: make frontend-dist API_BASE=http://1.2.3.4:8080/api/v1"; exit 1)
	cd frontend && NEXT_PUBLIC_API_BASE=$(API_BASE) npm run build
	rm -rf frontend-dist
	cp -R frontend/.next/standalone frontend-dist
	cp -R frontend/.next/static frontend-dist/.next/static
	cp -R frontend/public frontend-dist/public

fmt:
	cd backend && gofmt -w .
	cd frontend && npx --yes prettier -w "app/**/*.{ts,tsx}" "components/**/*.{ts,tsx}" "lib/**/*.{ts,tsx}" || true
