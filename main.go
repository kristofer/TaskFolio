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
	"net/url"
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
	IsCurrent   bool   // Whether student is in current cohort (true) or past (false)
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
	GroupID        *int       // ID of the group this task is assigned to (if any)
	Comments       string     // User's notes and thoughts about the task
	DisplayOrder   int        // Order of the task within its status column
}

type Group struct {
	ID          int
	Name        string
	Description string
	ProjectURL  string // URL for group project repository
	IsActive    bool   // Whether the group is currently active
	CreatedAt   time.Time
	Students    []int // Student IDs in this group
}

type TaskStats struct {
	Total      int
	Pending    int
	InProgress int
	Blocked    int
	Completed  int
}

type TaskTemplate struct {
	ID          int
	Title       string
	Description string
	TaskType    string // Same as Task task types
	ResourceURL string
	CreatedAt   time.Time
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

	// Enable WAL mode for better concurrent read performance and durability.
	// WAL (Write-Ahead Logging) allows readers to proceed without blocking writers
	// and vice versa. This is a no-op for in-memory databases.
	if _, err = db.Exec("PRAGMA journal_mode=WAL"); err != nil {
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
			is_suspended BOOLEAN DEFAULT 0,
			is_current BOOLEAN DEFAULT 1
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
			task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', 'research-analysis-design', 'organize', '')),
			resource_url TEXT,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			group_id INTEGER,
			comments TEXT,
			display_order INTEGER DEFAULT 0,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id),
			FOREIGN KEY (group_id) REFERENCES groups(id)
		)
	`)
	if err != nil {
		return err
	}

	// Create groups table
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			project_url TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		return err
	}

	// Create group_members table for many-to-many relationship
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS group_members (
			group_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			PRIMARY KEY (group_id, student_id),
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create task_templates table for instructor task template library
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS task_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', 'research-analysis-design', 'organize', '')),
			resource_url TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
		INSERT OR IGNORE INTO users (id, username, password, role, cohort_name, is_suspended, is_current) VALUES
		(1, 'instructor1', 'password123', 'instructor', NULL, 0, 1),
		(2, 'student1', 'password123', 'student', 'Java11.2', 0, 1),
		(3, 'student2', 'password123', 'student', 'Java11.2', 0, 1),
		(4, 'student3', 'password123', 'student', 'Data6.2', 0, 1)
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

	// Check and add is_current column to users table
	var isCurrentExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='is_current'
	`).Scan(&isCurrentExists)
	if err == nil && isCurrentExists == 0 {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN is_current BOOLEAN DEFAULT 1`)
		if err != nil {
			return fmt.Errorf("failed to add is_current column: %v", err)
		}
		log.Println("Added is_current column to users table")

		// Set existing students to is_current = 1 (true)
		_, err = db.Exec(`UPDATE users SET is_current = 1 WHERE role = 'student' AND is_current IS NULL`)
		if err != nil {
			log.Printf("Warning: failed to initialize is_current values: %v", err)
		} else {
			log.Println("Initialized is_current to true for existing students")
		}
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

	// Check and add group_id column to tasks table
	var groupIDExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='group_id'
	`).Scan(&groupIDExists)
	if err == nil && groupIDExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN group_id INTEGER REFERENCES groups(id)`)
		if err != nil {
			return fmt.Errorf("failed to add group_id column: %v", err)
		}
		log.Println("Added group_id column to tasks table")
	}

	// Check and add comments column to tasks table
	var commentsExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='comments'
	`).Scan(&commentsExists)
	if err == nil && commentsExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN comments TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add comments column: %v", err)
		}
		log.Println("Added comments column to tasks table")
	}

	// Check and add display_order column to tasks table
	var displayOrderExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='display_order'
	`).Scan(&displayOrderExists)
	if err == nil && displayOrderExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN display_order INTEGER DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add display_order column: %v", err)
		}
		log.Println("Added display_order column to tasks table")

		// Initialize display_order for existing tasks based on created_at
		// Order within each status/owner_id combination
		_, err = db.Exec(`
			UPDATE tasks SET display_order = (
				SELECT COUNT(*) FROM tasks t2 
				WHERE t2.owner_id = tasks.owner_id 
				AND t2.status = tasks.status 
				AND t2.created_at <= tasks.created_at
			)
		`)
		if err != nil {
			log.Printf("Warning: failed to initialize display_order values: %v", err)
		} else {
			log.Println("Initialized display_order values for existing tasks")
		}
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
			task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', 'research-analysis-design', 'organize', '')),
			resource_url TEXT,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			group_id INTEGER,
			comments TEXT,
			display_order INTEGER DEFAULT 0,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id),
			FOREIGN KEY (group_id) REFERENCES groups(id)
		)
	`)
	if err != nil {
		return err
	}

	// Copy data from old table to new table, mapping old task types to new ones
	// Old task types: 'reading', 'coding', 'submission', 'document-creation'
	// New task types: 'group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev'
	// Mapping: 'coding' -> 'individual-programming-lab', 'submission' -> '', 'document-creation' -> 'document'

	// Check which columns exist in the old table
	var groupIDExists, commentsExists, displayOrderExists int
	tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='group_id'`).Scan(&groupIDExists)
	tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='comments'`).Scan(&commentsExists)
	tx.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='display_order'`).Scan(&displayOrderExists)

	// Build SELECT statement based on existing columns
	groupIDSelect := "NULL"
	if groupIDExists > 0 {
		groupIDSelect = "group_id"
	}
	commentsSelect := "''"
	if commentsExists > 0 {
		commentsSelect = "COALESCE(comments, '')"
	}
	displayOrderSelect := "0"
	if displayOrderExists > 0 {
		displayOrderSelect = "COALESCE(display_order, 0)"
	}

	insertSQL := fmt.Sprintf(`
		INSERT INTO tasks_new (id, title, description, due_date, status, owner_id, created_by, assigned_to, created_at, task_type, resource_url, cohort_id, in_progress_date, done_date, group_id, comments, display_order)
		SELECT id, title, description, due_date, status, owner_id, created_by,
		       COALESCE(assigned_to, ''), created_at,
		       CASE COALESCE(task_type, '')
		           WHEN 'coding' THEN 'individual-programming-lab'
		           WHEN 'submission' THEN ''
		           WHEN 'document-creation' THEN 'document'
		           ELSE COALESCE(task_type, '')
		       END,
		       COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
		       in_progress_date, done_date, %s, %s, %s
		FROM tasks
	`, groupIDSelect, commentsSelect, displayOrderSelect)

	_, err = tx.Exec(insertSQL)
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
// or is missing new task types like 'demo-presentation', 'research-analysis-design', or 'organize'
func containsOldTaskTypes(sql string) bool {
	oldTypes := []string{"'coding'", "'submission'", "'document-creation'"}
	for _, oldType := range oldTypes {
		if strings.Contains(sql, oldType) {
			return true
		}
	}
	// Also check if the schema is missing any of the newer task types
	// This handles cases where the database was created before these types were added
	newTypes := []string{"'demo-presentation'", "'research-analysis-design'", "'organize'"}
	for _, newType := range newTypes {
		if strings.Contains(sql, "task_type") && !strings.Contains(sql, newType) {
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
		SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0), COALESCE(is_current, 1)
		FROM users WHERE role = 'student' ORDER BY username
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type StudentWithStats struct {
		User
		Stats TaskStats
	}

	var currentStudents []StudentWithStats
	var pastStudents []StudentWithStats
	for rows.Next() {
		var s User
		s.Role = "student"
		rows.Scan(&s.ID, &s.Username, &s.CohortName, &s.IsSuspended, &s.IsCurrent)

		// Get task statistics for this student
		stats, err := getTaskStatsForStudent(s.ID)
		if err != nil {
			log.Printf("Error getting task stats for student %d: %v", s.ID, err)
			stats = TaskStats{} // Use empty stats on error
		}

		studentWithStats := StudentWithStats{
			User:  s,
			Stats: stats,
		}

		if s.IsCurrent {
			currentStudents = append(currentStudents, studentWithStats)
		} else {
			pastStudents = append(pastStudents, studentWithStats)
		}
	}

	data := struct {
		User            User
		CurrentStudents []StudentWithStats
		PastStudents    []StudentWithStats
	}{
		User:            user,
		CurrentStudents: currentStudents,
		PastStudents:    pastStudents,
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
			SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0), COALESCE(is_current, 1)
			FROM users WHERE id = ? AND role = 'student'
		`, studentID).Scan(&student.ID, &student.Username, &cohortName, &student.IsSuspended, &student.IsCurrent)

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
	isCurrent := r.FormValue("is_current") == "on"
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Update student information
	// If password is provided, update it; otherwise, only update other fields
	if password != "" {
		_, err = db.Exec(`
			UPDATE users
			SET username = ?, cohort_name = ?, is_suspended = ?, is_current = ?, password = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, isCurrent, password, studentID)
	} else {
		_, err = db.Exec(`
			UPDATE users
			SET username = ?, cohort_name = ?, is_suspended = ?, is_current = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, isCurrent, studentID)
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
		INSERT INTO users (username, password, role, cohort_name, is_suspended, is_current)
		VALUES (?, ?, 'student', ?, 0, 1)
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

// moveStudentsHandler handles moving students between current and past status
func moveStudentsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can move students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get student IDs and action from the form
	err = r.ParseForm()
	if err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	studentIDs := r.Form["student_ids"]
	action := r.FormValue("action") // "move_to_past" or "move_to_current"

	if len(studentIDs) == 0 {
		http.Redirect(w, r, "/students", http.StatusSeeOther)
		return
	}

	// Determine the new is_current value based on action
	var isCurrent bool
	if action == "move_to_current" {
		isCurrent = true
	} else if action == "move_to_past" {
		isCurrent = false
	} else {
		http.Error(w, "Invalid action", http.StatusBadRequest)
		return
	}

	// Update students in a transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, studentID := range studentIDs {
		_, err = tx.Exec(`
			UPDATE users
			SET is_current = ?
			WHERE id = ? AND role = 'student'
		`, isCurrent, studentID)
		if err != nil {
			http.Error(w, "Failed to update student", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit changes", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// groupsHandler displays list of all groups for instructors to manage
func groupsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get all groups
	groups, err := getAllGroups()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User   User
		Groups []Group
	}{
		User:   user,
		Groups: groups,
	}

	templates.ExecuteTemplate(w, "groups.html", data)
}

// groupDetailHandler displays detailed information about a specific group
func groupDetailHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get group ID from query parameter
	groupIDStr := r.URL.Query().Get("id")
	if groupIDStr == "" {
		http.Error(w, "Group ID is required", http.StatusBadRequest)
		return
	}

	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Get group details
	group, err := getGroupByID(groupID)
	if err != nil {
		http.Error(w, "Group not found", http.StatusNotFound)
		return
	}

	// Get student details for all group members
	var students []User
	if len(group.Students) > 0 {
		placeholders := make([]string, len(group.Students))
		args := make([]interface{}, len(group.Students))
		for i, studentID := range group.Students {
			placeholders[i] = "?"
			args[i] = studentID
		}

		query := fmt.Sprintf(`
			SELECT id, username, COALESCE(cohort_name, ''), role
			FROM users WHERE id IN (%s) ORDER BY username
		`, strings.Join(placeholders, ","))

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var s User
			rows.Scan(&s.ID, &s.Username, &s.CohortName, &s.Role)
			students = append(students, s)
		}
	}

	// Get tasks assigned to this group
	var tasks []Task
	taskRows, err := db.Query(`
		SELECT DISTINCT id, title, due_date, COALESCE(task_type, '')
		FROM tasks
		WHERE group_id = ?
		ORDER BY created_at DESC
	`, groupID)
	if err == nil {
		defer taskRows.Close()
		for taskRows.Next() {
			var t Task
			taskRows.Scan(&t.ID, &t.Title, &t.DueDate, &t.TaskType)
			tasks = append(tasks, t)
		}
	}

	data := struct {
		User     User
		Group    Group
		Students []User
		Tasks    []Task
	}{
		User:     user,
		Group:    group,
		Students: students,
		Tasks:    tasks,
	}

	templates.ExecuteTemplate(w, "group_detail.html", data)
}

// addGroupHandler handles creating new groups
func addGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can add groups
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		// Get all students for selection
		rows, err := db.Query(`
			SELECT id, username, COALESCE(cohort_name, '')
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
			rows.Scan(&s.ID, &s.Username, &s.CohortName)
			students = append(students, s)
		}

		data := struct {
			User     User
			Students []User
			Error    string
		}{
			User:     user,
			Students: students,
		}
		templates.ExecuteTemplate(w, "add_group.html", data)
		return
	}

	// Handle POST - create the group
	name := r.FormValue("name")
	description := r.FormValue("description")
	projectURL := r.FormValue("project_url")
	isActive := r.FormValue("is_active") == "on"
	studentIDs := r.Form["student_ids"]

	// Validate required fields
	if name == "" {
		// Get students for re-rendering form
		rows, _ := db.Query(`
			SELECT id, username, COALESCE(cohort_name, '')
			FROM users WHERE role = 'student' ORDER BY username
		`)
		defer rows.Close()

		var students []User
		for rows.Next() {
			var s User
			rows.Scan(&s.ID, &s.Username, &s.CohortName)
			students = append(students, s)
		}

		data := struct {
			User     User
			Students []User
			Error    string
		}{
			User:     user,
			Students: students,
			Error:    "Group name is required",
		}
		templates.ExecuteTemplate(w, "add_group.html", data)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Insert the group
	result, err := tx.Exec(`
		INSERT INTO groups (name, description, project_url, is_active)
		VALUES (?, ?, ?, ?)
	`, name, description, projectURL, isActive)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert group members
	for _, studentIDStr := range studentIDs {
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			continue
		}
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}

// editGroupHandler handles editing group details and members
func editGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can edit groups
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	groupIDStr := r.URL.Query().Get("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	if r.Method == "GET" {
		group, err := getGroupByID(groupID)
		if err != nil {
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		// Get all students for selection
		rows, err := db.Query(`
			SELECT id, username, COALESCE(cohort_name, '')
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
			rows.Scan(&s.ID, &s.Username, &s.CohortName)
			students = append(students, s)
		}

		data := struct {
			User     User
			Group    Group
			Students []User
		}{
			User:     user,
			Group:    group,
			Students: students,
		}
		templates.ExecuteTemplate(w, "edit_group.html", data)
		return
	}

	// Handle POST - update the group
	name := r.FormValue("name")
	description := r.FormValue("description")
	projectURL := r.FormValue("project_url")
	isActive := r.FormValue("is_active") == "on"
	studentIDs := r.Form["student_ids"]

	// Validate required fields
	if name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
		return
	}

	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	// Update the group
	_, err = tx.Exec(`
		UPDATE groups
		SET name = ?, description = ?, project_url = ?, is_active = ?
		WHERE id = ?
	`, name, description, projectURL, isActive, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Delete existing group members
	_, err = tx.Exec(`DELETE FROM group_members WHERE group_id = ?`, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Insert new group members
	for _, studentIDStr := range studentIDs {
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			continue
		}
		_, err = tx.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/groups", http.StatusSeeOther)
}

// deleteGroupHandler handles deleting a group
func deleteGroupHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can delete groups
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	groupIDStr := r.URL.Query().Get("id")
	groupID, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Invalid group ID", http.StatusBadRequest)
		return
	}

	// Delete the group (group_members will cascade delete)
	_, err = db.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/groups", http.StatusSeeOther)
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
		SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
		       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
		       in_progress_date, done_date, group_id, COALESCE(comments, ''), COALESCE(display_order, 0)
		FROM tasks
		WHERE owner_id = ?
		ORDER BY status, display_order, created_at DESC
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
		var inProgressDate, doneDate sql.NullTime
		var groupID sql.NullInt64
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt,
			&task.TaskType, &task.ResourceURL, &task.CohortID, &inProgressDate, &doneDate, &groupID, &task.Comments, &task.DisplayOrder)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if inProgressDate.Valid {
			task.InProgressDate = &inProgressDate.Time
		}
		if doneDate.Valid {
			task.DoneDate = &doneDate.Time
		}
		if groupID.Valid {
			gid := int(groupID.Int64)
			task.GroupID = &gid
		}
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

	// Create a struct to hold task with group name for the template
	type TaskWithGroup struct {
		Task
		GroupName string
	}

	var tasksWithGroups []TaskWithGroup
	for _, task := range tasks {
		twg := TaskWithGroup{Task: task}
		if task.GroupID != nil {
			twg.GroupName = getGroupNameByID(*task.GroupID)
		}
		tasksWithGroups = append(tasksWithGroups, twg)
	}

	data := struct {
		User                 User
		Student              User
		Tasks                []TaskWithGroup
		TotalTasks           int
		CompletedTasks       int
		InProgressTasks      int
		PendingTasks         int
		CompletionPercentage float64
	}{
		User:                 user,
		Student:              student,
		Tasks:                tasksWithGroups,
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
		Title                string
		Description          string
		DueDate              string
		TaskType             string
		ResourceURL          string
		CreatedBy            int
		CreatorName          string
		CohortID             string
		CreatedAt            string
		TotalCount           int
		InProgress           int
		Completed            int
		CompletionPercentage int
	}

	rows, err := db.Query(`
		SELECT
			t.title,
			COALESCE(t.description, '') as description,
			COALESCE(t.due_date, '') as due_date,
			COALESCE(t.task_type, '') as task_type,
			COALESCE(t.resource_url, '') as resource_url,
			t.created_by,
			creator.username as creator_name,
			COALESCE(t.cohort_id, '') as cohort_id,
			MIN(t.created_at) as created_at,
			COUNT(*) as total_count,
			SUM(CASE WHEN t.status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks t
		JOIN users u ON t.owner_id = u.id
		JOIN users creator ON t.created_by = creator.id
		WHERE t.created_by IN (SELECT id FROM users WHERE role = 'instructor')
		AND u.role = 'student'
		GROUP BY t.title, t.created_by, creator.username, COALESCE(t.cohort_id, '')
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
			&gt.ResourceURL, &gt.CreatedBy, &gt.CreatorName, &gt.CohortID, &gt.CreatedAt,
			&gt.TotalCount, &gt.InProgress, &gt.Completed)
		if err != nil {
			log.Printf("Error scanning group task: %v", err)
			continue
		}
		// Calculate completion percentage
		if gt.TotalCount > 0 {
			gt.CompletionPercentage = (gt.Completed * 100) / gt.TotalCount
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

// exportGroupTaskHandler exports a single group task as CSV in the import format
func exportGroupTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can export tasks
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		selected := r.Form["task"]
		if len(selected) == 0 {
			http.Error(w, "No tasks selected", http.StatusBadRequest)
			return
		}

		stmt, err := db.Prepare(`
			SELECT title, COALESCE(description, ''), COALESCE(due_date, ''),
			       COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM tasks
			WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
			LIMIT 1
		`)
		if err != nil {
			http.Error(w, "Failed to prepare export query", http.StatusInternalServerError)
			return
		}
		defer stmt.Close()

		records := make([][]string, 0, len(selected))
		for _, item := range selected {
			parsed, err := url.ParseQuery(item)
			if err != nil {
				http.Error(w, "Invalid task selection", http.StatusBadRequest)
				return
			}

			title := parsed.Get("title")
			createdByStr := parsed.Get("created_by")
			cohortID := parsed.Get("cohort_id")
			if title == "" || createdByStr == "" {
				http.Error(w, "Missing required parameters", http.StatusBadRequest)
				return
			}

			createdBy, err := strconv.Atoi(createdByStr)
			if err != nil {
				http.Error(w, "Invalid created_by parameter", http.StatusBadRequest)
				return
			}

			var taskTitle, description, dueDate, taskType, resourceURL string
			err = stmt.QueryRow(title, createdBy, cohortID).Scan(&taskTitle, &description, &dueDate, &taskType, &resourceURL)
			if err != nil {
				http.Error(w, "Task not found", http.StatusNotFound)
				return
			}

			records = append(records, []string{taskTitle, description, dueDate, taskType, resourceURL})
		}

		filename := "task_export_selected_" + time.Now().Format("20060102") + ".csv"
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(filename)+"\"")

		csvWriter := csv.NewWriter(w)
		defer csvWriter.Flush()

		for _, record := range records {
			if err := csvWriter.Write(record); err != nil {
				log.Printf("Error writing CSV: %v", err)
				return
			}
		}

		return
	}

	// Get parameters
	title := r.URL.Query().Get("title")
	createdByStr := r.URL.Query().Get("created_by")
	cohortID := r.URL.Query().Get("cohort_id")

	if title == "" || createdByStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	createdBy, err := strconv.Atoi(createdByStr)
	if err != nil {
		http.Error(w, "Invalid created_by parameter", http.StatusBadRequest)
		return
	}

	// Get the task details (get first instance for the task data)
	var taskTitle, description, dueDate, taskType, resourceURL string
	err = db.QueryRow(`
		SELECT title, COALESCE(description, ''), COALESCE(due_date, ''), 
		       COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM tasks
		WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
		LIMIT 1
	`, title, createdBy, cohortID).Scan(&taskTitle, &description, &dueDate, &taskType, &resourceURL)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Set headers for CSV download
	// Sanitize filename by removing potentially dangerous characters and limiting to alphanumeric, spaces, dashes, and underscores
	safeTitle := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ' || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, taskTitle)
	safeTitle = strings.ReplaceAll(safeTitle, " ", "_")
	// Limit filename length to prevent issues
	if len(safeTitle) > 50 {
		safeTitle = safeTitle[:50]
	}
	filename := "task_export_" + safeTitle + ".csv"

	w.Header().Set("Content-Type", "text/csv")
	// Properly escape the filename for RFC 6266 compliance
	w.Header().Set("Content-Disposition", "attachment; filename=\""+url.PathEscape(filename)+"\"")

	// Write CSV data in the import format: Task Title, Description, DueDate, TaskType, ResourceURL
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write the single task row
	record := []string{taskTitle, description, dueDate, taskType, resourceURL}
	if err := csvWriter.Write(record); err != nil {
		log.Printf("Error writing CSV: %v", err)
		return
	}
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

// deleteGroupTaskHandler deletes all instances of a task assigned to multiple students
func deleteGroupTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can delete group tasks
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get task parameters
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

	// Verify the task was created by an instructor
	var creatorRole string
	err = db.QueryRow("SELECT role FROM users WHERE id = ?", createdBy).Scan(&creatorRole)
	if err != nil {
		http.Error(w, "Creator not found", http.StatusNotFound)
		return
	}

	if creatorRole != "instructor" {
		http.Error(w, "Can only delete instructor-created tasks", http.StatusForbidden)
		return
	}

	// Delete all tasks with matching title, created_by, and cohort_id
	// Use explicit comparison for cohort_id to handle both NULL and empty string cases correctly
	var result sql.Result
	if cohortID == "" {
		result, err = db.Exec(`
			DELETE FROM tasks
			WHERE title = ? AND created_by = ? AND (cohort_id IS NULL OR cohort_id = '')
		`, taskTitle, createdBy)
	} else {
		result, err = db.Exec(`
			DELETE FROM tasks
			WHERE title = ? AND created_by = ? AND cohort_id = ?
		`, taskTitle, createdBy, cohortID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("Warning: No tasks were deleted for task '%s' created by user %d with cohort_id '%s'", taskTitle, createdBy, cohortID)
		http.Error(w, "Task not found or already deleted", http.StatusNotFound)
		return
	}

	log.Printf("Deleted %d task instances for task '%s' created by user %d", rowsAffected, taskTitle, createdBy)

	http.Redirect(w, r, "/tasks/stats", http.StatusSeeOther)
}

// editGroupTaskHandler provides a form to edit all instances of a task assigned to multiple students
func editGroupTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can edit group tasks
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	taskTitle := r.URL.Query().Get("title")
	createdByStr := r.URL.Query().Get("created_by")
	cohortID := r.URL.Query().Get("cohort_id")

	if r.Method == "GET" {
		if taskTitle == "" || createdByStr == "" {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		createdBy, err := strconv.Atoi(createdByStr)
		if err != nil {
			http.Error(w, "Invalid created_by parameter", http.StatusBadRequest)
			return
		}

		// Get task information from the first matching task
		var task Task
		err = db.QueryRow(`
			SELECT id, title, COALESCE(description, ''), COALESCE(due_date, ''),
			       status, owner_id, created_by,
			       COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM tasks
			WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
			ORDER BY created_at ASC
			LIMIT 1
		`, taskTitle, createdBy, cohortID).Scan(&task.ID, &task.Title, &task.Description,
			&task.DueDate, &task.Status, &task.OwnerID, &task.CreatedBy,
			&task.TaskType, &task.ResourceURL)

		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		data := struct {
			User     User
			Task     Task
			IsGroup  bool
			CohortID string
		}{
			User:     user,
			Task:     task,
			IsGroup:  true,
			CohortID: cohortID,
		}
		templates.ExecuteTemplate(w, "edit_task.html", data)
		return
	}

	// POST request - update all matching tasks
	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")
	oldTitle := r.FormValue("old_title")
	createdByStr = r.FormValue("created_by")
	cohortID = r.FormValue("cohort_id")

	if oldTitle == "" || createdByStr == "" {
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	createdBy, err := strconv.Atoi(createdByStr)
	if err != nil {
		http.Error(w, "Invalid created_by parameter", http.StatusBadRequest)
		return
	}

	// Verify the task was created by an instructor
	var creatorRole string
	err = db.QueryRow("SELECT role FROM users WHERE id = ?", createdBy).Scan(&creatorRole)
	if err != nil {
		http.Error(w, "Creator not found", http.StatusNotFound)
		return
	}

	if creatorRole != "instructor" {
		http.Error(w, "Can only edit instructor-created tasks", http.StatusForbidden)
		return
	}

	// Update all tasks with matching old_title, created_by, and cohort_id
	// Use explicit comparison for cohort_id to handle both NULL and empty string cases correctly
	var result sql.Result
	if cohortID == "" {
		result, err = db.Exec(`
			UPDATE tasks
			SET title = ?, description = ?, due_date = ?, task_type = ?, resource_url = ?
			WHERE title = ? AND created_by = ? AND (cohort_id IS NULL OR cohort_id = '')
		`, title, description, dueDate, taskType, resourceURL, oldTitle, createdBy)
	} else {
		result, err = db.Exec(`
			UPDATE tasks
			SET title = ?, description = ?, due_date = ?, task_type = ?, resource_url = ?
			WHERE title = ? AND created_by = ? AND cohort_id = ?
		`, title, description, dueDate, taskType, resourceURL, oldTitle, createdBy, cohortID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Error getting rows affected: %v", err)
	}

	if rowsAffected == 0 {
		log.Printf("Warning: No tasks were updated for task '%s' created by user %d with cohort_id '%s'", oldTitle, createdBy, cohortID)
		http.Error(w, "Task not found or no changes made", http.StatusNotFound)
		return
	}

	log.Printf("Updated %d task instances for task '%s' created by user %d", rowsAffected, oldTitle, createdBy)

	http.Redirect(w, r, "/tasks/stats", http.StatusSeeOther)
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

// taskTemplatesHandler displays list of task templates for instructors
func taskTemplatesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Get all task templates
	rows, err := db.Query(`
		SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, ''), created_at
		FROM task_templates ORDER BY created_at DESC
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taskTemplates []TaskTemplate
	for rows.Next() {
		var t TaskTemplate
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.ResourceURL, &t.CreatedAt)
		taskTemplates = append(taskTemplates, t)
	}

	data := struct {
		User      User
		Templates []TaskTemplate
	}{
		User:      user,
		Templates: taskTemplates,
	}

	templates.ExecuteTemplate(w, "task_templates.html", data)
}

// addTaskTemplateHandler handles adding a new task template
func addTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can add task templates
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

	// Validate required fields
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Insert new task template
	_, err = db.Exec(`
		INSERT INTO task_templates (title, description, task_type, resource_url)
		VALUES (?, ?, ?, ?)
	`, title, description, taskType, resourceURL)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/task-templates", http.StatusSeeOther)
}

// editTaskTemplateHandler handles editing a task template
func editTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can edit task templates
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := r.FormValue("id")
	title := r.FormValue("title")
	description := r.FormValue("description")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

	// Validate required fields
	if title == "" {
		http.Error(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Update task template
	_, err = db.Exec(`
		UPDATE task_templates
		SET title = ?, description = ?, task_type = ?, resource_url = ?
		WHERE id = ?
	`, title, description, taskType, resourceURL, templateID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/task-templates", http.StatusSeeOther)
}

// deleteTaskTemplateHandler handles deleting a task template
func deleteTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can delete task templates
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	templateID := r.FormValue("id")

	// Delete task template
	_, err = db.Exec(`DELETE FROM task_templates WHERE id = ?`, templateID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/task-templates", http.StatusSeeOther)
}

// importTaskTemplatesHandler handles bulk import of task templates from CSV file
func importTaskTemplatesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can import task templates
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

	// Read and parse CSV
	var importedTemplates []string
	var failedImports []string
	lineNum := 0

	csvReader := csv.NewReader(file)
	csvReader.FieldsPerRecord = -1 // Allow variable fields (4 expected)
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

		// Validate CSV has expected number of fields (4)
		if len(record) != 4 {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Invalid format (expected 4 fields: Title, Description, TaskType, ResourceURL)", lineNum))
			continue
		}

		title := strings.TrimSpace(record[0])
		description := strings.TrimSpace(record[1])
		taskType := strings.TrimSpace(record[2])
		resourceURL := strings.TrimSpace(record[3])

		if title == "" {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Title is required", lineNum))
			continue
		}

		// Validate task type if provided
		validTaskTypes := map[string]bool{
			"group-programming-project":  true,
			"individual-programming-lab": true,
			"reading":                    true,
			"document":                   true,
			"resume":                     true,
			"prodev":                     true,
			"demo-presentation":          true,
			"research-analysis-design":   true,
			"organize":                   true,
			"":                           true, // Empty is allowed
		}
		if !validTaskTypes[taskType] {
			failedImports = append(failedImports, fmt.Sprintf("Line %d: Invalid task type '%s'", lineNum, taskType))
			continue
		}

		// Insert task template
		_, err = db.Exec(`
			INSERT INTO task_templates (title, description, task_type, resource_url)
			VALUES (?, ?, ?, ?)
		`, title, description, taskType, resourceURL)

		if err != nil {
			failedImports = append(failedImports, fmt.Sprintf("%s: %v", title, err))
		} else {
			importedTemplates = append(importedTemplates, title)
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
        <h1>Task Template Import Results</h1>
`)

	if len(importedTemplates) > 0 {
		fmt.Fprintf(w, `<div class="success">
            <h2>Successfully Imported: %d template(s)</h2>
            <ul class="list">`, len(importedTemplates))
		for _, title := range importedTemplates {
			fmt.Fprintf(w, "<li>%s</li>", html.EscapeString(title))
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

	if len(importedTemplates) == 0 && len(failedImports) == 0 {
		fmt.Fprintf(w, `<div class="error">No templates were imported. Please check your CSV file format.</div>`)
	}

	fmt.Fprintf(w, `
        <a href="/task-templates" class="btn">← Back to Task Templates</a>
    </div>
</body>
</html>`)
}

// exportTaskTemplatesHandler handles exporting selected task templates to CSV
func exportTaskTemplatesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can export task templates
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	// Get selected template IDs
	templateIDs := r.Form["template_ids"]
	if len(templateIDs) == 0 {
		http.Error(w, "No templates selected", http.StatusBadRequest)
		return
	}

	// Build query to get selected templates
	placeholders := make([]string, len(templateIDs))
	args := make([]interface{}, len(templateIDs))
	for i, id := range templateIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM task_templates WHERE id IN (%s)
		ORDER BY created_at DESC
	`, strings.Join(placeholders, ","))

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// Set CSV headers
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=task_templates.csv")

	// Create CSV writer
	csvWriter := csv.NewWriter(w)
	defer csvWriter.Flush()

	// Write data
	for rows.Next() {
		var title, description, taskType, resourceURL string
		rows.Scan(&title, &description, &taskType, &resourceURL)
		csvWriter.Write([]string{title, description, taskType, resourceURL})
	}
}

// assignTaskTemplateHandler handles assigning a task template to students
func assignTaskTemplateHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can assign task templates
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get parameters
	templateID := r.FormValue("template_id")
	dueDate := r.FormValue("due_date")
	studentIDs := r.Form["student_ids"]
	
	// Get edited values from form (these may be different from template values)
	title := r.FormValue("title")
	description := r.FormValue("description")
	resourceURL := r.FormValue("resource_url")

	if templateID == "" || len(studentIDs) == 0 {
		http.Error(w, "Template ID and student IDs are required", http.StatusBadRequest)
		return
	}

	// Get the template to retrieve task_type (which is not editable in the modal)
	var template TaskTemplate
	err = db.QueryRow(`
		SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM task_templates WHERE id = ?
	`, templateID).Scan(&template.ID, &template.Title, &template.Description, &template.TaskType, &template.ResourceURL)

	if err != nil {
		http.Error(w, "Template not found", http.StatusNotFound)
		return
	}
	
	// Use edited values if provided, otherwise fall back to template values
	if title == "" {
		title = template.Title
	}
	if description == "" {
		description = template.Description
	}
	if resourceURL == "" {
		resourceURL = template.ResourceURL
	}

	// Create tasks for each student
	tasksCreated := 0
	for _, studentIDStr := range studentIDs {
		studentID, err := strconv.Atoi(studentIDStr)
		if err != nil {
			continue
		}

		_, err = db.Exec(`
			INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, task_type, resource_url)
			VALUES (?, ?, ?, 'pending', ?, ?, ?, ?)
		`, title, description, dueDate, studentID, user.ID, template.TaskType, resourceURL)

		if err == nil {
			tasksCreated++
		}
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"tasksCreated": tasksCreated,
	})
}

// apiTaskTemplatesHandler provides JSON API for task templates (for search/autocomplete)
func apiTaskTemplatesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Only instructors can access this API
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get search query if provided
	searchQuery := r.URL.Query().Get("q")

	var rows *sql.Rows
	if searchQuery != "" {
		rows, err = db.Query(`
			SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM task_templates
			WHERE title LIKE ? OR description LIKE ?
			ORDER BY title
			LIMIT 50
		`, "%"+searchQuery+"%", "%"+searchQuery+"%")
	} else {
		rows, err = db.Query(`
			SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM task_templates
			ORDER BY created_at DESC
			LIMIT 50
		`)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var taskTemplates []TaskTemplate
	for rows.Next() {
		var t TaskTemplate
		rows.Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.ResourceURL)
		taskTemplates = append(taskTemplates, t)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(taskTemplates)
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

// updateTaskCommentsHandler handles AJAX requests to update task comments
func updateTaskCommentsHandler(w http.ResponseWriter, r *http.Request) {
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

	taskID := r.URL.Query().Get("id")
	comments := r.FormValue("comments")

	// Get task and verify ownership
	var task Task
	err = db.QueryRow(`
		SELECT owner_id
		FROM tasks WHERE id = ?
	`, taskID).Scan(&task.OwnerID)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Check permissions - users can only update comments on their own tasks
	if task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Update comments
	_, err = db.Exec(`
		UPDATE tasks
		SET comments = ?
		WHERE id = ?
	`, comments, taskID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": true}`))
}

// reorderTaskHandler handles reordering tasks within a column
func reorderTaskHandler(w http.ResponseWriter, r *http.Request) {
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

	// Parse the request body to get the task order
	var requestData struct {
		TaskID  int    `json:"taskId"`
		Status  string `json:"status"`
		TaskIDs []int  `json:"taskIds"` // New order of task IDs in the column
	}

	err = json.NewDecoder(r.Body).Decode(&requestData)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	// Verify all tasks belong to the user
	for _, taskID := range requestData.TaskIDs {
		var ownerID int
		err = db.QueryRow(`
			SELECT owner_id FROM tasks WHERE id = ? AND status = ?
		`, taskID, requestData.Status).Scan(&ownerID)

		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		if ownerID != user.ID {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
	}

	// Update display_order for each task
	tx, err := db.Begin()
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for i, taskID := range requestData.TaskIDs {
		_, err = tx.Exec(`
			UPDATE tasks
			SET display_order = ?
			WHERE id = ? AND owner_id = ?
		`, i, taskID, user.ID)

		if err != nil {
			http.Error(w, "Failed to update task order", http.StatusInternalServerError)
			return
		}
	}

	err = tx.Commit()
	if err != nil {
		http.Error(w, "Failed to commit changes", http.StatusInternalServerError)
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
	// IMPORTANT: All modes filter by is_current = 1 to prevent assignment to past students
	var targetStudentIDs []int
	switch assignType {
	case "all":
		// "All Students" mode: Frontend pre-selects all current students; instructor can deselect via Ctrl/Cmd-click
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
	case "cohort":
		cohortName := r.FormValue("cohort")
		if cohortName == "" {
			http.Error(w, "Please select a cohort", http.StatusBadRequest)
			return
		}
		// Query only current students in the cohort
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ? AND is_current = 1", cohortName)
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
		// Specific students mode - form data already parsed by ParseMultipartForm above
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
			"group-programming-project":  true,
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

		taskSuccessCount := 0
		for _, studentID := range targetStudentIDs {
			_, err = db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, taskType, resourceURL, cohortID)

			if err != nil {
				log.Printf("Failed to create task '%s' for student %d: %v", title, studentID, err)
			} else {
				tasksCreated++
				taskSuccessCount++
			}
		}

		// Only add to imported list if at least one instance was created successfully
		if taskSuccessCount > 0 {
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

// apiCohortsHandler returns list of unique cohorts as JSON.
// IMPORTANT: Only returns cohorts of CURRENT students (is_current = 1).
// This ensures import tasks functionality only shows active cohorts.
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

	// Query only cohorts from current students to prevent assignment to past students
	rows, err := db.Query(`
		SELECT DISTINCT cohort_name 
		FROM users 
		WHERE role = 'student' AND cohort_name IS NOT NULL AND cohort_name != '' AND is_current = 1
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

// apiStudentsHandler returns list of students as JSON.
// IMPORTANT: Only returns CURRENT students (is_current = 1).
// This ensures import tasks functionality only shows active students.
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

	// Query only current students to prevent assignment to past students
	rows, err := db.Query(`
		SELECT id, username, COALESCE(cohort_name, '') 
		FROM users 
		WHERE role = 'student' AND is_current = 1
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

// toFloat64 converts various numeric types to float64
func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case uint:
		return float64(val)
	case uint64:
		return float64(val)
	case uint32:
		return float64(val)
	default:
		return 0
	}
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Load templates with custom functions
	funcMap := template.FuncMap{
		"mul": func(a, b interface{}) float64 {
			aVal := toFloat64(a)
			bVal := toFloat64(b)
			return aVal * bVal
		},
		"div": func(a, b interface{}) float64 {
			aVal := toFloat64(a)
			bVal := toFloat64(b)
			if bVal == 0 {
				return 0
			}
			return aVal / bVal
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
		"normalizeURL": func(url string) string {
			// Normalize URLs by adding https:// if no protocol is present
			url = strings.TrimSpace(url)
			if url == "" {
				return ""
			}
			// Check if URL already has a protocol (http:// or https://)
			if !strings.HasPrefix(strings.ToLower(url), "http://") &&
				!strings.HasPrefix(strings.ToLower(url), "https://") {
				return "https://" + url
			}
			return url
		},
		"urlquery": func(s string) string {
			// URL encode a string for use in query parameters
			return url.QueryEscape(s)
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
	http.HandleFunc("/tasks/update-comments", updateTaskCommentsHandler)
	http.HandleFunc("/tasks/reorder", reorderTaskHandler)
	http.HandleFunc("/tasks/import", importTasksHandler)
	http.HandleFunc("/students", studentsHandler)
	http.HandleFunc("/students/add", addStudentHandler)
	http.HandleFunc("/students/import", importStudentsHandler)
	http.HandleFunc("/students/move", moveStudentsHandler)
	http.HandleFunc("/students/edit", editStudentHandler)
	http.HandleFunc("/students/progress", studentProgressHandler)
	http.HandleFunc("/groups", groupsHandler)
	http.HandleFunc("/groups/detail", groupDetailHandler)
	http.HandleFunc("/groups/add", addGroupHandler)
	http.HandleFunc("/groups/edit", editGroupHandler)
	http.HandleFunc("/groups/delete", deleteGroupHandler)
	http.HandleFunc("/tasks/stats", taskStatsHandler)
	http.HandleFunc("/tasks/export", exportGroupTaskHandler)
	http.HandleFunc("/tasks/detail", taskDetailHandler)
	http.HandleFunc("/tasks/delete-group", deleteGroupTaskHandler)
	http.HandleFunc("/tasks/edit-group", editGroupTaskHandler)
	http.HandleFunc("/instructors", instructorsHandler)
	http.HandleFunc("/instructors/add", addInstructorHandler)
	http.HandleFunc("/api/cohorts", apiCohortsHandler)
	http.HandleFunc("/api/students", apiStudentsHandler)
	http.HandleFunc("/task-templates", taskTemplatesHandler)
	http.HandleFunc("/task-templates/add", addTaskTemplateHandler)
	http.HandleFunc("/task-templates/edit", editTaskTemplateHandler)
	http.HandleFunc("/task-templates/delete", deleteTaskTemplateHandler)
	http.HandleFunc("/task-templates/import", importTaskTemplatesHandler)
	http.HandleFunc("/task-templates/export", exportTaskTemplatesHandler)
	http.HandleFunc("/task-templates/assign", assignTaskTemplateHandler)
	http.HandleFunc("/api/task-templates", apiTaskTemplatesHandler)

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

	// Create a struct to hold task with group name
	type TaskWithGroup struct {
		Task
		GroupName string
	}

	var tasksWithGroups []TaskWithGroup
	for _, task := range tasks {
		twg := TaskWithGroup{Task: task}
		if task.GroupID != nil {
			twg.GroupName = getGroupNameByID(*task.GroupID)
		}
		tasksWithGroups = append(tasksWithGroups, twg)
	}

	data := struct {
		User  User
		Tasks []TaskWithGroup
	}{
		User:  user,
		Tasks: tasksWithGroups,
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
		var groups []Group
		cohorts := make(map[string]bool)
		filteredStudents := r.URL.Query().Get("filtered_students")

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

			// Get all active groups
			groups, _ = getAllGroups()
		}

		// Convert cohorts map to slice
		var cohortList []string
		for cohort := range cohorts {
			cohortList = append(cohortList, cohort)
		}

		// Get pre-selected group ID if provided
		selectedGroupID := r.URL.Query().Get("group_id")

		data := struct {
			User             User
			Students         []User
			Cohorts          []string
			Groups           []Group
			FilteredStudents string
			SelectedGroupID  string
		}{
			User:             user,
			Students:         students,
			Cohorts:          cohortList,
			Groups:           groups,
			FilteredStudents: filteredStudents,
			SelectedGroupID:  selectedGroupID,
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
	filteredStudents := r.FormValue("filtered_students")

	// Students can only create tasks for themselves
	ownerID := user.ID
	cohortID := ""

	if user.Role == "instructor" && assignedTo != "" {
		// Check if assigning to filtered students
		if assignedTo == "filtered" {
			if filteredStudents == "" {
				http.Error(w, "No filtered students provided", http.StatusBadRequest)
				return
			}
			// Parse the comma-separated student IDs
			studentIDStrs := strings.Split(filteredStudents, ",")
			var studentIDs []int
			var args []interface{}

			for _, idStr := range studentIDStrs {
				studentID, err := strconv.Atoi(strings.TrimSpace(idStr))
				if err != nil {
					log.Printf("Invalid student ID in filtered list: %s", idStr)
					continue
				}
				studentIDs = append(studentIDs, studentID)
				args = append(args, studentID)
			}

			if len(studentIDs) == 0 {
				http.Error(w, "No valid student IDs provided", http.StatusBadRequest)
				return
			}

			// Build placeholders for IN clause - simpler approach
			placeholderSlice := make([]string, len(studentIDs))
			for i := range placeholderSlice {
				placeholderSlice[i] = "?"
			}
			placeholders := strings.Join(placeholderSlice, ",")

			// Validate all student IDs at once
			query := fmt.Sprintf("SELECT id FROM users WHERE id IN (%s) AND role = 'student'", placeholders)
			rows, err := db.Query(query, args...)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to validate student IDs: %v", err), http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			validStudentIDs := make(map[int]bool)
			for rows.Next() {
				var id int
				if err := rows.Scan(&id); err != nil {
					log.Printf("Failed to scan student ID: %v", err)
					continue
				}
				validStudentIDs[id] = true
			}

			// Check if any invalid IDs were provided
			invalidCount := 0
			for _, id := range studentIDs {
				if !validStudentIDs[id] {
					invalidCount++
					log.Printf("Invalid or non-student ID: %d", id)
				}
			}

			if invalidCount > 0 && len(validStudentIDs) == 0 {
				http.Error(w, "None of the provided IDs are valid students", http.StatusBadRequest)
				return
			}

			// Create tasks for valid students
			for id := range validStudentIDs {
				_, err = db.Exec(`
					INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id, group_id)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, title, description, dueDate, id, user.ID, "", taskType, resourceURL, "", nil)
				if err != nil {
					log.Printf("Failed to create task for student %d: %v", id, err)
				}
			}

			if invalidCount > 0 {
				log.Printf("Warning: %d invalid student IDs were skipped during task creation", invalidCount)
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		// Check if assigning to a group
		if len(assignedTo) > 6 && assignedTo[:6] == "group:" {
			groupIDStr := assignedTo[6:]
			groupID, err := strconv.Atoi(groupIDStr)
			if err != nil {
				http.Error(w, "Invalid group ID", http.StatusBadRequest)
				return
			}

			// Get group members
			group, err := getGroupByID(groupID)
			if err != nil {
				http.Error(w, "Group not found", http.StatusNotFound)
				return
			}

			// Create tasks for each group member
			for _, studentID := range group.Students {
				_, err := db.Exec(`
					INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id, group_id)
					VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
				`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, "", groupID)
				if err != nil {
					log.Printf("Failed to create task for student %d in group %d: %v", studentID, groupID, err)
				}
			}

			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

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
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id, group_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, cohortID, nil)
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
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id, group_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, "", nil)
			if err != nil {
				log.Printf("Failed to create task for student %d: %v", studentID, err)
			}
		}
	} else {
		_, err = db.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id, group_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, title, description, dueDate, ownerID, user.ID, assignedTo, taskType, resourceURL, cohortID, nil)

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
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(comments, '')
			FROM tasks WHERE id = ?
		`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.TaskType, &task.ResourceURL, &task.Comments)

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
	comments := r.FormValue("comments")

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
		SET title = ?, description = ?, due_date = ?, status = ?, task_type = ?, resource_url = ?, comments = ?
		WHERE id = ?
	`, title, description, dueDate, status, taskType, resourceURL, comments, taskID)

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

// getGroupByID retrieves a group by its ID
func getGroupByID(id int) (Group, error) {
	var group Group
	var projectURL sql.NullString
	var description sql.NullString

	err := db.QueryRow(`
		SELECT id, name, COALESCE(description, ''), COALESCE(project_url, ''), 
		       COALESCE(is_active, 1), created_at 
		FROM groups WHERE id = ?
	`, id).Scan(&group.ID, &group.Name, &description, &projectURL, &group.IsActive, &group.CreatedAt)

	if err != nil {
		return group, err
	}

	if description.Valid {
		group.Description = description.String
	}
	if projectURL.Valid {
		group.ProjectURL = projectURL.String
	}

	// Get student IDs for this group
	rows, err := db.Query(`
		SELECT student_id FROM group_members WHERE group_id = ?
	`, id)
	if err != nil {
		return group, err
	}
	defer rows.Close()

	group.Students = []int{}
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			return group, err
		}
		group.Students = append(group.Students, studentID)
	}

	return group, nil
}

// getAllGroups retrieves all groups
func getAllGroups() ([]Group, error) {
	rows, err := db.Query(`
		SELECT id, name, COALESCE(description, ''), COALESCE(project_url, ''), 
		       COALESCE(is_active, 1), created_at 
		FROM groups ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []Group
	for rows.Next() {
		var group Group
		var projectURL sql.NullString
		var description sql.NullString

		err := rows.Scan(&group.ID, &group.Name, &description, &projectURL, &group.IsActive, &group.CreatedAt)
		if err != nil {
			return nil, err
		}

		if description.Valid {
			group.Description = description.String
		}
		if projectURL.Valid {
			group.ProjectURL = projectURL.String
		}

		// Get student count for this group
		var count int
		db.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, group.ID).Scan(&count)
		group.Students = make([]int, count) // Just track the count for list view

		groups = append(groups, group)
	}

	return groups, nil
}

// getGroupNameByID retrieves a group's name by its ID
func getGroupNameByID(id int) string {
	var name string
	err := db.QueryRow(`SELECT name FROM groups WHERE id = ?`, id).Scan(&name)
	if err != nil {
		return ""
	}
	return name
}

// getTaskStatsForStudent calculates task statistics for a student
func getTaskStatsForStudent(studentID int) (TaskStats, error) {
	stats := TaskStats{}

	// Get counts for each status
	rows, err := db.Query(`
		SELECT status, COUNT(*) as count
		FROM tasks
		WHERE owner_id = ?
		GROUP BY status
	`, studentID)
	if err != nil {
		return stats, err
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return stats, err
		}

		stats.Total += count
		switch status {
		case "pending":
			stats.Pending = count
		case "in_progress":
			stats.InProgress = count
		case "blocked":
			stats.Blocked = count
		case "completed":
			stats.Completed = count
		}
	}

	return stats, nil
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
			       in_progress_date, done_date, group_id, COALESCE(comments, ''), COALESCE(display_order, 0)
			FROM tasks
			WHERE created_by = ? AND owner_id = ?
			ORDER BY status, display_order, created_at DESC
		`, user.ID, user.ID)
	} else {
		// Students see their own tasks
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, ''),
			       in_progress_date, done_date, group_id, COALESCE(comments, ''), COALESCE(display_order, 0)
			FROM tasks
			WHERE owner_id = ?
			ORDER BY status, display_order, created_at DESC
		`, user.ID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var task Task
		var inProgressDate, doneDate sql.NullTime
		var groupID sql.NullInt64
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt,
			&task.TaskType, &task.ResourceURL, &task.CohortID, &inProgressDate, &doneDate, &groupID, &task.Comments, &task.DisplayOrder)
		if err != nil {
			return nil, err
		}
		if inProgressDate.Valid {
			task.InProgressDate = &inProgressDate.Time
		}
		if doneDate.Valid {
			task.DoneDate = &doneDate.Time
		}
		if groupID.Valid {
			gid := int(groupID.Int64)
			task.GroupID = &gid
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}
