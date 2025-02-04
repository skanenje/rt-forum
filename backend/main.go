package main

import (
    "fmt"
    "log"
    "net/http"
    "ramfo/backend/models"
    "ramfo/backend/routes"
    "ramfo/backend/ws"
)

type cacheControlHandler struct {
    handler http.Handler
}

func (h *cacheControlHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Cache-Control", "max-age=3600, must-revalidate")
    h.handler.ServeHTTP(w, r)
}

func main() {
    // Initialize the SQLite database
    models.InitDB("db/forum.db")

    // Set up routes
    http.HandleFunc("/api/register", routes.Register)
    http.HandleFunc("/api/login", routes.Login)
    http.HandleFunc("/api/posts", routes.GetPosts)
    http.HandleFunc("/api/create-post", routes.CreatePost)
    http.HandleFunc("/ws/chat", ws.ChatHandler)

    // Serve static files (HTML, CSS, JS, Source Maps)
    fs := http.FileServer(http.Dir("../frontend"))
    staticHandler := &cacheControlHandler{handler: fs}
    http.Handle("/", staticHandler)
    http.Handle("/static/", http.StripPrefix("/static/", staticHandler))

    // Start the server
    port := ":8080"
    fmt.Printf("Server running on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}