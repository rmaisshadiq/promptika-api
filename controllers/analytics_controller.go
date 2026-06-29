package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rmaisshadiq/critical-prompt-api/config"
	"github.com/rmaisshadiq/critical-prompt-api/utils"
)

// ── Response Structs ──────────────────────────────────────────────────────────

// DailyTrendPoint represents a single data-point for the daily trend line chart.
type DailyTrendPoint struct {
	Date     string  `json:"date"`
	AvgScore float64 `json:"avg_score"`
}

// DashboardAnalyticsResponse is the full JSON payload returned to the client.
type DashboardAnalyticsResponse struct {
	TotalPrompts int64             `json:"total_prompts"`
	AverageScore float64           `json:"average_score"`
	LazyCount    int64             `json:"lazy_count"`
	CriticalCount int64            `json:"critical_count"`
	DailyTrend   []DailyTrendPoint `json:"daily_trend"`
}

// ── Internal scan targets ────────────────────────────────────────────────────

// aggregateResult is used to scan the single-row aggregate query.
type aggregateResult struct {
	TotalPrompts  int64   `gorm:"column:total_prompts"`
	AverageScore  float64 `gorm:"column:average_score"`
	LazyCount     int64   `gorm:"column:lazy_count"`
	CriticalCount int64   `gorm:"column:critical_count"`
}

// dailyTrendRow is used to scan each row of the GROUP BY date query.
type dailyTrendRow struct {
	Date     string  `gorm:"column:date"`
	AvgScore float64 `gorm:"column:avg_score"`
}

// ── Controller ───────────────────────────────────────────────────────────────

// GetDashboardAnalytics returns personal prompt analytics for the authenticated
// user within a given date range.
//
// GET /api/users/dashboard-analytics?start_date=YYYY-MM-DD&end_date=YYYY-MM-DD
func GetDashboardAnalytics(c *gin.Context) {
	// 1. Extract userID injected by the JWT middleware.
	rawUserID, exists := c.Get("userID")
	if !exists {
		utils.Error(c, http.StatusUnauthorized, "User not authenticated")
		return
	}
	userID := rawUserID.(uuid.UUID)

	// 2. Parse query params.
	startDateStr := c.DefaultQuery("start_date", "")
	endDateStr := c.DefaultQuery("end_date", "")

	// Default: last 30 days when no range is supplied.
	now := time.Now().UTC()
	var startDate, endDate time.Time
	var parseErr error

	if startDateStr == "" {
		startDate = now.AddDate(0, 0, -30)
	} else {
		startDate, parseErr = time.Parse("2006-01-02", startDateStr)
		if parseErr != nil {
			utils.Error(c, http.StatusBadRequest, "Invalid start_date format. Use YYYY-MM-DD")
			return
		}
	}

	if endDateStr == "" {
		endDate = now
	} else {
		endDate, parseErr = time.Parse("2006-01-02", endDateStr)
		if parseErr != nil {
			utils.Error(c, http.StatusBadRequest, "Invalid end_date format. Use YYYY-MM-DD")
			return
		}
	}

	// Make end_date inclusive by advancing to end-of-day.
	endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	// 3. Single-pass aggregate query (JOIN sessions → prompt_logs).
	var agg aggregateResult
	aggErr := config.DB.
		Table("prompt_logs").
		Select(`
			COUNT(prompt_logs.id)                                                          AS total_prompts,
			COALESCE(AVG(prompt_logs.criticality_score), 0)                               AS average_score,
			SUM(CASE WHEN prompt_logs.criticality_score < 0.5  THEN 1 ELSE 0 END)         AS lazy_count,
			SUM(CASE WHEN prompt_logs.criticality_score >= 0.5 THEN 1 ELSE 0 END)         AS critical_count
		`).
		Joins("JOIN sessions ON sessions.id = prompt_logs.session_id").
		Where("sessions.user_id = ? AND prompt_logs.created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Scan(&agg).Error

	if aggErr != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to fetch analytics data")
		return
	}

	// 4. Daily trend query — GROUP BY calendar date.
	var trendRows []dailyTrendRow
	trendErr := config.DB.
		Table("prompt_logs").
		Select("DATE(prompt_logs.created_at) AS date, AVG(prompt_logs.criticality_score) AS avg_score").
		Joins("JOIN sessions ON sessions.id = prompt_logs.session_id").
		Where("sessions.user_id = ? AND prompt_logs.created_at BETWEEN ? AND ?", userID, startDate, endDate).
		Group("DATE(prompt_logs.created_at)").
		Order("date ASC").
		Scan(&trendRows).Error

	if trendErr != nil {
		utils.Error(c, http.StatusInternalServerError, "Failed to fetch daily trend data")
		return
	}

	// 5. Map trend rows to the response DTO.
	dailyTrend := make([]DailyTrendPoint, 0, len(trendRows))
	for _, row := range trendRows {
		dailyTrend = append(dailyTrend, DailyTrendPoint{
			Date:     row.Date,
			AvgScore: row.AvgScore,
		})
	}

	// 6. Build and return response.
	utils.Success(c, DashboardAnalyticsResponse{
		TotalPrompts:  agg.TotalPrompts,
		AverageScore:  agg.AverageScore,
		LazyCount:     agg.LazyCount,
		CriticalCount: agg.CriticalCount,
		DailyTrend:    dailyTrend,
	})
}
