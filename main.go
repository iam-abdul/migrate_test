package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Todo represents a task
type Todo struct {
	ID     uint   `gorm:"primaryKey"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// DB instance
var DB *gorm.DB

// Initialize Database
func InitDB() {
	dsn := "my_user:my_password@tcp(127.0.0.1:3306)/space?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// AutoMigrate tables
	DB.AutoMigrate(&Todo{})
	fmt.Println("Database connected and migrated successfully")
}

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// Handler to create a todo
func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	var todo Todo

	// Decode the request body into the todo struct
	if err := json.NewDecoder(r.Body).Decode(&todo); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create the todo in the database
	if err := DB.Create(&todo).Error; err != nil {
		http.Error(w, "Failed to create todo", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(todo)
}

// Handler to list todos
func GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	var todos []Todo
	if err := DB.Find(&todos).Error; err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todos)
}

func main() {
	// Initialize database
	InitDB()

	// Set up router
	r := mux.NewRouter()
	r.Use(JSONMiddleware)
	r.HandleFunc("/todos", GetTodosHandler).Methods("GET")
	r.HandleFunc("/todos", CreateTodoHandler).Methods("POST")

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
