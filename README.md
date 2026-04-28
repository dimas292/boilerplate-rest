# URL Shortener

URL shortener API built with Go, featuring a **modular architecture** and **Go generics** for zero-boilerplate CRUD.

## Tech Stack

- **[Gin](https://github.com/gin-gonic/gin)** — HTTP framework
- **[GORM](https://gorm.io)** — ORM (PostgreSQL)
- **[go-redis](https://github.com/redis/go-redis)** — Redis client
- **Go Generics** — Reusable base layers

## Project Structure

```
url-short/
├── cmd/
│   └── main.go                        # Entry point
├── config.yml                         # App configuration (gitignored)
├── examole.config.yml                 # Example config
├── modules/                           # Feature modules
│   ├── url/                           # URL module (CRUD + custom)
│   │   ├── model.go
│   │   ├── repository.go
│   │   ├── service.go
│   │   └── router.go
│   └── tag/                           # Tag module (pure CRUD)
│       ├── model.go
│       └── module.go
├── pkg/                               # Shared packages
│   ├── config/                        # YAML config loader
│   ├── database/                      # PostgreSQL & Redis init
│   ├── model/
│   │   └── base_model.go             # BaseModel + generic constraints
│   ├── repository/
│   │   └── base_repository.go        # Generic CRUD repository
│   ├── service/
│   │   └── base_service.go           # Generic CRUD service
│   ├── handler/
│   │   └── base_handler.go           # Generic CRUD HTTP handler
│   ├── response/
│   │   └── base_response.go          # Standardized API response
│   ├── router/
│   │   └── base_router.go            # Module interface
│   └── server/
│       └── server.go                  # Server bootstrap
├── go.mod
└── go.sum
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Redis

### Setup

1. Clone the repo:

```bash
git clone https://github.com/dimas292/url_shortener.git
cd url_shortener
```

2. Copy and edit the config:

```bash
cp examole.config.yml config.yml
```

```yaml
app:
  name: urlshort
  port: ":4444"
  db:
    postgres:
      dbhost: localhost
      dbuser: postgres
      dbpassword: postgres
      dbname: postgres
    redis:
      host: localhost
      port: 6379
```

3. Install dependencies and run:

```bash
go mod tidy
go run cmd/main.go
```

Server runs at `http://localhost:4444`.

---

## Architecture Guide

### Base Layers (Generics)

Setiap layer menggunakan Go generics dengan pattern **pointer-element constraint**:

```
model.BaseModel  →  repository.BaseRepository  →  service.BaseService  →  handler.BaseHandler
```

| Package | File | Fungsi |
|---|---|---|
| `pkg/model` | `base_model.go` | BaseModel (ID, timestamps, soft-delete), `Model` interface, `ModelPtr` constraint |
| `pkg/repository` | `base_repository.go` | Generic CRUD: `Create`, `FindByID`, `FindAll` (paginated), `Update`, `Delete` |
| `pkg/service` | `base_service.go` | Business logic layer, delegates ke repository |
| `pkg/handler` | `base_handler.go` | HTTP handler CRUD + `RegisterCRUD()` untuk register 5 routes sekaligus |
| `pkg/response` | `base_response.go` | Standardized JSON response + pagination |
| `pkg/router` | `base_router.go` | `Module` interface: `RegisterRoutes(rg *gin.RouterGroup)` |

---

## Cara Menambahkan Module Baru

### Skenario 1: Pure CRUD (tanpa custom logic)

Cukup **2 file**. Contoh: module `Tag`.

#### 1. Buat model (`modules/tag/model.go`)

```go
package tag

import "github.com/dimas292/url_shortener/pkg/model"

type Tag struct {
    model.BaseModel
    Name  string `json:"name" gorm:"uniqueIndex" binding:"required"`
    Color string `json:"color" gorm:"default:'#000000'"`
}
```

#### 2. Buat module (`modules/tag/module.go`)

```go
package tag

import (
    "github.com/dimas292/url_shortener/pkg/handler"
    "github.com/dimas292/url_shortener/pkg/repository"
    "github.com/dimas292/url_shortener/pkg/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type TagModule struct {
    handler *handler.BaseHandler[Tag, *Tag]
}

func NewTagModule(db *gorm.DB) *TagModule {
    db.AutoMigrate(&Tag{})

    repo := repository.NewBaseRepository[Tag, *Tag](db)
    svc := service.NewBaseService[Tag, *Tag](repo)
    h := handler.NewBaseHandler[Tag, *Tag](svc)

    return &TagModule{handler: h}
}

func (m *TagModule) RegisterRoutes(rg *gin.RouterGroup) {
    m.handler.RegisterCRUD(rg.Group("/tags"))
}
```

#### 3. Register di `cmd/main.go`

```go
srv.RegisterModules(
    tagmodule.NewTagModule(srv.DB),
)
```

**Selesai!** Otomatis dapat:

| Method | Endpoint | Fungsi |
|---|---|---|
| `POST` | `/api/v1/tags` | Create |
| `GET` | `/api/v1/tags` | List (paginated) |
| `GET` | `/api/v1/tags/:id` | Get by ID |
| `PUT` | `/api/v1/tags/:id` | Update |
| `DELETE` | `/api/v1/tags/:id` | Soft delete |

---

### Skenario 2: CRUD + Custom Logic

Ketika butuh endpoint atau logic tambahan. Contoh: module `URL` punya `Shorten` dan `Resolve`.

#### 1. Buat model + DTO (`modules/url/model.go`)

```go
package url

import "github.com/dimas292/url_shortener/pkg/model"

type URL struct {
    model.BaseModel
    OriginalURL string `json:"original_url" gorm:"type:text;not null" binding:"required,url"`
    ShortCode   string `json:"short_code" gorm:"type:varchar(10);uniqueIndex;not null"`
    Clicks      int64  `json:"clicks" gorm:"default:0"`
}

type CreateURLRequest struct {
    OriginalURL string `json:"original_url" binding:"required,url"`
}
```

#### 2. Extend repository — tambah custom query (`modules/url/repository.go`)

```go
package url

import (
    "github.com/dimas292/url_shortener/pkg/repository"
    "gorm.io/gorm"
)

type URLRepository struct {
    *repository.BaseRepository[URL, *URL]
}

// Custom query — tidak ada di BaseRepository
func (r *URLRepository) FindByShortCode(code string) (*URL, error) {
    var url URL
    err := r.DB.Where("short_code = ?", code).First(&url).Error
    return &url, err
}

func (r *URLRepository) IncrementClicks(id uint) error {
    return r.DB.Model(&URL{}).Where("id = ?", id).
        UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error
}
```

#### 3. Extend service — tambah business logic (`modules/url/service.go`)

```go
package url

import "github.com/dimas292/url_shortener/pkg/service"

type URLService struct {
    *service.BaseService[URL, *URL]
    repo *URLRepository
}

func NewURLService(repo *URLRepository) *URLService {
    return &URLService{
        BaseService: service.NewBaseService[URL, *URL](repo.BaseRepository),
        repo:        repo,
    }
}

// Custom business logic
func (s *URLService) Shorten(originalURL string) (*URL, error) {
    // generate short code, create URL...
}

func (s *URLService) Resolve(code string) (*URL, error) {
    // find by code, increment clicks...
}
```

#### 4. Buat module — gabungkan generic CRUD + custom endpoint (`modules/url/router.go`)

```go
package url

import (
    "github.com/dimas292/url_shortener/pkg/handler"
    "github.com/dimas292/url_shortener/pkg/repository"
    "github.com/dimas292/url_shortener/pkg/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type URLModule struct {
    crud    *handler.BaseHandler[URL, *URL]
    service *URLService
}

func NewURLModule(db *gorm.DB) *URLModule {
    db.AutoMigrate(&URL{})

    // Generic CRUD
    repo := repository.NewBaseRepository[URL, *URL](db)
    svc := service.NewBaseService[URL, *URL](repo)
    crudHandler := handler.NewBaseHandler[URL, *URL](svc)

    // Custom layers
    urlRepo := &URLRepository{BaseRepository: repo}
    urlService := NewURLService(urlRepo)

    return &URLModule{crud: crudHandler, service: urlService}
}

func (m *URLModule) RegisterRoutes(rg *gin.RouterGroup) {
    // Generic CRUD
    m.crud.RegisterCRUD(rg.Group("/urls"))

    // Custom endpoints
    rg.POST("/shorten", m.handleShorten)
    rg.GET("/r/:code", m.handleResolve)
}
```

---

## API Response Format

Semua endpoint menggunakan format response yang sama:

### Success Response

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": { ... }
}
```

### Created Response

```json
{
  "status": 201,
  "message": "created successfully",
  "data": { ... }
}
```

### Paginated Response

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": [ ... ],
  "meta": {
    "page": 1,
    "per_page": 10,
    "total": 50,
    "total_page": 5
  }
}
```

### Error Response

```json
{
  "status": 404,
  "message": "not found"
}
```

### Pagination Query Parameters

| Parameter | Default | Min | Max | Description |
|---|---|---|---|---|
| `page` | 1 | 1 | - | Halaman |
| `per_page` | 10 | 1 | 100 | Jumlah data per halaman |

Contoh: `GET /api/v1/tags?page=2&per_page=20`

---

## API Endpoints

### URL Module

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/urls` | Create URL (generic CRUD) |
| `GET` | `/api/v1/urls` | List URLs (paginated) |
| `GET` | `/api/v1/urls/:id` | Get URL by ID |
| `PUT` | `/api/v1/urls/:id` | Update URL |
| `DELETE` | `/api/v1/urls/:id` | Soft delete URL |
| `POST` | `/api/v1/shorten` | Shorten URL (custom) |
| `GET` | `/api/v1/r/:code` | Redirect to original URL (custom) |

### Tag Module

| Method | Endpoint | Description |
|---|---|---|
| `POST` | `/api/v1/tags` | Create tag |
| `GET` | `/api/v1/tags` | List tags (paginated) |
| `GET` | `/api/v1/tags/:id` | Get tag by ID |
| `PUT` | `/api/v1/tags/:id` | Update tag |
| `DELETE` | `/api/v1/tags/:id` | Soft delete tag |

---

## Quick Reference

```
Tambah module pure CRUD:
  1. modules/<name>/model.go     → define struct (embed model.BaseModel)
  2. modules/<name>/module.go    → wire repo → service → handler, RegisterRoutes
  3. cmd/main.go                 → srv.RegisterModules(...)

Tambah module dengan custom logic:
  1. modules/<name>/model.go     → define struct + DTOs
  2. modules/<name>/repository.go → extend BaseRepository (custom queries)
  3. modules/<name>/service.go    → extend BaseService (business logic)
  4. modules/<name>/router.go     → BaseHandler CRUD + custom handlers
  5. cmd/main.go                  → srv.RegisterModules(...)
```
