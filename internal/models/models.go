package models

type Document struct {
	ID      string `json:"id"`
	Content string `json:"content"`
	// Add metadata fields as needed
}

type QueryRequest struct {
	Query string `json:"query"`
}

type QueryResponse struct {
	Answer string `json:"answer"`
}