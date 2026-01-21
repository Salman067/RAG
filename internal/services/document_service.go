package services

import (
	"database/sql"
	"rag/internal/models"
)

type DocumentService struct {
	db *sql.DB
}

func NewDocumentService(db *sql.DB) *DocumentService {
	// Create table if not exists
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS documents (
			id TEXT PRIMARY KEY,
			content TEXT,
			filename TEXT
		)
	`)
	if err != nil {
		panic("Failed to create documents table: " + err.Error())
	}
	return &DocumentService{db: db}
}

func (ds *DocumentService) SaveDocument(doc models.Document, filename string) error {
	_, err := ds.db.Exec("INSERT INTO documents (id, content, filename) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET content = EXCLUDED.content, filename = EXCLUDED.filename", doc.ID, doc.Content, filename)
	return err
}

func (ds *DocumentService) LoadAllDocuments() ([]models.Document, error) {
	rows, err := ds.db.Query("SELECT id, content FROM documents")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var docs []models.Document
	for rows.Next() {
		var doc models.Document
		err := rows.Scan(&doc.ID, &doc.Content)
		if err != nil {
			return nil, err
		}
		docs = append(docs, doc)
	}
	return docs, nil
}
