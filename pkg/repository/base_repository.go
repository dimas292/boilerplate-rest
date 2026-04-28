package repository

import (
	"fmt"

	"github.com/dimas292/url_shortener/pkg/model"
	"gorm.io/gorm"
)

// Repository defines the generic CRUD contract.
type Repository[T any, PT model.ModelPtr[T]] interface {
	Create(entity PT) error
	FindByID(id string) (PT, error)
	FindAll(page, perPage int) ([]T, int64, error)
	Update(entity PT) error
	Delete(id string) error
}

// BaseRepository is a generic GORM-backed repository implementing CRUD operations.
// T is the value type (e.g. URL), PT is *T satisfying Model.
type BaseRepository[T any, PT model.ModelPtr[T]] struct {
	DB *gorm.DB
}

// NewBaseRepository creates a new BaseRepository for the given model type.
func NewBaseRepository[T any, PT model.ModelPtr[T]](db *gorm.DB) *BaseRepository[T, PT] {
	return &BaseRepository[T, PT]{DB: db}
}

// Create inserts a new record.
func (r *BaseRepository[T, PT]) Create(entity PT) error {
	if err := r.DB.Create(entity).Error; err != nil {
		return fmt.Errorf("repository create: %w", err)
	}
	return nil
}

// FindByID retrieves a single record by primary key.
func (r *BaseRepository[T, PT]) FindByID(id string) (PT, error) {
	entity := PT(new(T))
	if err := r.DB.First(entity, "id = ?", id).Error; err != nil {
		return nil, fmt.Errorf("repository find by id: %w", err)
	}
	return entity, nil
}

// FindAll retrieves paginated records and total count.
func (r *BaseRepository[T, PT]) FindAll(page, perPage int) ([]T, int64, error) {
	var entities []T
	var total int64

	if err := r.DB.Model(PT(new(T))).Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("repository count: %w", err)
	}

	offset := (page - 1) * perPage
	if err := r.DB.Offset(offset).Limit(perPage).Find(&entities).Error; err != nil {
		return nil, 0, fmt.Errorf("repository find all: %w", err)
	}

	return entities, total, nil
}

// Update saves changes to an existing record.
func (r *BaseRepository[T, PT]) Update(entity PT) error {
	if err := r.DB.Save(entity).Error; err != nil {
		return fmt.Errorf("repository update: %w", err)
	}
	return nil
}

// Delete soft-deletes a record by primary key.
func (r *BaseRepository[T, PT]) Delete(id string) error {
	entity := PT(new(T))
	if err := r.DB.Where("id = ?", id).Delete(entity).Error; err != nil {
		return fmt.Errorf("repository delete: %w", err)
	}
	return nil
}
