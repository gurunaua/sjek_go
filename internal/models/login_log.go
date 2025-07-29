package models

import (
	"time"
	"github.com/google/uuid"
)

type LoginLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	UserID    uuid.UUID `json:"user_id" gorm:"type:uuid;not null"`
	Username  string    `json:"username" gorm:"not null"`
	Email     string    `json:"email"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
	LoginTime time.Time `json:"login_time" gorm:"default:CURRENT_TIMESTAMP"`
	Status    string    `json:"status" gorm:"type:varchar(20);default:'SUCCESS'"` // SUCCESS, FAILED
	Message   string    `json:"message,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Relationship
	User User `json:"user,omitempty" gorm:"foreignKey:UserID"`
}

const (
	LoginStatusSuccess = "SUCCESS"
	LoginStatusFailed  = "FAILED"
)