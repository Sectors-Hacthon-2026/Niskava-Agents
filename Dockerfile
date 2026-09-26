# =============================================================================
# Stage 1: Build pure-Go Niskava Core Binary
# =============================================================================
FROM golang:1.24-alpine AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/niskava ./cmd/niskava

# =============================================================================
# Stage 2: Runtime Container (Python 3.12 Slim + Quant Engine + Niskava)
# =============================================================================
FROM python:3.12-slim

ENV PYTHONUNBUFFERED=1 \
    PYTHONIOENCODING=utf-8 \
    PYTHONUTF8=1 \
    NISKAVA_PORT=8080

WORKDIR /app

# Install system utilities
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

# Install Python quantitative dependencies
COPY backend/engine/requirements.txt /app/backend/engine/requirements.txt
RUN pip install --no-cache-dir --upgrade pip && \
    pip install --no-cache-dir -r /app/backend/engine/requirements.txt

# Copy engine and project code
COPY backend /app/backend
COPY --from=builder /bin/niskava /usr/local/bin/niskava

# Expose web workspace & REST daemon
EXPOSE 8080

VOLUME ["/root/.niskava"]

ENTRYPOINT ["niskava"]
CMD ["serve", "--port", "8080", "--open=false"]
