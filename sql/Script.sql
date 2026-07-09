DROP TABLE IF EXISTS geo_events;

CREATE TABLE IF NOT EXISTS geo_events (
	id varchar(255) NOT NULL,
	source varchar(10) NOT NULL ,
	event_type varchar(50) NOT NULL,
	title text NOT NULL,
	magnitude NUMERIC(6, 2) NOT NULL,
	depth NUMERIC(6, 2) NOT NULL,
	event_timestamp timestamptz NOT NULL,
	country varchar(10) NOT NULL,
	location text NOT NULL,
	longitude NUMERIC(6, 3) NOT NULL,
	latitude NUMERIC(6, 3) NOT NULL,
	geom GEOMETRY(Point, 4326),
	created_at timestamptz DEFAULT now(),
	details jsonb NOT NULL,
	PRIMARY KEY (id, event_type)
) PARTITION BY LIST (event_type);

-- Partition sub table
CREATE TABLE geo_events_earthquakes PARTITION OF geo_events
    FOR VALUES IN ('Earthquake');

CREATE TABLE geo_events_rainfalls PARTITION OF geo_events
    FOR VALUES IN ('Rain');

CREATE TABLE geo_events_tsunamis PARTITION OF geo_events
    FOR VALUES IN ('Tsunami');
    
CREATE TABLE geo_events_weather PARTITION OF geo_events
    FOR VALUES IN ('SevereWeather');

CREATE TABLE geo_events_volcanoes PARTITION OF geo_events
    FOR VALUES IN ('Volcano');

CREATE TABLE geo_events_default PARTITION OF geo_events DEFAULT;

-- Create index
CREATE INDEX IF NOT EXISTS idx_event_timestamp ON
geo_events(event_timestamp DESC);

CREATE INDEX IF NOT EXISTS idx_geo_events_geom ON geo_events USING GIST (geom);

CREATE EXTENSION IF NOT EXISTS postgis;

