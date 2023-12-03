.PHONY: generate test lint

generate:
	@go generate ./...

test:
	@TODOIST__TOKEN=TEST go test -v ./...

lint:
	golangci-lint run ./...
