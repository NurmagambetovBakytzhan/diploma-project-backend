package v1

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"mime/multipart"
	"net/http"
	"time"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
	"tourism-backend/pkg/payment"
	"tourism-backend/utils"
)

type tourismRoutes struct {
	t usecase.TourismInterface
	l logger.Interface
	p *payment.PaymentProcessor
}

// newTourismRoutes initializes tourism routes.
// @title Tourism API
// @version 1.0
// @description API for managing tourism-related data (tours, images, videos).
// @host localhost:8080
// @BasePath /api
func newTourismRoutes(handler *gin.RouterGroup, t usecase.TourismInterface, l logger.Interface, csbn *casbin.Enforcer, payment *payment.PaymentProcessor) {
	r := &tourismRoutes{t, l, payment}

	h := handler.Group("/tours")
	{
		h.GET("/", r.GetTours)
		h.GET("/:id", r.GetTourByID)
		h.GET("/categories", r.GetAllCategories)
		h.GET("/tour-events", r.GetFilteredTourEvents)
		pay := h.Group("/payment")
		pay.Use(utils.JWTAuthMiddleware())
		{
			pay.POST("/", r.PayTourEvent)
		}

		protected := h.Group("/provider")
		protected.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
		{
			protected.POST("/", r.CreateTour)
			protected.POST("/tour-event", r.CreateTourEvent)
			protected.POST("/tour-category", r.CreateTourCategory)
			protected.POST("/tour-location", r.CreateTourLocation)
			protected.GET("/tour-location/:id", r.GetTourLocationByID)
		}
	}
}

// GetFilteredTourEvents retrieves tour events based on filters.
// @Summary Get filtered tour events
// @Description Fetches a list of tour events based on filters like date, price, and category.
// @Tags tours
// @Produce json
// @Param start_date query string false "Start date (YYYY-MM-DD)"
// @Param end_date query string false "End date (YYYY-MM-DD)"
// @Param category_ids query []string false "Category IDs"
// @Param min_price query number false "Minimum price"
// @Param max_price query number false "Maximum price"
// @Success 200 {array} entity.TourEvent "List of filtered tour events"
// @Router /tours/tour-events [get]
func (r *tourismRoutes) GetFilteredTourEvents(c *gin.Context) {
	var filter entity.TourEventFilter

	if err := c.BindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	categoryIDs := c.QueryArray("category_ids")

	for _, id := range categoryIDs {
		parsedID, err := uuid.Parse(id)
		if err == nil {
			filter.CategoryIDs = append(filter.CategoryIDs, parsedID)
		}
	}

	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	if startDateStr != "" && endDateStr != "" {
		startDate, err1 := time.Parse("2006-01-02", startDateStr)
		endDate, err2 := time.Parse("2006-01-02", endDateStr)
		if err1 == nil && err2 == nil {
			filter.StartDate = startDate
			filter.EndDate = endDate
		}
	}
	if minPrice := c.Query("min_price"); minPrice != "" {
		filter.MinPrice = utils.ParseFloat(minPrice)
	}
	if maxPrice := c.Query("max_price"); maxPrice != "" {
		filter.MaxPrice = utils.ParseFloat(maxPrice)
	}

	tourEvents, err := r.t.GetFilteredTourEvents(&filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tour events"})
		return
	}

	c.JSON(http.StatusOK, tourEvents)
}

// GetTourLocationByID retrieves a tour location by ID.
// @Summary Get tour location by ID
// @Description Fetches details of a specific tour location.
// @Tags provider
// @Produce json
// @Param id path string true "Tour Location ID"
// @Security BearerAuth
// @Success 200 {object} entity.TourLocation "Tour location details"
// @Router /tours/provider/tour-location/{id} [get]
func (r *tourismRoutes) GetTourLocationByID(c *gin.Context) {
	userID := utils.GetUserIDFromContext(c)

	tourID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}
	if !r.t.CheckTourOwner(tourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}
	tourLocation, err := r.t.GetTourLocationByID(tourID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tourLocation)

}

// CreateTourLocation creates a new tour location.
// @Summary Create a new tour location
// @Description Adds a new location for tours.
// @Tags provider
// @Accept json
// @Produce json
// @Param location body entity.CreateTourLocationDTO true "Tour location details"
// @Security BearerAuth
// @Success 201 {object} entity.TourLocation "Created tour location"
// @Router /tours/provider/tour-location [post]
func (r *tourismRoutes) CreateTourLocation(c *gin.Context) {
	var createTourLocationDTO entity.CreateTourLocationDTO
	if err := c.ShouldBind(&createTourLocationDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if !r.t.CheckTourOwner(createTourLocationDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	createdTourLocation, err := r.t.CreateTourLocation(&createTourLocationDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour Location created successfully!", "Tour Location": createdTourLocation})
}

// CreateTourCategory creates a new tour category.
// @Summary Create a new tour category
// @Description Adds a new category for tours.
// @Tags provider
// @Accept json
// @Produce json
// @Param category body entity.CreateTourCategoryDTO true "Tour category details"
// @Security BearerAuth
// @Success 201 {object} entity.TourCategory "Created tour category"
// @Router /tours/provider/tour-category [post]
func (r *tourismRoutes) CreateTourCategory(c *gin.Context) {
	var createTourCategoryDTO entity.CreateTourCategoryDTO
	if err := c.ShouldBind(&createTourCategoryDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := utils.GetUserIDFromContext(c)
	if !r.t.CheckTourOwner(createTourCategoryDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	createdTourCategory, err := r.t.CreateTourCategory(&createTourCategoryDTO)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "Tour Category created successfully!", "Tour Category": createdTourCategory})

}

// GetAllCategories retrieves all tour categories.
// @Summary Get all tour categories
// @Description Fetches a list of all available tour categories.
// @Tags tours
// @Produce json
// @Success 200 {array} entity.TourCategory "List of tour categories"
// @Router /tours/categories [get]
func (r *tourismRoutes) GetAllCategories(c *gin.Context) {
	categories, err := r.t.GetAllCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"Categories": categories})
}

// PayTourEvent processes a payment for a tour event.
// @Summary Pay for a tour event
// @Description Processes a payment for a selected tour event.
// @Tags payment
// @Accept json
// @Produce json
// @Param payment body entity.TourPurchaseRequest true "Payment details"
// @Security BearerAuth
// @Success 200 {object} entity.Purchase "Purchase details"
// @Router /tours/payment [post]
func (r *tourismRoutes) PayTourEvent(c *gin.Context) {
	var purchaseRaw entity.TourPurchaseRequest
	if err := c.ShouldBindJSON(&purchaseRaw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	UserID := utils.GetUserIDFromContext(c)

	purchase := entity.Purchase{
		TourEventID: purchaseRaw.TourEventID,
		UserID:      UserID,
		Status:      "Processing",
	}

	processingPurchase, err := r.t.CreatePurchase(&purchase)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	r.p.PurchaseQueue <- processingPurchase

	c.JSON(http.StatusOK, gin.H{"Purchase": processingPurchase})
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
	userID := utils.GetUserIDFromContext(c)

	if !r.t.CheckTourOwner(createTourEventDTO.TourID, userID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Unauthorized: You are not owner of this tour"})
		return
	}

	tour := &entity.TourEvent{
		TourID:         createTourEventDTO.TourID,
		Date:           createTourEventDTO.Date,
		Price:          createTourEventDTO.Price,
		Place:          createTourEventDTO.Place,
		AmountOfPlaces: createTourEventDTO.AmountOfPlaces,
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
