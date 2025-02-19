package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"time"
)

type TourEvent struct {
	gorm.Model
	ID     uuid.UUID `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Tour   Tour
	Date   time.Time `json:"Date" gorm:"not null"`
	Price  float64   `json:"Price" gorm:"not null"`
	Place  string    `json:"Place" gorm:"not null"`
	Amount float64   `json:"Amount" gorm:"not null"`
	TourID uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
}
