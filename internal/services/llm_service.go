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
	prompt := "You are an AI assistant answering questions based on a person's CV/resume. Use the provided context to give accurate, specific answers in a clear and well-formatted manner.\n\nContext:\n" + context + "\n\nQuestion: " + query + "\n\nAnswer the question directly and accurately based only on the context. Format the response exactly as:\n**Skills**\n- Category:\n  - Item\n  - Item\n- Another Category:\n  - Item\nDo not add any additional notes, comments, or information not present in the context."

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
