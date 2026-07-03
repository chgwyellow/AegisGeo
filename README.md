# AegisGeo 🌍

[![Go Version](https://img.shields.io/github/go-mod/go-version/chgwyellow/AegisGeo)](https://github.com/chgwyellow/AegisGeo)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

AegisGeo (Shield of the Earth) is an industrial-grade, highly concurrent global natural disaster and geological anomaly monitoring backend engine built in Go.

The system leverages Go's native concurrency primitives (`Goroutines`, `Channels`, `sync.Mutex`, and `sync.WaitGroup`) to simultaneously ingest live, real-time data streams from multiple global meteorological and geological agencies, standardizing heterogeneous raw payloads into a unified stream telemetry cache.

---

## Key Features

- **Multi-Source Concurrent Ingestion**: Spawns independent, isolated Goroutines with deadlock protection to monitor and ingest live HTTP/WebSocket/SSE streams from international agencies (e.g., CWA, USGS, JMA).
- **Idempotent Deduplication (Anti-Loop Defense)**: Utilizes an internal thread-safe tracking memoization table powered by `sync.Mutex` to guarantee no event payload is processed or broadcast twice.
- **Unified Event Ontology**: Cleanses and transforms chaotic, multi-lingual, and format-clashing JSON structures into a highly structured, single-source-of-truth telemetry model.
- **High-Performance In-Memory Cache**: Thread-safe global cache storage ensuring instantaneous `O(1)` read access for downstream consumer applications and mobile APIs.
- **Zero-Dependency Lightweight Binary**: Compiled into a lightweight executable with minimized memory footprint (≈ 15MB under high traffic concurrency), bypassing heavy JVM or Python runtimes.

---

## Architecture & Data Flow

AegisGeo adopts an **Event-Driven Streaming Architecture**. The data pipeline is divided into three isolated layers: **Ingestion**, **Standardization**, and **Storage/API**.

```mermaid
graph TD
    %% Source Definitions
    subgraph sources ["Data Sources (Global Open Data / Live Streams)"]
        USGS[USGS API - USA]
        CWA[CWA Webhook - Taiwan]
        JMA[JMA Stream - Japan]
    end

    %% Ingestion Layer
    subgraph ingestion ["Ingestion Layer (Isolated Goroutines)"]
        G1[Goroutine: usgs.go]
        G2[Goroutine: cwa.go]
        G3[Goroutine: jma.go]
    end

    %% Connection to Ingestion
    USGS -->|Pull JSON| G1
    CWA -->|Push Event| G2
    JMA -->|Stream SSE| G3

    %% Concurrency & Logic Layer
    subgraph core ["Core Telemetry Engine (Standardization & Protection)"]
        M_Check{Anti-Loop Board <br> sync.Mutex}
        T_Transform["Data Cleansing & <br> Ontology Mapping"]
    end

    G1 --> M_Check
    G2 --> M_Check
    G3 --> M_Check

    M_Check -->|If New Event| T_Transform
    M_Check -->|If Duplicated| Drop["Discard & wg.Done"]

    %% Storage & API Layer
    subgraph storage ["Storage & Outbound Layer"]
        Cache[("Thread-Safe In-Memory Cache")]
        API["Gin Web Server / HTTP API"]
    end

    T_Transform -->|Safe Write| Cache
    Cache -->|"O(1) Read"| API
    API -->|JSON Endpoint| App["Mobile App / Dashboard Frontend"]

    %% Styles
    style M_Check fill:#f9f,stroke:#333,stroke-width:2px
    style Cache fill:#bbf,stroke:#333,stroke-width:2px
    style App fill:#bfb,stroke:#333,stroke-width:2px
```

## Project Structure

```text
aegisgeo/
├── cmd/
│   └── server/
│       └── main.go       # Application entrypoint (Initializes global WaitGroup)
├── internal/
│   ├── ingestion/        # Ingestion Layer: Isolated stream workers (usgs, cwa, jma)
│   ├── model/            # Unified Data Domain: Core ontology structs
│   └── store/            # Storage Layer: Thread-safe in-memory cache using sync.Mutex
├── .gitignore            # Strict network/environment production ignore-list
├── go.mod                # Go module definition and dependencies
└── README.md             # System documentation
```

