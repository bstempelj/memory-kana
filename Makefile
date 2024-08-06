.DEFAULT_GOAL := help

.PHONY: help
help: ## Display all Makefile commands
	@grep -E '^[a-z.A-Z_-]+.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: docker/build docker/tag/% docker/push/%
docker/build: ## Build a docker image
	@docker build -t memory-kana .

docker/tag/%: ## Tag the built docker image
	@docker tag memory-kana blazstempelj/memory-kana:$*

docker/push/%: ## Push the tagged docker image to docker hub
	@docker image push blazstempelj/memory-kana:$*

.PHONY: compose/dev/up compose/dev/down dev compose/prod/up compose/prod/down prod
compose/dev/up:
	@docker compose --profile dev up -d

compose/dev/down:
	@docker compose --profile dev down

dev: compose/dev/up

compose/prod/up:
	@docker compose --profile prod up -d

compose/prod/down:
	@docker compose --profile prod down

prod: compose/prod/up
