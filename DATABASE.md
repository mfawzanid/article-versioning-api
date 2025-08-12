# Database Schema

This document describes the database schema for the **Article Versioning API**.

---

## **users**
Stores system users and their roles.

| Column     | Type         | Constraints                                           | Description                                 |
|------------|--------------|-------------------------------------------------------|---------------------------------------------|
| id         | SERIAL       | PRIMARY KEY                                           | Auto-incremented ID                         |
| username   | VARCHAR(50)  | NOT NULL, UNIQUE                                      | Unique username                             |
| role       | VARCHAR(50)  | NOT NULL                                              | User role: `reader`, `admin`, `writer`, `editor` |
| hash       | TEXT         | NOT NULL                                              | Password hash                               |

---

## **articles**
Represents an article container. Each article can have multiple versions.

| Column     | Type         | Constraints                     | Description                  |
|------------|--------------|---------------------------------|------------------------------|
| id         | SERIAL       | PRIMARY KEY                     | Auto-incremented ID          |
| serial     | VARCHAR(25)  | NOT NULL, UNIQUE                 | Unique article identifier    |
| created_at | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP | Creation timestamp           |
| updated_at | TIMESTAMP    |                                 | Last update timestamp        |
| deleted_at | TIMESTAMP    |                                 | Soft delete timestamp        |

---

## **versions**
Stores individual versions of articles.

| Column                 | Type         | Constraints                                                                 | Description                           |
|------------------------|--------------|-------------------------------------------------------------------------------|---------------------------------------|
| id                     | SERIAL       | PRIMARY KEY                                                                 | Auto-incremented ID                   |
| serial                 | VARCHAR(25)  | NOT NULL, UNIQUE                                                             | Unique version identifier             |
| author_username        | VARCHAR(50)  | NOT NULL REFERENCES users(username)                                          | Author's username                     |
| version_number         | INT          | NOT NULL                                                                    | Version number                        |
| article_serial         | VARCHAR(25)  | NOT NULL REFERENCES articles(serial), UNIQUE(article_serial, serial)         | Related article serial                 |
| status                 | VARCHAR(50)  |                                                                             | `draft`, `published`, `archived`      |
| title                  | TEXT         | NOT NULL                                                                    | Version title                         |
| content                | TEXT         | NOT NULL                                                                    | Version content                       |
| created_at             | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP                                          | Creation timestamp                    |
| updated_at             | TIMESTAMP    |                                                                             | Last update timestamp                 |
| deleted_at             | TIMESTAMP    |                                                                             | Soft delete timestamp                  |
| published_at           | TIMESTAMP    |                                                                             | Publish timestamp                     |
| tag_relationship_score | FLOAT        | DEFAULT 0                                                                   | Relationship score between tags       |

**Index:**
- `one_published_per_article`: Ensures only one published version per article.

---

## **tags**
Represents tags assigned to article versions.

| Column     | Type         | Constraints                                           | Description              |
|------------|--------------|-------------------------------------------------------|--------------------------|
| id         | SERIAL       | PRIMARY KEY                                           | Auto-incremented ID      |
| serial     | VARCHAR(25)  | NOT NULL, UNIQUE                                      | Unique tag identifier    |
| name       | TEXT         | NOT NULL                                              | Tag name                 |
| created_at | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP                    | Creation timestamp       |

---

## **version_tags**
Many-to-many relationship between versions and tags.

| Column         | Type         | Constraints                                                 | Description                |
|----------------|--------------|-------------------------------------------------------------|----------------------------|
| version_serial | VARCHAR(25)  | NOT NULL REFERENCES versions(serial)                        | Linked version serial      |
| tag_serial     | VARCHAR(25)  | NOT NULL REFERENCES tags(serial)                            | Linked tag serial          |
| **Primary Key**|              | (version_serial, tag_serial)                                | Unique combination         |

---

## **tag_stats**
Stores statistics for individual tags.

| Column                  | Type         | Constraints                              | Description                           |
|-------------------------|--------------|------------------------------------------|---------------------------------------|
| tag_serial              | VARCHAR(25)  | PRIMARY KEY REFERENCES tags(serial)      | Linked tag serial                     |
| usage_count             | INT          | DEFAULT 0                                | Number of published articles using it |
| trending_score          | FLOAT        | DEFAULT 0                                | Trending score                        |
| usage_count_updated_at  | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP       | Last usage count update timestamp     |
| trending_score_updated_at | TIMESTAMP  | NOT NULL DEFAULT CURRENT_TIMESTAMP       | Last trending score update timestamp  |

---

## **tag_pair_stats**
Stores statistics for pairs of tags.

| Column       | Type         | Constraints                                     | Description                          |
|--------------|--------------|-------------------------------------------------|--------------------------------------|
| tag1_serial  | VARCHAR(25)  | NOT NULL REFERENCES tags(serial)                | First tag serial                     |
| tag2_serial  | VARCHAR(25)  | NOT NULL REFERENCES tags(serial)                | Second tag serial                    |
| usage_count  | INT          | DEFAULT 0                                       | Number of published articles with both tags |
| updated_at   | TIMESTAMP    | NOT NULL DEFAULT CURRENT_TIMESTAMP              | Last update timestamp                |
| **Primary Key** |           | (tag1_serial, tag2_serial)                      | Unique combination                   |
