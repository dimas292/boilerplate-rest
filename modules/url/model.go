package url

import (
	"github.com/dimas292/url_shortener/pkg/model"
)

// URL is the domain model for shortened URLs.
type URL struct {
	model.BaseModel
	OriginalURL string `json:"original_url" gorm:"type:text;not null" binding:"required,url"`
	ShortCode   string `json:"short_code" gorm:"type:varchar(10);uniqueIndex;not null"`
	Clicks      int64  `json:"clicks" gorm:"default:0"`
}

// CreateURLRequest is the DTO for creating a new shortened URL.
type CreateURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
}

// UpdateURLRequest is the DTO for updating a shortened URL.
type UpdateURLRequest struct {
	OriginalURL string `json:"original_url" binding:"required,url"`
}
