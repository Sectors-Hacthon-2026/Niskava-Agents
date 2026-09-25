#!/usr/bin/env bash
# =============================================================================
# NISKAVA AGENT — Universal Installer (Linux & macOS)
# Automates Go verification, Python virtualenv bootstrap, and CLI compilation.
# =============================================================================

set -euo pipefail

# ANSI Color codes
BOLD="\033[1m"
GREEN="\033[0;32m"
YELLOW="\033[0;33m"
RED="\033[0;31m"
CYAN="\033[0;36m"
RESET="\033[0m"

echo -e "\n${BOLD}${CYAN}=== Niskava Agent Universal Installer (Linux/macOS) ===${RESET}\n"

# 1. Check Go compiler
if ! command -v go >/dev/null 2>&1; then
    echo -e "${RED}[✗] Go compiler is not installed or not found on PATH.${RESET}"
    echo -e "    Please install Go (>=1.22):"
    echo -e "      • Ubuntu/Debian : sudo apt install golang"
    echo -e "      • macOS (Homebrew): brew install go"
    echo -e "      • Official Site   : https://go.dev/dl/\n"
    exit 1
fi
GO_VER=$(go version | awk '{print $3}')
echo -e "${GREEN}[✓] Go compiler detected:${RESET} ${GO_VER}"

# 2. Check Python 3.11+
PYTHON_BIN=""
for candidate in python3 python py; do
    if command -v "$candidate" >/dev/null 2>&1; then
        VER_STR=$("$candidate" -c "import sys; print(f'{sys.version_info.major}.{sys.version_info.minor}')" 2>/dev/null || true)
        if [ -n "$VER_STR" ]; then
            MAJOR=$(echo "$VER_STR" | cut -d. -f1)
            MINOR=$(echo "$VER_STR" | cut -d. -f2)
            if [ "$MAJOR" -ge 3 ] && [ "$MINOR" -ge 11 ]; then
                PYTHON_BIN="$candidate"
                break
            fi
        fi
    fi
done

if [ -z "$PYTHON_BIN" ]; then
    echo -e "${RED}[✗] Python 3.11+ is required but was not found.${RESET}"
    echo -e "    Please install Python 3.11 or higher:"
    echo -e "      • Ubuntu/Debian : sudo apt install python3 python3-venv python3-pip"
    echo -e "      • macOS (Homebrew): brew install python@3.12"
    echo -e "      • Official Site   : https://www.python.org/downloads/\n"
    exit 1
fi
echo -e "${GREEN}[✓] Compatible Python runtime detected:${RESET} $($PYTHON_BIN --version)"

# 3. Create or update virtual environment
echo -e "\n${BOLD}Setting up Python isolated environment (.venv)...${RESET}"
if [ ! -d ".venv" ]; then
    echo -e "  • Creating virtual environment in .venv/"
    "$PYTHON_BIN" -m venv .venv
else
    echo -e "  • Virtual environment .venv/ already exists."
fi

VENV_PIP=".venv/bin/pip"
VENV_PY=".venv/bin/python3"
if [ ! -f "$VENV_PY" ]; then
    VENV_PY=".venv/bin/python"
fi

echo -e "  • Upgrading pip in .venv..."
"$VENV_PIP" install --upgrade pip --quiet

echo -e "  • Installing quantitative engine dependencies (NumPy, Pandas, NetworkX, Trafilatura)..."
"$VENV_PIP" install -r backend/engine/requirements.txt --quiet
echo -e "${GREEN}[✓] Quantitative dependencies installed successfully.${RESET}"

# 4. Compile Go Core Binary
echo -e "\n${BOLD}Compiling Niskava standalone executable...${RESET}"
mkdir -p bin
go build -o bin/niskava ./cmd/niskava
chmod +x bin/niskava
echo -e "${GREEN}[✓] Binary compiled successfully:${RESET} bin/niskava"

# 5. Optional PATH installation
if [ -d "$HOME/.local/bin" ]; then
    cp bin/niskava "$HOME/.local/bin/niskava"
    echo -e "${GREEN}[✓] Copied executable to:${RESET} $HOME/.local/bin/niskava"
    if [[ ":$PATH:" != *":$HOME/.local/bin:"* ]]; then
        echo -e "${YELLOW}[!] Note: Add ~/.local/bin to your PATH to run 'niskava' from any directory:${RESET}"
        echo -e "    export PATH=\"\$HOME/.local/bin:\$PATH\""
    fi
fi

echo -e "\n${BOLD}${GREEN}=== Installation Complete! ===${RESET}"
echo -e "Next recommended steps:"
echo -e "  1. Run diagnostic check : ${CYAN}./bin/niskava doctor${RESET}"
echo -e "  2. Configure credentials: ${CYAN}./bin/niskava setup${RESET}"
echo -e "  3. Start Web Workspace  : ${CYAN}./bin/niskava serve${RESET}"
echo -e "  4. Launch Terminal HUD  : ${CYAN}./bin/niskava${RESET}\n"
