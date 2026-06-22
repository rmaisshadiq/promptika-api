package controllers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/models"
	"github.com/rmaisshadiq/critical-prompt-api/services"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// GenerateReportInput represents the request body for report generation.
type GenerateReportInput struct {
	SessionID string `json:"session_id" binding:"required,uuid"`
}

// GenerateReport fetches all prompt logs for a session, sends them to
// the Gemini API, saves the generated report, and updates the session status.
// POST /api/reports/generate
func GenerateReport(c *gin.Context) {
	fmt.Println("[DEBUG-1] Masuk ke controller GenerateReport")
	userID, exists := c.Get("userID")
	if !exists {
		fmt.Println("[DEBUG-FAIL] User tidak terautentikasi (JWT missing)")
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var input GenerateReportInput
	if err := c.ShouldBindJSON(&input); err != nil {
		fmt.Println("[DEBUG-FAIL] Gagal bind JSON:", err.Error())
		utils.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	fmt.Println("[DEBUG-2] JSON berhasil diterima, SessionID:", input.SessionID)

	sessionID, err := uuid.Parse(input.SessionID)
	if err != nil {
		fmt.Println("[DEBUG-FAIL] Format UUID SessionID salah:", input.SessionID)
		utils.Error(c, http.StatusBadRequest, "Invalid session ID format")
		return
	}

	var session models.Session
	if result := config.DB.Where("id = ? AND user_id = ?", sessionID, userID).First(&session); result.Error != nil {
		fmt.Println("[DEBUG-FAIL] Session tidak ditemukan di DB atau bukan milik User ini!")
		utils.Error(c, http.StatusNotFound, "Session not found")
		return
	}
	fmt.Println("[DEBUG-3] Session tervalidasi milik user")

	var existingReport models.Report
	if result := config.DB.Where("session_id = ?", sessionID).First(&existingReport); result.Error == nil {
		fmt.Println("[DEBUG-FAIL] Report untuk session ini SUDAH ADA!")
		utils.Error(c, http.StatusConflict, "A report already exists for this session")
		return
	}
	fmt.Println("[DEBUG-4] Belum ada report lama, lanjut fetch prompt logs...")

	var prompts []models.PromptLog
	if result := config.DB.Where("session_id = ?", sessionID).Order("created_at ASC").Find(&prompts); result.Error != nil {
		fmt.Println("[DEBUG-FAIL] Gagal query prompt logs:", result.Error)
		utils.Error(c, http.StatusInternalServerError, "Failed to fetch prompts")
		return
	}

	if len(prompts) == 0 {
		fmt.Println("[DEBUG-FAIL] Jumlah prompt log KOSONG (0)")
		utils.Error(c, http.StatusBadRequest, "No prompts found for this session")
		return
	}
	fmt.Println("[DEBUG-5] Ditemukan", len(prompts), "prompts. Memulai service generate (Mock/Gemini)...")

	reportResult, err := services.GenerateReport(prompts)
	if err != nil {
		fmt.Println("[DEBUG-FAIL] Service Gemini/Mock Error:", err.Error())
		utils.Error(c, http.StatusInternalServerError, "Failed to generate report: "+err.Error())
		return
	}
	fmt.Println("[DEBUG-6] Service berhasil merender laporan! Menyiapkan INSERT ke DB...")

	report := models.Report{
		ID:            uuid.New(), // Pastikan Primary Key Report diisi!
		SessionID:     sessionID,
		ReportContent: reportResult.ReportContent,
		OverallScore:  reportResult.OverallScore,
	}

	fmt.Println("[DEBUG-7] Eksekusi DB.Create...")
	if result := config.DB.Create(&report); result.Error != nil {
		fmt.Println("[DEBUG-FAIL] Gagal INSERT ke tabel reports:", result.Error)
		utils.Error(c, http.StatusInternalServerError, "Failed to save report")
		return
	}

	fmt.Println("[DEBUG-SUCCESS] Laporan berhasil disimpan!")
	utils.Created(c, report)
}

// GetReportHistory fetches all reports generated for the authenticated user's sessions.
// GET /api/reports
func GetReportHistory(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var reports []models.Report
	result := config.DB.
		Joins("JOIN sessions ON sessions.id = reports.session_id").
		Where("sessions.user_id = ?", userID.(uuid.UUID)).
		Order("reports.created_at DESC").
		Find(&reports)

	if result.Error != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to fetch report history")
		return
	}

	utils.Success(c, reports)
}