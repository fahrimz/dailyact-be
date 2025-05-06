package models

import (
	"time"
)

type AppFeedback struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Feedback  string    `json:"feedback" gorm:"not null"`
	CreatedAt time.Time `json:"created_at"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	User      User      `json:"user" gorm:"foreignKey:UserID"`
}
