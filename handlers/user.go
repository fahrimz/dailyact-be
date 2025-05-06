package handlers

import (
	"dailyact/models"
	"dailyact/types"
	"net/http"
	"strings"

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
	var query struct {
		types.PaginationQuery
		Name string `form:"name"`
	}
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_QUERY",
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	var total int64
	db := h.db.Model(&models.User{})
	if query.Name != "" {
		db = db.Where("LOWER(name) LIKE ? OR LOWER(email) LIKE ?", "%"+strings.ToLower(query.Name)+"%", "%"+strings.ToLower(query.Name)+"%")
	}
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to count users",
			err.Error(),
		))
		return
	}

	var users []models.User
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&users).Error; err != nil {
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

func (h *UserHandler) ChangeRole(c *gin.Context) {
	var user models.User
	if err := h.db.First(&user, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"User not found",
			err.Error(),
		))
		return
	}

	var req models.ChangeRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_REQUEST",
			"Invalid request body",
			err.Error(),
		))
		return
	}

	user.Role = req.Role
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to update user role",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"User role updated successfully",
		user,
		nil,
	))
}
