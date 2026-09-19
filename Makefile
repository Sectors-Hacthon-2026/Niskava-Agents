.PHONY: all build run test test-go test-python clean lint venv

all: build

venv:
	@test -d .venv || python3 -m venv .venv
	@.venv/bin/pip install --upgrade pip
	@.venv/bin/pip install -r engine/requirements.txt

build:
	@mkdir -p bin
	go build -o bin/niskava ./cmd/niskava

run: build
	./bin/niskava

test: test-go test-python

test-go:
	go test -v ./internal/config ./internal/db ./internal/ipc ./internal/server

test-python:
	PYTHONPATH=. .venv/bin/pytest tests/

test-mcp:
	PYTHONPATH=. .venv/bin/pytest tests/test_unified_mcp_server.py tests/test_sectors_mcp_server.py -v

lint:
	go vet ./...
	gofmt -s -l internal/ cmd/
	PYTHONPATH=. .venv/bin/python3 -m py_compile engine/*.py engine/*/*.py engine/*/*/*.py

clean:
	rm -rf bin/
	rm -rf __pycache__ engine/__pycache__ engine/*/__pycache__ engine/*/*/__pycache__
	rm -rf .pytest_cache tests/__pycache__
