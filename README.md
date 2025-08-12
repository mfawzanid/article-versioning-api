# Article Versioning API

A RESTful API for managing articles with versioning support, tag management, and role-based access control (RBAC).  
It supports JWT authentication and enforces user roles to determine which APIs are accessible.  

## Features

- **User Authentication & Roles**  
  - JWT-based authentication with username and password.  
  - Role-based authorization (`admin`, `editor`, `writer`, `reader`).  

- **Article Versioning**  
  - Multiple versions per article (draft, published, archived).  
  - Only one published version per article at a time.  
  - Ability to rollback or view version history.  

- **Tag Management & Analytics**  
    Each article can have tags. Each tag has two kinds of scores:
    - **Usage Count**: number of published articles using the tag.
    - **Trending Score**: computed using **exponential decay** based on usage over time (updated periodically by a background worker).

    Each article has a **Tag Relationship Score**, which is computed using **Positive PMI** to measure how closely tags are related.

### Database Schema
- [View Database Schema Detail](./DATABASE.md)
- [View Database Schema Image](./db_schema.png)

## API Endpoints

### Authentication
| Method | Endpoint        | Description | Auth Required | Roles |
|--------|----------------|-------------|--------------|-------|
| POST   | `/user/register`     | Register a new user | No | - |
| POST   | `/user/login`        | Login and get JWT token | No | - |

---

### Articles

#### Writer Only
| Method | Endpoint              | Description |
|--------|----------------------|-------------|
| POST   | `/articles`          | Create a new article |
| POST   | `/articles/:serial/versions` | Create a new version for an article |

#### Admin, Editor, Writer
| Method | Endpoint                                | Description |
|--------|----------------------------------------|-------------|
| PATCH  | `/articles/:serial/versions/:versionSerial/status` | Update article version status (publish/draft/delete) |
| DELETE | `/articles/:serial`                    | Delete an article |
| GET    | `/articles/:serial/latest-details`             | Get latest article details |
| GET    | `/articles/:serial/versions`             | Get all versions of an article |
| GET    | `/articles/versions/:versionSerial`             | Get version details by serial |

---

### Tags

#### Admin, Editor, Writer
| Method | Endpoint             | Description |
|--------|---------------------|-------------|
| POST   | `/tags`             | Create a new tag |
| GET    | `/tags`             | Get list of tags |
| GET    | `/tags/:serial`     | Get tag details by serial |

---

### Public (Unauthenticated)
| Method | Endpoint  | Description |
|--------|-----------|-------------|
| GET    | `/articles` | Get list of published articles (supports pagination, sorting, filtering) |

---

## Tag Scoring Logic

- **Usage Count**  
  Incremented when a tag is associated with a **published** version of an article.  
  Decremented when the article is drafted or deleted.

- **Trending Score**  
  Calculated using **exponential decay** to prioritize recent activity:
  ```markdown
  trendingScore = usageCount * exp(-λ * Δt)
  - λ: decay rate  
  - Δt: hours since last update
  ```
  Updated periodically by a worker or triggered on `usage_count` change.

- **Tag Relationship Score**  
    Calculated using **Positive PMI (Pointwise Mutual Information)**:

  ```markdown
  PMI(i,j) = log2( (C(i,j) * N) / (C(i) * C(j)) )
  PMI+ = max(PMI, 0)  
  ```
  Where:  
  - `C(i,j)` = count of articles with both tags i and j.  
  - `C(i)` and `C(j)` = count of articles with each tag.  
  - `N` = total number of published articles.  

---

## Tech Stack

- **Backend**: Go (Golang), `gorm.io` ORM
- **Database**: PostgreSQL
- **Auth**: JWT
- **Containerization**: Docker & Docker Compose

## Running the project
Use Docker to run the project:
```
docker compose up --build
```

The API server will be available at http://localhost:8080 and the worker will start automatically.