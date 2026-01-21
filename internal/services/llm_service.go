package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"rag/config"
)

type LLMService struct {
	model string
}

func NewLLMService(cfg *config.Config) *LLMService {
	return &LLMService{
		model: "llama3.2", // Better model for accurate answers
	}
}

func (ls *LLMService) GenerateAnswer(query string, context string) (string, error) {
	prompt := "Context:\n" + context + "\n\nQuestion: " + query + "\n\nAnswer ONLY this specific question based on the context. For multiple-choice questions, format your answer as: 'The correct answer [option]. [answer]'. For example: 'The correct answer B. Mobile'. Provide only this format, no additional text."

	reqBody := map[string]interface{}{
		"model":  ls.model,
		"prompt": prompt,
		"stream": false,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:11434/api/generate", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if response, ok := result["response"].(string); ok {
		return response, nil
	}
	return "", fmt.Errorf("failed to get response")
}
