package handlers

import (
	"bytes"
	"io"
	"net/http"
	"path/filepath"
	"strings"

	"rag/internal/models"
	"rag/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
)

func IngestHandler(c *gin.Context, vectorSvc *services.VectorService, documentSvc *services.DocumentService) {
	var doc models.Document
	if err := c.ShouldBindJSON(&doc); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := vectorSvc.StoreDocument(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := documentSvc.SaveDocument(doc, ""); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Document ingested"})
}

func UploadHandler(c *gin.Context, vectorSvc *services.VectorService, documentSvc *services.DocumentService) {
	file, header, err := c.Request.FormFile("pdf")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()

	// Check if PDF
	if filepath.Ext(header.Filename) != ".pdf" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only PDF files allowed"})
		return
	}

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}

	// Extract text from PDF
	text, err := extractTextFromPDF(content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to extract text from PDF"})
		return
	}

	// Create document
	doc := models.Document{
		ID:      header.Filename,
		Content: text,
	}

	if err := vectorSvc.StoreDocument(doc); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := documentSvc.SaveDocument(doc, header.Filename); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "PDF uploaded and ingested", "filename": header.Filename})
}

func QueryHandler(c *gin.Context, llmSvc *services.LLMService, vectorSvc *services.VectorService, cacheSvc *services.CacheService) {
	var req models.QueryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check cache first
	cachedAnswer, err := cacheSvc.GetAnswer(req.Query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cachedAnswer != nil {
		c.JSON(http.StatusOK, models.QueryResponse{Answer: *cachedAnswer})
		return
	}

	// Not in cache, proceed with retrieval and generation
	docs, err := vectorSvc.Search(req.Query, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	var contextDocs []string
	for _, doc := range docs {
		contextDocs = append(contextDocs, doc.Content)
	}
	ctx := strings.Join(contextDocs, "\n")
	if len(ctx) > 2000 {
		ctx = ctx[:2000]
	}
	answer, err := llmSvc.GenerateAnswer(req.Query, ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Cache the answer
	if err := cacheSvc.SetAnswer(req.Query, answer); err != nil {
		// Log error but don't fail the request
		// You might want to add logging here
	}

	c.JSON(http.StatusOK, models.QueryResponse{Answer: answer})
}

func HealthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func extractTextFromPDF(content []byte) (string, error) {
	reader, err := pdf.NewReader(bytes.NewReader(content), int64(len(content)))
	if err != nil {
		return "", err
	}

	var text string
	for i := 1; i <= reader.NumPage(); i++ {
		page := reader.Page(i)
		if page.V.IsNull() {
			continue
		}
		pageText, err := page.GetPlainText(nil)
		if err != nil {
			continue
		}
		text += pageText
	}
	return text, nil
}
