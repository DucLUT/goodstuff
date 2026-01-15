package controllers

import (
	"net/http"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func GetServices(c *gin.Context) {
	var services []models.Service

	query := config.DB.Preload("Category").Where("is_active = ?", true)

	// Filter by category if provided
	if categoryID := c.Query("category_id"); categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if err := query.Find(&services).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch services")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Services retrieved", services)
}

func GetServiceByID(c *gin.Context) {
	id := c.Param("id")
	serviceID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid service ID")
		return
	}

	var service models.Service
	if err := config.DB.Preload("Category").First(&service, "id = ?", serviceID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Service not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Service retrieved", service)
}

func GetCategories(c *gin.Context) {
	var categories []models.ServiceCategory

	if err := config.DB.Where("is_active = ?", true).Find(&categories).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch categories")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Categories retrieved", categories)
}

// Admin endpoints for managing services
func CreateService(c *gin.Context) {
	var service models.Service
	if err := c.ShouldBindJSON(&service); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := config.DB.Create(&service).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create service")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Service created", service)
}

func CreateCategory(c *gin.Context) {
	var category models.ServiceCategory
	if err := c.ShouldBindJSON(&category); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := config.DB.Create(&category).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create category")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Category created", category)
}
