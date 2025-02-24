package repo

import (
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"tourism-backend/internal/entity"
	"tourism-backend/pkg/postgres"
)

const _defaultEntityCap = 64

// TourismRepo -.
type TourismRepo struct {
	PG *postgres.Postgres
}

// New -.
func NewTourismRepo(pg *postgres.Postgres) *TourismRepo {
	return &TourismRepo{pg}
}

func (r *TourismRepo) GetTourLocationByID(tourLocationID uuid.UUID) (*entity.TourLocation, error) {
	var tourLocation entity.TourLocation
	err := r.PG.Conn.Where("tour_id = ?", tourLocationID).First(&tourLocation).Error
	if err != nil {
		return nil, fmt.Errorf("get tour location by id: %w", err)
	}
	return &tourLocation, nil
}

func (r *TourismRepo) CreateTourLocation(tourLocation *entity.CreateTourLocationDTO) (*entity.TourLocation, error) {
	tourLocationEntity := &entity.TourLocation{
		TourID:    tourLocation.TourID,
		Longitude: tourLocation.Longitude,
		Latitude:  tourLocation.Latitude,
	}

	err := r.PG.Conn.Create(&tourLocationEntity).Error
	if err != nil {
		return nil, fmt.Errorf("Tour Location not Found")
	}
	return tourLocationEntity, nil

}

func (r *TourismRepo) GetAllCategories() ([]entity.Category, error) {
	var categories []entity.Category
	err := r.PG.Conn.Find(&categories).Error
	if err != nil {
		return nil, fmt.Errorf("GetAllCategories: %w", err)
	}
	return categories, nil
}

func (r *TourismRepo) CreateTourCategory(tourCategory *entity.CreateTourCategoryDTO) (*entity.TourCategory, error) {

	category := &entity.TourCategory{
		TourID:     tourCategory.TourID,
		CategoryID: tourCategory.CategoryID,
	}

	// Check if the record already exists
	existingCategory := &entity.TourCategory{}
	err := r.PG.Conn.Where("tour_id = ? AND category_id = ?", category.TourID, category.CategoryID).First(existingCategory).Error
	if err == nil {
		return existingCategory, fmt.Errorf("tour Category already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("Tour Category not Found")
	}

	err = r.PG.Conn.Create(&category).Error
	if err != nil {
		return nil, fmt.Errorf("Tour Category or Tour not Found")
	}
	return category, nil
}

func (r *TourismRepo) CreatePurchase(purchase *entity.Purchase) (*entity.Purchase, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Check tour event record in the database
		var tourEvent entity.TourEvent

		// Fetch the tour event to verify conditions before updating
		if err := tx.Where("id = ? AND is_opened = ? AND amount_of_places > 0",
			purchase.TourEventID, true).First(&tourEvent).Error; err != nil {
			return fmt.Errorf("tour event not found or closed: %w", err)
		}

		// Decrease the available places count
		if err := tx.Model(&entity.TourEvent{}).
			Where("id = ?", purchase.TourEventID).
			UpdateColumn("amount_of_places", gorm.Expr("amount_of_places - 1")).Error; err != nil {
			return fmt.Errorf("failed to update amount_of_places: %w", err)
		}

		// Create the purchase record in the database
		if err := tx.Create(purchase).Error; err != nil {
			return fmt.Errorf("create purchase failed: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("create purchase transaction failed: %w", err)
	}

	// Reload purchase with related data
	err = r.PG.Conn.Preload("User").Preload("TourEvent.Tour").
		First(purchase, "id = ?", purchase.ID).Error
	if err != nil {
		return nil, fmt.Errorf("failed to preload purchase data: %w", err)
	}

	return purchase, nil
}

func (r *TourismRepo) PayTourEvent(purchase *entity.Purchase) error {

	result := r.PG.Conn.Model(&entity.Purchase{}).
		Where("id = ? AND status = ?", purchase.ID, "Processing").
		Update("status", "Paid")

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (r *TourismRepo) CheckTourOwner(tourID uuid.UUID, userID uuid.UUID) bool {
	var tourOwnerID string
	err := r.PG.Conn.Table("tours").
		Select("owner_id").
		Where("id = ?", tourID).
		Scan(&tourOwnerID).Error

	if err != nil {
		fmt.Println("Error checking tour owner:", err)
		return false
	}

	// Convert string to UUID
	ownerUUID, err := uuid.Parse(tourOwnerID)
	if err != nil {
		fmt.Println("Error parsing owner UUID:", err)
		return false
	}

	return ownerUUID == userID
}

func (r *TourismRepo) CreateTourEvent(tourEvent *entity.TourEvent) (*entity.TourEvent, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Create the tour record in the database
		var count int64
		if err := tx.Model(&entity.Tour{}).Where("id = ?", tourEvent.TourID).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return fmt.Errorf("tour with id %s does not exist", tourEvent.TourID)
		}

		// Create the tour event record in the database
		if err := tx.Create(&tourEvent).Error; err != nil {
			return err
		}

		return nil
	})
	err = r.PG.Conn.Preload("Tour").First(tourEvent, "id = ?", tourEvent.ID).Error

	if err != nil {
		return nil, err
	}

	return tourEvent, nil
}

func (r *TourismRepo) GetTourByID(tourID string) (*entity.Tour, error) {
	var tour entity.Tour

	err := r.PG.Conn.Preload("TourImages").Preload("TourVideos").First(&tour, "id = ?", tourID).Error
	if err != nil {
		return nil, err
	}

	return &tour, nil
}

func (r *TourismRepo) GetTours() ([]entity.Tour, error) {
	var tours []entity.Tour
	err := r.PG.Conn.Preload("TourImages").Preload("TourVideos").Find(&tours).Error
	if err != nil {
		return nil, err
	}
	return tours, nil
}

func (r *TourismRepo) CreateTour(tour *entity.Tour, imageFiles []*multipart.FileHeader, videoFiles []*multipart.FileHeader) (*entity.Tour, error) {
	err := r.PG.Conn.Transaction(func(tx *gorm.DB) error {
		// Create the tour record in the database
		if err := tx.Create(&tour).Error; err != nil {
			return err
		}

		// Save images inside the transaction
		var imagePaths []entity.Image
		for _, file := range imageFiles {
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filespath := "./uploads/images/" + filename
			// Save the image file
			if err := r.saveFile(file, filespath); err != nil {
				return err
			}
			// Append the image record to the list
			imagePaths = append(imagePaths, entity.Image{ID: uuid.New(), ImageURL: filespath, TourID: tour.ID})
		}

		// Save videos inside the transaction
		var videoPaths []entity.Video
		for _, file := range videoFiles {
			filename := uuid.New().String() + filepath.Ext(file.Filename)
			filespath := "./uploads/videos/" + filename
			// Save the video file
			if err := r.saveFile(file, filespath); err != nil {
				return err
			}
			// Append the video record to the list
			videoPaths = append(videoPaths, entity.Video{ID: uuid.New(), VideoURL: filespath, TourID: tour.ID})
		}

		// Insert image and video records into the database
		if len(imagePaths) > 0 {
			if err := tx.Create(&imagePaths).Error; err != nil {
				return err
			}
		}
		if len(videoPaths) > 0 {
			if err := tx.Create(&videoPaths).Error; err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tour, nil
}

// Helper function to save the file to the disk
func (r *TourismRepo) saveFile(file *multipart.FileHeader, path string) error {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	// Create the destination file
	dst, err := os.Create(path)
	if err != nil {
		return err
	}
	defer dst.Close()

	// Copy the contents of the source file to the destination file
	if _, err := io.Copy(dst, src); err != nil {
		return err
	}

	return nil
}

//
//// GetHistory -.
//func (r *TranslationRepo) GetHistory(ctx context.Context) ([]entity.Translation, error) {
//	sql, _, err := r.Builder.
//		Select("source, destination, original, translation").
//		From("history").
//		ToSql()
//	if err != nil {
//		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Builder: %w", err)
//	}
//
//	rows, err := r.Pool.Query(ctx, sql)
//	if err != nil {
//		return nil, fmt.Errorf("TranslationRepo - GetHistory - r.Pool.Query: %w", err)
//	}
//	defer rows.Close()
//
//	entities := make([]entity.Translation, 0, _defaultEntityCap)
//
//	for rows.Next() {
//		e := entity.Translation{}
//
//		err = rows.Scan(&e.Source, &e.Destination, &e.Original, &e.Translation)
//		if err != nil {
//			return nil, fmt.Errorf("TranslationRepo - GetHistory - rows.Scan: %w", err)
//		}
//
//		entities = append(entities, e)
//	}
//
//	return entities, nil
//}
//
//// Store -.
//func (r *TranslationRepo) Store(ctx context.Context, t entity.Translation) error {
//	sql, args, err := r.Builder.
//		Insert("history").
//		Columns("source, destination, original, translation").
//		Values(t.Source, t.Destination, t.Original, t.Translation).
//		ToSql()
//	if err != nil {
//		return fmt.Errorf("TranslationRepo - Store - r.Builder: %w", err)
//	}
//
//	_, err = r.Pool.Exec(ctx, sql, args...)
//	if err != nil {
//		return fmt.Errorf("TranslationRepo - Store - r.Pool.Exec: %w", err)
//	}
//
//	return nil
//}
