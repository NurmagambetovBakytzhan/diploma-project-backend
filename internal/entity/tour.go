package entity

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Tour struct {
	gorm.Model
	ID          uuid.UUID   `json:"ID" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Description string      `json:"description"`
	Route       string      `json:"route"`
	TourImages  []Image     `json:"tour_images" gorm:"foreignKey:TourID;references:ID"`
	TourVideos  []Video     `json:"tour_videos" gorm:"foreignKey:TourID;references:ID"`
	TourEvents  []TourEvent `json:"tour_events" gorm:"foreignKey:TourID;references:ID"`
	OwnerID     uuid.UUID   `json:"owner_id" gorm:"type:uuid;index"`
}

type Image struct {
	gorm.Model
	ID       uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID   uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	ImageURL string    `json:"image_url"`
}

type Video struct {
	gorm.Model
	ID       uuid.UUID `json:"ID" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	TourID   uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	VideoURL string    `json:"video_url"`
}
