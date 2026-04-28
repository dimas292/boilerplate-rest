package url

import (
	"fmt"

	"github.com/dimas292/url_shortener/pkg/repository"
	"gorm.io/gorm"
)

// URLRepository extends BaseRepository with URL-specific queries.
type URLRepository struct {
	*repository.BaseRepository[URL, *URL]
}

// FindByShortCode retrieves a URL by its short code.
func (r *URLRepository) FindByShortCode(code string) (*URL, error) {
	var url URL
	if err := r.DB.Where("short_code = ?", code).First(&url).Error; err != nil {
		return nil, fmt.Errorf("repository find by short code: %w", err)
	}
	return &url, nil
}

// IncrementClicks atomically increments the click counter.
func (r *URLRepository) IncrementClicks(id uint) error {
	if err := r.DB.Model(&URL{}).Where("id = ?", id).
		UpdateColumn("clicks", gorm.Expr("clicks + 1")).Error; err != nil {
		return fmt.Errorf("repository increment clicks: %w", err)
	}
	return nil
}
