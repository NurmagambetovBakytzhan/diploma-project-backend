package v1

import (
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"net/http"
	"tourism-backend/internal/usecase"
	"tourism-backend/pkg/logger"
	"tourism-backend/utils"
)

type adminRoutes struct {
	t usecase.AdminInterface
	l logger.Interface
}

// newUserRoutes initializes User routes.
// @title Tourism API
// @version 1.0
// @host localhost:8080
// @BasePath /api
func newAdminRoutes(handler *gin.RouterGroup, t usecase.AdminInterface, l logger.Interface, csbn *casbin.Enforcer) {
	r := &adminRoutes{t, l}

	h := handler.Group("/admin")
	h.Use(utils.JWTAuthMiddleware(), utils.CasbinMiddleware(csbn))
	{
		h.GET("/users", r.GetUsers)
	}
}

func (r *adminRoutes) GetUsers(c *gin.Context) {
	users, err := r.t.GetUsers()

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, &users)
}
