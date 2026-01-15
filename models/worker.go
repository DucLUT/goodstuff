package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Worker struct {
	ID           uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	UserID       uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex" json:"user_id"`
	User         User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Bio          string         `gorm:"type:text" json:"bio"`
	HourlyRate   float64        `gorm:"not null;default:0" json:"hourly_rate"`
	Rating       float64        `gorm:"default:0" json:"rating"`
	TotalJobs    int            `gorm:"default:0" json:"total_jobs"`
	TotalReviews int            `gorm:"default:0" json:"total_reviews"`
	IsVerified   bool           `gorm:"default:false" json:"is_verified"`
	IsAvailable  bool           `gorm:"default:true" json:"is_available"`
	ServiceAreas string         `gorm:"type:text" json:"service_areas"` // JSON array of areas
	WorkingHours string         `gorm:"type:text" json:"working_hours"` // JSON schedule
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

func (w *Worker) BeforeCreate(tx *gorm.DB) error {
	if w.ID == uuid.Nil {
		w.ID = uuid.New()
	}
	return nil
}
