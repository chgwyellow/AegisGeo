# AegisGeo

AegisGeo is a global natural disaster and geological anomaly monitoring backend engine built in Go. The system leverages concurrency to simultaneously fetch real-time data from multiple geological monitoring agencies (e.g., CWA, USGS, JMA), cleanse and format the inputs into a unified model, perform spatial collision deduplication via PostgreSQL + PostGIS, and save the events to both a database and an in-memory cache.

---

## Tech Stack

- **Programming Language**: Go 1.26.4
- **Database**: PostgreSQL (requires **PostGIS** extension for physical distance calculation of geographical coordinates)
- **Database Driver**: `github.com/jackc/pgx/v5` (`pgxpool` connection pool)
- **Environment Variables**: `github.com/joho/godotenv`
- **Concurrency Control**: Native `sync.WaitGroup` and `sync.RWMutex` (for thread-safe memory cache operations)

---

## Architecture and Data Flow

The core data pipeline of AegisGeo can be divided into four steps: initialization, concurrent ingestion, spatial collision deduplication, and storage/caching.

```mermaid
graph TD
    %% Source Definitions
    subgraph sources ["Data Sources (HTTP API)"]
        USGS_API[USGS API]
        CWA_API[CWA API]
        JMA_API[JMA API]
    end

    %% Ingestion Layer
    subgraph ingestion ["Ingestion Layer (Goroutines)"]
        G1[Goroutine: UsgsClient]
        G2[Goroutine: CwaClient]
        G3[Goroutine: JmaClient]
    end

    %% Connection to Ingestion
    USGS_API -->|HTTP GET / JSON| G1
    CWA_API -->|HTTP GET / Authorization Header| G2
    JMA_API -->|HTTP GET / JSON| G3

    %% Core Process
    subgraph core ["Core Telemetry Engine (Standardization & Verification)"]
        Cache["In-Memory Cache MemoryCache <br>(sync.RWMutex)"]
        DB_Check{PostgreSQL + PostGIS <br> Spatial Collision Check}
        SaveDB[(PostgreSQL)]
    end

    G1 -->|Standardize to Event| Cache
    G2 -->|Standardize to Event| Cache
    G3 -->|Standardize to Event| Cache

    G1 -->|Write/Verify| DB_Check
    G2 -->|Write/Verify| DB_Check
    G3 -->|Write/Verify| DB_Check

    DB_Check -->|Distance > 50km <br> or Time Diff > 60s| SaveDB
    DB_Check -->|Duplicate USGS Event| Drop[Discard USGS Event]
    DB_Check -->|ID Conflict| Upsert["ON CONFLICT <br> DO UPDATE"]
```

### Deduplication and Conflict Resolution Logic

1. **Spatial and Temporal Collision Detection**: Before storing an event into the database, the system calculates its physical distance to existing events using PostGIS's `ST_DistanceSphere` function. If another event occurred within **60 seconds** and is within **50 kilometers**, it is identified as a collision.
2. **Source-Based Deduplication**: When a collision is detected, if the database already contains a local high-precision event from `CWA` or `JMA`, and the new event is sourced from `USGS`, the system will filter out and discard the `USGS` event to ensure data accuracy.
3. **Upsert (ON CONFLICT)**: If an event `ID` already exists, the database triggers `ON CONFLICT (id) DO UPDATE` to update variables such as magnitude, depth, timestamp, title, and location.

---

## Project Structure

```text
AegisGeo/
├── cmd/
│   └── server/
│       └── main.go       # Application entry point (Loads env, initializes pgxpool, spawns ingestion, and outputs cache summary)
├── internal/
│   ├── database/
│   │   └── postgres.go   # PostgreSQL client using pgxpool for queries, spatial collision deduplication, and Upsert operations
│   ├── ingestion/
│   │   ├── client.go     # IngestionClient interface definition
│   │   ├── cwa.go        # Central Weather Administration (CWA) fetcher and parser (requires authorization token)
│   │   ├── jma.go        # Japan Meteorological Agency (JMA) fetcher and parser
│   │   └── usgs.go       # United States Geological Survey (USGS) fetcher, parser, and country translation logic
│   ├── models/
│   │   └── disaster.go   # Unified Event domain model struct definition
│   └── store/
│       └── cache.go      # Thread-safe in-memory cache using sync.RWMutex, sorted by timestamp descending
├── sql/
│   └── Script.sql        # Database schema script (enables postgis and creates geo_events table with GIST index)
├── .env                  # Environment configuration file (URLs, tokens, and DB URL)
├── go.mod                # Go module definition and dependencies
└── README.md             # Project documentation
```

---

## How to Run

### 1. Database Setup

Make sure you have PostgreSQL installed with the `postgis` extension enabled. Use `sql/Script.sql` to initialize tables and indices:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
-- Run Script.sql to create the geo_events table and spatial indexes
```

### 2. Configure Environment Variables

Create a `.env` file in the project root directory:

```env
DATABASE_URL=postgres://user:password@localhost:5432/aegisgeo?sslmode=disable
CWA_API_URL=https://opendata.cwa.gov.tw/api/v1/rest/datastore/E-A0015-001
CWA_TOKEN=your_cwa_token
USGS_API_URL=https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/all_hour.geojson
JMA_API_URL=https://www.jma.go.jp/bosai/otpath/data/otpath.json
```

### 3. Run the Server

Execute the following command in the root directory:

```bash
go run cmd/server/main.go
```

The application will:

1. Load variables from `.env`.
2. Connect and ping the PostgreSQL database.
3. Spawn isolated Goroutines for `CWA`, `USGS`, and `JMA` clients concurrently.
4. Fetch raw JSON payloads, convert them to standard events, write to the database (performing PostGIS deduplication), and cache them in memory.
5. Print all processed events cached in memory (ordered by timestamp descending) to the console.
