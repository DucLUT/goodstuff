package controllers

import (
	"net/http"
	"time"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateBookingInput struct {
	ServiceID     uuid.UUID `json:"service_id" binding:"required"`
	ScheduledAt   time.Time `json:"scheduled_at" binding:"required"`
	DurationHours float64   `json:"duration_hours" binding:"required,min=1"`
	Address       string    `json:"address" binding:"required"`
	Notes         string    `json:"notes"`
}

type CancelBookingInput struct {
	Reason string `json:"reason"`
}

func CreateBooking(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input CreateBookingInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Validate scheduled time is in the future
	if input.ScheduledAt.Before(time.Now()) {
		utils.ErrorResponse(c, http.StatusBadRequest, "Scheduled time must be in the future")
		return
	}

	// Get service to calculate price
	var service models.Service
	if err := config.DB.First(&service, "id = ?", input.ServiceID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Service not found")
		return
	}

	// Calculate total price
	totalPrice := service.BasePrice + (service.PricePerHour * input.DurationHours)

	booking := models.Booking{
		CustomerID:    userID,
		ServiceID:     input.ServiceID,
		ScheduledAt:   input.ScheduledAt,
		DurationHours: input.DurationHours,
		Address:       input.Address,
		Notes:         input.Notes,
		TotalPrice:    totalPrice,
		Status:        models.StatusPending,
	}

	if err := config.DB.Create(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create booking")
		return
	}

	// Reload with relations
	config.DB.Preload("Service").Preload("Customer").First(&booking, "id = ?", booking.ID)

	utils.SuccessResponse(c, http.StatusCreated, "Booking created", booking)
}

func GetBookings(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	userRole := c.MustGet("userRole").(string)

	var bookings []models.Booking
	query := config.DB.Preload("Service").Preload("Customer").Preload("Worker.User")

	if userRole == string(models.RoleCustomer) {
		query = query.Where("customer_id = ?", userID)
	} else if userRole == string(models.RoleWorker) {
		// Get worker ID
		var worker models.Worker
		if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
			utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
			return
		}
		query = query.Where("worker_id = ?", worker.ID)
	}

	// Optional status filter
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Order("created_at DESC").Find(&bookings).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch bookings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Bookings retrieved", bookings)
}

func GetBookingByID(c *gin.Context) {
	id := c.Param("id")
	bookingID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var booking models.Booking
	if err := config.DB.Preload("Service").Preload("Customer").Preload("Worker.User").First(&booking, "id = ?", bookingID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking retrieved", booking)
}

func GetPendingBookings(c *gin.Context) {
	var bookings []models.Booking

	if err := config.DB.Preload("Service").Preload("Customer").
		Where("status = ? AND worker_id IS NULL", models.StatusPending).
		Order("scheduled_at ASC").
		Find(&bookings).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch bookings")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Pending bookings retrieved", bookings)
}

func AcceptBooking(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id := c.Param("id")
	bookingID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	// Get worker
	var worker models.Worker
	if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	var booking models.Booking
	if err := config.DB.First(&booking, "id = ?", bookingID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
		return
	}

	if booking.Status != models.StatusPending {
		utils.ErrorResponse(c, http.StatusBadRequest, "Booking is not pending")
		return
	}

	booking.WorkerID = &worker.ID
	booking.Status = models.StatusConfirmed

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to accept booking")
		return
	}

	config.DB.Preload("Service").Preload("Customer").Preload("Worker.User").First(&booking, "id = ?", bookingID)
	utils.SuccessResponse(c, http.StatusOK, "Booking accepted", booking)
}

func StartBooking(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id := c.Param("id")
	bookingID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var worker models.Worker
	if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	var booking models.Booking
	if err := config.DB.First(&booking, "id = ? AND worker_id = ?", bookingID, worker.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found or not assigned to you")
		return
	}

	if booking.Status != models.StatusConfirmed {
		utils.ErrorResponse(c, http.StatusBadRequest, "Booking is not confirmed")
		return
	}

	now := time.Now()
	booking.Status = models.StatusInProgress
	booking.StartedAt = &now

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to start booking")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking started", booking)
}

func CompleteBooking(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id := c.Param("id")
	bookingID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var worker models.Worker
	if err := config.DB.First(&worker, "user_id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Worker profile not found")
		return
	}

	var booking models.Booking
	if err := config.DB.First(&booking, "id = ? AND worker_id = ?", bookingID, worker.ID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found or not assigned to you")
		return
	}

	if booking.Status != models.StatusInProgress {
		utils.ErrorResponse(c, http.StatusBadRequest, "Booking is not in progress")
		return
	}

	now := time.Now()
	booking.Status = models.StatusCompleted
	booking.CompletedAt = &now

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to complete booking")
		return
	}

	// Update worker stats
	config.DB.Model(&worker).Updates(map[string]interface{}{
		"total_jobs": worker.TotalJobs + 1,
	})

	utils.SuccessResponse(c, http.StatusOK, "Booking completed", booking)
}

func CancelBooking(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	id := c.Param("id")
	bookingID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid booking ID")
		return
	}

	var input CancelBookingInput
	c.ShouldBindJSON(&input)

	var booking models.Booking
	if err := config.DB.First(&booking, "id = ?", bookingID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
		return
	}

	// Check ownership (customer or assigned worker)
	var worker models.Worker
	isWorker := config.DB.First(&worker, "user_id = ?", userID).Error == nil

	if booking.CustomerID != userID && (!isWorker || booking.WorkerID == nil || *booking.WorkerID != worker.ID) {
		utils.ErrorResponse(c, http.StatusForbidden, "Not authorized to cancel this booking")
		return
	}

	if booking.Status == models.StatusCompleted || booking.Status == models.StatusCancelled {
		utils.ErrorResponse(c, http.StatusBadRequest, "Cannot cancel this booking")
		return
	}

	now := time.Now()
	booking.Status = models.StatusCancelled
	booking.CancelledAt = &now
	booking.CancelReason = input.Reason

	if err := config.DB.Save(&booking).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to cancel booking")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Booking cancelled", booking)
}
