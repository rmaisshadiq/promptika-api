package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/models"
	"github.com/rmaisshadiq/critical-prompt-api/services"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// RegisterInput represents the request body for user registration.
type RegisterInput struct {
	NIM      string `json:"nim" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}

// LoginInput represents the request body for user login.
type LoginInput struct {
	NIM      string `json:"nim" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register creates a new user account and returns a JWT token.
// POST /api/auth/register
func Register(c *gin.Context) {
	var input RegisterInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	// Check if NIM already exists
	var existing models.User
	if result := config.DB.Where("nim = ?", input.NIM).First(&existing); result.Error == nil {
		utils.Error(c, http.StatusConflict, "NIM is already registered")
		return
	}

	// Hash the password
	hashedPassword, err := services.HashPassword(input.Password)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	user := models.User{
		NIM:          input.NIM,
		PasswordHash: hashedPassword,
		Name:         input.Name,
	}

	if result := config.DB.Create(&user); result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create user")
		return
	}

	// Generate JWT
	token, err := services.GenerateToken(user.ID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.Created(c, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login authenticates a user by NIM and password, returning a JWT token.
// POST /api/auth/login
func Login(c *gin.Context) {
	var input LoginInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	var user models.User
	if result := config.DB.Where("nim = ?", input.NIM).First(&user); result.Error != nil {
		utils.Error(c, http.StatusUnauthorized, "Invalid NIM or password")
		return
	}

	if !services.CheckPassword(user.PasswordHash, input.Password) {
		utils.Error(c, http.StatusUnauthorized, "Invalid NIM or password")
		return
	}

	token, err := services.GenerateToken(user.ID)
	if err != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	utils.Success(c, gin.H{
		"user":  user,
		"token": token,
	})
}
