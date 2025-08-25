CREATE DATABASE IF NOT EXISTS movies;
CREATE TABLE IF NOT EXISTS movies.analytics (
    user_id UUID,
    movie_id String,
    timestamp_ms Int32
) ENGINE = MergeTree()
ORDER BY movie_id;
