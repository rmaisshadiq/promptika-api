package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/models"
	"github.com/rmaisshadiq/critical-prompt-api/services"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// PromptInput represents the request body for submitting a prompt.
type PromptInput struct {
	SessionID  string `json:"session_id" binding:"required,uuid"`
	PromptText string `json:"prompt_text" binding:"required"`
}

// CreatePrompt receives a prompt from the browser extension, forwards it to
// the external FastAPI prediction service, saves the classified result to
// the database, and returns it to the caller.
// POST /api/prompts
func CreatePrompt(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var input PromptInput
	if err := c.ShouldBindJSON(&input); err != nil {
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	sessionID, err := uuid.Parse(input.SessionID)
	if err != nil {
		utils.Error(c, http.StatusBadRequest, "Invalid session ID format")
		return
	}

	// Verify the session belongs to the authenticated user and is active
	var session models.Session
	if result := config.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session); result.Error != nil {
		utils.Error(c, http.StatusNotFound, "Session not found")
		return
	}

	if session.Status != "active" {
		utils.Error(c, http.StatusBadRequest, "Session is not active")
		return
	}

	// Forward prompt to the external FastAPI prediction service
	prediction, err := services.CallPredictService(input.PromptText)
	if err != nil {
		utils.Error(c, http.StatusBadGateway, "Failed to get prediction: "+err.Error())
		return
	}

	// Save the scored prompt log
	promptLog := models.PromptLog{
		SessionID:        sessionID,
		PromptText:       input.PromptText,
		CriticalityScore: prediction.CriticalityScore,
	}

	if result := config.DB.Create(&promptLog); result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to save prompt log")
		return
	}

	utils.Created(c, promptLog)
}
