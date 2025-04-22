package handlers

import (
	"net/http"
	"dailyact/models"
	"dailyact/types"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Handler struct {
	db *gorm.DB
}

func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Category handlers
func (h *Handler) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_INPUT",
			"Invalid input data",
			err.Error(),
		))
		return
	}

	if err := h.db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to create category",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, types.NewSuccessResponse(
		"Category created successfully",
		category,
		nil,
	))
}

func (h *Handler) GetCategories(c *gin.Context) {
	var query types.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_QUERY",
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	var total int64
	if err := h.db.Model(&models.Category{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to count categories",
			err.Error(),
		))
		return
	}

	var categories []models.Category
	offset := (query.Page - 1) * query.PageSize
	if err := h.db.Offset(offset).Limit(query.PageSize).Find(&categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to fetch categories",
			err.Error(),
		))
		return
	}

	pagination := types.NewPaginationResponse(query.Page, query.PageSize, total)
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Categories retrieved successfully",
		categories,
		&pagination,
	))
}

// Activity handlers
func (h *Handler) CreateActivity(c *gin.Context) {
	user, _ := c.Get("user")
	var activity models.Activity
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_INPUT",
			"Invalid input data",
			err.Error(),
		))
		return
	}

	activity.UserID = user.(models.User).ID
	if err := h.db.Create(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to create activity",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusCreated, types.NewSuccessResponse(
		"Activity created successfully",
		activity,
		nil,
	))
}

func (h *Handler) GetActivities(c *gin.Context) {
	user, _ := c.Get("user")
	var query types.PaginationQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_QUERY",
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	var total int64
	db := h.db.Model(&models.Activity{})
	// If not admin, only show user's activities
	if user.(models.User).Role != models.RoleAdmin {
		db = db.Where("user_id = ?", user.(models.User).ID)
	}

	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to count activities",
			err.Error(),
		))
		return
	}

	var activities []models.Activity
	offset := (query.Page - 1) * query.PageSize
	db = h.db.Preload("Category").Preload("User")
	// If not admin, only show user's activities
	if user.(models.User).Role != models.RoleAdmin {
		db = db.Where("user_id = ?", user.(models.User).ID)
	}

	if err := db.Offset(offset).Limit(query.PageSize).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to fetch activities",
			err.Error(),
		))
		return
	}

	pagination := types.NewPaginationResponse(query.Page, query.PageSize, total)
	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Activities retrieved successfully",
		activities,
		&pagination,
	))
}

func (h *Handler) GetActivityByID(c *gin.Context) {
	var activity models.Activity
	if err := h.db.Preload("Category").First(&activity, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"Activity not found",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Activity retrieved successfully",
		activity,
		nil,
	))
}

func (h *Handler) UpdateActivity(c *gin.Context) {
	// Check if activity exists
	var activity models.Activity
	if err := h.db.First(&activity, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"Activity not found",
			err.Error(),
		))
		return
	}

	// Bind new data
	if err := c.ShouldBindJSON(&activity); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_INPUT",
			"Invalid input data",
			err.Error(),
		))
		return
	}

	// Save changes
	if err := h.db.Save(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to update activity",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Activity updated successfully",
		activity,
		nil,
	))
}

func (h *Handler) DeleteActivity(c *gin.Context) {
	// Check if activity exists
	var activity models.Activity
	if err := h.db.First(&activity, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"Activity not found",
			err.Error(),
		))
		return
	}

	// Delete activity
	if err := h.db.Delete(&activity).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to delete activity",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Activity deleted successfully",
		nil,
		nil,
	))
}
