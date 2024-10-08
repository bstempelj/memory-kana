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

.PHONY: dev dev/down prod prod/down
dev: ## Run compose with dev profile
	@docker compose --profile dev up -d

dev/down: ## Stop compose with dev profile
	@docker compose --profile dev down

prod: ## Run compose with prod profile
	@docker compose --profile prod up -d

prod/down: ## Stop compose with prod profilel
	@docker compose --profile prod down
