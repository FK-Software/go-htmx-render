package main

import (
    "os"
    "fmt"
    "errors"
    "database/sql"
    "net/http"
	"html/template"
	"path/filepath"
	"bytes"
	"time"
	"strconv"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

var db *sql.DB
var errEmptyString error = errors.New("empty string")

type Task struct {
	Id int
	Title string
	CreatedAt sql.NullString
	UpdatedAt sql.NullString
}

func getTasks(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query(`SELECT id, title, created_at, updated_at FROM "tasks"`)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to get tasks: %v", err), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		if err := rows.Scan(
			&t.Id, 
			&t.Title, 
			&t.CreatedAt, 
			&t.UpdatedAt,
		); err != nil {
			http.Error(w, fmt.Sprintf("failed to scan value: %v", err), http.StatusInternalServerError)
			return
		}

		tasks = append(tasks, t)
	}
	if err = rows.Err(); err != nil {
		http.Error(w, fmt.Sprintf("failed while iterating: %v", err), http.StatusInternalServerError)
		return
	}

	t, err := template.New("tasks").ParseFiles(
		filepath.Join("templates", "index.html"),
	)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
		return
	}

	var buf bytes.Buffer
	if err = t.Execute(&buf, map[string]interface{}{
		"tasks": tasks,
	}); err != nil {
		http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
		return
	}

	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(buf.Bytes())
}

func createTask(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var title string = r.Form.Get("title")
	if err := checkEmptyString(title); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form: %v", err), http.StatusBadRequest)
		return
	}

	if _, err := db.Exec(
		`INSERT INTO "tasks" (title, created_at) VALUES ($1, $2)`,
		title, time.Now().Format(time.RFC3339),
	); err != nil {
		http.Error(w, fmt.Sprintf("failed to create task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "get-tasks")
	w.WriteHeader(http.StatusOK)
}

func editTask(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if len(idParam) == 0 {
		http.Error(w, fmt.Sprintf("failed to get id: %v", errEmptyString), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, fmt.Sprintf("failed to parse form: %v", err), http.StatusBadRequest)
		return
	}

	var title string = r.Form.Get("title")
	if err := checkEmptyString(title); err != nil {
		http.Error(w, fmt.Sprintf("failed to validate form: %v", err), http.StatusBadRequest)
		return
	}

	if _, err = db.Exec(
		`UPDATE "tasks" SET title=$1, updated_at=$2 WHERE id=$3`,
		title, time.Now().Format(time.RFC3339), id,
	); err != nil {
		http.Error(w, fmt.Sprintf("failed to edit task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "get-tasks")
	w.WriteHeader(http.StatusOK)
}

func deleteTask(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if len(idParam) == 0 {
		http.Error(w, fmt.Sprintf("failed to get id: %v", errEmptyString), http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		http.Error(w, fmt.Sprintf("failed to parse id: %v", err), http.StatusBadRequest)
		return
	}

	if _, err = db.Exec(`DELETE from "tasks" WHERE id=$1`, id); err != nil {
		http.Error(w, fmt.Sprintf("failed to delete task: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Add("HX-Trigger", "get-tasks")
	w.WriteHeader(http.StatusOK)
}

func checkEmptyString(str ...string) error {
    for _, s := range str {
        if len(s) == 0 {
            return errEmptyString
        }
    }
    return nil
}

func main() {
    if os.Getenv("ENV") == "dev" {
        err := godotenv.Load()
        if err != nil {
            fmt.Fprintln(os.Stdout, err.Error())
            os.Exit(1)
        }
    }

    err := checkEmptyString(os.Getenv("PORT"), os.Getenv("DATABASE_URL"))
    if err != nil {
        fmt.Fprintln(os.Stdout, err.Error())
        os.Exit(1)
    }

    db, err = sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        fmt.Fprintln(os.Stdout, err.Error())
        os.Exit(1)
    }

    mux := http.NewServeMux()
    mux.Handle(
        "/static/", 
        http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))),
    )

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        t, err := template.New("index").ParseFiles(
            filepath.Join("templates", "index.html"),
        )
        if err != nil {
            http.Error(w, fmt.Sprintf("failed to parse template: %v", err), http.StatusInternalServerError)
            return
        }

        var buf bytes.Buffer
        if err := t.Execute(&buf, nil); err != nil {
            http.Error(w, fmt.Sprintf("failed to execute template: %v", err), http.StatusInternalServerError)
            return
        }

        w.Header().Set("Content-Type", "text/html; charset=utf-8")
        w.Write(buf.Bytes())
    })

    mux.HandleFunc("/tasks", getTasks)
    mux.HandleFunc("/task/create", createTask)
    mux.HandleFunc("/task/edit", editTask)
    mux.HandleFunc("/task/delete", deleteTask)
    
    err = http.ListenAndServe(":"+os.Getenv("PORT"), mux)
    if err != nil {
        fmt.Fprintln(os.Stdout, err.Error())
        os.Exit(1)
    }
}
