.PHONY: valkey up trash-valkey trash

valkey:
	docker compose --profile db up -d

up:
	docker compose --profile all up -d

trash-valkey:
	docker compose --profile db down

trash:
	docker compose --profile all down
