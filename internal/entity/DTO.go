package entity

import (
	"github.com/google/uuid"
	"time"
)

type CreateTourDTO struct {
	Description string `json:"description"`
	Route       string `json:"route"`
	Price       int    `json:"price"`
	//TourImages  []Image   `json:"tour_images"`
	//TourVideos  []Video   `json:"tour_videos"`
}

type CreateUserDTO struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Role     string `json:"role"` // Optional: "user" (default) or "admin"
}

type LoginUserDTO struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type CreateTourEventDTO struct {
	Date   time.Time `json:"date" gorm:"not null"`
	Price  float64   `json:"price" gorm:"not null"`
	Place  string    `json:"place" gorm:"not null"`
	TourID uuid.UUID `json:"tour_id" gorm:"type:uuid;index"`
	Amount float64   `json:"amount" gorm:"not null"`
}
