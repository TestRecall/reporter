SHELL=bash

test:
	go test -race -count=1 -v ./...

.PHONY: setup
setup:
	go install github.com/jstemmer/go-junit-report@latest
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
