package handlers

import (
	"dailyact/models"
	"dailyact/types"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

// GetUsers returns a paginated list of users (admin only)
func (h *UserHandler) GetUsers(c *gin.Context) {
	var query types.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_QUERY",
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			"UNAUTHORIZED",
			"User not found in context",
			"Please login again",
		))
		return
	}

	if user.(models.User).Role != models.RoleAdmin && user.(models.User).Role != models.RoleSuperAdmin {
		c.JSON(http.StatusForbidden, types.NewErrorResponse(
			"FORBIDDEN",
			"Only admin and superadmin can access this endpoint",
			"You don't have permission to access this resource",
		))
		return
	}

	var total int64
	if err := h.db.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to count users",
			err.Error(),
		))
		return
	}

	var users []models.User
	offset := (query.Page - 1) * query.PageSize
	if err := h.db.Offset(offset).Limit(query.PageSize).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to fetch users",
			err.Error(),
		))
		return
	}

	pagination := types.NewPaginationResponse(query.Page, query.PageSize, total)
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Users retrieved successfully",
		users,
		&pagination,
	))
}

// GetUserByID returns details of a specific user (admin only)
func (h *UserHandler) GetUserByID(c *gin.Context) {
	var user models.User
	if err := h.db.Preload("Activities.Category").First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"User not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"User retrieved successfully",
		user,
		nil,
	))
}
