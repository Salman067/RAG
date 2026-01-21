package main

import (
	"rag/config"
	"rag/internal/handlers"
	"rag/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	cacheSvc, err := services.NewCacheService(cfg.DatabaseURL)
	if err != nil {
		panic("Failed to connect to database: " + err.Error())
	}
	defer cacheSvc.Close()

	documentSvc := services.NewDocumentService(cacheSvc.DB)
	vectorSvc := services.NewVectorService(cfg, documentSvc)
	llmSvc := services.NewLLMService(cfg)

	r := gin.Default()

	r.GET("/health", handlers.HealthHandler)
	r.POST("/ingest", func(c *gin.Context) { handlers.IngestHandler(c, vectorSvc, documentSvc) })
	r.POST("/upload", func(c *gin.Context) { handlers.UploadHandler(c, vectorSvc, documentSvc) })
	r.POST("/query", func(c *gin.Context) { handlers.QueryHandler(c, llmSvc, vectorSvc, cacheSvc) })

	r.Run(":8080")
}
