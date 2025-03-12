package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/go-gormigrate/gormigrate/v2"
)

// Todo represents a task
type User struct {
	ID    uint   `gorm:"primaryKey"`
	Name  string `json:"name"`
	Todos []Todo `gorm:"many2many:user_todos;"`
}

type Todo struct {
	ID       uint   `gorm:"primaryKey"`
	TaskName string `json:"task_name"`
	Status   string `json:"status"`
	Users    []User `gorm:"many2many:user_todos;"`
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
	// DB.AutoMigrate(&User{}, &Todo{}, &UserTodo{}, Department{}, DepartmentTodo{})

	m := gormigrate.New(DB, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "20250312_rename_title_column",
			Migrate: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE todos CHANGE COLUMN title task_name VARCHAR(255);").Error
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Exec("ALTER TABLE todos CHANGE COLUMN task_name title VARCHAR(255);").Error
			},
		},
	})

	if err := m.Migrate(); err != nil {
		log.Fatal("Could not apply migration:", err)
	}

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
	todo := Todo{TaskName: "Sample Todo", Status: "pending"}
	DB.Create(&todo)
	fmt.Fprintf(w, "Todo created: %+v", todo)
}

// Handler to list todos
func GetTodosHandler(w http.ResponseWriter, r *http.Request) {
	var todos []Todo
	DB.Find(&todos)
	fmt.Fprintf(w, "Todos: %+v", todos)
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
