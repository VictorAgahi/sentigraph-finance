# SentiGraph Finance

A high-performance cryptocurrency market analysis suite built with Rust, Go, and Next.js. Optimized for Apple Silicon (M1/M2/M3).

## 🏗 Architecture

The system is a distributed microservices architecture designed for real-time data flow:

1.  **SentiGraph-Harvester (Rust)**: 
    *   Ingests real-time BTC/USDT trades from Binance.
    *   Uses **Zero-copy deserialization** and **Batching** for HFT performance.
    *   Streams data to DragonFlyDB (Redis) with structured JSON logging.
2.  **SentiGraph-Analyst (Go)**:
    *   Orchestrates a **3-worker Goroutine pool** for AI analysis.
    *   Queries a local **Ollama** instance (Llama 3) for financial sentiment.
    *   Persists market data into **TimescaleDB** using high-speed Batch Inserts.
    *   Exposes a **WebSocket** stream for the frontend.
3.  **SentiGraph-Dashboard (Next.js)**:
    *   Real-time terminal interface with **0 security vulnerabilities**.
    *   Interactive price charts using `lightweight-charts`.
    *   Live sentiment gauge and price flash effects.

## 🛠 Tech Stack

*   **Languages**: Rust, Go, TypeScript (Node 24).
*   **Databases**: DragonFlyDB (Cache/PubSub), TimescaleDB (Time-series SQL).
*   **AI**: Ollama (Llama 3).
*   **Observability**: Structured JSON Logging (Tracing in Rust, Slog in Go).

## 🚀 Quick Start

1. **Prerequisites**:
   *   Docker & Docker Compose.
   *   [Ollama](https://ollama.ai/) running locally with `ollama run llama3`.

2. **Launch**:
   ```bash
   docker-compose up --build
   ```

3. **Access**:
   *   **Dashboard**: `http://localhost:3000`
   *   **Analyst API**: `http://localhost:8080/dashboard`
   *   **History API**: `http://localhost:8080/api/history`

---

Built for high-concurrency environments and real-time market insights.