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
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `json:"name"`
	Todos []Todo `gorm:"many2many:user_todos;"`
}

type Todo struct {
	ID          uint         `gorm:"primaryKey"`
	Title       string       `json:"title"`
	Status      string       `json:"status"`
	Users       []User       `gorm:"many2many:user_todos;"`
	Departments []Department `gorm:"many2many:department_todos;"`
}

type Department struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `json:"name"`
	Todos []Todo `gorm:"many2many:department_todos;"`
}

type DepartmentTodo struct {
	ID           uint `gorm:"primaryKey"`
	DepartmentID uint `gorm:"index"` // Foreign key for Department
	TodoID       uint `gorm:"index"` // Foreign key for Todo
}
type UserTodo struct {
	ID     uint `gorm:"primaryKey"`
	UserID uint `gorm:"index"` // Foreign key for User
	TodoID uint `gorm:"index"` // Foreign key for Todo
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
	DB.AutoMigrate(&User{}, &Todo{}, &UserTodo{}, Department{}, DepartmentTodo{})
	fmt.Println("Database connected and migrated successfully")
}

func JSONMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func CreateDepartmentHandler(w http.ResponseWriter, r *http.Request) {
	var department Department

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create department in the database
	if err := DB.Create(&department).Error; err != nil {
		http.Error(w, "Failed to create department", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(department)
}

// Handler to create a todo
func CreateTodoHandler(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		Title        string `json:"title"`
		Status       string `json:"status"`
		UserID       uint   `json:"user_id"`
		DepartmentID uint   `json:"department_id"`
	}

	// Decode the JSON request body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Find the user by ID
	var user User
	if err := DB.First(&user, requestBody.UserID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Find the department by ID
	var department Department
	if err := DB.First(&department, requestBody.DepartmentID).Error; err != nil {
		http.Error(w, "Department not found", http.StatusNotFound)
		return
	}

	// Create a new Todo and associate it with the user and department
	todo := Todo{
		Title:       requestBody.Title,
		Status:      requestBody.Status,
		Users:       []User{user},             // Associate user with the todo
		Departments: []Department{department}, // Associate department with the todo
	}

	// Save the todo and automatically add entries in the join tables
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
	if err := DB.Preload("Users").Preload("Departments").Find(&todos).Error; err != nil {
		http.Error(w, "Failed to fetch todos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(todos)
}
func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var user User

	// Decode JSON request body
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Create user in the database
	if err := DB.Create(&user).Error; err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	// Send JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user)
}

func main() {
	// Initialize database
	InitDB()

	// Set up router
	r := mux.NewRouter()
	r.Use(JSONMiddleware)
	r.HandleFunc("/todos", GetTodosHandler).Methods("GET")
	r.HandleFunc("/todos", CreateTodoHandler).Methods("POST")
	r.HandleFunc("/users", CreateUserHandler).Methods("POST")
	r.HandleFunc("/departments", CreateDepartmentHandler).Methods("POST")

	fmt.Println("Server running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
