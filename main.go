package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates *template.Template

type User struct {
	ID       int
	Username string
	Password string
	Role     string // "student" or "instructor"
}

type Task struct {
	ID          int
	Title       string
	Description string
	DueDate     string
	Status      string
	OwnerID     int
	CreatedBy   int
	AssignedTo  string // For group assignments, comma-separated usernames
	CreatedAt   time.Time
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./taskfolio.db")
	if err != nil {
		return err
	}

	// Create users table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor'))
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed')),
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Insert demo users if not exists
	_, err = db.Exec(`
		INSERT OR IGNORE INTO users (id, username, password, role) VALUES 
		(1, 'instructor1', 'password123', 'instructor'),
		(2, 'student1', 'password123', 'student'),
		(3, 'student2', 'password123', 'student'),
		(4, 'student3', 'password123', 'student')
	`)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Load templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Setup routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/add", addTaskHandler)
	http.HandleFunc("/tasks/edit", editTaskHandler)
	http.HandleFunc("/tasks/delete", deleteTaskHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("TaskFolio server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	tasks, err := getTasks(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User  User
		Tasks []Task
	}{
		User:  user,
		Tasks: tasks,
	}

	templates.ExecuteTemplate(w, "home.html", data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	err := db.QueryRow("SELECT id, username, role FROM users WHERE username = ? AND password = ?",
		username, password).Scan(&user.ID, &user.Username, &user.Role)

	if err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid credentials"})
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: fmt.Sprintf("%d", user.ID),
		Path:  "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		// Get all students if instructor
		var students []User
		if user.Role == "instructor" {
			rows, err := db.Query("SELECT id, username FROM users WHERE role = 'student'")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var s User
					rows.Scan(&s.ID, &s.Username)
					students = append(students, s)
				}
			}
		}

		data := struct {
			User     User
			Students []User
		}{
			User:     user,
			Students: students,
		}
		templates.ExecuteTemplate(w, "add_task.html", data)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	assignedTo := r.FormValue("assigned_to")

	// Students can only create tasks for themselves
	ownerID := user.ID
	if user.Role == "instructor" && assignedTo != "" {
		// Instructor assigning to student(s)
		if assignedTo == "all" {
			ownerID = 0 // Special case: assign to all students
		} else {
			// Parse student ID
			ownerID, _ = strconv.Atoi(assignedTo)
		}
	}

	if ownerID == 0 {
		// Assign to all students
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student'")
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var studentID int
				rows.Scan(&studentID)
				_, err = db.Exec(`
					INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to)
					VALUES (?, ?, ?, ?, ?, ?)
				`, title, description, dueDate, studentID, user.ID, "")
			}
		}
	} else {
		_, err = db.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to)
			VALUES (?, ?, ?, ?, ?, ?)
		`, title, description, dueDate, ownerID, user.ID, assignedTo)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	taskID := r.URL.Query().Get("id")

	if r.Method == "GET" {
		var task Task
		err := db.QueryRow(`
			SELECT id, title, description, due_date, status, owner_id, created_by
			FROM tasks WHERE id = ?
		`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy)

		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Check permissions
		if user.Role == "student" && task.OwnerID != user.ID {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		data := struct {
			User User
			Task Task
		}{
			User: user,
			Task: task,
		}
		templates.ExecuteTemplate(w, "edit_task.html", data)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	status := r.FormValue("status")

	// Verify ownership or instructor role
	var task Task
	err = db.QueryRow("SELECT owner_id, created_by FROM tasks WHERE id = ?", taskID).
		Scan(&task.OwnerID, &task.CreatedBy)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec(`
		UPDATE tasks 
		SET title = ?, description = ?, due_date = ?, status = ?
		WHERE id = ?
	`, title, description, dueDate, status, taskID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	taskID := r.URL.Query().Get("id")

	// Verify ownership or instructor role
	var task Task
	err = db.QueryRow("SELECT owner_id, created_by FROM tasks WHERE id = ?", taskID).
		Scan(&task.OwnerID, &task.CreatedBy)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type Session struct {
	UserID int
}

func getSession(r *http.Request) (Session, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return Session{}, err
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return Session{}, err
	}

	return Session{UserID: userID}, nil
}

func getUserByID(id int) (User, error) {
	var user User
	err := db.QueryRow("SELECT id, username, role FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Role)
	return user, err
}

func getTasks(user User) ([]Task, error) {
	var tasks []Task
	var rows *sql.Rows
	var err error

	if user.Role == "instructor" {
		// Instructors see all tasks they created
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at
			FROM tasks 
			WHERE created_by = ?
			ORDER BY created_at DESC
		`, user.ID)
	} else {
		// Students see their own tasks
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at
			FROM tasks 
			WHERE owner_id = ?
			ORDER BY created_at DESC
		`, user.ID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
