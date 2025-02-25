package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"tourism-backend/internal/entity"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
	"tourism-backend/utils"
)

type userRoutes struct {
	t usecase.UserInterface
	l logger.Interface
}

// newUserRoutes initializes User routes.
// @title Tourism API
// @version 1.0
// @host localhost:8080
// @BasePath /api
func newUserRoutes(handler *gin.RouterGroup, t usecase.UserInterface, l logger.Interface) {
	r := &userRoutes{t, l}

	h := handler.Group("/users")
	{
		//h.GET("/", r.GetTours)
		h.POST("/", r.RegisterUser)
		h.POST("/login", r.LoginUser)
	}
}

// LoginUser authenticates a user and returns a token.
// @Summary Login a user
// @Description Authenticates a user with their credentials and returns an access token.
// @Tags users
// @Accept json
// @Produce json
// @Param credentials body entity.LoginUserDTO true "User login credentials"
// @Success 200 {object} map[string]string "Authentication successful"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/login [post]
func (r *userRoutes) LoginUser(c *gin.Context) {
	var input entity.LoginUserDTO

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := r.t.LoginUser(&input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

// RegisterUser registers a new user.
// @Summary Register a new user
// @Description Creates a new user account with the provided details.
// @Tags users
// @Accept json
// @Produce json
// @Param user body entity.CreateUserDTO true "User registration data"
// @Success 201 {object} map[string]interface{} "User registered successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [post]
func (r *userRoutes) RegisterUser(c *gin.Context) {
	var createUserDTO entity.CreateUserDTO
	if err := c.ShouldBindJSON(&createUserDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hashedPassword, err := utils.HashPassword(createUserDTO.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error hashing password"})
		return
	}

	user := entity.User{
		Username: createUserDTO.Username,
		Email:    createUserDTO.Email,
		Password: hashedPassword,
		Role:     "user",
	}

	createdUser, err := r.t.RegisterUser(&user)

	c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "User": createdUser})
}
