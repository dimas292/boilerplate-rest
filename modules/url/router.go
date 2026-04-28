package url

import (
	"net/http"

	"github.com/dimas292/url_shortener/pkg/handler"
	"github.com/dimas292/url_shortener/pkg/repository"
	"github.com/dimas292/url_shortener/pkg/response"
	"github.com/dimas292/url_shortener/pkg/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// URLModule implements router.Module for the URL feature.
// It combines generic CRUD (from BaseHandler) with custom endpoints.
type URLModule struct {
	crud    *handler.BaseHandler[URL, *URL]
	service *URLService
}

// NewURLModule wires up the URL module's dependencies.
func NewURLModule(db *gorm.DB) *URLModule {
	// Auto-migrate
	db.AutoMigrate(&URL{})

	// Base layers (generic CRUD)
	repo := repository.NewBaseRepository[URL, *URL](db)
	svc := service.NewBaseService[URL, *URL](repo)
	crudHandler := handler.NewBaseHandler[URL, *URL](svc)

	// Domain-specific layers
	urlRepo := &URLRepository{BaseRepository: repo}
	urlService := NewURLService(urlRepo)

	return &URLModule{
		crud:    crudHandler,
		service: urlService,
	}
}

// RegisterRoutes registers all URL routes.
//
// CRUD (generic):
//
//	POST   /urls        → Create
//	GET    /urls        → FindAll (paginated)
//	GET    /urls/:id    → FindByID
//	PUT    /urls/:id    → Update
//	DELETE /urls/:id    → Delete
//
// Custom:
//
//	POST   /shorten     → Shorten (generate short code)
//	GET    /r/:code     → Resolve (redirect to original URL)
func (m *URLModule) RegisterRoutes(rg *gin.RouterGroup) {
	// Generic CRUD routes
	urls := rg.Group("/urls")
	m.crud.RegisterCRUD(urls)

	// Custom endpoints
	rg.POST("/shorten", m.handleShorten)
	rg.GET("/r/:code", m.handleResolve)
}

// handleShorten handles POST /shorten — create a shortened URL.
func (m *URLModule) handleShorten(c *gin.Context) {
	var req CreateURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	url, err := m.service.Shorten(req.OriginalURL)
	if err != nil {
		response.Error(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.Created(c, "url shortened successfully", url)
}

// handleResolve handles GET /r/:code — redirect to the original URL.
func (m *URLModule) handleResolve(c *gin.Context) {
	code := c.Param("code")

	url, err := m.service.Resolve(code)
	if err != nil {
		response.Error(c, http.StatusNotFound, "short url not found")
		return
	}

	c.Redirect(http.StatusMovedPermanently, url.OriginalURL)
}
