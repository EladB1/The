.PHONY: build test run test-coverage test-coverage-html integration update-snaps-lexer update-snaps-parser update-snaps-semantic update-snaps-irgen update-snaps-codegen update-snaps-integration

build:
	go build -o the ./cmd/the

test:
	go test -v ./internal/...

IN_FILE ?= ''

run: build
	./the $(IN_FILE)

# Example: make run IN_FILE=examples/src/strings.the

test-coverage:
	go test -coverprofile cover.out ./internal/...

test-coverage-html:
	go test -coverprofile cover.out ./internal/...; go tool cover -html=cover.out

integration: build
	go test -tags=integration -v ./cmd/the/...

update-snaps-lexer: 
	UPDATE_SNAPS=true go test ./internal/lexer/...

update-fixtures-parser:
	UPDATE_FIXTURES=true go test ./internal/parser/... -run=TestGenerateFixtures
	$(MAKE) update-snaps-parser

update-snaps-parser:
	UPDATE_SNAPS=true go test ./internal/parser/...

update-fixtures-semantic:
	UPDATE_FIXTURES=true go test ./internal/semantic/... -run=TestGenerateFixtures
	$(MAKE) update-snaps-semantic

update-snaps-semantic:
	UPDATE_SNAPS=true go test ./internal/semantic/...

update-fixtures-irgen:
	UPDATE_FIXTURES=true go test ./internal/irgen/... -run=TestGenerateFixtures
	$(MAKE) update-snaps-irgen

update-snaps-irgen:
	UPDATE_SNAPS=true go test ./internal/irgen/...

update-fixtures-codegen:
	UPDATE_FIXTURES=true go test ./internal/codegen/... -run=TestGenerateFixtures
	$(MAKE) update-snaps-codegen

update-snaps-codegen:
	UPDATE_SNAPS=true go test ./internal/codegen/...

update-snaps-integration: build
	UPDATE_SNAPS=true go test -tags=integration ./cmd/the/...
