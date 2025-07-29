package models

import (
    "time"
    "github.com/google/uuid"
)

type UserType string
type UserStatus string

const (
    UserTypeAdmin  UserType = "ADMIN"
    UserTypeDriver UserType = "DRIVER"
)

const (
    UserStatusActive   UserStatus = "ACTIVE"
    UserStatusInactive UserStatus = "INACTIVE"
)

type User struct {
    ID             uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:uuid_generate_v4()"`
    Username       string     `json:"username" gorm:"unique;not null"`
    Email          string     `json:"email" gorm:"type:varchar(1500);unique"`
    Password       string     `json:"password,omitempty" gorm:"not null"`
    Type           UserType   `json:"type" gorm:"type:varchar(20);default:'DRIVER'"`
    Status         UserStatus `json:"status" gorm:"type:varchar(20);default:'ACTIVE'"`
    ActivatedDate  *time.Time `json:"activated_date,omitempty"`
    InactiveDate   *time.Time `json:"inactive_date,omitempty"`
    CreatedAt      time.Time  `json:"created_at"`
    UpdatedAt      time.Time  `json:"updated_at"`
    Roles          []Role     `json:"roles,omitempty" gorm:"many2many:map_user_role;"`
}