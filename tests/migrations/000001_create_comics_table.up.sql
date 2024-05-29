CREATE TABLE IF NOT EXISTS comics (
    id INT PRIMARY KEY,
    url TEXT NOT NULL,
    keywords JSONB 
);

CREATE INDEX comics_keywords_idx ON comics USING GIN (keywords);

CREATE TABLE IF NOT EXISTS keywords (
    id SERIAL PRIMARY KEY,
    keyword TEXT UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS keyword_comics_map (
    keyword_id INT REFERENCES keywords(id),
    comics_id INT REFERENCES comics(id),
    PRIMARY KEY (keyword_id, comics_id)
);