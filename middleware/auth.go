package middleware

import (
	"dailyact/models"
	"dailyact/types"
	"dailyact/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthMiddleware struct {
	db *gorm.DB
}

func NewAuthMiddleware(db *gorm.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.NewErrorResponse(
				"UNAUTHORIZED",
				"No authorization header",
				"Missing Authorization header",
			))
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := utils.ValidateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.NewErrorResponse(
				"UNAUTHORIZED",
				"Invalid token",
				err.Error(),
			))
			return
		}

		// Get user from database
		var user models.User
		if err := m.db.First(&user, claims.UserID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.NewErrorResponse(
				"UNAUTHORIZED",
				"User not found",
				err.Error(),
			))
		
			return
		}

		// Set user in context
		c.Set("user", user)
		c.Next()
	}
}

func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.NewErrorResponse(
				"UNAUTHORIZED",
				"User not found in context",
				"Please login again",
			))
			return
		}

		if user.(models.User).Role != models.RoleAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, types.NewErrorResponse(
				"FORBIDDEN",
				"Admin access required",
				"You don't have permission to access this resource",
			))
			return
		}

		c.Next()
	}
}

func (m *AuthMiddleware) RequireOwnershipOrAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, types.NewErrorResponse(
				"UNAUTHORIZED",
				"User not found in context",
				"Please login again",
			))
			return
		}

		// Admin can access all activities
		if user.(models.User).Role == models.RoleAdmin {
			c.Next()
			return
		}

		// For non-admin users, check ownership
		activityID := c.Param("id")
		if activityID == "" {
			c.Next() // For list endpoints, filtering will be done in the handler
			return
		}

		var activity models.Activity
		if err := m.db.First(&activity, activityID).Error; err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, types.NewErrorResponse(
				"NOT_FOUND",
				"Activity not found",
				err.Error(),
			))
			return
		}

		if activity.UserID != user.(models.User).ID {
			c.AbortWithStatusJSON(http.StatusForbidden, types.NewErrorResponse(
				"FORBIDDEN",
				"Access denied",
				"You don't have permission to access this activity",
			))
			return
		}

		c.Next()
	}
}
