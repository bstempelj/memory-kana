.DEFAULT_GOAL := help
.PHONY: help docker/build docker/run compose/up compose/down lint/compose

help: ## Display all Makefile commands
	@grep -E '^[a-z.A-Z_-]+.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

docker/build: ## Build a docker image
	@docker build -t memory-kana .

docker/tag/%: ## Tag the built docker image
	@docker tag memory-kana blazstempelj/memory-kana:$*

docker/push/%: ## Push the tagged docker image to docker hub
	docker image push blazstempelj/memory-kana:$*
