package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Service struct {
	ID           uuid.UUID       `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CategoryID   uuid.UUID       `gorm:"type:uuid;not null" json:"category_id"`
	Category     ServiceCategory `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Name         string          `gorm:"not null" json:"name"`
	Description  string          `gorm:"type:text" json:"description"`
	BasePrice    float64         `gorm:"not null" json:"base_price"`
	PricePerHour float64         `gorm:"not null" json:"price_per_hour"`
	MinDuration  int             `gorm:"default:60" json:"min_duration"`  // minutes
	MaxDuration  int             `gorm:"default:480" json:"max_duration"` // minutes
	Image        string          `json:"image,omitempty"`
	IsActive     bool            `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	DeletedAt    gorm.DeletedAt  `gorm:"index" json:"-"`
}

func (s *Service) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}
