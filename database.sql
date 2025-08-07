CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL,
    hash VARCHAR(255) NOT NULL,
    UNIQUE(username)
);