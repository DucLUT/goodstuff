package controllers

import (
	"net/http"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateProfileInput struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
	Avatar  string `json:"avatar"`
}

type ChangePasswordInput struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

func GetProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Profile retrieved", user)
}

func UpdateProfile(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input UpdateProfileInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	updates := map[string]interface{}{}
	if input.Name != "" {
		updates["name"] = input.Name
	}
	if input.Phone != "" {
		updates["phone"] = input.Phone
	}
	if input.Address != "" {
		updates["address"] = input.Address
	}
	if input.Avatar != "" {
		updates["avatar"] = input.Avatar
	}

	if err := config.DB.Model(&user).Updates(updates).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to update profile")
		return
	}

	config.DB.First(&user, "id = ?", userID)
	utils.SuccessResponse(c, http.StatusOK, "Profile updated", user)
}

func ChangePassword(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)

	var input ChangePasswordInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := config.DB.First(&user, "id = ?", userID).Error; err != nil {
		utils.ErrorResponse(c, http.StatusNotFound, "User not found")
		return
	}

	if !utils.CheckPassword(input.OldPassword, user.PasswordHashed) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Current password is incorrect")
		return
	}

	hashedPassword, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	if err := config.DB.Model(&user).Update("password_hashed", hashedPassword).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to change password")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Password changed successfully", nil)
}
