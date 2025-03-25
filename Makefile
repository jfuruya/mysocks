COMPOSE := docker compose -f docker/dev/docker-compose.yml -p mysocks

.PHONY: up down shell

up:
	$(COMPOSE) up -d

down:
	$(COMPOSE) down

shell:
	$(COMPOSE) exec dev bash
