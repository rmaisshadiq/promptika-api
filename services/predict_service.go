package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// PredictRequest is the payload sent to the external FastAPI /predict endpoint.
type PredictRequest struct {
	PromptText string `json:"prompt_text"`
}

// PredictResponse is the response received from the FastAPI /predict endpoint.
// The IndoBERT model uses a regression approach, returning a continuous
// criticality_score in the range [0, 1] where:
//   - 0 = lazy prompting
//   - 1 = critical prompting
// An intent_label (0 or 1) is also derived from a threshold (>= 0.5 → 1).
type PredictResponse struct {
	IntentLabel      int     `json:"intent_label"`
	CriticalityScore float64 `json:"criticality_score"`
}

// CallPredictService sends a prompt text to the external FastAPI prediction
// service and returns the regression result.
func CallPredictService(promptText string) (*PredictResponse, error) {
	baseURL := os.Getenv("PREDICT_SERVICE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8000"
	}

	payload := PredictRequest{PromptText: promptText}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal predict request: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Post(baseURL+"/predict", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to call predict service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("predict service returned status %d", resp.StatusCode)
	}

	var result PredictResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode predict response: %w", err)
	}

	return &result, nil
}
