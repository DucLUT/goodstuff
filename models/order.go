package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BookingStatus string

const (
	StatusPending    BookingStatus = "pending"
	StatusConfirmed  BookingStatus = "confirmed"
	StatusInProgress BookingStatus = "in_progress"
	StatusCompleted  BookingStatus = "completed"
	StatusCancelled  BookingStatus = "cancelled"
)

type Booking struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	CustomerID    uuid.UUID      `gorm:"type:uuid;not null" json:"customer_id"`
	Customer      User           `gorm:"foreignKey:CustomerID" json:"customer,omitempty"`
	WorkerID      *uuid.UUID     `gorm:"type:uuid" json:"worker_id,omitempty"`
	Worker        *Worker        `gorm:"foreignKey:WorkerID" json:"worker,omitempty"`
	ServiceID     uuid.UUID      `gorm:"type:uuid;not null" json:"service_id"`
	Service       Service        `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Status        BookingStatus  `gorm:"type:varchar(20);not null;default:'pending'" json:"status"`
	ScheduledAt   time.Time      `gorm:"not null" json:"scheduled_at"`
	DurationHours float64        `gorm:"not null" json:"duration_hours"`
	Address       string         `gorm:"not null" json:"address"`
	Notes         string         `gorm:"type:text" json:"notes,omitempty"`
	TotalPrice    float64        `gorm:"not null" json:"total_price"`
	StartedAt     *time.Time     `json:"started_at,omitempty"`
	CompletedAt   *time.Time     `json:"completed_at,omitempty"`
	CancelledAt   *time.Time     `json:"cancelled_at,omitempty"`
	CancelReason  string         `gorm:"type:text" json:"cancel_reason,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (b *Booking) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}
