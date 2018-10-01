.PHONY: help build clean test lint

help:
	@printf "Command:\n"
	@printf "  build\t\tCompiles source code to binaries\n"
	@printf "  clean\t\tRemove the binaries file\n"
	@printf "  test\t\tTest all packages\n"
	@printf "  lint\t\tLinting all source code\n"

build:
	@mkdir -p bin && go build -i -o ./bin/proxy ./cmd/...

lint-misspell:
	@[ -z "$(shell command -v misspell)" ] && go get -u github.com/client9/misspell/cmd/misspell || true
	@find . -not -path '*/vendor/*' -type f \( -name '*.md' -o -name '*.go' \) -exec misspell -error {} +

lint-golint:
	@[ -z "$(shell command -v golint)" ] && go get -u golang.org/x/lint/golint || true
	@go list ./... | xargs -L1 golint

lint: lint-misspell lint-golint

clean:
	@find "$(PWD)" -type f -name "*.out" -delete
	@rm -rf "$(PWD)/bin"

test: lint
	@go test -covermode=count -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out
