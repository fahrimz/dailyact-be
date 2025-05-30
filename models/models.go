package models

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null;unique"`
	Description string     `json:"description"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	Activities  []Activity `json:"activities,omitempty" gorm:"foreignKey:CategoryID"`
}

type Activity struct {
	ID          uint            `json:"id" gorm:"primaryKey"`
	Date        time.Time       `json:"date" gorm:"not null;index"`
	StartTime   time.Time       `json:"start_time" gorm:"not null"`
	EndTime     time.Time       `json:"end_time" gorm:"not null"`
	Duration    int             `json:"duration" gorm:"not null"` // in second
	Description EncryptedString `json:"description" gorm:"type:text;not null"`
	Notes       EncryptedString `json:"notes" gorm:"type:text"`
	CategoryID  uint            `json:"category_id" gorm:"not null"`
	Category    Category        `json:"category" gorm:"foreignKey:CategoryID"`
	UserID      uint            `json:"user_id" gorm:"not null"`
	User        User            `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func (a *Activity) BeforeCreate(tx *gorm.DB) (err error) {
	// set date and duration automatically by calculating start and end time
	a.Date = a.StartTime.UTC().Truncate(24 * time.Hour)
	a.Duration = int(a.EndTime.Sub(a.StartTime).Seconds())

	if a.Duration < 0 {
		return errors.New("duration cannot be negative")
	}
	return nil
}

// Add BeforeUpdate hook to handle updates
func (a *Activity) BeforeUpdate(tx *gorm.DB) (err error) {
    // set date and duration automatically by calculating start and end time
    a.Date = a.StartTime.UTC().Truncate(24 * time.Hour)
    a.Duration = int(a.EndTime.Sub(a.StartTime).Seconds())

    if a.Duration < 0 {
        return errors.New("duration cannot be negative")
    }
    return nil
}
