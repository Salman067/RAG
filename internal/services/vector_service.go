package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"rag/config"
	"rag/internal/models"
)

type Chunk struct {
	Text      string
	Embedding []float32
}

type VectorService struct {
	chunks          [][]Chunk // per document
	documentService *DocumentService
}

func NewVectorService(cfg *config.Config, documentService *DocumentService) *VectorService {
	vs := &VectorService{
		chunks:          [][]Chunk{},
		documentService: documentService,
	}

	// Load existing documents and embed them
	docs, err := documentService.LoadAllDocuments()
	if err != nil {
		fmt.Printf("Warning: Failed to load documents: %v\n", err)
		return vs
	}

	for _, doc := range docs {
		err := vs.StoreDocument(doc)
		if err != nil {
			fmt.Printf("Warning: Failed to embed document %s: %v\n", doc.ID, err)
		}
	}

	return vs
}

func (vs *VectorService) EmbedText(text string) ([]float32, error) {
	// Call Ollama embeddings API
	reqBody := map[string]string{
		"model":  "nomic-embed-text",
		"prompt": text,
	}
	jsonData, _ := json.Marshal(reqBody)

	resp, err := http.Post("http://localhost:11434/api/embeddings", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if embedding, ok := result["embedding"].([]interface{}); ok {
		var emb []float32
		for _, v := range embedding {
			emb = append(emb, float32(v.(float64)))
		}
		return emb, nil
	}
	return nil, fmt.Errorf("failed to get embedding")
}

func (vs *VectorService) StoreDocument(doc models.Document) error {
	chunks := splitText(doc.Content, 300, 50) // Smaller chunks for better retrieval
	var docChunks []Chunk
	for _, chunkText := range chunks {
		embedding, err := vs.EmbedText(chunkText)
		if err != nil {
			return err
		}
		docChunks = append(docChunks, Chunk{Text: chunkText, Embedding: embedding})
	}
	vs.chunks = append(vs.chunks, docChunks)
	return nil
}

func (vs *VectorService) Search(query string, limit uint64) ([]models.Document, error) {
	embedding, err := vs.EmbedText(query)
	if err != nil {
		return nil, err
	}
	// Find top chunks
	type scoredChunk struct {
		text  string
		score float64
	}
	var allScored []scoredChunk
	for _, docChunks := range vs.chunks {
		for _, chunk := range docChunks {
			score := cosineSimilarity(embedding, chunk.Embedding)
			allScored = append(allScored, scoredChunk{text: chunk.Text, score: score})
		}
	}
	// Sort by score descending
	for i := 0; i < len(allScored)-1; i++ {
		for j := i + 1; j < len(allScored); j++ {
			if allScored[i].score < allScored[j].score {
				allScored[i], allScored[j] = allScored[j], allScored[i]
			}
		}
	}
	// Collect top texts as separate documents
	var results []models.Document
	for i := 0; i < int(limit) && i < len(allScored); i++ {
		results = append(results, models.Document{Content: allScored[i].text})
	}
	return results, nil
}

func cosineSimilarity(a, b []float32) float64 {
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i] * b[i])
		normA += float64(a[i] * a[i])
		normB += float64(b[i] * b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

func splitText(text string, chunkSize, overlap int) []string {
	var chunks []string
	runes := []rune(text)
	for i := 0; i < len(runes); i += chunkSize - overlap {
		end := i + chunkSize
		if end > len(runes) {
			end = len(runes)
		}
		chunks = append(chunks, string(runes[i:end]))
		if end == len(runes) {
			break
		}
	}
	return chunks
}
