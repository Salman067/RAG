package services

import (
	"database/sql"

	_ "github.com/lib/pq"
)

type CacheService struct {
	DB *sql.DB
}

func NewCacheService(databaseURL string) (*CacheService, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, err
	}

	// Create table if not exists
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS question_cache (
			id SERIAL PRIMARY KEY,
			question TEXT UNIQUE,
			answer TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		return nil, err
	}

	return &CacheService{DB: db}, nil
}

func (c *CacheService) GetAnswer(question string) (*string, error) {
	var answer string
	err := c.DB.QueryRow("SELECT answer FROM question_cache WHERE question = $1", question).Scan(&answer)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, err
	}
	return &answer, nil
}

func (c *CacheService) SetAnswer(question, answer string) error {
	_, err := c.DB.Exec("INSERT INTO question_cache (question, answer) VALUES ($1, $2) ON CONFLICT (question) DO UPDATE SET answer = EXCLUDED.answer", question, answer)
	return err
}

func (c *CacheService) Close() error {
	return c.DB.Close()
}
