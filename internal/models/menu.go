package models

import (
	"time"
	"github.com/google/uuid"
)

type Menu struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
	Name        string    `json:"name" gorm:"not null"`
	Path        string    `json:"path" gorm:"not null"`
	Icon        string    `json:"icon,omitempty"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" gorm:"type:uuid"`
	Sequence    int       `json:"sequence" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	
	// Relationships
	Parent   *Menu   `json:"parent,omitempty" gorm:"foreignKey:ParentID"`
	Children []Menu  `json:"children,omitempty" gorm:"foreignKey:ParentID"`
	Roles    []Role  `json:"roles,omitempty" gorm:"many2many:map_role_menu;"`
}