DROP VIEW IF EXISTS v_earthquakes, v_rainfalls, v_geo_events CASCADE;
DROP TABLE IF EXISTS geo_events CASCADE;

CREATE EXTENSION IF NOT EXISTS postgis;


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

CREATE OR REPLACE VIEW v_geo_events AS
SELECT 
    *,
    CASE 
        WHEN event_type = 'Earthquake' AND magnitude >= 6.0 THEN '#E74C3C'
        WHEN event_type = 'Earthquake' AND magnitude >= 4.0 THEN '#E67E22'
        WHEN event_type = 'Earthquake' THEN '#F1C40F'
        WHEN event_type = 'Rain' AND magnitude >= 100.0 THEN '#AF56D6'
        WHEN event_type = 'Rain' AND magnitude >= 40.0 THEN '#1B4F72'
        WHEN event_type = 'Rain' THEN '#3498DB'
        WHEN event_type = 'Volcano' THEN '#D35400'
        ELSE '#95A5A6'
    END AS color,
    CONCAT(event_type, ' (', magnitude, '): ', location) AS label
FROM geo_events;

-- Earthquake
CREATE OR REPLACE VIEW v_earthquakes AS
SELECT * FROM v_geo_events 
WHERE event_type = 'Earthquake';

-- Rain
CREATE OR REPLACE VIEW v_rainfalls AS
SELECT * FROM v_geo_events 
WHERE event_type = 'Rain';