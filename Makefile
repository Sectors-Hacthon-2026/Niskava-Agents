.PHONY: all build run test test-go test-python test-mcp clean lint venv doctor

ifeq ($(OS),Windows_NT)
    PYTHON ?= python
    VENV_PIP := .venv\Scripts\pip.exe
    VENV_PY := .venv\Scripts\python.exe
    VENV_PYTEST := .venv\Scripts\pytest.exe
    BIN := bin\niskava.exe
    MKDIR := powershell -Command "New-Item -ItemType Directory -Force -Path bin"
    RM := powershell -Command "Remove-Item -Recurse -Force -ErrorAction SilentlyContinue"
else
    PYTHON ?= python3
    VENV_PIP := .venv/bin/pip
    VENV_PY := .venv/bin/python3
    VENV_PYTEST := .venv/bin/pytest
    BIN := bin/niskava
    MKDIR := mkdir -p bin
    RM := rm -rf
endif

all: build

venv:
	@$(PYTHON) -m venv .venv
	@$(VENV_PIP) install --upgrade pip
	@$(VENV_PIP) install -r backend/engine/requirements.txt

build:
	@$(MKDIR)
	go build -o $(BIN) ./cmd/niskava

run: build
	$(BIN)

doctor: build
	$(BIN) doctor

test: test-go test-python

test-go:
	go test -v -race ./backend/core/... ./clients/cli/...

test-python:
	PYTHONPATH=backend $(VENV_PYTEST) backend/engine/tests/

test-mcp:
	PYTHONPATH=backend $(VENV_PYTEST) backend/engine/tests/test_unified_mcp_server.py backend/engine/tests/test_sectors_mcp_server.py -v

lint:
	go vet ./...
	gofmt -s -l backend/ clients/ cmd/
	PYTHONPATH=backend $(VENV_PY) -m py_compile backend/engine/*.py backend/engine/*/*.py backend/engine/*/*/*.py

clean:
	@$(RM) bin __pycache__ backend/engine/__pycache__ .pytest_cache
