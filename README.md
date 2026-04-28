# Go Modular CRUD Template

Template project Go dengan **modular architecture**, **Go generics** untuk CRUD tanpa boilerplate, dan **JWT authentication & authorization**.

## Tech Stack

- **[Gin](https://github.com/gin-gonic/gin)** — HTTP framework
- **[GORM](https://gorm.io)** — ORM (PostgreSQL)
- **[go-redis](https://github.com/redis/go-redis)** — Redis client
- **[golang-jwt](https://github.com/golang-jwt/jwt)** — JWT authentication
- **[bcrypt](https://pkg.go.dev/golang.org/x/crypto/bcrypt)** — Password hashing
- **Go Generics** — Reusable base layers

## Project Structure

```
project/
├── cmd/
│   └── main.go                        # Entry point
├── config.yml                         # App config (gitignored)
├── example.config.yml                 # Example config
├── modules/
│   └── auth/                          # Auth module (built-in)
│       ├── model.go                   # User model + DTOs
│       ├── service.go                 # Register, Login, Profile
│       └── module.go                  # Routes + wiring
├── pkg/
│   ├── auth/
│   │   ├── base_jwt.go               # JWT generation & validation
│   │   └── base_middleware.go         # Auth & role middleware
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
│   │   └── base_response.go          # Standardized JSON response
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

1. Clone template:

```bash
git clone https://github.com/dimas292/url_shortener.git -b template my-project
cd my-project
```

2. Ganti module name di `go.mod`:

```
module github.com/username/my-project
```

3. Find & replace semua import path:

```bash
grep -rl "github.com/dimas292/url_shortener" . --include="*.go" | xargs sed -i '' 's|github.com/dimas292/url_shortener|github.com/username/my-project|g'
```

4. Copy dan edit config:

```bash
cp example.config.yml config.yml
```

```yaml
app:
  name: myapp
  port: ":4444"
  jwt:
    secret: "ganti-dengan-secret-key-yang-kuat"
    expiration: 24  # jam
  db:
    postgres:
      dbhost: localhost
      dbuser: postgres
      dbpassword: postgres
      dbname: myapp
    redis:
      host: localhost
      port: 6379
```

5. Run:

```bash
go mod tidy
go run cmd/main.go
```

---

## Authentication & Authorization

### Built-in Auth Endpoints

| Method | Endpoint | Auth | Description |
|---|---|---|---|
| `POST` | `/api/v1/auth/register` | - | Register user baru |
| `POST` | `/api/v1/auth/login` | - | Login, dapat JWT token |
| `GET` | `/api/v1/auth/profile` | Bearer | Get profile user saat ini |

### Register

```bash
curl -X POST http://localhost:4444/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name": "John", "email": "john@example.com", "password": "secret123"}'
```

Response:
```json
{
  "status": 201,
  "message": "registered successfully",
  "data": {
    "user": {
      "id": 1,
      "name": "John",
      "email": "john@example.com",
      "role": "user"
    },
    "token": "eyJhbGciOiJIUzI1NiIs..."
  }
}
```

### Login

```bash
curl -X POST http://localhost:4444/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "john@example.com", "password": "secret123"}'
```

### Menggunakan Token

Tambahkan header `Authorization: Bearer <token>` di setiap request ke endpoint yang dilindungi:

```bash
curl http://localhost:4444/api/v1/auth/profile \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIs..."
```

### Melindungi Routes di Module Lain

#### Semua routes butuh login (authentication)

```go
func (m *ProductModule) RegisterRoutes(rg *gin.RouterGroup) {
    products := rg.Group("/products")
    products.Use(auth.AuthMiddleware(m.jwtService))  // semua route dilindungi
    {
        m.crud.RegisterCRUD(products)
    }
}
```

#### Sebagian route public, sebagian protected

```go
func (m *ProductModule) RegisterRoutes(rg *gin.RouterGroup) {
    products := rg.Group("/products")

    // Public
    products.GET("", m.crud.FindAll)
    products.GET("/:id", m.crud.FindByID)

    // Protected (butuh login)
    protected := products.Group("")
    protected.Use(auth.AuthMiddleware(m.jwtService))
    {
        protected.POST("", m.crud.Create)
        protected.PUT("/:id", m.crud.Update)
        protected.DELETE("/:id", m.crud.Delete)
    }
}
```

#### Role-based authorization

```go
func (m *AdminModule) RegisterRoutes(rg *gin.RouterGroup) {
    admin := rg.Group("/admin")
    admin.Use(auth.AuthMiddleware(m.jwtService))       // harus login
    admin.Use(auth.RoleMiddleware("admin"))             // harus role admin
    {
        // hanya user dengan role "admin" yang bisa akses
        admin.GET("/dashboard", m.handleDashboard)
    }
}
```

#### Multiple roles

```go
// user ATAU admin bisa akses
rg.Use(auth.RoleMiddleware("admin", "user"))

// hanya admin dan moderator
rg.Use(auth.RoleMiddleware("admin", "moderator"))
```

### Helper Functions

Ambil info user dari gin context di handler manapun (setelah AuthMiddleware):

```go
func (m *MyModule) handleSomething(c *gin.Context) {
    userID := auth.GetUserID(c)   // uint
    email  := auth.GetEmail(c)    // string
    role   := auth.GetRole(c)     // string
}
```

### JWT Claims

Token berisi informasi berikut:

| Field | Type | Description |
|---|---|---|
| `user_id` | uint | ID user |
| `email` | string | Email user |
| `role` | string | Role user (default: "user") |
| `exp` | timestamp | Token expiration |
| `iat` | timestamp | Token issued at |

---

## Cara Menambahkan Module

### Skenario 1: Pure CRUD (tanpa custom logic)

Cukup **2 file**, langsung dapat 5 endpoint CRUD.

#### 1. Buat model — `modules/product/model.go`

```go
package product

import "github.com/username/my-project/pkg/model"

type Product struct {
    model.BaseModel
    Name  string  `json:"name" gorm:"not null" binding:"required"`
    Price float64 `json:"price" gorm:"not null" binding:"required,min=0"`
    Stock int     `json:"stock" gorm:"default:0"`
}
```

#### 2. Buat module — `modules/product/module.go`

```go
package product

import (
    "github.com/username/my-project/pkg/handler"
    "github.com/username/my-project/pkg/repository"
    "github.com/username/my-project/pkg/service"
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"
)

type ProductModule struct {
    handler *handler.BaseHandler[Product, *Product]
}

func NewProductModule(db *gorm.DB) *ProductModule {
    db.AutoMigrate(&Product{})

    repo := repository.NewBaseRepository[Product, *Product](db)
    svc := service.NewBaseService[Product, *Product](repo)
    h := handler.NewBaseHandler[Product, *Product](svc)

    return &ProductModule{handler: h}
}

func (m *ProductModule) RegisterRoutes(rg *gin.RouterGroup) {
    m.handler.RegisterCRUD(rg.Group("/products"))
}
```

#### 3. Register di `cmd/main.go`

```go
srv.RegisterModules(
    authmodule.NewAuthModule(srv.DB, srv.JWT),
    productmodule.NewProductModule(srv.DB),
)
```

**Selesai!** Otomatis dapat:

| Method | Endpoint |
|---|---|
| `POST` | `/api/v1/products` |
| `GET` | `/api/v1/products` |
| `GET` | `/api/v1/products/:id` |
| `PUT` | `/api/v1/products/:id` |
| `DELETE` | `/api/v1/products/:id` |

---

### Skenario 2: CRUD + Auth + Custom Logic

#### 1. Model + DTO

```go
type Product struct {
    model.BaseModel
    Name     string  `json:"name" binding:"required"`
    Price    float64 `json:"price" binding:"required"`
    Category string  `json:"category"`
}
```

#### 2. Custom Repository

```go
type ProductRepository struct {
    *repository.BaseRepository[Product, *Product]
}

func (r *ProductRepository) FindByCategory(category string) ([]Product, error) {
    var products []Product
    err := r.DB.Where("category = ?", category).Find(&products).Error
    return products, err
}
```

#### 3. Custom Service

```go
type ProductService struct {
    *service.BaseService[Product, *Product]
    repo *ProductRepository
}

func (s *ProductService) GetByCategory(category string) ([]Product, error) {
    return s.repo.FindByCategory(category)
}
```

#### 4. Module dengan Auth

```go
type ProductModule struct {
    crud       *handler.BaseHandler[Product, *Product]
    service    *ProductService
    jwtService *pkgauth.JWTService
}

func NewProductModule(db *gorm.DB, jwt *pkgauth.JWTService) *ProductModule {
    // ... wire dependencies
    return &ProductModule{crud: h, service: svc, jwtService: jwt}
}

func (m *ProductModule) RegisterRoutes(rg *gin.RouterGroup) {
    products := rg.Group("/products")

    // Public: siapa saja bisa lihat
    products.GET("", m.crud.FindAll)
    products.GET("/:id", m.crud.FindByID)

    // Protected: harus login
    protected := products.Group("")
    protected.Use(pkgauth.AuthMiddleware(m.jwtService))
    {
        protected.POST("", m.crud.Create)
        protected.PUT("/:id", m.crud.Update)
        protected.DELETE("/:id", m.crud.Delete)
    }
}
```

---

## API Response Format

### Success

```json
{
  "status": 200,
  "message": "retrieved successfully",
  "data": { ... }
}
```

### Paginated

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

### Error

```json
{
  "status": 401,
  "message": "invalid token"
}
```

### Pagination Parameters

| Parameter | Default | Range |
|---|---|---|
| `page` | 1 | min: 1 |
| `per_page` | 10 | 1–100 |

Contoh: `GET /api/v1/products?page=2&per_page=20`

---

## Quick Reference

```
Pure CRUD (2 file):
  1. modules/<name>/model.go     → struct (embed model.BaseModel)
  2. modules/<name>/module.go    → wire repo → service → handler
  3. cmd/main.go                 → srv.RegisterModules(...)

CRUD + Auth (3 file):
  1. modules/<name>/model.go     → struct
  2. modules/<name>/module.go    → wire + auth middleware
  3. cmd/main.go                 → NewModule(srv.DB, srv.JWT)

CRUD + Auth + Custom (4+ file):
  1. modules/<name>/model.go      → struct + DTOs
  2. modules/<name>/repository.go → embed BaseRepository + custom queries
  3. modules/<name>/service.go    → embed BaseService + business logic
  4. modules/<name>/module.go     → BaseHandler CRUD + custom handlers + auth
  5. cmd/main.go                  → srv.RegisterModules(...)
```
