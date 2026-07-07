DROP TABLE IF EXISTS geo_events;

CREATE TABLE IF NOT EXISTS geo_events (
	id varchar(64) PRIMARY KEY,
	source varchar(10) NOT NULL ,
	event_type varchar(50) NOT NULL,
	title text NOT NULL,
	magnitude NUMERIC(3, 1) NOT NULL,
	depth NUMERIC(6, 2) NOT NULL,
	event_timestamp timestamptz NOT NULL,
	country varchar(10) NOT NULL,
	location varchar(255) NOT NULL,
	geom GEOMETRY(Point, 4326),
	created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_event_timestamp ON
geo_events(event_timestamp DESC);

TRUNCATE TABLE geo_events;

SELECT * FROM geo_events ge ;

CREATE EXTENSION IF NOT EXISTS postgis;

CREATE INDEX IF NOT EXISTS idx_geo_events_geom ON geo_events USING GIST (geom);