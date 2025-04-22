package models

import (
	"time"
)

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	Email         string    `json:"email" gorm:"unique;not null"`
	Name          string    `json:"name" gorm:"not null"`
	Picture       string    `json:"picture"`
	GoogleID      string    `json:"google_id" gorm:"unique;not null"`
	Role          Role      `json:"role" gorm:"type:varchar(10);default:user"`
	Activities    []Activity `json:"activities,omitempty" gorm:"foreignKey:UserID"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	LastLoginAt   time.Time `json:"last_login_at"`
}
