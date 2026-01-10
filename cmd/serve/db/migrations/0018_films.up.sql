CREATE TABLE films (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    tmdb_id INTEGER UNIQUE,
    title VARCHAR NOT NULL,
    year INTEGER,
    overview TEXT,
    runtime INTEGER,
    published BOOLEAN DEFAULT false
);

CREATE INDEX idx_films_tmdb_id ON films(tmdb_id);

CREATE TABLE film_reviews (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    film_id INTEGER NOT NULL REFERENCES films(id) ON DELETE CASCADE,
    watched_date DATE NOT NULL DEFAULT CURRENT_DATE,
    rating INTEGER NOT NULL,
    is_rewatch BOOLEAN DEFAULT false,
    has_spoilers BOOLEAN DEFAULT false,
    review_text TEXT,
    published BOOLEAN DEFAULT false
);

CREATE INDEX idx_film_reviews_film_id ON film_reviews(film_id);

