package ws

import (
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"ramfo/backend/models"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn     *websocket.Conn
	UserID   int
	Username string
	Mutex    sync.Mutex
}

type Message struct {
	ID         int       `json:"id"`
	SenderID   int       `json:"sender_id"`
	ReceiverID int       `json:"receiver_id"`
	Content    string    `json:"content"`
	Timestamp  time.Time `json:"timestamp"`
	Username   string    `json:"username"`
}

var (
	clients   = make(map[int]*Client)
	broadcast = make(chan Message)
	upgrader  = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func ChatHandler(w http.ResponseWriter, r *http.Request) {
	// Extract user ID from session cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		http.Error(w, "Invalid session", http.StatusUnauthorized)
		return
	}

	// Fetch username from database
	var username string
	err = models.DB.QueryRow("SELECT nickname FROM users WHERE id = ?", userID).Scan(&username)
	if err != nil {
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		Username: username,
	}

	clients[userID] = client

	go listenForMessages(client)
	go handleBroadcast()
}

func listenForMessages(client *Client) {
	defer func() {
		delete(clients, client.UserID)
		client.Conn.Close()
	}()

	for {
		var msg Message
		err := client.Conn.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		msg.SenderID = client.UserID
		msg.Username = client.Username
		msg.Timestamp = time.Now()

		// Store message in database
		query := `INSERT INTO messages (sender_id, receiver_id, content, timestamp) VALUES (?, ?, ?, ?)`
		result, err := models.DB.Exec(query, msg.SenderID, msg.ReceiverID, msg.Content, msg.Timestamp)
		if err != nil {
			log.Printf("Error storing message: %v", err)
			continue
		}

		msgID, _ := result.LastInsertId()
		msg.ID = int(msgID)

		broadcast <- msg
	}
}

func handleBroadcast() {
	for msg := range broadcast {
		// Send to specific receiver if exists
		if receiver, exists := clients[msg.ReceiverID]; exists {
			receiver.Mutex.Lock()
			err := receiver.Conn.WriteJSON(msg)
			receiver.Mutex.Unlock()

			if err != nil {
				log.Printf("Error sending message: %v", err)
				delete(clients, msg.ReceiverID)
			}
		}
	}
}

func GetPrivateMessages(userID, otherUserID int, offset int) ([]Message, error) {
	query := `
		SELECT id, sender_id, receiver_id, content, timestamp 
		FROM messages 
		WHERE (sender_id = ? AND receiver_id = ?) OR (sender_id = ? AND receiver_id = ?) 
		ORDER BY timestamp DESC 
		LIMIT 10 OFFSET ?
	`
	rows, err := models.DB.Query(query, userID, otherUserID, otherUserID, userID, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		err := rows.Scan(&msg.ID, &msg.SenderID, &msg.ReceiverID, &msg.Content, &msg.Timestamp)
		if err != nil {
			return nil, err
		}
		messages = append(messages, msg)
	}

	return messages, nil
}
