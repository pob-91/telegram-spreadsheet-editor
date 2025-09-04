.PHONY: redis
redis:
	docker compose --profile db up -d

.PHONY trash
trash:
	docker compose down -v --remove-oprhans
