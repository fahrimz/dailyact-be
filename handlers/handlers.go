package handlers

import (
	"time"
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

func (h *Handler) UpdateCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	// Check if category exists
	if err := h.db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"Category not found",
			err.Error(),
		))
		return
	}

	// Bind update data
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_INPUT",
			"Invalid input data",
			err.Error(),
		))
		return
	}

	// Update category
	if err := h.db.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to update category",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Category updated successfully",
		category,
		nil,
	))
}

func (h *Handler) DeleteCategory(c *gin.Context) {
	id := c.Param("id")
	var category models.Category

	// Check if category exists
	if err := h.db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, types.NewErrorResponse(
			"NOT_FOUND",
			"Category not found",
			err.Error(),
		))
		return
	}

	// Check if category is being used by any activities
	var activityCount int64
	if err := h.db.Model(&models.Activity{}).Where("category_id = ?", id).Count(&activityCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to check category usage",
			err.Error(),
		))
		return
	}

	if activityCount > 0 {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"CATEGORY_IN_USE",
			"Cannot delete category that is being used by activities",
			"Category is being used by activities",
		))
		return
	}

	// Delete category
	if err := h.db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to delete category",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, types.NewSuccessResponse(
		"Category deleted successfully",
		nil,
		nil,
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
	var filter types.ActivityFilter

	// Bind pagination query
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			"INVALID_QUERY",
			"Invalid pagination parameters",
			err.Error(),
		))
		return
	}

	// Start building base query
	db := h.db.Model(&models.Activity{}).Preload("Category").Preload("User")
	db = db.Where("user_id = ?", user.(models.User).ID)

	// Bind and apply filters
	if err := c.ShouldBindQuery(&filter); err == nil {
		if filter.CategoryID != nil {
			db = db.Where("category_id = ?", *filter.CategoryID)
		}
		if filter.StartDate != nil {
			if start, err := time.Parse("2006-01-02", *filter.StartDate); err == nil {
				db = db.Where("date >= ?", start)
			}
		}
		if filter.EndDate != nil {
			if end, err := time.Parse("2006-01-02", *filter.EndDate); err == nil {
				db = db.Where("date <= ?", end)
			}
		}
	}

	// Count total items with applied filters
	var total int64
	if err := db.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to count activities",
			err.Error(),
		))
		return
	}

	// Fetch paginated activities
	var activities []models.Activity
	offset := (query.Page - 1) * query.PageSize
	if err := db.Offset(offset).Limit(query.PageSize).Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			"DB_ERROR",
			"Failed to fetch activities",
			err.Error(),
		))
		return
	}

	// Prepare pagination response
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
