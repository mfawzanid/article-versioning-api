# API Contract

## Register User
Registers a new user with the specified username, password, and role.

### Endpoint:
```bash
POST /users/register
```

### Request

#### Body
| Field     | Type   | Required | Description                                     | Example   |
|-----------|--------|----------|-------------------------------------------------|-----------|
| username  | string | Yes      | Unique username for the user.                   | `writer1` |
| password  | string | Yes      | Password for the account.                       | `writer1` |
| role      | string | Yes      | Role assigned to the user (`reader`, `writer`, `editor`, `admin`). | `writer`  |

Example:
```json
{
    "username": "writer1",
    "password": "writer1",
    "role": "writer"
}
```

## Login
Authenticates a user with their username and password, returning a JWT token if the credentials are valid.

### Request

#### Body
| Field     | Type   | Required | Description                         | Example   |
|-----------|--------|----------|-------------------------------------|-----------|
| username  | string | Yes      | The username of the account.        | `writer1` |
| password  | string | Yes      | The password of the account.        | `writer1` |

Example:
```json
{
    "username": "writer1",
    "password": "writer1"
}
```
#### Response
```json
{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQ5ODYwMTgsInJvbGUiOiJ3cml0ZXIiLCJ1c2VybmFtZSI6IndyaXRlcjEifQ.yTDMJwPZrTGV-bkGmhcUi0TCaw25dsXrv98WhTMjn04"
}
```

## Create Tag
Creates a new tag.

### Endpoint:
```bash
POST /tags
```

#### Body
| Field | Type   | Required | Description | Example |
|-------|--------|----------|-------------|---------|
| name  | string | Yes      | Tag name.   | `tag1`  |

Example:
```json
{
    "name": "tag1"
}
```

## Create Article
Creates a new article with an initial version in `draft` status.

### Endpoint:
```bash
POST /articles
```

#### Body
| Field       | Type     | Required | Description                                | Example       |
|-------------|----------|----------|--------------------------------------------|---------------|
| title       | string   | Yes      | The title of the article.                  | `title2`      |
| content     | string   | Yes      | The content/body of the article.           | `content2`    |
| tagSerials  | string[] | Yes      | List of tag serials associated with the article. | `["TAG-J1KNW7"]` |


Example:
```json
{
    "title": "title2",
    "content": "content2",
    "tagSerials": [
        "TAG-J1KNW7"
    ]
}
```

#### Response
```json
{
    "articleSerial": "ART-NZ42JZ",
    "authorUsername": "writer1",
    "version": {
        "serial": "VER-2E1YU0",
        "authorUsername": "writer1",
        "versionNumber": 1,
        "articleSerial": "ART-NZ42JZ",
        "title": "title2",
        "content": "content2",
        "status": "draft",
        "createdAt": "0001-01-01T00:00:00Z",
        "updatedAt": null,
        "deletedAt": null,
        "publishedAt": null,
        "tagRelationshipScore": 0,
        "tags": null
    }
}
```

## Update Article Version Status
Updates the status of a specific article version.  
Only users with `admin`, `editor`, or `writer` roles are allowed to perform this action.

### Endpoint:
```bash
PATCH /articles/{articleSerial}/versions/{versionSerial}/status
```

### Path Parameters
| Parameter       | Type   | Required | Description                            | Example     |
|-----------------|--------|----------|----------------------------------------|-------------|
| articleSerial   | string | Yes      | The serial of the article.             | `ART-NZ42JZ`|
| versionSerial   | string | Yes      | The serial of the article version.     | `VER-2E1YU0`|

### Request Body
| Field      | Type   | Required | Description                                       | Example   |
|------------|--------|----------|---------------------------------------------------|-----------|
| newStatus  | string | Yes      | The new status for the version (`draft`, `published`, `archived`). | `draft`   |

## Delete Article
Deletes an article by its serial. This action is restricted to users with roles `admin`, `editor`, or `writer`.

### Endpoint:
```bash
DELETE /articles/{articleSerial}
```

### Path Parameters
| Parameter     | Type   | Required | Description                    | Example       |
|---------------|--------|----------|--------------------------------|---------------|
| articleSerial | string | Yes      | The serial of the article.     | `ART-SSNC2W`  |

## Create Article Version
Creates a new version for an existing article.

### Endpoint:
```bash
POST /articles/{articleSerial}/version
```

### Path Parameters
| Parameter     | Type   | Required | Description                | Example       |
|---------------|--------|----------|----------------------------|---------------|
| articleSerial | string | Yes      | The serial of the article. | `ART-DN42E1`  |

### Request Body
| Field       | Type     | Required | Description                                   | Example       |
|-------------|----------|----------|-----------------------------------------------|---------------|
| title       | string   | Yes      | Title of the new article version.             | `title2`      |
| content     | string   | Yes      | Content of the new article version.           | `content2`    |
| tagSerials  | string[] | Yes      | List of tag serials to associate with version. | `["TAG-YV0MIT"]` |

Example:
```json
{
    "title": "title2",
    "content": "content2",
    "tagSerials": [
        "TAG-YV0MIT"
    ]
}
```

### Response
| Field                       | Type     | Description                                                                 |
|-----------------------------|----------|-----------------------------------------------------------------------------|
| articleSerial               | string   | Serial identifier for the article.                                          |
| authorId                    | string   | Username of the author who created the version.                             |
| version                     | object   | Details of the created article version.                                     |
| version.serial              | string   | Serial identifier for the version.                                          |
| version.authorUsername      | string   | Username of the author for this version.                                    |
| version.versionNumber       | integer  | Version number of the article.                                              |
| version.articleSerial       | string   | Serial identifier of the related article.                                   |
| version.title               | string   | Title of the version.                                                       |
| version.content             | string   | Content of the version.                                                     |
| version.status              | string   | Status of the version (`draft`, `published`, `archived`).                   |
| version.createdAt           | time   | Timestamp when the version was created (RFC3339 format).                     |
| version.updatedAt           | time   | Timestamp when the version was last updated (nullable).                     |
| version.deletedAt           | time   | Timestamp when the version was deleted (nullable).                          |
| version.publishedAt         | time   | Timestamp when the version was published (nullable).                        |
| version.tagRelationshipScore| float    | Relationship score between tags in this version.                            |
| version.tags                | array    | List of tags associated with the version (nullable if no tags are assigned).|

Example:
```json
{
    "articleSerial": "ART-93WEE9",
    "authorId": "writer1",
    "version": {
        "serial": "VER-7CKQ5M",
        "authorUsername": "writer1",
        "versionNumber": 2,
        "articleSerial": "ART-93WEE9",
        "title": "title2",
        "content": "content2",
        "status": "draft",
        "createdAt": "0001-01-01T00:00:00Z",
        "updatedAt": null,
        "deletedAt": null,
        "publishedAt": null,
        "tagRelationshipScore": 0,
        "tags": null
    }
}
```

## Get Articles
Retrieves a paginated list of articles.

### Endpoint:
```bash
GET /articles
```

### Query Parameters
| Field           | Type   | Required | Description                                                                                   | Example       |
|-----------------|--------|----------|-----------------------------------------------------------------------------------------------|---------------|
| page            | int    | No       | Page number for pagination. Defaults to 1.                                                   | `1`           |
| pageSize        | int    | No       | Number of items per page. Defaults to 10.                                                    | `1`           |
| authorUsername  | string | No       | Filter articles by the author's username.                                                    | `writer1`     |
| tagSerial       | string | No       | Filter articles that contain a specific tag by its serial.                                   | `TAG-TX7D3E`  |
| sortBy          | string | No       | Field to sort by. Supported values: `created_at`, `updated_at`, `published_at`, `title`.     | `created_at`  |
| sortType        | string | No       | Sort order. Accepted values: `asc` (ascending) or `desc` (descending).                       | `desc`        |

Example:
```bash
curl --location 'localhost:8080/articles?page=1&pageSize=1&authorUsername=writer1&tagSerial=TAG-TX7D3E&sortBy=created_at&sortType=desc' \
--header 'Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3NTQ5ODc3MzAsInJvbGUiOiJ3cml0ZXIiLCJ1c2VybmFtZSI6IndyaXRlcjEifQ.4fhWsGstbRBzfRlSB0JyksIjM47xSkCjnUOPF27drdU'
```

### Response
Example
```json
{
    "version": [
        {
            "serial": "VER-16Q0KT",
            "authorUsername": "writer1",
            "versionNumber": 1,
            "articleSerial": "ART-CC0OYK",
            "title": "title1",
            "content": "content1",
            "status": "published",
            "createdAt": "2025-08-12T06:40:06.111097Z",
            "updatedAt": "2025-08-12T06:41:55.176946Z",
            "deletedAt": null,
            "publishedAt": null,
            "tagRelationshipScore": 0,
            "tags": [
                {
                    "serial": "TAG-J1KNW7",
                    "name": "tag1"
                }
            ]
        }
    ],
    "pagination": {
        "page": 1,
        "pageSize": 1,
        "totalPage": 1,
        "total": 1
    }
}
```

## Get Article Latest Detail
Retrieves the latest published version and the latest version (regardless of status) for a given article.

### Endpoint
```bash
GET /articles/{articleSerial}/latest-details
```

### Path Parameters

| Parameter      | Type   | Required | Description                              | Example       |
|----------------|--------|----------|------------------------------------------|---------------|
| articleSerial  | string | Yes      | Unique serial of the article to retrieve | `ART-GSJ9OK`  |

### Response
Example:
```json
{
    "publishedVersion": {
        "serial": "VER-K8WR9S",
        "authorUsername": "writer1",
        "versionNumber": 1,
        "articleSerial": "ART-GSJ9OK",
        "title": "title3",
        "content": "content3",
        "status": "published",
        "createdAt": "2025-08-12T07:35:49.76614Z",
        "updatedAt": "2025-08-12T07:37:07.529764Z",
        "deletedAt": null,
        "publishedAt": "2025-08-12T07:37:07.529764Z",
        "tagRelationshipScore": 0,
        "tags": [
            {
                "serial": "TAG-TX7D3E",
                "name": "tag1"
            }
        ]
    },
    "latestVersion": {
        "serial": "VER-K8WR9S",
        "authorUsername": "writer1",
        "versionNumber": 1,
        "articleSerial": "ART-GSJ9OK",
        "title": "title3",
        "content": "content3",
        "status": "published",
        "createdAt": "2025-08-12T07:35:49.76614Z",
        "updatedAt": "2025-08-12T07:37:07.529764Z",
        "deletedAt": null,
        "publishedAt": "2025-08-12T07:37:07.529764Z",
        "tagRelationshipScore": 0,
        "tags": [
            {
                "serial": "TAG-TX7D3E",
                "name": "tag1"
            }
        ]
    }
}
```

## Get Article Versions
Retrieves all versions of a specific article.

### Endpoint:
```bash
GET /articles/{articleSerial}/versions
```

### Path Parameters
| Parameter      | Type   | Required | Description                              | Example       |
|----------------|--------|----------|------------------------------------------|---------------|
| articleSerial  | string | Yes      | Unique serial of the article to retrieve versions for | `ART-GSJ9OK`  |

### Response
```json
{
    "versions": [
        {
            "serial": "VER-K8WR9S",
            "authorUsername": "writer1",
            "versionNumber": 1,
            "articleSerial": "ART-GSJ9OK",
            "title": "title3",
            "content": "content3",
            "status": "published",
            "createdAt": "2025-08-12T07:35:49.76614Z",
            "updatedAt": "2025-08-12T07:37:07.529764Z",
            "deletedAt": null,
            "publishedAt": "2025-08-12T07:37:07.529764Z",
            "tagRelationshipScore": 0,
            "tags": [
                {
                    "serial": "TAG-TX7D3E",
                    "name": "tag1"
                }
            ]
        },
        {
            "serial": "VER-JYYC77",
            "authorUsername": "writer1",
            "versionNumber": 2,
            "articleSerial": "ART-GSJ9OK",
            "title": "title111",
            "content": "content111",
            "status": "draft",
            "createdAt": "2025-08-12T07:43:38.700003Z",
            "updatedAt": null,
            "deletedAt": null,
            "publishedAt": null,
            "tagRelationshipScore": 0,
            "tags": [
                {
                    "serial": "TAG-TX7D3E",
                    "name": "tag1"
                }
            ]
        }
    ]
}
```

## Get Article Version
Retrieves the details of a specific article version based on its serial.

### Endpoint:
```bash
GET /articles/versions/{versionSerial}
```

### Path Parameters
| Parameter     | Type   | Required | Description                             | Example       |
|---------------|--------|----------|-----------------------------------------|---------------|
| versionSerial | string | Yes      | Unique serial of the article version to retrieve | `VER-K8WR9S`  |

### Response
Example
```json
{
    "serial": "VER-K8WR9S",
    "authorUsername": "writer1",
    "versionNumber": 1,
    "articleSerial": "ART-GSJ9OK",
    "title": "title3",
    "content": "content3",
    "status": "published",
    "createdAt": "2025-08-12T07:35:49.76614Z",
    "updatedAt": "2025-08-12T07:37:07.529764Z",
    "deletedAt": null,
    "publishedAt": "2025-08-12T07:37:07.529764Z",
    "tagRelationshipScore": 0,
    "tags": [
        {
            "serial": "TAG-TX7D3E",
            "name": "tag1"
        }
    ]
}
```

## Get All Tags
Retrieves a paginated list of all tags with their usage count and trending score.

### Endpoint:
```bash
GET /tags
```

### Response
Example:
```json
{
    "tags": [
        {
            "serial": "TAG-6J4KI5",
            "name": "tag3",
            "usageCount": 0,
            "trendingScore": 0
        },
        {
            "serial": "TAG-KB5GAM",
            "name": "tag2",
            "usageCount": 0,
            "trendingScore": 0
        },
        {
            "serial": "TAG-TX7D3E",
            "name": "tag1",
            "usageCount": 1,
            "trendingScore": 0.9993211
        }
    ],
    "pagination": {
        "page": 1,
        "pageSize": 3,
        "totalPage": 1,
        "total": 3
    }
}
```

## Get Tag
Retrieves detailed information about a specific tag by its serial.

### Endpoint:
```bash
GET /tags/{serial}
```

#### Response
Example
```json
{
    "serial": "TAG-TX7D3E",
    "name": "tag1",
    "usageCount": 1,
    "trendingScore": 0.9991836
}
```

## Update All Tag Trending Score
Updates the trending score for all tags.  
This API is intended to be called by a worker periodically.

### Endpoint:
```bash
PUT /articles/tags/trending-score
```