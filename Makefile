SHELL=bash

test:
	go test -race ./...

.PHONY: setup
setup:
	go install github.com/jstemmer/go-junit-report/v2@latest
	go install github.com/goreleaser/goreleaser@latest
	go mod tidy

.PHONY: release
release: ## example: make release V=0.0.0
	@read -p "Press enter to confirm and push to origin ..."
	git tag v$(V)
	git push origin v$(V)

.PHONY: build
build:
	goreleaser --snapshot --skip-publish --rm-dist
