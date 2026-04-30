.PHONY: help run build up up-d down restart logs ps db-shell clean

help:
	@echo "Available Docker commands:"
	@echo "  make run       Start all services with logs"
	@echo "  make build     Build docker images"
	@echo "  make up        Start services with logs"
	@echo "  make up-d      Start services in background"
	@echo "  make down      Stop and remove containers"
	@echo "  make restart   Restart all services"
	@echo "  make logs      Follow service logs"
	@echo "  make ps        Show service status"
	@echo "  make db-shell  Open Postgres shell"
	@echo "  make clean     Stop services and remove volumes"

run: up

build:
	docker compose build

up:
	docker compose up --build

up-d:
	docker compose up -d --build

down:
	docker compose down

restart:
	docker compose restart

logs:
	docker compose logs -f

ps:
	docker compose ps

db-shell:
	docker compose exec postgres psql -U agent -d agent_corp

clean:
	docker compose down --volumes --remove-orphans
