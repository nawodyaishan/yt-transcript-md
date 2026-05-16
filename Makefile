.PHONY: help sync format lint typecheck test check build clean

help:
	@echo "Available targets:"
	@echo "  sync       Install project and development dependencies"
	@echo "  format     Format source and tests with Ruff"
	@echo "  lint       Lint source and tests with Ruff"
	@echo "  typecheck  Type check source with mypy"
	@echo "  test       Run the test suite"
	@echo "  check      Run format check, lint, typecheck, and tests"
	@echo "  build      Build package distributions"
	@echo "  clean      Remove generated caches and build artifacts"

sync:
	uv sync --all-groups

format:
	uv run ruff format .

lint:
	uv run ruff check .

typecheck:
	uv run mypy src

test:
	uv run pytest

check:
	uv run ruff format --check .
	uv run ruff check .
	uv run mypy src
	uv run pytest

build:
	uv build

clean:
	rm -rf build dist htmlcov .coverage .coverage.* .pytest_cache .mypy_cache .ruff_cache *.egg-info
