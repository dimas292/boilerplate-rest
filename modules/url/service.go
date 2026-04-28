package url

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/dimas292/url_shortener/pkg/service"
)

// URLService extends BaseService with URL-specific business logic.
type URLService struct {
	*service.BaseService[URL, *URL]
	repo *URLRepository
}

// NewURLService creates a new URLService.
func NewURLService(repo *URLRepository) *URLService {
	return &URLService{
		BaseService: service.NewBaseService[URL, *URL](repo.BaseRepository),
		repo:        repo,
	}
}

// Shorten creates a new shortened URL with a generated short code.
func (s *URLService) Shorten(originalURL string) (*URL, error) {
	code, err := generateShortCode(6)
	if err != nil {
		return nil, fmt.Errorf("service shorten: %w", err)
	}

	url := &URL{
		OriginalURL: originalURL,
		ShortCode:   code,
	}

	if err := s.Repo.Create(url); err != nil {
		return nil, err
	}

	return url, nil
}

// Resolve finds the original URL by short code and increments clicks.
func (s *URLService) Resolve(code string) (*URL, error) {
	url, err := s.repo.FindByShortCode(code)
	if err != nil {
		return nil, fmt.Errorf("service resolve: %w", err)
	}

	if err := s.repo.IncrementClicks(url.ID); err != nil {
		return nil, fmt.Errorf("service resolve increment: %w", err)
	}

	url.Clicks++
	return url, nil
}

// generateShortCode creates a URL-safe random string of the given length.
func generateShortCode(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate short code: %w", err)
	}

	code := base64.URLEncoding.EncodeToString(bytes)
	code = strings.TrimRight(code, "=")

	if len(code) > length {
		code = code[:length]
	}

	return code, nil
}
