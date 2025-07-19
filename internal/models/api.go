package models

import (
    "time"
    "github.com/google/uuid"
)

type API struct {
    ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Path        string    `json:"path" gorm:"not null"`
    Method      string    `json:"method" gorm:"not null"`
    Description string    `json:"description"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
    Roles       []Role    `json:"roles,omitempty" gorm:"many2many:map_role_api;"`
}