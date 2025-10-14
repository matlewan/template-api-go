package main

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "os"
    _ "github.com/lib/pq"
    "github.com/joho/godotenv"
	"github.com/rs/cors"
)

type Todo struct {
    ID          int    `json:"id"`
    Title       string `json:"title"`
    Description string `json:"description"`
    Priority    string `json:"priority"`
    CreatedAt   string `json:"created_at"`
}

var db *sql.DB

func main() {
    godotenv.Load()
    dsn := os.Getenv("DATABASE_URL")
    port := os.Getenv("PORT")

    var err error
    db, err = sql.Open("postgres", dsn)
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()

    mux := http.NewServeMux()
    mux.HandleFunc("/todos", handleTodos)
    mux.HandleFunc("/todos/", handleTodoByID)

    handler := cors.New(cors.Options{
        AllowedOrigins: []string{
            "https://matlewan.github.io",
            "http://127.0.0.1:3000",
            "http://localhost:3000",
        },
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Content-Type"},
        AllowCredentials: true,
    }).Handler(mux)

    fmt.Println("Server running on port", port)
    log.Fatal(http.ListenAndServe(":"+port, handler))
}

// GET /todos & POST /todos
func handleTodos(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    if r.Method == http.MethodGet {
        rows, err := db.Query(`SELECT id, title, description, priority, created_at FROM todos ORDER BY id`)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        defer rows.Close()

        todos := []Todo{}
        for rows.Next() {
            var t Todo
            rows.Scan(&t.ID, &t.Title, &t.Description, &t.Priority, &t.CreatedAt)
            todos = append(todos, t)
        }
        json.NewEncoder(w).Encode(todos)
        return
    }

    if r.Method == http.MethodPost {
        var t Todo
        json.NewDecoder(r.Body).Decode(&t)

        err := db.QueryRow(
            `INSERT INTO todos (title, description, priority) VALUES ($1, $2, $3) RETURNING id, created_at`,
            t.Title, t.Description, t.Priority,
        ).Scan(&t.ID, &t.CreatedAt)

        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }

        json.NewEncoder(w).Encode(t)
        return
    }

    http.Error(w, "Method not allowed", 405)
}

// PUT /todos/:id & DELETE /todos/:id
func handleTodoByID(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    var id int
    _, err := fmt.Sscanf(r.URL.Path, "/todos/%d", &id)
    if err != nil {
        http.Error(w, "Invalid ID", 400)
        return
    }

    if r.Method == http.MethodPut {
        var t Todo
        json.NewDecoder(r.Body).Decode(&t)
        _, err := db.Exec(`UPDATE todos SET title=$1, description=$2, priority=$3 WHERE id=$4`,
            t.Title, t.Description, t.Priority, id)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        t.ID = id
        json.NewEncoder(w).Encode(t)
        return
    }

    if r.Method == http.MethodDelete {
        _, err := db.Exec(`DELETE FROM todos WHERE id=$1`, id)
        if err != nil {
            http.Error(w, err.Error(), 500)
            return
        }
        w.WriteHeader(http.StatusNoContent)
        return
    }

    http.Error(w, "Method not allowed", 405)
}
