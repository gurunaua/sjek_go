package models

import (
    "time"
    "github.com/google/uuid"
)

type User struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Username  string    `json:"username" gorm:"unique;not null"`
    Password  string    `json:"password,omitempty" gorm:"not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Roles     []Role    `json:"roles,omitempty" gorm:"many2many:map_user_role;"`
}