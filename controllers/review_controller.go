package controllers

import (
	"net/http"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateReviewInput struct {
	BookingID uuid.UUID `json:"booking_id" binding:"required"`
	Rating    int       `json:"rating" binding:"required,min=1,max=5"`
	Comment   string    `json:"comment"`
}

func CreateReview(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input CreateReviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Verify booking exists and belongs to user
	var booking models.Booking
	if err := config.DB.First(&booking, "id = ? AND customer_id = ?", input.BookingID, userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "Booking not found")
		return
	}

	// Check booking is completed
	if booking.Status != models.StatusCompleted {
		utils.ErrorResponse(c, http.StatusBadRequest, "Can only review completed bookings")
		return
	}

	// Check if already reviewed
	var existingReview models.Review
	if err := config.DB.First(&existingReview, "booking_id = ?", input.BookingID).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Booking already reviewed")
		return
	}

	review := models.Review{
		BookingID: input.BookingID,
		Rating:    input.Rating,
		Comment:   input.Comment,
	}

	if err := config.DB.Create(&review).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create review")
		return
	}

	// Update worker rating
	if booking.WorkerID != nil {
		var worker models.Worker
		if err := config.DB.First(&worker, "id = ?", *booking.WorkerID).Error; err == nil {
			// Calculate new average rating
			newTotalReviews := worker.TotalReviews + 1
			newRating := ((worker.Rating * float64(worker.TotalReviews)) + float64(input.Rating)) / float64(newTotalReviews)

			config.DB.Model(&worker).Updates(map[string]interface{}{
				"rating":        newRating,
				"total_reviews": newTotalReviews,
			})
		}
	}

	utils.SuccessResponse(c, http.StatusCreated, "Review submitted", review)
}

func GetWorkerReviews(c *gin.Context) {
	id := c.Param("id")
	workerID, err := uuid.Parse(id)
	if err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, "Invalid worker ID")
		return
	}

	var reviews []models.Review
	if err := config.DB.
		Joins("JOIN bookings ON reviews.booking_id = bookings.id").
		Where("bookings.worker_id = ?", workerID).
		Preload("Booking.Customer").
		Find(&reviews).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to fetch reviews")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Reviews retrieved", reviews)
}
