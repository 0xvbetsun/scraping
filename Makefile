.PHONY: help

help:
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.DEFAULT_GOAL := help

lint: ## Runs linters
	go vet ./... && gocyclo -over 25  . && revive ./... && staticcheck ./... && exportloopref ./...

format: ## Runs formatter
	gofumpt -l -w .

test: ## Runs test for app
	go test ./... -race -coverprofile=cover.out -covermode=atomic

cover: ## Gets percents of code coverage
	go tool cover -func cover.out | grep total:
