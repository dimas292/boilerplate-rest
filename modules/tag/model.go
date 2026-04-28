package tag

import (
	"github.com/dimas292/url_shortener/pkg/model"
)

// Tag is the domain model for URL tags/categories.
type Tag struct {
	model.BaseModel
	Name  string `json:"name" gorm:"type:varchar(50);uniqueIndex;not null" binding:"required"`
	Color string `json:"color" gorm:"type:varchar(7);default:'#000000'" binding:"omitempty,hexcolor"`
}
