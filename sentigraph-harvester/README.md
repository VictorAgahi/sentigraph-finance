# SentiGraph-Harvester

SentiGraph-Harvester is a high-performance, HFT-ready Rust microservice designed to ingest real-time Bitcoin market data from the Binance WebSocket API and pipeline it into DragonFlyDB.

## Features

- **HFT-Optimized Event Loop**: Built on Tokio with zero-copy deserialization using lifetimes (`'input`) to eliminate heap allocations on the hot path.
- **Batch Pipelining**: Efficient data ingestion into Redis/DragonFlyDB using a batching strategy (50 messages or 10ms interval) to minimize syscalls.
- **DragonFlyDB Integration**: Fully compatible with DragonFlyDB's multi-threaded architecture for maximum throughput.
- **Robust Storage Architecture**: Utilizes `XADD` (Streams) for historical data and `PUBLISH` (Pub/Sub) for real-time downstream consumers.
- **Strict Configuration**: Fail-fast environment variable management using `envy` and `.env` files.
- **Containerized**: Production-ready `Docker` and `docker-compose` setup included.

## Technical Stack

- **Language**: Rust (Stable)
- **Async Runtime**: [Tokio](https://tokio.rs/)
- **WebSocket**: `tokio-tungstenite`
- **Serialization**: `serde` / `serde_json` (Zero-copy)
- **Database**: `redis-rs` (Compatible with DragonFlyDB)
- **Config**: `dotenvy` / `envy`

## Getting Started

### Prerequisites

- Rust toolchain (1.75+)
- Docker and Docker Compose (for running DragonFlyDB)

### Configuration

1. Copy the example environment file:
```bash
cp .env.example .env
```
2. Adjust the values in `.env` if necessary.

### Running with Docker (Recommended)

To launch the entire stack (DragonFlyDB + Harvester):

```bash
docker-compose up --build
```

### Manual Execution

1. Start DragonFlyDB:
```bash
docker-compose up -d redis
```

2. Run the harvester:
```bash
cargo run --release
```

## Internal Architecture

The application is modularized for clarity and performance:

- **`config.rs`**: Centralized, type-safe environment configuration.
- **`models.rs`**: Zero-copy data structures for Binance market events.
- **`harvester.rs`**: Core engine handling the WebSocket stream and the dual-trigger batcher (count/time).
- **`storage.rs`**: Asynchronous Redis pipeline manager using `ConnectionManager` for multiplexed throughput.

The harvester loop uses `tokio::select!` to multiplex between receiving messages and flushing the buffer, ensuring no data is held back during low-volatility periods while maintaining high efficiency during high-frequency spikes.
