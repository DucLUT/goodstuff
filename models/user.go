package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleWorker   UserRole = "worker"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email          string         `gorm:"uniqueIndex;not null" json:"email"`
	Phone          string         `gorm:"uniqueIndex;not null" json:"phone"`
	PasswordHashed string         `gorm:"not null" json:"-"`
	Name           string         `gorm:"not null" json:"name"`
	Role           UserRole       `gorm:"type:varchar(20);not null;default:'customer'" json:"role"`
	Avatar         string         `json:"avatar,omitempty"`
	Address        string         `json:"address,omitempty"`
	IsActive       bool           `gorm:"default:true" json:"is_active"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}
