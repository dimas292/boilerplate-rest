package model

import (
	"time"

	"gorm.io/gorm"
)

// Model is the generic constraint that all domain models must satisfy.
// Any struct embedding BaseModel automatically satisfies this interface.
type Model interface {
	GetID() uint
	SetID(id uint)
}

// ModelPtr is a constraint ensuring T is a pointer to a type embedding BaseModel.
// Usage: BaseRepository[T, PT ModelPtr[T]] ensures *YourStruct satisfies the interface.
type ModelPtr[T any] interface {
	Model
	*T
}

// BaseModel provides common fields for all GORM models.
// Embed this in your domain structs to get ID, timestamps, and soft-delete.
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}

// GetID returns the model's primary key.
func (b *BaseModel) GetID() uint {
	return b.ID
}

// SetID sets the model's primary key.
func (b *BaseModel) SetID(id uint) {
	b.ID = id
}
