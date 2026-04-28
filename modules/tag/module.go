package tag

import (
	"github.com/dimas292/url_shortener/pkg/handler"
	"github.com/dimas292/url_shortener/pkg/repository"
	"github.com/dimas292/url_shortener/pkg/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TagModule implements router.Module for the Tag feature.
type TagModule struct {
	handler *handler.BaseHandler[Tag, *Tag]
}

// NewTagModule wires up a full CRUD module for Tag.
func NewTagModule(db *gorm.DB) *TagModule {
	db.AutoMigrate(&Tag{})

	repo := repository.NewBaseRepository[Tag, *Tag](db)
	svc := service.NewBaseService[Tag, *Tag](repo)
	h := handler.NewBaseHandler[Tag, *Tag](svc)

	return &TagModule{handler: h}
}

// RegisterRoutes registers CRUD routes under /tags.
func (m *TagModule) RegisterRoutes(rg *gin.RouterGroup) {
	m.handler.RegisterCRUD(rg.Group("/tags"))
}
