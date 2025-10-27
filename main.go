package main

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates *template.Template

type User struct {
	ID          int
	Username    string
	Password    string
	Role        string // "student" or "instructor"
	CohortName  string // e.g., "Java11.2", "Data6.2"
	IsSuspended bool   // For account suspension
}

type Task struct {
	ID             int
	Title          string
	Description    string
	DueDate        string
	Status         string
	OwnerID        int
	CreatedBy      int
	AssignedTo     string // For group assignments, comma-separated usernames
	CreatedAt      time.Time
	TaskType       string     // "group-programming-project", "individual-programming-lab", "reading", "document", "resume", "prodev"
	ResourceURL    string     // URL pointing to resources for the task
	CohortID       string     // Cohort name for cohort-based assignments
	InProgressDate *time.Time // Date when task was first moved to in_progress
	DoneDate       *time.Time // Date when task was moved to done
}

func initDB() error {
	var err error
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "./taskfolio.db"
	}
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return err
	}

	// Create users table with new fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table with new fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', '')),
			resource_url TEXT,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Migrate existing database by adding new columns if they don't exist
	if err := migrateDatabase(); err != nil {
		return err
	}

	// Insert demo users if not exists
	_, err = db.Exec(`
		INSERT OR IGNORE INTO users (id, username, password, role, cohort_name, is_suspended) VALUES
		(1, 'instructor1', 'password123', 'instructor', NULL, 0),
		(2, 'student1', 'password123', 'student', 'Java11.2', 0),
		(3, 'student2', 'password123', 'student', 'Java11.2', 0),
		(4, 'student3', 'password123', 'student', 'Data6.2', 0)
	`)
	if err != nil {
		return err
	}

	return nil
}

// migrateDatabase adds new columns to existing tables if they don't exist
func migrateDatabase() error {
	// Check and add cohort_name column to users table
	var cohortNameExists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='cohort_name'
	`).Scan(&cohortNameExists)
	if err == nil && cohortNameExists == 0 {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN cohort_name TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add cohort_name column: %v", err)
		}
		log.Println("Added cohort_name column to users table")
	}

	// Check and add is_suspended column to users table
	var isSuspendedExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='is_suspended'
	`).Scan(&isSuspendedExists)
	if err == nil && isSuspendedExists == 0 {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN is_suspended BOOLEAN DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add is_suspended column: %v", err)
		}
		log.Println("Added is_suspended column to users table")
	}

	// Check and add task_type column to tasks table
	var taskTypeExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='task_type'
	`).Scan(&taskTypeExists)
	if err == nil && taskTypeExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN task_type TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add task_type column: %v", err)
		}
		log.Println("Added task_type column to tasks table")
	}

	// Check and add resource_url column to tasks table
	var resourceURLExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='resource_url'
	`).Scan(&resourceURLExists)
	if err == nil && resourceURLExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN resource_url TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add resource_url column: %v", err)
		}
		log.Println("Added resource_url column to tasks table")
	}

	// Check and add cohort_id column to tasks table
	var cohortIDExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='cohort_id'
	`).Scan(&cohortIDExists)
	if err == nil && cohortIDExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN cohort_id TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add cohort_id column: %v", err)
		}
		log.Println("Added cohort_id column to tasks table")
	}

	// Check and add in_progress_date column to tasks table
	var inProgressDateExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='in_progress_date'
	`).Scan(&inProgressDateExists)
	if err == nil && inProgressDateExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN in_progress_date DATETIME`)
		if err != nil {
			return fmt.Errorf("failed to add in_progress_date column: %v", err)
		}
		log.Println("Added in_progress_date column to tasks table")
	}

	// Check and add done_date column to tasks table
	var doneDateExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='done_date'
	`).Scan(&doneDateExists)
	if err == nil && doneDateExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN done_date DATETIME`)
		if err != nil {
			return fmt.Errorf("failed to add done_date column: %v", err)
		}
		log.Println("Added done_date column to tasks table")
	}

	// Migrate task_type CHECK constraint to new values
	// We need to check if the old constraint exists and recreate the table if needed
	if err := migrateTaskTypeConstraint(); err != nil {
		return fmt.Errorf("failed to migrate task_type constraint: %v", err)
	}

	return nil
}

// migrateTaskTypeConstraint recreates the tasks table with the updated CHECK constraint
func migrateTaskTypeConstraint() error {
	// Check if tasks table exists
	var tableExists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='tasks'
	`).Scan(&tableExists)
	if err != nil || tableExists == 0 {
		// Table doesn't exist, no migration needed
		return nil
	}

	// Get the current table schema
	var sql string
	err = db.QueryRow(`
		SELECT sql FROM sqlite_master WHERE type='table' AND name='tasks'
	`).Scan(&sql)
	if err != nil {
		return err
	}

	// Check if the old constraint exists (contains 'coding', 'submission', or 'document-creation')
	needsMigration := false
	if containsOldTaskTypes(sql) {
		needsMigration = true
	}

	if !needsMigration {
		// No migration needed
		return nil
	}

	log.Println("Migrating tasks table to update task_type CHECK constraint...")

	// Start a transaction
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Create a new table with the correct constraint
	_, err = tx.Exec(`
		CREATE TABLE tasks_new (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', '')),
			resource_url TEXT,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Copy data from old table to new table, mapping old task types to new ones
	// Old task types: 'reading', 'coding', 'submission', 'document-creation'
	// New task types: 'group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev'
	// Mapping: 'coding' -> 'individual-programming-lab', 'submission' -> '', 'document-creation' -> 'document'
	_, err = tx.Exec(`
		INSERT INTO tasks_new (id, title, description, due_date, status, owner_id, created_by, assigned_to, created_at, task_type, resource_url, cohort_id, in_progress_date, done_date)
		SELECT id, title, description, due_date, status, owner_id, created_by,
		       COALESCE(assigned_to, ''), created_at,
		       CASE COALESCE(task_type, '')
		           WHEN 'coding' THEN 'individual-programming-lab'
		           WHEN 'submission' THEN ''
		           WHEN 'document-creation' THEN 'document'
		           ELSE COALESCE(task_type, '')
		       END,
		       COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
		       in_progress_date, done_date
		FROM tasks
	`)
	if err != nil {
		return err
	}

	// Drop the old table
	_, err = tx.Exec(`DROP TABLE tasks`)
	if err != nil {
		return err
	}

	// Rename the new table
	_, err = tx.Exec(`ALTER TABLE tasks_new RENAME TO tasks`)
	if err != nil {
		return err
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return err
	}

	log.Println("Successfully migrated tasks table with new task_type values")
	return nil
}

// containsOldTaskTypes checks if the table schema contains old task type values
func containsOldTaskTypes(sql string) bool {
	oldTypes := []string{"'coding'", "'submission'", "'document-creation'"}
	for _, oldType := range oldTypes {
		if strings.Contains(sql, oldType) {
			return true
		}
	}
	return false
}

// studentsHandler displays list of students for instructors to manage
func studentsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can access this page
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get all students
	rows, err := db.Query(`
		SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0)
		FROM users WHERE role = 'student' ORDER BY username
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []User
	for rows.Next() {
		var s User
		s.Role = "student"
		rows.Scan(&s.ID, &s.Username, &s.CohortName, &s.IsSuspended)
		students = append(students, s)
	}

	data := struct {
		User     User
		Students []User
	}{
		User:     user,
		Students: students,
	}

	templates.ExecuteTemplate(w, "students.html", data)
}

// editStudentHandler handles editing student cohort and suspension status
func editStudentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can edit students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	studentID := r.URL.Query().Get("id")

	if r.Method == "GET" {
		var student User
		var cohortName sql.NullString
		err := db.QueryRow(`
			SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0)
			FROM users WHERE id = ? AND role = 'student'
		`, studentID).Scan(&student.ID, &student.Username, &cohortName, &student.IsSuspended)

		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		if cohortName.Valid {
			student.CohortName = cohortName.String
		}

		data := struct {
			User    User
			Student User
		}{
			User:    user,
			Student: student,
		}
		templates.ExecuteTemplate(w, "edit_student.html", data)
		return
	}

	cohortName := r.FormValue("cohort_name")
	isSuspended := r.FormValue("is_suspended") == "on"
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Update student information
	// If password is provided, update it; otherwise, only update other fields
	if password != "" {
		_, err = db.Exec(`
			UPDATE users
			SET username = ?, cohort_name = ?, is_suspended = ?, password = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, password, studentID)
	} else {
		_, err = db.Exec(`
			UPDATE users
			SET username = ?, cohort_name = ?, is_suspended = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, studentID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// addStudentHandler handles adding new students
func addStudentHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can add students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		data := struct {
			User  User
			Error string
		}{
			User: user,
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	cohortName := r.FormValue("cohort_name")

	// Validate required fields
	if username == "" || password == "" {
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username and password are required",
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	// Insert new student
	_, err = db.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'student', ?, 0)
	`, username, password, cohortName)

	if err != nil {
		// Check if it's a unique constraint violation
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username already exists. Please choose a different username.",
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// importStudentsHandler handles bulk import of students from CSV file
func importStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can import students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 1MB for CSV files)
	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csvfile")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read and parse CSV using encoding/csv for proper CSV handling
	var importedStudents []string
	var failedImports []string
	lineNum := 0

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1 // Allow 2-3 fields (username, password, optional cohort)
	csvReader.TrimLeadingSpace = true

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			http.Error(w, "Failed to read CSV file", http.StatusBadRequest)
			return
		}

		lineNum++
		
		// Skip empty lines
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Validate CSV has expected number of fields (2-3)
		if len(record) < 2 || len(record) > 3 {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Invalid format (expected 2-3 fields: Username, Password, Cohort)", lineNum))
			continue
		}

		username := strings.TrimSpace(record[0])
		password := strings.TrimSpace(record[1])
		cohort := ""
		if len(record) >= 3 {
			cohort = strings.TrimSpace(record[2])
		}

		if username == "" || password == "" {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Username or password is empty", lineNum))
			continue
		}

		// Insert student into database
		// Note: Passwords are stored in plaintext as per existing application design
		_, err = db.Exec(`
			INSERT INTO users (username, password, role, cohort_name, is_suspended)
			VALUES (?, ?, 'student', ?, 0)
		`, username, password, cohort)

		if err != nil {
			failedImports = append(failedImports, fmt.Sprintf("%s: %v", username, err))
		} else {
			importedStudents = append(importedStudents, username)
		}
	}

	// Generate report
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Import Results - Tasking</title>
    <style>
        body { font-family: sans-serif; padding: 20px; background: #f5f7fa; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { color: #2C5F7C; }
        .success { color: #065f46; background: #d1fae5; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .error { color: #991b1b; background: #fee2e2; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .list { margin: 10px 0; padding-left: 20px; }
        .btn { 
            display: inline-block;
            padding: 10px 20px; 
            background: #667eea; 
            color: white; 
            text-decoration: none; 
            border-radius: 5px; 
            margin-top: 20px;
        }
        .btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Import Results</h1>
`)

	if len(importedStudents) > 0 {
		fmt.Fprintf(w, `<div class="success">
            <h2>Successfully Imported: %d student(s)</h2>
            <ul class="list">`, len(importedStudents))
		for _, username := range importedStudents {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(username))
		}
		fmt.Fprintf(w, "</ul></div>")
	}

	if len(failedImports) > 0 {
		fmt.Fprintf(w, `<div class="error">
            <h2>Failed Imports: %d</h2>
            <ul class="list">`, len(failedImports))
		for _, failure := range failedImports {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(failure))
		}
		fmt.Fprintf(w, "</ul></div>")
	}

	if len(importedStudents) == 0 && len(failedImports) == 0 {
		fmt.Fprintf(w, `<div class="error">No students were imported. Please check your CSV file format.</div>`)
	}

	fmt.Fprintf(w, `
        <a href="/students" class="btn">← Back to Students</a>
    </div>
</body>
</html>`)
}

// studentProgressHandler shows task completion for a specific student
func studentProgressHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can view student progress
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	studentID := r.URL.Query().Get("id")

	// Get student info
	var student User
	var cohortName sql.NullString
	err = db.QueryRow(`
		SELECT id, username, COALESCE(cohort_name, '')
		FROM users WHERE id = ? AND role = 'student'
	`, studentID).Scan(&student.ID, &student.Username, &cohortName)

	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	if cohortName.Valid {
		student.CohortName = cohortName.String
	}

	// Get all tasks for this student
	rows, err := db.Query(`
		SELECT id, title, description, due_date, status, created_at,
		       COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM tasks
		WHERE owner_id = ?
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	totalTasks := 0
	completedTasks := 0
	inProgressTasks := 0
	pendingTasks := 0

	for rows.Next() {
		var task Task
		rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.CreatedAt, &task.TaskType, &task.ResourceURL)
		tasks = append(tasks, task)
		totalTasks++

		switch task.Status {
		case "completed":
			completedTasks++
		case "in_progress":
			inProgressTasks++
		case "pending":
			pendingTasks++
		}
	}

	completionPercentage := 0.0
	if totalTasks > 0 {
		completionPercentage = float64(completedTasks) / float64(totalTasks) * 100
	}

	data := struct {
		User                 User
		Student              User
		Tasks                []Task
		TotalTasks           int
		CompletedTasks       int
		InProgressTasks      int
		PendingTasks         int
		CompletionPercentage float64
	}{
		User:                 user,
		Student:              student,
		Tasks:                tasks,
		TotalTasks:           totalTasks,
		CompletedTasks:       completedTasks,
		InProgressTasks:      inProgressTasks,
		PendingTasks:         pendingTasks,
		CompletionPercentage: completionPercentage,
	}

	templates.ExecuteTemplate(w, "student_progress.html", data)
}

// taskStatsHandler shows completion statistics for group/cohort tasks
func taskStatsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can view task statistics
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get all unique group tasks (tasks created by instructors that were assigned to multiple students)
	// We identify these by looking at tasks with the same title, created_by, and cohort_id
	type GroupTask struct {
		Title       string
		Description string
		DueDate     string
		TaskType    string
		ResourceURL string
		CreatedBy   int
		CohortID    string
		CreatedAt   string
		TotalCount  int
		InProgress  int
		Completed   int
	}

	rows, err := db.Query(`
		SELECT
			t.title,
			COALESCE(t.description, '') as description,
			COALESCE(t.due_date, '') as due_date,
			COALESCE(t.task_type, '') as task_type,
			COALESCE(t.resource_url, '') as resource_url,
			t.created_by,
			COALESCE(t.cohort_id, '') as cohort_id,
			MIN(t.created_at) as created_at,
			COUNT(*) as total_count,
			SUM(CASE WHEN t.status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks t
		JOIN users u ON t.owner_id = u.id
		WHERE t.created_by IN (SELECT id FROM users WHERE role = 'instructor')
		AND u.role = 'student'
		GROUP BY t.title, t.created_by, COALESCE(t.cohort_id, '')
		HAVING COUNT(*) > 1
		ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var groupTasks []GroupTask
	for rows.Next() {
		var gt GroupTask
		err := rows.Scan(&gt.Title, &gt.Description, &gt.DueDate, &gt.TaskType,
			&gt.ResourceURL, &gt.CreatedBy, &gt.CohortID, &gt.CreatedAt,
			&gt.TotalCount, &gt.InProgress, &gt.Completed)
		if err != nil {
			log.Printf("Error scanning group task: %v", err)
			continue
		}
		groupTasks = append(groupTasks, gt)
	}

	data := struct {
		User       User
		GroupTasks []GroupTask
	}{
		User:       user,
		GroupTasks: groupTasks,
	}

	templates.ExecuteTemplate(w, "task_stats.html", data)
}

// taskDetailHandler shows a specific task and all students assigned to it
func taskDetailHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can view task details
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get task title and creator from query params
	taskTitle := r.URL.Query().Get("title")
	createdByStr := r.URL.Query().Get("created_by")
	cohortID := r.URL.Query().Get("cohort_id")

	if taskTitle == "" || createdByStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	createdBy, err := strconv.Atoi(createdByStr)
	if err != nil {
		http.Error(w, "Invalid created_by parameter", http.StatusBadRequest)
		return
	}

	// Get task details
	type TaskAssignment struct {
		TaskID         int
		StudentID      int
		StudentName    string
		CohortName     string
		Status         string
		CreatedAt      time.Time
		InProgressDate *time.Time
		DoneDate       *time.Time
	}

	var taskInfo struct {
		Title       string
		Description string
		DueDate     string
		TaskType    string
		ResourceURL string
		CreatedBy   int
		CreatorName string
		CohortID    string
		CreatedAt   time.Time
	}

	// Get task information from the first matching task
	err = db.QueryRow(`
		SELECT t.title, COALESCE(t.description, ''), COALESCE(t.due_date, ''),
		       COALESCE(t.task_type, ''), COALESCE(t.resource_url, ''),
		       t.created_by, u.username, COALESCE(t.cohort_id, ''), t.created_at
		FROM tasks t
		JOIN users u ON t.created_by = u.id
		WHERE t.title = ? AND t.created_by = ? AND COALESCE(t.cohort_id, '') = ?
		ORDER BY t.created_at ASC
		LIMIT 1
	`, taskTitle, createdBy, cohortID).Scan(&taskInfo.Title, &taskInfo.Description, &taskInfo.DueDate,
		&taskInfo.TaskType, &taskInfo.ResourceURL, &taskInfo.CreatedBy, &taskInfo.CreatorName,
		&taskInfo.CohortID, &taskInfo.CreatedAt)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Get all students assigned to this task
	rows, err := db.Query(`
		SELECT t.id, t.owner_id, u.username, COALESCE(u.cohort_name, ''),
		       t.status, t.created_at, t.in_progress_date, t.done_date
		FROM tasks t
		JOIN users u ON t.owner_id = u.id
		WHERE t.title = ? AND t.created_by = ? AND COALESCE(t.cohort_id, '') = ?
		ORDER BY u.username
	`, taskTitle, createdBy, cohortID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var assignments []TaskAssignment
	statusCounts := map[string]int{
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
		"blocked":     0,
	}

	for rows.Next() {
		var ta TaskAssignment
		var inProgressDate, doneDate sql.NullTime
		err := rows.Scan(&ta.TaskID, &ta.StudentID, &ta.StudentName, &ta.CohortName,
			&ta.Status, &ta.CreatedAt, &inProgressDate, &doneDate)
		if err != nil {
			log.Printf("Error scanning task assignment: %v", err)
			continue
		}
		if inProgressDate.Valid {
			ta.InProgressDate = &inProgressDate.Time
		}
		if doneDate.Valid {
			ta.DoneDate = &doneDate.Time
		}
		assignments = append(assignments, ta)
		statusCounts[ta.Status]++
	}

	data := struct {
		User         User
		TaskInfo     interface{}
		Assignments  []TaskAssignment
		StatusCounts map[string]int
		TotalCount   int
	}{
		User:         user,
		TaskInfo:     taskInfo,
		Assignments:  assignments,
		StatusCounts: statusCounts,
		TotalCount:   len(assignments),
	}

	templates.ExecuteTemplate(w, "task_detail.html", data)
}

// instructorsHandler displays list of instructors for instructors to manage
func instructorsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can access this page
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get all instructors
	rows, err := db.Query(`
		SELECT id, username
		FROM users WHERE role = 'instructor' ORDER BY username
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var instructors []User
	for rows.Next() {
		var i User
		i.Role = "instructor"
		rows.Scan(&i.ID, &i.Username)
		instructors = append(instructors, i)
	}

	data := struct {
		User        User
		Instructors []User
	}{
		User:        user,
		Instructors: instructors,
	}

	templates.ExecuteTemplate(w, "instructors.html", data)
}

// addInstructorHandler handles adding new instructors
func addInstructorHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can add instructors
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		data := struct {
			User  User
			Error string
		}{
			User: user,
		}
		templates.ExecuteTemplate(w, "add_instructor.html", data)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// Validate required fields
	if username == "" || password == "" {
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username and password are required",
		}
		templates.ExecuteTemplate(w, "add_instructor.html", data)
		return
	}

	// Insert new instructor
	_, err = db.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'instructor', NULL, 0)
	`, username, password)

	if err != nil {
		// Check if it's a unique constraint violation
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username already exists. Please choose a different username.",
		}
		templates.ExecuteTemplate(w, "add_instructor.html", data)
		return
	}

	http.Redirect(w, r, "/instructors", http.StatusSeeOther)
}

// updateTaskStatusHandler handles AJAX requests to update task status
func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	session, err := getSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	taskID := r.FormValue("task_id")
	newStatus := r.FormValue("status")

	// Validate status
	validStatuses := map[string]bool{
		"pending":     true,
		"blocked":     true,
		"in_progress": true,
		"completed":   true,
	}
	if !validStatuses[newStatus] {
		http.Error(w, "Invalid status", http.StatusBadRequest)
		return
	}

	// Get current task
	var task Task
	var inProgressDate, doneDate sql.NullTime
	err = db.QueryRow(`
		SELECT id, owner_id, status, in_progress_date, done_date
		FROM tasks WHERE id = ?
	`, taskID).Scan(&task.ID, &task.OwnerID, &task.Status, &inProgressDate, &doneDate)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Check permissions - students can only update their own tasks
	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Manage date fields based on status change
	now := time.Now()

	// Check what needs to be updated based on status transition
	if task.Status == "completed" && newStatus != "completed" {
		// Clear done_date when moving out of completed
		if newStatus == "in_progress" && !inProgressDate.Valid {
			// Also set in_progress_date if moving to in_progress for first time
			_, err = db.Exec(`
				UPDATE tasks
				SET status = ?, done_date = NULL, in_progress_date = ?
				WHERE id = ?
			`, newStatus, now, taskID)
		} else {
			_, err = db.Exec(`
				UPDATE tasks
				SET status = ?, done_date = NULL
				WHERE id = ?
			`, newStatus, taskID)
		}
	} else if newStatus == "completed" {
		// Set done_date when moving to completed
		_, err = db.Exec(`
			UPDATE tasks
			SET status = ?, done_date = ?
			WHERE id = ?
		`, newStatus, now, taskID)
	} else if newStatus == "in_progress" && !inProgressDate.Valid {
		// Set in_progress_date when moving to in_progress for the first time
		_, err = db.Exec(`
			UPDATE tasks
			SET status = ?, in_progress_date = ?
			WHERE id = ?
		`, newStatus, now, taskID)
	} else {
		// Just update status
		_, err = db.Exec(`
			UPDATE tasks
			SET status = ?
			WHERE id = ?
		`, newStatus, taskID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

// importTasksHandler handles bulk import of tasks from CSV file
func importTasksHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can import tasks
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 1MB for CSV files)
	err = r.ParseMultipartForm(1 << 20)
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get assignment type and targets
	assignType := r.FormValue("assign_type")
	if assignType == "" {
		http.Error(w, "Please select an assignment type", http.StatusBadRequest)
		return
	}

	// Get target student IDs based on assignment type
	var targetStudentIDs []int
	switch assignType {
	case "all":
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student'")
		if err != nil {
			http.Error(w, "Failed to query students", http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			var studentID int
			rows.Scan(&studentID)
			targetStudentIDs = append(targetStudentIDs, studentID)
		}
		rows.Close()
	case "cohort":
		cohortName := r.FormValue("cohort")
		if cohortName == "" {
			http.Error(w, "Please select a cohort", http.StatusBadRequest)
			return
		}
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ?", cohortName)
		if err != nil {
			http.Error(w, "Failed to query cohort students", http.StatusInternalServerError)
			return
		}
		for rows.Next() {
			var studentID int
			rows.Scan(&studentID)
			targetStudentIDs = append(targetStudentIDs, studentID)
		}
		rows.Close()
	case "students":
		r.ParseForm()
		studentIDs := r.Form["students"]
		if len(studentIDs) == 0 {
			http.Error(w, "Please select at least one student", http.StatusBadRequest)
			return
		}
		for _, idStr := range studentIDs {
			id, err := strconv.Atoi(idStr)
			if err == nil {
				targetStudentIDs = append(targetStudentIDs, id)
			}
		}
	default:
		http.Error(w, "Invalid assignment type", http.StatusBadRequest)
		return
	}

	if len(targetStudentIDs) == 0 {
		http.Error(w, "No students found for assignment", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("csvfile")
	if err != nil {
		http.Error(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read and parse CSV using encoding/csv for proper CSV handling
	var importedTasks []string
	var failedImports []string
	lineNum := 0
	tasksCreated := 0

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1 // Allow variable fields (5 expected)
	csvReader.TrimLeadingSpace = true

	for {
		record, err := csvReader.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			http.Error(w, "Failed to read CSV file", http.StatusBadRequest)
			return
		}

		lineNum++

		// Skip empty lines
		if len(record) == 0 || (len(record) == 1 && strings.TrimSpace(record[0]) == "") {
			continue
		}

		// Validate CSV has expected number of fields (5)
		if len(record) != 5 {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Invalid format (expected 5 fields: Task Title, Description, DueDate, TaskType, ResourceURL)", lineNum))
			continue
		}

		title := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])
		dueDate := strings.TrimSpace(record[2])
		taskType := strings.TrimSpace(record[3])
		resourceURL := strings.TrimSpace(record[4])

		if title == "" {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Task title is required", lineNum))
			continue
		}

		// Validate task type if provided
		validTaskTypes := map[string]bool{
			"group-programming-project": true,
			"individual-programming-lab": true,
			"reading":                    true,
			"document":                   true,
			"resume":                     true,
			"prodev":                     true,
			"":                           true, // Empty is allowed
		}
		if !validTaskTypes[taskType] {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Invalid task type '%s'", lineNum, taskType))
			continue
		}

		// Create task for each target student
		cohortID := ""
		if assignType == "cohort" {
			cohortID = r.FormValue("cohort")
		}

		for _, studentID := range targetStudentIDs {
			_, err = db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, taskType, resourceURL, cohortID)

			if err != nil {
				log.Printf("Failed to create task '%s' for student %d: %v", title, studentID, err)
			} else {
				tasksCreated++
			}
		}

		if tasksCreated > 0 {
			importedTasks = append(importedTasks, title)
		}
	}

	// Generate report
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, `<!DOCTYPE html>
<html>
<head>
    <title>Import Results - Tasking</title>
    <style>
        body { font-family: sans-serif; padding: 20px; background: #f5f7fa; }
        .container { max-width: 800px; margin: 0 auto; background: white; padding: 30px; border-radius: 10px; box-shadow: 0 2px 8px rgba(0,0,0,0.1); }
        h1 { color: #2C5F7C; }
        .success { color: #065f46; background: #d1fae5; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .error { color: #991b1b; background: #fee2e2; padding: 15px; border-radius: 5px; margin: 20px 0; }
        .list { margin: 10px 0; padding-left: 20px; }
        .btn { 
            display: inline-block;
            padding: 10px 20px; 
            background: #667eea; 
            color: white; 
            text-decoration: none; 
            border-radius: 5px; 
            margin-top: 20px;
        }
        .btn:hover { opacity: 0.9; }
    </style>
</head>
<body>
    <div class="container">
        <h1>Task Import Results</h1>
`)

	if len(importedTasks) > 0 {
		fmt.Fprintf(w, `<div class="success">
            <h2>Successfully Imported: %d task(s) assigned to %d student(s)</h2>
            <p><strong>Total tasks created:</strong> %d</p>
            <h3>Tasks imported:</h3>
            <ul class="list">`, len(importedTasks), len(targetStudentIDs), tasksCreated)
		for _, taskTitle := range importedTasks {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(taskTitle))
		}
		fmt.Fprintf(w, "</ul></div>")
	}

	if len(failedImports) > 0 {
		fmt.Fprintf(w, `<div class="error">
            <h2>Failed Imports: %d</h2>
            <ul class="list">`, len(failedImports))
		for _, failure := range failedImports {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(failure))
		}
		fmt.Fprintf(w, "</ul></div>")
	}

	if len(importedTasks) == 0 && len(failedImports) == 0 {
		fmt.Fprintf(w, `<div class="error">No tasks were imported. Please check your CSV file format.</div>`)
	}

	fmt.Fprintf(w, `
        <a href="/" class="btn">← Back to Tasks</a>
    </div>
</body>
</html>`)
}

// apiCohortsHandler returns list of unique cohorts as JSON
func apiCohortsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil || user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT DISTINCT cohort_name 
		FROM users 
		WHERE role = 'student' AND cohort_name IS NOT NULL AND cohort_name != '' 
		ORDER BY cohort_name
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var cohorts []string
	for rows.Next() {
		var cohort string
		rows.Scan(&cohort)
		cohorts = append(cohorts, cohort)
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cohorts); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// apiStudentsHandler returns list of students as JSON
func apiStudentsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil || user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	rows, err := db.Query(`
		SELECT id, username, COALESCE(cohort_name, '') 
		FROM users 
		WHERE role = 'student' 
		ORDER BY username
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []map[string]interface{}
	for rows.Next() {
		var id int
		var username, cohort string
		rows.Scan(&id, &username, &cohort)
		students = append(students, map[string]interface{}{
			"id":       id,
			"username": username,
			"cohort":   cohort,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(students); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Load templates with custom functions
	funcMap := template.FuncMap{
		"mul": func(a, b float64) float64 { return a * b },
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"formatDate": func(v interface{}) string {
			// Handle both string and time.Time inputs
			switch val := v.(type) {
			case string:
				// Parse SQLite timestamp format and convert to friendly format
				t, err := time.Parse("2006-01-02 15:04:05", val)
				if err != nil {
					// Try alternate format
					t, err = time.Parse(time.RFC3339, val)
					if err != nil {
						return val // Return as-is if parsing fails
					}
				}
				return t.Format("Jan 2, 2006")
			case time.Time:
				return val.Format("Jan 2, 2006")
			case *time.Time:
				if val != nil {
					return val.Format("Jan 2, 2006")
				}
				return ""
			default:
				return fmt.Sprintf("%v", v)
			}
		},
	}
	templates = template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))

	// Setup routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/add", addTaskHandler)
	http.HandleFunc("/tasks/edit", editTaskHandler)
	http.HandleFunc("/tasks/delete", deleteTaskHandler)
	http.HandleFunc("/tasks/update-status", updateTaskStatusHandler)
	http.HandleFunc("/tasks/import", importTasksHandler)
	http.HandleFunc("/students", studentsHandler)
	http.HandleFunc("/students/add", addStudentHandler)
	http.HandleFunc("/students/import", importStudentsHandler)
	http.HandleFunc("/students/edit", editStudentHandler)
	http.HandleFunc("/students/progress", studentProgressHandler)
	http.HandleFunc("/tasks/stats", taskStatsHandler)
	http.HandleFunc("/tasks/detail", taskDetailHandler)
	http.HandleFunc("/instructors", instructorsHandler)
	http.HandleFunc("/instructors/add", addInstructorHandler)
	http.HandleFunc("/api/cohorts", apiCohortsHandler)
	http.HandleFunc("/api/students", apiStudentsHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Tasking server starting on port %s...", port)
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
	var isSuspended bool
	err := db.QueryRow("SELECT id, username, role, COALESCE(is_suspended, 0) FROM users WHERE username = ? AND password = ?",
		username, password).Scan(&user.ID, &user.Username, &user.Role, &isSuspended)

	if err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid credentials"})
		return
	}

	// Check if account is suspended
	if isSuspended {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Account suspended. Please contact an instructor."})
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
		// Get all students and cohorts if instructor
		var students []User
		cohorts := make(map[string]bool)
		if user.Role == "instructor" {
			rows, err := db.Query("SELECT id, username, COALESCE(cohort_name, '') FROM users WHERE role = 'student'")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var s User
					rows.Scan(&s.ID, &s.Username, &s.CohortName)
					students = append(students, s)
					if s.CohortName != "" {
						cohorts[s.CohortName] = true
					}
				}
			}
		}

		// Convert cohorts map to slice
		var cohortList []string
		for cohort := range cohorts {
			cohortList = append(cohortList, cohort)
		}

		data := struct {
			User     User
			Students []User
			Cohorts  []string
		}{
			User:     user,
			Students: students,
			Cohorts:  cohortList,
		}
		templates.ExecuteTemplate(w, "add_task.html", data)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	assignedTo := r.FormValue("assigned_to")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

	// Students can only create tasks for themselves
	ownerID := user.ID
	cohortID := ""

	if user.Role == "instructor" && assignedTo != "" {
		// Check if assigning to a cohort
		if len(assignedTo) > 7 && assignedTo[:7] == "cohort:" {
			cohortID = assignedTo[7:]
			ownerID = 0 // Special case: assign to cohort
		} else if assignedTo == "all" {
			ownerID = 0 // Special case: assign to all students
		} else {
			// Parse student ID and validate it's actually a student
			ownerID, _ = strconv.Atoi(assignedTo)
			// Validate that the target user is a student, not an instructor
			var targetRole string
			err = db.QueryRow("SELECT role FROM users WHERE id = ?", ownerID).Scan(&targetRole)
			if err != nil || targetRole != "student" {
				http.Error(w, "Invalid assignment: can only assign tasks to students", http.StatusBadRequest)
				return
			}
		}
	}

	if ownerID == 0 && cohortID != "" {
		// Assign to all students in a cohort
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ?", cohortID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query cohort students: %v", err), http.StatusInternalServerError)
			return
		}
		// Collect all student IDs first
		var studentIDs []int
		for rows.Next() {
			var studentID int
			if err := rows.Scan(&studentID); err != nil {
				log.Printf("Failed to scan student ID: %v", err)
				continue
			}
			studentIDs = append(studentIDs, studentID)
		}
		rows.Close()

		// Now create tasks for each student
		for _, studentID := range studentIDs {
			_, err := db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, cohortID)
			if err != nil {
				log.Printf("Failed to create task for student %d: %v", studentID, err)
			}
		}
	} else if ownerID == 0 {
		// Assign to all students
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student'")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query students: %v", err), http.StatusInternalServerError)
			return
		}
		// Collect all student IDs first
		var studentIDs []int
		for rows.Next() {
			var studentID int
			if err := rows.Scan(&studentID); err != nil {
				log.Printf("Failed to scan student ID: %v", err)
				continue
			}
			studentIDs = append(studentIDs, studentID)
		}
		rows.Close()

		// Now create tasks for each student
		for _, studentID := range studentIDs {
			_, err := db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, "")
			if err != nil {
				log.Printf("Failed to create task for student %d: %v", studentID, err)
			}
		}
	} else {
		_, err = db.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, title, description, dueDate, ownerID, user.ID, assignedTo, taskType, resourceURL, cohortID)

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
			SELECT id, title, description, due_date, status, owner_id, created_by,
			       COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM tasks WHERE id = ?
		`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.TaskType, &task.ResourceURL)

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
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

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
		SET title = ?, description = ?, due_date = ?, status = ?, task_type = ?, resource_url = ?
		WHERE id = ?
	`, title, description, dueDate, status, taskType, resourceURL, taskID)

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
	var creatorRole string
	err = db.QueryRow(`
		SELECT t.owner_id, t.created_by, u.role
		FROM tasks t
		JOIN users u ON t.created_by = u.id
		WHERE t.id = ?
	`, taskID).Scan(&task.OwnerID, &task.CreatedBy, &creatorRole)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Students can only delete their own tasks
	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Instructors can delete tasks created by any instructor, or their own tasks
	if user.Role == "instructor" && creatorRole != "instructor" && task.OwnerID != user.ID {
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
	var cohortName sql.NullString
	err := db.QueryRow("SELECT id, username, role, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0) FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Role, &cohortName, &user.IsSuspended)
	if cohortName.Valid {
		user.CohortName = cohortName.String
	}
	return user, err
}

func getTasks(user User) ([]Task, error) {
	var tasks []Task
	var rows *sql.Rows
	var err error

	if user.Role == "instructor" {
		// Instructors see only tasks they created AND assigned to themselves
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
			       in_progress_date, done_date
			FROM tasks
			WHERE created_by = ? AND owner_id = ?
			ORDER BY created_at DESC
		`, user.ID, user.ID)
	} else {
		// Students see their own tasks
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
			       in_progress_date, done_date
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
		var inProgressDate, doneDate sql.NullTime
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt,
			&task.TaskType, &task.ResourceURL, &task.CohortID, &inProgressDate, &doneDate)
		if err != nil {
			return nil, err
		}
		if inProgressDate.Valid {
			task.InProgressDate = &inProgressDate.Time
		}
		if doneDate.Valid {
			task.DoneDate = &doneDate.Time
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
