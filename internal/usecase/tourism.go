package usecase

import (
	"fmt"
	"github.com/google/uuid"
	"mime/multipart"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase/repo"
)

// TranslationUseCase -.
type TourismUseCase struct {
	repo *repo.TourismRepo
}

// NewTourismUseCase -.
func NewTourismUseCase(r *repo.TourismRepo) *TourismUseCase {
	return &TourismUseCase{
		repo: r,
	}
}

func (r *TourismUseCase) GetTourLocationByID(tourLocationID uuid.UUID) (*entity.TourLocation, error) {
	return r.repo.GetTourLocationByID(tourLocationID)
}

func (r *TourismUseCase) CreateTourLocation(tourLocation *entity.CreateTourLocationDTO) (*entity.TourLocation, error) {
	return r.repo.CreateTourLocation(tourLocation)
}

func (r *TourismUseCase) CreateTourCategory(tourCategory *entity.CreateTourCategoryDTO) (*entity.TourCategory, error) {
	return r.repo.CreateTourCategory(tourCategory)
}

func (r *TourismUseCase) GetAllCategories() ([]entity.Category, error) {
	categories, err := r.repo.GetAllCategories()
	if err != nil {
		return nil, err
	}
	return categories, nil
}

func (t *TourismUseCase) CreatePurchase(purchase *entity.Purchase) (*entity.Purchase, error) {
	return t.repo.CreatePurchase(purchase)
}

func (t *TourismUseCase) PayTourEvent(purchase *entity.Purchase) error {
	return t.repo.PayTourEvent(purchase)
}

func (t *TourismUseCase) CheckTourOwner(tourID uuid.UUID, userID uuid.UUID) bool {
	return t.repo.CheckTourOwner(tourID, userID)
}

func (t *TourismUseCase) CreateTourEvent(tourEvent *entity.TourEvent) (*entity.TourEvent, error) {
	tourEvent, err := t.repo.CreateTourEvent(tourEvent)
	if err != nil {
		return nil, fmt.Errorf("create tour event: %w", err)
	}
	return tourEvent, nil
}

func (t *TourismUseCase) CreateTour(tour *entity.Tour, imageFiles []*multipart.FileHeader, videoFiles []*multipart.FileHeader) (*entity.Tour, error) {
	tour, err := t.repo.CreateTour(tour, imageFiles, videoFiles)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

func (t *TourismUseCase) GetTourByID(id string) (*entity.Tour, error) {
	tour, err := t.repo.GetTourByID(id)
	if err != nil {
		return nil, err
	}
	return tour, nil
}

func (t *TourismUseCase) GetTours() ([]entity.Tour, error) {
	tours, err := t.repo.GetTours()
	if err != nil {
		return nil, err
	}
	return tours, nil
}
