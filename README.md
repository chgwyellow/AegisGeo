# AegisGeo

![Go](https://img.shields.io/badge/Go-1.26.4-00ADD8?logo=go&logoColor=white)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-PostGIS-4169E1?logo=postgresql&logoColor=white)
![Version](https://img.shields.io/badge/version-1.7.0-brightgreen)
![Platform](https://img.shields.io/badge/platform-Windows-0078D6?logo=windows&logoColor=white)
![Started](https://img.shields.io/badge/started-July%202026-blue)

AegisGeo is a global natural disaster and meteorological/geological anomaly monitoring backend engine built in Go. The system leverages concurrency to simultaneously fetch real-time data from multiple monitoring agencies (CWA, USGS, JMA, NOAA, NWS), cleanse and format the inputs into a unified model, perform spatial collision deduplication via PostgreSQL + PostGIS, and save the events to both a database and an in-memory cache.

---

## Architecture and Data Flow

The core data pipeline of AegisGeo can be divided into four steps: initialization, concurrent ingestion, spatial collision deduplication, and storage/caching.

```mermaid
graph TD
    USGS_API[USGS Earthquake API] --> G1[UsgsClient]
    CWA_EQ_API[CWA Earthquake API] --> G2[CwaClient]
    JMA_API[JMA Earthquake API] --> G3[JmaClient]
    NOAA_API[NOAA Tsunami API] --> G4[TsunamiClient]
    CWA_RAIN_API[CWA Rainfall API] --> G5[CwaRainClient]
    NWS_API[NWS Severe Weather API] --> G6[NwsSevereWeatherClient]
    VOLCANO_API[USGS Volcano API] --> G7[VolcanoClient]

    G1 --> Std[Standardize to Event]
    G2 --> Std
    G3 --> Std
    G4 --> Std
    G5 --> Std
    G6 --> Std
    G7 --> Std

    Std --> Cache["In-Memory Cache (MemoryCache)"]
    Cache --> Check{"Collision Check\n(Distance <= 50km\nTime Diff < 60s)"}

    Check -->|USGS Duplicate| Drop[Discard USGS Event]
    Check -->|New Event| DB[(PostgreSQL)]
    Check -->|ID Conflict| Upsert[ON CONFLICT DO UPDATE]

    %% Source API Colors
    style USGS_API fill:#4a7c59,stroke:#2d5a3a,color:#fff
    style CWA_EQ_API fill:#8b6914,stroke:#6b4f10,color:#fff
    style JMA_API fill:#c0392b,stroke:#962d22,color:#fff
    style NOAA_API fill:#2471a3,stroke:#1a5276,color:#fff
    style CWA_RAIN_API fill:#5b7fa5,stroke:#3d5a80,color:#fff
    style NWS_API fill:#7d3c98,stroke:#5b2c6f,color:#fff
    style VOLCANO_API fill:#d35400,stroke:#a04000,color:#fff

    %% Client Goroutine Colors
    style G1 fill:#5dade2,stroke:#2e86c1,color:#fff
    style G2 fill:#f0b27a,stroke:#e67e22,color:#fff
    style G3 fill:#e74c3c,stroke:#c0392b,color:#fff
    style G4 fill:#3498db,stroke:#2471a3,color:#fff
    style G5 fill:#85c1e9,stroke:#5dade2,color:#fff
    style G6 fill:#a569bd,stroke:#7d3c98,color:#fff
    style G7 fill:#e67e22,stroke:#d35400,color:#fff

    %% Core Process Colors
    style Std fill:#1abc9c,stroke:#16a085,color:#fff
    style Cache fill:#2ecc71,stroke:#27ae60,color:#fff
    style Check fill:#f39c12,stroke:#d68910,color:#fff
    style Drop fill:#95a5a6,stroke:#7f8c8d,color:#fff
    style DB fill:#2c3e50,stroke:#1a252f,color:#fff
    style Upsert fill:#34495e,stroke:#2c3e50,color:#fff
```

### Ingestion Clients and Data Sources

1. **CWA (Central Weather Administration)**:
   - **Earthquake Client**: Fetches Taiwan's latest earthquake reports from the CWA Open Data API (requires authorization token). Event Type: `Earthquake`.
   - **Rainfall Client**: Fetches real-time precipitation measurements from CWA weather stations across Taiwan (requires authorization token). Event Type: `Rain`.
2. **USGS (United States Geological Survey)**:
   - **Earthquake Client**: Fetches global earthquake events in GeoJSON format. Translates geographic keywords/states to ISO country codes. Event Type: `Earthquake`.
   - **Volcano Hazards Program**: Fetches real-time CAP (Common Alerting Protocol) volcanic activity alerts (XML format) from the HANS service. Event Type: `Volcano`.
3. **JMA (Japan Meteorological Agency)**:
   - Fetches recent earthquake and tsunami history events for Japan. Event Type: `Earthquake`.
4. **NOAA (National Oceanic and Atmospheric Administration)**:
   - Fetches historical global tsunami event data from the NOAA/NCEI HAZEL hazard-service API. This source is not an active tsunami warning feed, so new records may appear sparsely. Event Type: `Tsunami`.
5. **NWS (National Weather Service)**:
   - Fetches active severe weather alerts (e.g., Tornado, Severe Thunderstorm watches and warnings) across the US. Identifies requests via a contact email address in the `User-Agent` header, configured via the `Email` environment variable. Event Type: `SevereWeather`.

### Deduplication and Conflict Resolution Logic

1. **Spatial and Temporal Collision Detection**: Before storing an event into the database, the system calculates its physical distance to existing events of the same `event_type` using PostGIS's `ST_DistanceSphere` function. If another event occurred within **60 seconds** and is within **50 kilometers**, it is identified as a collision.
2. **Source-Based Deduplication**: When an earthquake collision is detected, if the database already contains a local high-precision event from `CWA` or `JMA`, and the new event is sourced from `USGS`, the system will filter out and discard the `USGS` event to ensure data accuracy.
3. **Upsert (ON CONFLICT)**: If an event `ID` and `event_type` combination already exists, the database triggers `ON CONFLICT (id, event_type) DO UPDATE` to update variables such as magnitude, depth, timestamp, title, country, location, geom, and custom telemetry `details` (stored as JSONB).

### Storage & Performance Features

- **PostgreSQL Table Partitioning**: The primary table `geo_events` is partitioned by list (`PARTITION BY LIST (event_type)`) into sub-tables for enhanced query performance:
  - `geo_events_earthquakes` (`Earthquake`)
  - `geo_events_rainfalls` (`Rain`)
  - `geo_events_tsunamis` (`Tsunami`)
  - `geo_events_weather` (`SevereWeather`)
  - `geo_events_volcanoes` (`Volcano`)
  - `geo_events_default` (DEFAULT fallback)
- **Indexing**: Optimized using a GIST spatial index (`ST_SetSRID(ST_MakePoint(longitude, latitude), 4326)`) and B-Tree index on `event_timestamp DESC`.
- **Transaction & Batch Ingestion**: Ingestion operations run inside database transactions utilizing `pgx.Batch` for high-throughput batch updates.

---

## REST API Specification

AegisGeo provides a RESTful API server (`cmd/server/main.go`) to query processed disaster events and verify database status.

### Endpoints Overview

| Method | Endpoint | Authorization | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/status` | None | Database health check endpoint |
| `GET` | `/api/events` | `X-API-KEY` Header | Returns recent disaster events with filtering options |

---

### 1. `GET /api/status`

Checks whether the PostgreSQL database connection pool is active.

- **Request Headers**: None required.
- **Success Response**: `HTTP 200 OK`
  - Body: `OK`
- **Failure Response**: `HTTP 500 Internal Server Error`
  - Body: `Database is down`

---

### 2. `GET /api/events`

Fetches recent disaster events from the PostgreSQL database.

- **Request Headers**:
  - `X-API-KEY`: Required. Must match the `SECRET_KEY` value configured in `.env`.
- **Query Parameters**:

  | Parameter | Type | Required | Default Value | Description |
  | :--- | :--- | :--- | :--- | :--- |
  | `type` | String | No | *(All Types)* | Filter by event type: `Earthquake`, `Rain`, `Tsunami`, `SevereWeather`, `Volcano` |
  | `start` | String | No | `30 days ago` | Filter start date in `YYYY-MM-DD` format (Timezone: CST / UTC+8) |
  | `end` | String | No | `Current Date` | Filter end date in `YYYY-MM-DD` format (Timezone: CST / UTC+8) |

- **HTTP Status Codes**:
  - `200 OK`: Request succeeded. Returns JSON array of events.
  - `400 Bad Request`: Invalid date format for `start` or `end` parameter (expected `YYYY-MM-DD`).
  - `401 Unauthorized`: Missing or incorrect `X-API-KEY` header (`Unauthorized: Access Denied`).
  - `500 Internal Server Error`: Database query or JSON encoding failure.

#### Example Requests & Responses

##### Fetch Latest Disaster Summaries (Default Top 20)

```bash
curl -H "X-API-KEY: your_secret_api_key" http://localhost:8080/api/events
```

##### Fetch Earthquakes with Date Range Filter

```bash
curl -H "X-API-KEY: your_secret_api_key" \
     "http://localhost:8080/api/events?type=Earthquake&start=2026-07-01&end=2026-07-21"
```

##### Response Body (`200 OK`)

```json
[
  {
    "id": "us7000m123",
    "title": "M 5.4 - 15 km ESE of Hualien, Taiwan",
    "source": "USGS",
    "event_type": "Earthquake",
    "magnitude": 5.4,
    "depth": 12.5,
    "event_timestamp": "2026-07-20T18:45:00+08:00",
    "country": "TW",
    "location": "Hualien County, Taiwan"
  },
  {
    "id": "CWA-EQ-115042",
    "title": "07/20 18:44 第042號顯著有感地震",
    "source": "CWA",
    "event_type": "Earthquake",
    "magnitude": 5.5,
    "depth": 10.2,
    "event_timestamp": "2026-07-20T18:44:12+08:00",
    "country": "TW",
    "location": "花蓮縣近海"
  }
]
```

---

## Project Structure

```text
AegisGeo/
├── cmd/
│   ├── server/
│   │   └── main.go       # API server entry point (Sets up REST endpoints, connects to DB, and starts HTTP server)
│   ├── ingest/
│   │   └── main.go       # Telemetry ingestion job entry point (Single-cycle client fetches, PostGIS deduplication, DB save)
│   └── health/
│       └── main.go       # Data Health Check CLI tool (Inspects upstream client connectivity, latency, and event counts without DB write side effects)
├── internal/
│   ├── api/
│   │   ├── handlers.go   # REST API route handlers (/api/events, /api/status)
│   │   └── handlers_test.go # Unit tests for API handlers
│   ├── database/
│   │   └── postgres.go   # PostgreSQL client using pgxpool for queries, spatial collision deduplication, and Upsert operations
│   ├── health/
│   │   ├── health.go     # Data health check logic, latency measurement, and CLI table formatting
│   │   └── health_test.go# Unit tests for health check report generation
│   ├── ingestion/
│   │   ├── client.go     # IngestionClient interface definition
│   │   ├── cwa.go        # Central Weather Administration (CWA) Earthquake client
│   │   ├── cwa_rain.go   # Central Weather Administration (CWA) Rainfall client
│   │   ├── geodict.go    # Global geographic country dictionary and US states mapping
│   │   ├── jma.go        # Japan Meteorological Agency (JMA) client
│   │   ├── noaa_tsunami.go # National Oceanic and Atmospheric Administration (NOAA) Tsunami client
│   │   ├── nws_severe_weather.go # National Weather Service (NWS) Severe Weather client
│   │   ├── volcano.go    # USGS Volcano client (XML Parser)
│   │   └── usgs.go       # United States Geological Survey (USGS) client
│   ├── models/
│   │   └── disaster.go   # Unified Event domain model struct definition
│   └── store/
│       └── cache.go      # Thread-safe in-memory cache using sync.RWMutex, sorted by timestamp descending
├── sql/
│   └── Script.sql        # Database schema script (enables postgis, creates list partitioned geo_events table, GIST index, and GIS views)
├── .env                  # Environment configuration file (URLs, tokens, and DB URL)
├── go.mod                # Go module definition and dependencies
└── README.md             # Project documentation
```

---

## How to Run

### 1. Database Setup

Make sure you have PostgreSQL installed with the `postgis` extension enabled. Use `sql/Script.sql` to initialize tables, partitions, indices, and mapping views:

```sql
CREATE EXTENSION IF NOT EXISTS postgis;
-- Run Script.sql to create the geo_events tables, partitions, and GIS views
```

### 2. Configure Environment Variables

Create a `.env` file in the project root directory:

```env
DATABASE_URL=postgres://postgres:password@localhost:5432/aegisgeo?sslmode=disable
CWA_EQK_URL=https://opendata.cwa.gov.tw/api/v1/rest/datastore/E-A0015-001
CWA_RAIN_URL=https://opendata.cwa.gov.tw/api/v1/rest/datastore/O-A0002-001
CWA_TOKEN=your_cwa_token
USGS_API_URL=https://earthquake.usgs.gov/earthquakes/feed/v1.0/summary/2.5_day.geojson
JMA_API_URL=https://api.p2pquake.net/v2/history?codes=551&limit=30
NOAA_API_URL=https://www.ngdc.noaa.gov/hazel/hazard-service/api/v1/tsunamis/events?minYear=2020
NWS_API_URL=https://api.weather.gov/alerts/active?event=Tornado%20Watch,Tornado%20Warning,Severe%20Thunderstorm%20Watch,Severe%20Thunderstorm%20Warning
VOLCANO_API_URL=https://volcanoes.usgs.gov/hans-public/rss/cap/
Email=your_email@example.com
SECRET_KEY=your_secret_api_key
```

### 3. Run the Applications

#### Run the Telemetry Ingestion Job

To execute a single telemetry data ingestion cycle (optimized for cron schedulers, serverless functions like AWS Lambda, or CI/CD pipelines like GitHub Actions):

```bash
go run cmd/ingest/main.go
```

The ingestion job will:

1. Load variables from `.env`.
2. Connect and ping the PostgreSQL database.
3. Spawn isolated Goroutines for `CWA Earthquake`, `CWA Rain`, `USGS`, `JMA`, `NOAA Tsunami`, `NWS Severe Weather`, and `USGS Volcano` clients concurrently.
4. Fetch raw payloads, convert them to standard events, write to the database (performing PostGIS deduplication and batch inserts), and cache them in memory.
5. Print the latest 5 anomaly events cached in memory (ordered by timestamp descending) to the console and exit.

#### Run the Data Health Check CLI Tool

To inspect real-time operational status, event counts, API response latency, and latest event timestamps across all 7 upstream monitoring sources (without database side effects):

```bash
go run cmd/health/main.go
```

The health check CLI will:

1. Load configuration from `.env`.
2. Initialize client instances for `CWA Earthquake`, `USGS Earthquake`, `JMA Earthquake`, `NOAA Tsunami`, `CWA Rain Station`, `NWS Severe Weather`, and `USGS Volcano`.
3. Fetch data across all sources while recording response latency.
4. Print a formatted summary report to stdout:

```text
AegisGeo Data Health Check
Generated at: 2026-07-22 16:35:41 CST
------------------------------------------------------------------------------------------------------
Source               Status Count  Latest Event Time    Duration     Error
------------------------------------------------------------------------------------------------------
CWA                  OK     16     2026-07-15 22:44:30 CST 352.66ms     -
USGS                 OK     38     2026-07-22 14:46:40 CST 507.658ms    -
JMA                  OK     29     2026-07-22 10:55:00 CST 194.2663ms   -
NOAA-Tsunami         OK     125    2026-07-17 14:48:39 CST 618.49ms     -
CWA-RainStation      OK     1318   2026-07-22 16:20:00 CST 332.866ms    -
NWS-SevereWeather    OK     230    2026-07-22 00:19:00 CST 1.1545253s   -
USGS-Volcano         OK     2      2026-07-21 12:12:32 CST 821.0272ms   -
Summary: 7 sources checked, 7 OK, 0 FAIL
------------------------------------------------------------------------------------------------------
```

#### Run the API Server

To start the REST API server to serve event data:

```bash
go run cmd/server/main.go
```

The server will:

1. Load variables from `.env` and connect to the PostgreSQL database.
2. Register endpoints (`/api/status` and `/api/events`).
3. Listen and serve requests on port `8080` (blocks until terminated).

### 4. Viewing Spatial Data in DBeaver

The database setup automatically creates three pre-configured database views for GIS visualization:

- `v_geo_events`: All events with dynamic color-coding and tooltips.
- `v_earthquakes`: Earthquake-only events categorized by magnitude.
- `v_rainfalls`: Rainfall-only events categorized by precipitation.

To view these on a map in DBeaver:

1. In the database navigator, expand your connection ➡️ **Views**.
2. Double-click any view (e.g., `v_geo_events`).
3. Switch to the **Data** tab in the main editor.
4. On the right-side vertical toolbar of the data grid, select the **Spatial** panel.
5. The map will render with **dynamic color-coding** (e.g., Red/Orange/Yellow for earthquakes by magnitude, Dark/Light Blue for rain by precipitation level) and custom labels on hover.

## Tech Stack

- **Database Driver**: `github.com/jackc/pgx/v5` (`pgxpool` connection pool, batching transactions)
- **Environment Variables**: `github.com/joho/godotenv`
- **Concurrency Control**: Native `sync.WaitGroup` and `sync.RWMutex` (for thread-safe memory cache operations)

---

## Deploying to GitHub Actions with Cron-Job.org

The project includes a pre-configured GitHub Actions workflow located in `.github/workflows/ingest.yml` to run the ingestion cycle. Because GitHub Actions' native scheduler can be delayed during high-traffic periods, we trigger it precisely using [Cron-Job.org](https://cron-job.org/).

### Setup Instructions

1. **Create a GitHub Repository**: Push your project to GitHub.
2. **Configure Repository Secrets**:
   Go to your repository settings: **Settings ➡️ Secrets and variables ➡️ Actions** and add the following **Repository Secrets**:
   - `DATABASE_URL`: Your Neon connection string (e.g., `postgres://neondb_owner:password@host/neondb?sslmode=require`).
   - `CWA_TOKEN`: Your Central Weather Administration API token.
   - `EMAIL`: Your contact email (used in the User-Agent header for NWS requests).
3. **Generate a GitHub Personal Access Token (PAT)**:
   - Go to your GitHub **Settings ➡️ Developer Settings ➡️ Personal Access Tokens ➡️ Tokens (classic)**.
   - Generate a new classic token, selecting the `workflow` scope. Copy the generated token (`ghp_...`).
4. **Configure [Cron-Job.org](https://cron-job.org/)**:
   - Create a free account on [Cron-Job.org](https://cron-job.org/).
   - Click **Create Cron Job**:
     - **Title**: `AegisGeo Ingestion`
     - **URL**: `https://api.github.com/repos/{owner}/{repo}/actions/workflows/ingest.yml/dispatches` (Replace `{owner}` and `{repo}` with your GitHub username and repository name).
     - **Request Method**: `POST`
     - **Headers**:
       - `Authorization`: `Bearer <your_copied_ghp_token>` (Make sure there is a space after `Bearer`).
       - `Accept`: `application/vnd.github+json`
       - `X-GitHub-Api-Version`: `2022-11-28`
       - `User-Agent`: `Cron-Job.org`
       - `Content-Type`: `application/json`
     - **Body** (raw/JSON):

       ```json
       {
         "ref": "main"
       }
       ```

     - **Schedule**: Set to execute every 10 minutes.
