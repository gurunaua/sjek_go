package models

import (
    "time"
    "github.com/google/uuid"
)

type Role struct {
    ID        uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Name      string    `json:"name" gorm:"unique;not null"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
    Users     []User    `json:"users,omitempty" gorm:"many2many:map_user_role;"`
    APIs      []API     `json:"apis,omitempty" gorm:"many2many:map_role_api;"`
    Menus     []Menu    `json:"menus,omitempty" gorm:"many2many:map_role_menu;"`
}