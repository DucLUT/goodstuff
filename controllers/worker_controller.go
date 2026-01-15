package controllers

import (
	"net/http"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateWorkerInput struct {
	Bio          string  `json:"bio"`
	HourlyRate   float64 `json:"hourly_rate"`
	ServiceAreas string  `json:"service_areas"`
	WorkingHours string  `json:"working_hours"`
	IsAvailable  *bool   `json:"is_available"`
}

func GetWorkers(c *gin.Context) {
	var workers []models.Worker

	query := config.DB.Preload("User").Where("is_available = ?", true)

	// Filter by verified status
	if verified := c.Query("verified"); verified == "true" {
		query = query.Where("is_verified = ?", true)
	}

	if err := query.Find(&workers).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch workers")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Workers retrieved", workers)
}

func GetWorkerByID(c *gin.Context) {
	id := c.Param("id")
	workerID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid worker ID")
		return
	}

	var worker models.Worker
	if err := config.DB.Preload("User").First(&worker, "id = ?", workerID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Worker retrieved", worker)
}

func GetWorkerProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var worker models.Worker
	if err := config.DB.Preload("User").First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Worker profile retrieved", worker)
}

func UpdateWorkerProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateWorkerInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var worker models.Worker
	if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	updates := map[string]interface{}{}
	if input.Bio != "" {
		updates["bio"] = input.Bio
	}
	if input.HourlyRate > 0 {
		updates["hourly_rate"] = input.HourlyRate
	}
	if input.ServiceAreas != "" {
		updates["service_areas"] = input.ServiceAreas
	}
	if input.WorkingHours != "" {
		updates["working_hours"] = input.WorkingHours
	}
	if input.IsAvailable != nil {
		updates["is_available"] = *input.IsAvailable
	}

	if err := config.DB.Model(&worker).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update worker profile")
		return
	}

	config.DB.Preload("User").First(&worker, "user_id = ?", userID)
	utils.SuccessResponse(c, http.StatusOK, "Worker profile updated", worker)
}

func SetAvailability(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input struct {
		IsAvailable bool `json:"is_available"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var worker models.Worker
	if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	if err := config.DB.Model(&worker).Update("is_available", input.IsAvailable).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update availability")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Availability updated", map[string]bool{"is_available": input.IsAvailable})
}
