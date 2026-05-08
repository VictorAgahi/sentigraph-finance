# SentiGraph Finance

SentiGraph Finance is a high-performance market intelligence platform designed for real-time cryptocurrency sentiment analysis and order flow monitoring. The system leverages a polyglot microservices architecture (Rust, Go, TypeScript) to ensure low-latency data ingestion and scalable AI-driven analysis.

## System Architecture

The platform is composed of four primary services interconnected via a high-performance Redis backbone and a time-series database.

### SentiGraph Harvester (Rust)
* High-frequency data ingestion engine built in Rust.
* Connects to Binance Futures WebSocket streams for real-time ticker, market depth, and liquidation data.
* Implements batch processing and pipeline publishing to Redis (DragonflyDB) to minimize network overhead.
* Features automatic reconnection logic with exponential backoff.

### SentiGraph Analyst (Go)
* Core processing service written in Go using a clean internal/cmd structure.
* Orchestrates a pool of workers for asynchronous sentiment analysis using Ollama.
* Handles real-time market data state and broadcasts updates via WebSockets.
* Manages long-term persistence in TimescaleDB using optimized batch inserts.
* Exposes REST endpoints for historical data retrieval.

### SentiGraph Scout (Go)
* RSS-based market intelligence gatherer.
* Monitors multiple financial news sources and Nitter instances for real-time updates.
* Dispatches news items to the Analyst service for sentiment scoring.

### SentiGraph Dashboard (Next.js)
* Real-time monitoring terminal.
* High-performance visualization using lightweight-charts.
* WebSocket-driven UI updates for sub-second market reaction monitoring.

## Prerequisites

* Docker and Docker Compose
* Ollama (for sentiment analysis)
* Node.js 20+ (for local dashboard development)
* Go 1.24+ (for local analyst/scout development)
* Rust 1.80+ (for local harvester development)

## Setup and Installation

### 1. Ollama Configuration
SentiGraph requires a local Ollama instance for AI-driven sentiment analysis.

Install Ollama from [ollama.com](https://ollama.com) and pull the required model:
```bash
ollama pull llama3.2:1b
```

Ensure Ollama is accessible. On macOS, it typically runs on `http://localhost:11434`.

### 2. Environment Configuration
The system uses environment variables for configuration. While defaults are provided in `docker-compose.yml`, you can create a `.env` file in the root directory for customizations:

```env
REDIS_URL=redis:6379
DATABASE_URL=postgres://postgres:password@timescale:5432/postgres
OLLAMA_URL=http://host.docker.internal:11434/api/generate
DISCORD_WEBHOOK_URL=your_webhook_url
```

### 3. Deployment via Docker Compose
The simplest way to start the entire stack is using Docker Compose:

```bash
docker-compose up --build
```

This will initialize:
* DragonflyDB (Redis replacement) on port 6379
* TimescaleDB on port 5432
* Harvester services for BTC, ETH, and SOL
* Analyst service on port 8080
* Scout news service
* Dashboard on port 3000

### 4. Database Initialization
TimescaleDB is automatically initialized with the required schemas on first run. If you need to manually inspect the data:
```bash
docker exec -it sentigraph-timescale psql -U postgres
```

## API Reference

### WebSocket Feed
* **URL**: `ws://localhost:8080/ws`
* **Description**: Real-time stream of market states and liquidation events.

### Historical Data
* **Endpoint**: `GET /api/history/:symbol`
* **Description**: Returns the last 1000 data points for the specified symbol.

## Development

### Directory Structure
```text
.
├── sentigraph-analyst/     # Go Analyst Service (cmd/internal structure)
├── sentigraph-scout/       # Go Scout News Service
├── sentigraph-harvester/   # Rust Data Ingestion Service
├── sentigraph-dashboard/   # Next.js Frontend
└── docker-compose.yml      # Orchestration
```

### CI/CD
Building and image integrity are verified via GitHub Actions on every push and pull request to the main branch.