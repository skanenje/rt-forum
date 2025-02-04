package routes

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"ramfo/backend/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func validateRegistration(user struct {
	Nickname  string `json:"nickname"`
	Email     string `json:"email"`
	Password  string `json:"password"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}) []string {
	var errors []string

	// Nickname validation
	if len(user.Nickname) < 3 {
		errors = append(errors, "Nickname must be at least 3 characters")
	}

	// Email validation
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(user.Email) {
		errors = append(errors, "Invalid email format")
	}

	// Password validation
	if len(user.Password) < 8 {
		errors = append(errors, "Password must be at least 8 characters")
	}

	// Age validation
	if user.Age < 13 {
		errors = append(errors, "Must be at least 13 years old")
	}

	// Gender validation
	validGenders := []string{"male", "female", "other", "prefer not to say"}
	user.Gender = strings.ToLower(user.Gender)
	if !contains(validGenders, user.Gender) {
		errors = append(errors, "Invalid gender selection")
	}

	return errors
}

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
		return
	}

	// Declare credentials struct
	var credentials struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}

	// Decode request body
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Enhanced login with more specific error responses
	var loginResponse = struct {
		UserID    int    `json:"user_id"`
		Nickname  string `json:"nickname"`
		Success   bool   `json:"success"`
		ErrorType string `json:"error_type"`
	}{}

	// Query to find user
	var userId int
	var hashedPassword string
	var nickname string
	query := `SELECT id, nickname, password FROM users WHERE nickname = ? OR email = ?`
	err = models.DB.QueryRow(query, credentials.Identifier, credentials.Identifier).Scan(&userId, &nickname, &hashedPassword)
	if err != nil {
		loginResponse.Success = false
		loginResponse.ErrorType = "user_not_found"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(loginResponse)
		return
	}

	// Compare passwords
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(credentials.Password))
	if err != nil {
		loginResponse.Success = false
		loginResponse.ErrorType = "incorrect_password"
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(loginResponse)
		return
	}

	// Successful login
	loginResponse.Success = true
	loginResponse.UserID = userId
	loginResponse.Nickname = nickname

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    strconv.Itoa(userId),
		Path:     "/",
		HttpOnly: true, // Improved security
	})

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(loginResponse)
}

func Register(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Invalid method", http.StatusMethodNotAllowed)
        return
    }

    // Read the body once
    body, err := io.ReadAll(r.Body)
    if err != nil {
        log.Printf("Error reading body: %v", err)
        http.Error(w, "Error reading request body", http.StatusBadRequest)
        return
    }
    log.Printf("Received body: %s", string(body))

    var user struct {
        Nickname  string `json:"nickname"`
        Email     string `json:"email"`
        Password  string `json:"password"`
        Age       int    `json:"age"`
        Gender    string `json:"gender"`
        FirstName string `json:"first_name"`
        LastName  string `json:"last_name"`
    }

    // Use the body we already read
    err = json.Unmarshal(body, &user)
    if err != nil {
        log.Printf("Error parsing JSON: %v", err)
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Rest of the function remains the same...
    validationErrors := validateRegistration(user)
    if len(validationErrors) > 0 {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(map[string][]string{"errors": validationErrors})
        return
    }

    // Hash password
    hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    // Insert user
    query := `INSERT INTO users (nickname, email, password, first_name, last_name, age, gender) VALUES (?, ?, ?, ?, ?, ?, ?)`
    _, err = models.DB.Exec(query, user.Nickname, user.Email, hashedPassword, user.FirstName, user.LastName, user.Age, user.Gender)
    if err != nil {
        http.Error(w, "User already exists or database error", http.StatusConflict)
        return
    }

    // Set session cookie
    http.SetCookie(w, &http.Cookie{
        Name:     "session",
        Value:    uuid.NewString(),
        Path:     "/",
        HttpOnly: true,
    })

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Registration successful"})
}