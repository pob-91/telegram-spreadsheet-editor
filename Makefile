.PHONY: redis trash

redis:
	docker compose --profile db up -d

trash:
	docker compose --profile db down
