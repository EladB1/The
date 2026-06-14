.PHONY: build test test-coverage integration update-snaps-lexer update-snaps-parser update-snaps-semantic update-snaps-irgen update-snaps-codegen update-snaps-integration

build:
	go build -o the ./cmd/the

test:
	go test -v ./internal/...

test-coverage:
	go test -cover ./internal/...

integration: build
	go test -tags=integration -v ./cmd/the/...

update-snaps-lexer: 
	UPDATE_SNAPS=true go test ./internal/lexer/...

update-snaps-parser:
	UPDATE_SNAPS=true go test ./internal/parser/...

update-snaps-semantic:
	UPDATE_SNAPS=true go test ./internal/semantic/...

update-snaps-irgen:
	UPDATE_SNAPS=true go test ./internal/irgen/...

update-snaps-codegen:
	UPDATE_SNAPS=true go test ./internal/codegen/...

update-snaps-integration: build
	UPDATE_SNAPS=true go test -tags=integration ./cmd/the/...
