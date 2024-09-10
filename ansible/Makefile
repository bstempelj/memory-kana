.DEFAULT_GOAL := help

.PHONY: help
help: ## Display all Makefile commands
	@grep -E '^[a-z.A-Z_-]+.*?## .*$$' $(MAKEFILE_LIST) \
	| awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

sshkey: ## Create ssh key for ansible user
	@ssh-keygen -t ed25519 -f ansible-sshkey -C ansible

user: ## Create ansible user
	@ansible-playbook playbooks/01-create-user.yaml

project: ## Setup project
	@ansible-playbook playbooks/02-setup-project.yaml --private-key=ansible-sshkey --ask-become-pass
