package controllers

import (
	"net/http"

	"github.com/DucLUT/goodstuff/config"
	"github.com/DucLUT/goodstuff/models"
	"github.com/DucLUT/goodstuff/utils"
	"github.com/gin-gonic/gin"
)

type RegisterInput struct {
	Email    string          `json:"email" binding:"required,email"`
	Phone    string          `json:"phone" binding:"required"`
	Password string          `json:"password" binding:"required,min=6"`
	Name     string          `json:"name" binding:"required"`
	Role     models.UserRole `json:"role" binding:"required,oneof=customer worker"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type AuthResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
}

func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if user exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", input.Email).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Email already registered")
		return
	}

	if err := config.DB.Where("phone = ?", input.Phone).First(&existingUser).Error; err == nil {
		utils.ErrorResponse(c, http.StatusConflict, "Phone number already registered")
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		Email:          input.Email,
		Phone:          input.Phone,
		PasswordHashed: hashedPassword,
		Name:           input.Name,
		Role:           input.Role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// If worker, create worker profile
	if input.Role == models.RoleWorker {
		worker := models.Worker{
			UserID: user.ID,
		}
		config.DB.Create(&worker)
	}

	// Generate token
	token, err := utils.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusCreated, "Registration successful", AuthResponse{
		Token: token,
		User:  user,
	})
}

func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.ErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !utils.CheckPassword(input.Password, user.PasswordHashed) {
		utils.ErrorResponse(c, http.StatusUnauthorized, "Invalid email or password")
		return
	}

	if !user.IsActive {
		utils.ErrorResponse(c, http.StatusForbidden, "Account is deactivated")
		return
	}

	token, err := utils.GenerateToken(user.ID, string(user.Role))
	if err != nil {
		utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.SuccessResponse(c, http.StatusOK, "Login successful", AuthResponse{
		Token: token,
		User:  user,
	})
}
