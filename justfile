# Set shell
set shell := ["bash", "-uc"]

binary := "mnemo"
main := "./main.go"
packages := "./internal/..."
tmp_html := `mktemp -p /tmp mnemosynefsXXXXXXXXXXXX.html`

default:
    @just --list

# Generate code (mocks, etc.)
generate:
    @echo "Running go generate..."
    go generate ./...

# Run linter
lint:
    @echo "Running golangci-lint..."
    golangci-lint run ./...

# Run tests with coverage
test: generate
    @echo "Running tests..."
    go test -coverpkg={{packages}} -coverprofile=coverage.out {{packages}}
    @echo "Coverage report: coverage.out"
    @go tool cover -func=coverage.out | tail -n 1

# Run tests with HTML output
cover: test
    @covreport -i coverage.out -o {{tmp_html}}
    @xdg-open {{tmp_html}}

# Debug build (symbols, no optimization)
debug: generate
    @echo "Building debug binary..."
    go build -o build/debug/{{binary}}-debug {{main}}

# Release build (optimized and stripped)
release: generate
    @echo "Building release binary..."
    go build -ldflags="-s -w" -o build/release/{{binary}} {{main}}

# Cleanup artifacts
clean:
    @echo "Cleaning build artifacts..."
    go clean -testcache
    rm -rf build coverage.out
