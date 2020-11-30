test:
	go test -race -count=1 -v ./...

.PHONY: setup
setup:
	go get -u github.com/jstemmer/go-junit-report
	go mod download

.PHONY: release
release: ## example: make release V=0.0.0
	@read -p "Press enter to confirm and push to origin ..."
	git tag v$(V)
	git push origin v$(V)

.PHONY: build
build:
	goreleaser --snapshot --skip-publish --rm-dist
