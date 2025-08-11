CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(50) NOT NULL,
    role VARCHAR(50) NOT NULL, -- reader, admin, writer, editor
    hash TEXT NOT NULL,
    UNIQUE(username)
);

CREATE TABLE articles (
    id SERIAL PRIMARY KEY,
    serial VARCHAR(25) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(serial)
);

CREATE TABLE versions (
    id SERIAL PRIMARY KEY,
    serial VARCHAR(25) NOT NULL,
    author_username VARCHAR(50) NOT NULL REFERENCES users(username),
    version_number INT NOT NULL,
    article_serial VARCHAR(25) NOT NULL REFERENCES articles(serial),
    status VARCHAR(50), -- draft, pubslished, archived
    title TEXT NOT NULL,
    content TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    published_at TIMESTAMP,
    tag_relationship_score FLOAT DEFAULT 0,
    UNIQUE(serial),
    UNIQUE(article_serial, serial)
);

CREATE UNIQUE INDEX one_published_per_article ON versions(article_serial) WHERE status = 'published';

CREATE TABLE tags (
    id SERIAL PRIMARY KEY,
    serial VARCHAR(25) NOT NULL,
    name TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(serial)
);

CREATE TABLE version_tags (
    version_serial VARCHAR(25) NOT NULL REFERENCES versions(serial),
    tag_serial VARCHAR(25) NOT NULL REFERENCES tags(serial),
    PRIMARY KEY(version_serial, tag_serial)
);

CREATE TABLE tag_stats (
    tag_serial VARCHAR(25) PRIMARY KEY REFERENCES tags(serial),
    usage_count INT DEFAULT 0,
    trending_score FLOAT DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE tag_pair_stats (
    tag1_serial VARCHAR(25) NOT NULL REFERENCES tags(serial),
    tag2_serial VARCHAR(25) NOT NULL REFERENCES tags(serial),
    usage_count INT DEFAULT 0,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tag1_serial, tag2_serial)
);