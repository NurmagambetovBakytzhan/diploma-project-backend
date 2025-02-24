// Package postgres implements postgres connection.
package postgres

import (
	"fmt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
	"tourism-backend/config"
	"tourism-backend/internal/entity"
)

const (
	_defaultMaxPoolSize  = 1
	_defaultConnAttempts = 10
	_defaultConnTimeout  = time.Second
)

// Postgres -.
type Postgres struct {
	Conn *gorm.DB
}

// New -.
func New(url string, opts ...Option) (*Postgres, error) {

	db, err := gorm.Open(postgres.Open(url), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	pg := &Postgres{
		Conn: db,
	}

	return pg, nil
}

func (p *Postgres) Connect(cfg *config.Config) error {
	conn, err := gorm.Open(postgres.Open(cfg.URL), &gorm.Config{})
	if err != nil {
		return err
	}

	p.Conn = conn
	return nil
}

func (p *Postgres) Migrate() error {
	err := p.Conn.AutoMigrate(
		&entity.Tour{},
		&entity.Image{},
		&entity.Video{},
		&entity.User{},
		&entity.TourEvent{},
		&entity.Purchase{},
		&entity.TourCategory{},
		&entity.TourLocation{},
	)
	if err != nil {
		fmt.Errorf("Migrating entities to Postgres - err: %w", err)
		return err
	}
	return nil
}
