package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/models"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// StartSessionInput represents the request body for starting a session.
type StartSessionInput struct {
	// No fields required — session is created for the authenticated user.
}

// EndSessionInput represents the request body for ending a session.
type EndSessionInput struct {
	SessionID string `json:"session_id" binding:"required,uuid"`
}

// StartSession creates a new active session for the authenticated user.
// POST /api/sessions/start
func StartSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	session := models.Session{
		UserID:    userID.(uuid.UUID),
		StartTime: time.Now(),
		Status:    "active",
	}

	if result := config.DB.Create(&session); result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to create session")
		return
	}

	utils.Created(c, session)
}

// EndSession marks an active session as completed and sets the end time.
// POST /api/sessions/end
func EndSession(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var input EndSessionInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	sessionID, err := uuid.Parse(input.SessionID)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid session ID format")
		return
	}

	var session models.Session
	if result := config.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session); result.Error != nil {
		utils.Error(c, http.StatusNotFound, "Session not found")
		return
	}

	if session.Status == "completed" {
		utils.Error(c, http.StatusBadRequest, "Session is already completed")
		return
	}

	now := time.Now()
	session.EndTime = &now
	session.Status = "completed"

	if result := config.DB.Save(&session); result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to update session")
		return
	}

	utils.Success(c, session)
}
