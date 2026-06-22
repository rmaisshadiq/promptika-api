package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"github.com/rmaisshadiq/critical-prompt-api/models"
	"google.golang.org/api/option"
)

// GeminiReportResult holds the processed Gemini API response.
type GeminiReportResult struct {
	ReportContent string  `json:"report_content"`
	OverallScore  float64 `json:"overall_score"`
}

// GenerateReport takes a list of prompt logs and produces a qualitative analysis report via Gemini API.
func GenerateReport(prompts []models.PromptLog) (*GeminiReportResult, error) {
	if len(prompts) == 0 {
		return nil, fmt.Errorf("tidak ada prompt yang ditemukan untuk sesi ini")
	}

	// 1. Hitung Average Score dan susun riwayat prompt untuk dikirim ke Gemini
	var totalScore float64
	var promptHistory strings.Builder

	promptHistory.WriteString("Berikut adalah riwayat prompt yang dikirimkan oleh mahasiswa selama satu sesi ke Generative AI. Mohon buatkan laporan evaluasi kualitatif berdasarkan data berikut:\n\n")

	for i, p := range prompts {
		totalScore += p.CriticalityScore
		// Menyertakan teks dan skor IndoBERT agar Gemini punya konteks kuantitatif
		promptHistory.WriteString(fmt.Sprintf("Prompt %d: \"%s\" (Skor Kritis IndoBERT: %.2f/1.0)\n", i+1, p.PromptText, p.CriticalityScore))
	}
	
	averageScore := totalScore / float64(len(prompts))

	// 2. Setup Koneksi ke Gemini API
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY belum di-set pada file .env")
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat instance Gemini client: %v", err)
	}
	defer client.Close()

	// 3. Konfigurasi Model dan System Instruction
	model := client.GenerativeModel("gemini-3.1-flash-lite-preview")
	model.Temperature = genai.Ptr(float32(0.4)) // Dibuat rendah agar balasan tetap objektif dan analitis
	
	// System Instruction: Memberikan persona dan aturan main kepada LLM
	model.SystemInstruction = genai.NewUserContent(genai.Text(`
Kamu adalah Asisten Evaluator Akademik bernama "Promptika". Tugasmu adalah menganalisis riwayat prompt mahasiswa dan menyusun laporan sintesis kualitatif.
Gunakan pedoman penalaran analitis (mirip dengan standar Belmawa) untuk mengevaluasi apakah mahasiswa melakukan "Lazy Prompting" (pencarian instan) atau "Critical Thinking" (berpikir tingkat tinggi).
Berikan laporan dalam format Markdown yang rapi dengan struktur berikut:
1. Ringkasan Kinerja: Evaluasi umum dari keseluruhan sesi.
2. Temuan Positif: Sorot prompt yang memiliki pemikiran kritis yang baik.
3. Area Perbaikan: Tunjukkan prompt yang terlalu instan (skor IndoBERT rendah) dan berikan saran bagaimana memperbaikinya secara akademis.
Gunakan bahasa Indonesia yang baku, profesional, dan membangun. Jangan membuat skor angka sendiri, cukup analisis teksnya.
`))

	// 4. Eksekusi Pemanggilan API
	resp, err := model.GenerateContent(ctx, genai.Text(promptHistory.String()))
	if err != nil {
		return nil, fmt.Errorf("gagal mendapatkan respon dari Gemini API: %v", err)
	}

	// 5. Ekstraksi Teks dari Respon Gemini
	var reportContent string
	if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
		if part, ok := resp.Candidates[0].Content.Parts[0].(genai.Text); ok {
			reportContent = string(part)
		}
	} else {
		return nil, fmt.Errorf("respon dari Gemini API kosong atau tidak sesuai format")
	}

	// Kembalikan teks Markdown dari Gemini dan skor rata-rata yang dihitung murni di sisi backend (Golang)
	return &GeminiReportResult{
		ReportContent: reportContent,
		OverallScore:  averageScore,
	}, nil
}