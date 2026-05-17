.PHONY: help tidy tidy-check mod-verify vet lint test build verify clean tag

help:
	@echo "Available targets:"
	@echo "  tidy       Run go mod tidy"
	@echo "  tidy-check Check for go mod tidy drift"
	@echo "  mod-verify Verify go modules"
	@echo "  vet        Run go vet"
	@echo "  lint       Run golangci-lint"
	@echo "  test       Run tests"
	@echo "  build      Build binary"
	@echo "  verify     Run all quality checks"
	@echo "  clean      Remove build artifacts"
	@echo "  tag        Tag a new version (use V=v0.1.0 MSG=\"...\")"

tidy:
	./scripts/tidy.sh

tidy-check:
	./scripts/tidy-check.sh

mod-verify:
	./scripts/mod-verify.sh

vet:
	./scripts/vet.sh

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh

build:
	./scripts/build.sh

verify:
	./scripts/verify.sh

clean:
	./scripts/clean.sh

tag:
	@if [ -z "$(V)" ]; then echo "V is required (e.g. V=v0.1.0)"; exit 1; fi
	@if [ -z "$(MSG)" ]; then echo "MSG is required"; exit 1; fi
	git tag -a $(V) -m "$(MSG)"
	@echo "Tagged $(V)"
