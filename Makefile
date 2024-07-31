.DEFAULT_GOAL := help
.PHONY: help docker/build docker/run compose/up compose/down lint/compose

help: ## Display all Makefile commands
	@grep -E '^[a-z.A-Z_-]+.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

docker/build: ## Build a docker image
	@docker build -t memory-kana .

docker/run: ## Run the built docker image
	@docker run -p 1234:1234 --rm --name memory-kana memory-kana

compose/up: ## Run and build docker compose in background
	@docker compose up --build -d

compose/down: ## Stop docker compose
	@docker compose down

lint/compose: ## Lint compose yaml file
	@yamllint docker-compose.yml
