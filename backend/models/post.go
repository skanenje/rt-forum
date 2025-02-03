package models

type Post struct {
    ID        int       `json:"id"`
    Title     string    `json:"title"`
    Content   string    `json:"content"`
    Category  string    `json:"category"`
    CreatedAt string    `json:"created_at"` // Convert time to string for JSON
}