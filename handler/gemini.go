package handler

import (
	"cloud.google.com/go/vertexai/genai"
	"context"
	"encoding/json"
	"fmt"
	"hackathon-backend/models"
	"log"
	"net/http"
)

const (
	location  = "asia-northeast1"
	modelName = "gemini-1.5-flash-002"
	projectID = "term6-masataka-nakamura" // ① 自分のプロジェクトIDを指定する
)

type GenerateResponse struct {
	Response string `json:"response"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content struct {
			Role  string   `json:"Role"`
			Parts []string `json:"Parts"`
		} `json:"Content"`
	} `json:"Candidates"`
}

// handleGenerateはPOSTリクエストを受け取り、Geminiにプロンプトを送信して結果を返す
func HandleGenerate(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	switch r.Method {
	case http.MethodOptions:
		w.WriteHeader(http.StatusNoContent)
		return
	case http.MethodPost:
		var req models.Gemini_Request
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			log.Printf("fail: json.NewDecoder.Decode, %v\n", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		log.Printf("ポストはされている")
		log.Printf("aaaaaaaa%s", req.Prompt)

		// Geminiにプロンプトを送信
		resp, err := generateContentFromText(projectID, req.Prompt)
		if err != nil {
			log.Printf("Error generating content: %v\n", err)
			http.Error(w, "Failed to generate content", http.StatusInternalServerError)
			return
		}

		log.Printf("まさか")

		// 応答をJSON形式で返す
		response := GenerateResponse{Response: resp}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)

	default:
		log.Printf("fail: HTTP Method is %s\n", r.Method)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
}
func generateContentFromText(projectID, promptText string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, projectID, location)
	if err != nil {
		return "", fmt.Errorf("error creating client: %w", err)
	}
	log.Printf("生成開始")
	// Geminiにプロンプトを送信
	gemini := client.GenerativeModel(modelName)
	prompt := genai.Text(promptText)
	log.Printf("おれか")
	resp, err := gemini.GenerateContent(ctx, prompt)
	log.Printf("おまえか")
	if err != nil {
		return "", fmt.Errorf("error generating content: %w", err)
	}
	log.Printf("hogehoge")
	// デバッグ用に返り値をJSON形式で出力
	respJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return "", fmt.Errorf("error marshalling response: %w", err)
	}

	log.Printf("Gemini Response:\n%s", string(respJSON))

	// JSONを構造体にデコード
	var geminiResp GeminiResponse
	if err := json.Unmarshal(respJSON, &geminiResp); err != nil {
		return "", fmt.Errorf("error unmarshalling response: %w", err)
	}

	// 必要な内容を取り出して返す
	if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
		log.Printf("answer is %s", geminiResp.Candidates[0].Content.Parts[0])
		return geminiResp.Candidates[0].Content.Parts[0], nil
	}

	return "", fmt.Errorf("no content generated")
}
