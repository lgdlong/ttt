db-up:
	@docker compose up -d postgres

db-down:
	@docker compose down -v postgres