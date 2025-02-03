package main

import (
    "fmt"
    "log"
    "net/http"

    "github.com/gorilla/websocket"
    "ramfo/backend/models"
    "ramfo/backend/routes"
    "ramfo/backend/ws"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true // Allow all origins (for simplicity; you can restrict this in production)
    },
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

    // Serve static files (HTML, CSS, JS)
    fs := http.FileServer(http.Dir("../frontend"))
    http.Handle("/", fs)

    // Start the server
    port := ":8080"
    fmt.Printf("Server running on port %s\n", port)
    log.Fatal(http.ListenAndServe(port, nil))
}