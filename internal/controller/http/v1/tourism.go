package v1

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
	"tourism-backend/utils"
)

type tourismRoutes struct {
	t usecase.TourismInterface
	l logger.Interface
}

// newTourismRoutes initializes tourism routes.
// @title Tourism API
// @version 1.0
// @description API for managing tourism-related data (tours, images, videos).
// @host localhost:8080
// @BasePath /api
func newTourismRoutes(handler *gin.RouterGroup, t usecase.TourismInterface, l logger.Interface, csbn *casbin.Enforcer) {
	r := &tourismRoutes{t, l}

	h := handler.Group("/tours")
	{
		h.GET("/", r.GetTours)
		h.GET("/:id", r.GetTourByID)
		protected := h.Group("/provider")
		protected.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
		{
			protected.POST("/", r.CreateTour)
			protected.POST("/tour-event", r.CreateTourEvent)
		}
	}
}

// CreateTourEvent handles the creation of a new tour event related to some specific tour with images and videos.
// @Summary Create a new tour event
// @Description Create a new tour event.
// @Tags tours
// @Accept multipart/form-data
// @Produce json
// @Param description formData string true "Tour Description"
// @Param route formData string true "Tour Route"
// @Param price formData int true "Tour Price"
// @Success 201 {object} entity.TourDocs
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tours [post]
func (r *tourismRoutes) CreateTourEvent(c *gin.Context) {
	var createTourEventDTO entity.CreateTourEventDTO

	if err := c.ShouldBindJSON(&createTourEventDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get User ID from JWTMiddleware
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert user_id string to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	if !r.t.CheckTourOwner(createTourEventDTO.TourID, userID) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	tour := &entity.TourEvent{
		TourID: createTourEventDTO.TourID,
		Date:   createTourEventDTO.Date,
		Price:  createTourEventDTO.Price,
		Place:  createTourEventDTO.Place,
		Amount: createTourEventDTO.Amount,
	}

	createdTourEvent, err := r.t.CreateTourEvent(tour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour. Make sure the tour with such ID exists"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour event created successfully!", "Tour Event": createdTourEvent})

}

// GetStaticFiles serves static files (images and videos) for a given tour.
// @Summary Get static files for a tour
// @Description Fetches images and videos for a specific tour by ID.Example http://localhost:8080/uploads/videos/4f72a1cb-6ed4-4f01-b38b-b605d3062236.mp4.
// @Tags tours
// @Produce json
// @Param id path string true "Tour ID"
// @Success 200 {object} map[string]interface{} "Returns a list of image and video URLs."
// @Failure 400 {object} map[string]string "Invalid Tour ID"
// @Failure 404 {object} map[string]string "Tour not found"
// @Router /tours/{id}/ [get]
func GetStaticFiles(c *gin.Context) {

}

// GetTourByID retrieves a specific tour by ID.
// @Summary Get a tour by ID
// @Description Fetch details of a specific tour by its UUID.
// @Tags tours
// @Produce json
// @Param id path string true "Tour ID"
// @Success 200 {object} entity.TourDocs
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tours/{id} [get]
func (r *tourismRoutes) GetTourByID(c *gin.Context) {
	tour, err := r.t.GetTourByID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tour)
}

// GetTours retrieves all tours.
// @Summary Get all tours
// @Description Fetch a list of all available tours.
// @Tags tours
// @Produce json
// @Success 200 {array} entity.TourDocs
// @Failure 500 {object} map[string]string
// @Router /tours [get]
func (r *tourismRoutes) GetTours(c *gin.Context) {

	tours, err := r.t.GetTours()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

// CreateTour handles the creation of a new tour with images and videos.
// @Summary Create a new tour
// @Description Create a new tour with images and videos.
// @Tags tours
// @Accept multipart/form-data
// @Produce json
// @Param description formData string true "Tour Description"
// @Param route formData string true "Tour Route"
// @Param images formData file false "Tour Images (multiple allowed)"
// @Param videos formData file false "Tour Videos (multiple allowed)"
// @Success 201 {object} entity.TourDocs
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /tours [post]
func (r *tourismRoutes) CreateTour(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 200<<20) // 200MB

	if err := c.Request.ParseMultipartForm(200 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size too large"})
		return
	}

	// Get User ID from JWTMiddleware
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Convert user_id string to UUID
	userID, err := uuid.Parse(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	description := c.PostForm("description")
	route := c.PostForm("route")

	form, _ := c.MultipartForm()
	var imageFiles []*multipart.FileHeader
	var videoFiles []*multipart.FileHeader

	if files, exists := form.File["images"]; exists {
		imageFiles = files
	}
	if files, exists := form.File["videos"]; exists {
		videoFiles = files
	}

	tour := &entity.Tour{
		ID:          uuid.New(),
		Description: description,
		Route:       route,
		OwnerID:     userID,
	}

	createdTour, err := r.t.CreateTour(tour, imageFiles, videoFiles)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tour"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour created successfully", "tour": createdTour})

}
