package routes

import (
	"encoding/json"
	"log"
	"net/http"

	"ramfo/backend/models"
)

func CreatePost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	var post struct {
		Title    string `json:"title"`
		Content  string `json:"content"`
		Category string `json:"category"`
	}

	err := json.NewDecoder(r.Body).Decode(&post)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	session := r.Cookies()[0].Value
	query := `INSERT INTO posts (user_id, title, content, category) VALUES (?, ?, ?, ?)`
	_, err = models.DB.Exec(query, session, post.Title, post.Content, post.Category)
	if err != nil {
		http.Error(w, "Error creating post", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func GetPosts(w http.ResponseWriter, r *http.Request) {
	query := `SELECT id, title, content, category, created_at FROM posts ORDER BY created_at DESC`
	rows, err := models.DB.Query(query)
	if err != nil {
		http.Error(w, "Error fetching posts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var posts []models.Post
	for rows.Next() {
		var post models.Post
		err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreatedAt)
		if err != nil {
			log.Println("Error scanning row:", err)
			continue
		}

		// Convert CreatedAt to string format if needed
		// post.CreatedAt = post.CreatedAt.Format(time.RFC3339) // Optional: Format timestamp

		posts = append(posts, post)
	}

	// Encode the posts as JSON and send them in the response
	json.NewEncoder(w).Encode(posts)
}
