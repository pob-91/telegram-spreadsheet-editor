.PHONY: redis up trash-redis trash

redis:
	docker compose --profile db up -d

up:
	docker compose --profile all up -d

trash-redis:
	docker compose --profile db down

trash:
	docker compose --profile all down
