.PHONY: all build run test test-go test-python clean lint venv

all: build

venv:
	@test -d .venv || python3 -m venv .venv
	@.venv/bin/pip install --upgrade pip
	@.venv/bin/pip install -r backend/engine/requirements.txt

build:
	@mkdir -p bin
	go build -o bin/niskava ./cmd/niskava

run: build
	./bin/niskava

test: test-go test-python

test-go:
	go test -v ./backend/core/... ./clients/cli/...

test-python:
	PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/

test-mcp:
	PYTHONPATH=backend .venv/bin/pytest backend/engine/tests/test_unified_mcp_server.py backend/engine/tests/test_sectors_mcp_server.py -v

lint:
	go vet ./...
	gofmt -s -l backend/ clients/ cmd/
	PYTHONPATH=backend .venv/bin/python3 -m py_compile backend/engine/*.py backend/engine/*/*.py backend/engine/*/*/*.py

clean:
	rm -rf bin/
	rm -rf __pycache__ backend/engine/__pycache__ backend/engine/*/__pycache__ backend/engine/*/*/__pycache__
	rm -rf .pytest_cache backend/engine/tests/__pycache__
