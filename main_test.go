package main

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// createTestDB creates an in-memory test database with all necessary tables
func createTestDB(t *testing.T) *sql.DB {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Enable foreign keys for SQLite
	_, err = testDB.Exec("PRAGMA foreign_keys = ON")
	if err != nil {
		t.Fatalf("Failed to enable foreign keys: %v", err)
	}

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			task_type TEXT,
			resource_url TEXT,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			group_id INTEGER,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id),
			FOREIGN KEY (group_id) REFERENCES groups(id)
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Create groups table
	_, err = testDB.Exec(`
		CREATE TABLE groups (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			project_url TEXT,
			is_active BOOLEAN DEFAULT 1,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create groups table: %v", err)
	}

	// Create group_members table
	_, err = testDB.Exec(`
		CREATE TABLE group_members (
			group_id INTEGER NOT NULL,
			student_id INTEGER NOT NULL,
			PRIMARY KEY (group_id, student_id),
			FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
			FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create group_members table: %v", err)
	}

	return testDB
}

func TestDatabaseInitialization(t *testing.T) {
	// Create an in-memory database for testing
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor'))
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
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
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Verify tables exist
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users', 'tasks')").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query tables: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 tables, got %d", count)
	}
}

func TestUserCreation(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor'))
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert test users
	tests := []struct {
		username string
		password string
		role     string
	}{
		{"student1", "password123", "student"},
		{"instructor1", "password123", "instructor"},
	}

	for _, tt := range tests {
		_, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			tt.username, tt.password, tt.role)
		if err != nil {
			t.Errorf("Failed to insert user %s: %v", tt.username, err)
		}
	}

	// Verify users were created
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count users: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 users, got %d", count)
	}
}

func TestTaskCreation(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tables
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test user
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"student1", "password123", "student")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Insert test task
	_, err = testDB.Exec(`
		INSERT INTO tasks (title, description, due_date, owner_id, created_by)
		VALUES (?, ?, ?, ?, ?)
	`, "Test Task", "Test Description", "2025-10-20", 1, 1)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Verify task was created
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = 1").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 task, got %d", count)
	}

	// Verify task details
	var title, description string
	err = testDB.QueryRow("SELECT title, description FROM tasks WHERE id = 1").Scan(&title, &description)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	if title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", title)
	}

	if description != "Test Description" {
		t.Errorf("Expected description 'Test Description', got '%s'", description)
	}
}

func TestTaskStatusUpdate(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tables
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test task
	_, err = testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
		"Test Task", 1, 1)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Update task status
	_, err = testDB.Exec("UPDATE tasks SET status = ? WHERE id = ?", "completed", 1)
	if err != nil {
		t.Fatalf("Failed to update task: %v", err)
	}

	// Verify status was updated
	var status string
	err = testDB.QueryRow("SELECT status FROM tasks WHERE id = 1").Scan(&status)
	if err != nil {
		t.Fatalf("Failed to query task status: %v", err)
	}

	if status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", status)
	}
}

func TestCohortAssignment(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table with cohort support
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert test users with cohorts
	tests := []struct {
		username   string
		cohortName string
	}{
		{"student1", "Java11.2"},
		{"student2", "Java11.2"},
		{"student3", "Data6.2"},
	}

	for _, tt := range tests {
		_, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
			tt.username, "password123", "student", tt.cohortName)
		if err != nil {
			t.Errorf("Failed to insert user %s: %v", tt.username, err)
		}
	}

	// Verify cohorts were assigned
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE cohort_name = 'Java11.2'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count users in cohort: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected 2 users in Java11.2 cohort, got %d", count)
	}

	// Verify cohort retrieval
	var cohortName string
	err = testDB.QueryRow("SELECT cohort_name FROM users WHERE username = 'student1'").Scan(&cohortName)
	if err != nil {
		t.Fatalf("Failed to query cohort: %v", err)
	}

	if cohortName != "Java11.2" {
		t.Errorf("Expected cohort 'Java11.2', got '%s'", cohortName)
	}
}

func TestAccountSuspension(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table with suspension support
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert test user
	_, err = testDB.Exec("INSERT INTO users (username, password, role, is_suspended) VALUES (?, ?, ?, ?)",
		"student1", "password123", "student", false)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Suspend account
	_, err = testDB.Exec("UPDATE users SET is_suspended = ? WHERE username = ?", true, "student1")
	if err != nil {
		t.Fatalf("Failed to suspend account: %v", err)
	}

	// Verify suspension
	var isSuspended bool
	err = testDB.QueryRow("SELECT is_suspended FROM users WHERE username = 'student1'").Scan(&isSuspended)
	if err != nil {
		t.Fatalf("Failed to query suspension status: %v", err)
	}

	if !isSuspended {
		t.Errorf("Expected account to be suspended")
	}

	// Unsuspend account
	_, err = testDB.Exec("UPDATE users SET is_suspended = ? WHERE username = ?", false, "student1")
	if err != nil {
		t.Fatalf("Failed to unsuspend account: %v", err)
	}

	// Verify unsuspension
	err = testDB.QueryRow("SELECT is_suspended FROM users WHERE username = 'student1'").Scan(&isSuspended)
	if err != nil {
		t.Fatalf("Failed to query suspension status: %v", err)
	}

	if isSuspended {
		t.Errorf("Expected account to be active")
	}
}

func TestTaskTypes(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tasks table with task types
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			task_type TEXT,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Test different task types
	taskTypes := []string{"group-programming-project", "individual-programming-lab", "reading", "document", "resume", "prodev", "demo-presentation"}

	for i, taskType := range taskTypes {
		_, err := testDB.Exec("INSERT INTO tasks (title, task_type, owner_id, created_by) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("Task %d", i+1), taskType, 1, 1)
		if err != nil {
			t.Errorf("Failed to insert task with type %s: %v", taskType, err)
		}
	}

	// Verify all task types were created
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if count != len(taskTypes) {
		t.Errorf("Expected %d tasks, got %d", len(taskTypes), count)
	}

	// Verify specific task type
	var taskType string
	err = testDB.QueryRow("SELECT task_type FROM tasks WHERE title = 'Task 1'").Scan(&taskType)
	if err != nil {
		t.Fatalf("Failed to query task type: %v", err)
	}

	if taskType != "group-programming-project" {
		t.Errorf("Expected task type 'group-programming-project', got '%s'", taskType)
	}
}

func TestTaskResourceURL(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tasks table with resource URL
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			resource_url TEXT,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert task with resource URL
	testURL := "https://example.com/resource"
	_, err = testDB.Exec("INSERT INTO tasks (title, resource_url, owner_id, created_by) VALUES (?, ?, ?, ?)",
		"Test Task", testURL, 1, 1)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Verify resource URL
	var resourceURL string
	err = testDB.QueryRow("SELECT resource_url FROM tasks WHERE id = 1").Scan(&resourceURL)
	if err != nil {
		t.Fatalf("Failed to query resource URL: %v", err)
	}

	if resourceURL != testURL {
		t.Errorf("Expected resource URL '%s', got '%s'", testURL, resourceURL)
	}
}

func TestCohortTaskAssignment(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			cohort_id TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Verify tasks table exists
	var tableExists int
	err = testDB.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='tasks'").Scan(&tableExists)
	if err != nil || tableExists == 0 {
		t.Fatalf("Tasks table not found after creation: %v", err)
	}

	// Insert test users in same cohort
	cohortName := "Java11.2"
	for i := 1; i <= 3; i++ {
		_, err := testDB.Exec("INSERT INTO users (username, role, cohort_name) VALUES (?, ?, ?)",
			fmt.Sprintf("student%d", i), "student", cohortName)
		if err != nil {
			t.Fatalf("Failed to insert user: %v", err)
		}
	}

	// Simulate assigning task to cohort (creating task for each student in cohort)
	rows, err := testDB.Query("SELECT id FROM users WHERE cohort_name = ?", cohortName)
	if err != nil {
		t.Fatalf("Failed to query cohort students: %v", err)
	}

	// Collect student IDs first
	var studentIDs []int
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			t.Errorf("Failed to scan student ID: %v", err)
			continue
		}
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	// Now create tasks for each student
	taskTitle := "Cohort Assignment"
	instructorID := 1
	tasksCreated := 0

	for _, studentID := range studentIDs {
		_, err := testDB.Exec("INSERT INTO tasks (title, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?)",
			taskTitle, studentID, instructorID, cohortName)
		if err != nil {
			t.Errorf("Failed to create task for student %d: %v", studentID, err)
		}
		tasksCreated++
	}

	if tasksCreated != 3 {
		t.Errorf("Expected 3 tasks created, got %d", tasksCreated)
	}

	// Verify tasks were assigned to cohort
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE cohort_id = ?", cohortName).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count cohort tasks: %v", err)
	}

	if count != 3 {
		t.Errorf("Expected 3 tasks for cohort, got %d", count)
	}
}

func TestTaskCompletionStats(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	studentID := 1
	instructorID := 2

	// Insert tasks with different statuses
	statuses := []string{"completed", "completed", "in_progress", "pending", "pending"}
	for i, status := range statuses {
		_, err := testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("Task %d", i+1), status, studentID, instructorID)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}
	}

	// Calculate completion stats
	var totalTasks, completedTasks int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = ?", studentID).Scan(&totalTasks)
	if err != nil {
		t.Fatalf("Failed to count total tasks: %v", err)
	}

	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = ? AND status = 'completed'", studentID).Scan(&completedTasks)
	if err != nil {
		t.Fatalf("Failed to count completed tasks: %v", err)
	}

	if totalTasks != 5 {
		t.Errorf("Expected 5 total tasks, got %d", totalTasks)
	}

	if completedTasks != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", completedTasks)
	}

	completionPercentage := float64(completedTasks) / float64(totalTasks) * 100
	expectedPercentage := 40.0

	if completionPercentage != expectedPercentage {
		t.Errorf("Expected completion percentage %.1f%%, got %.1f%%", expectedPercentage, completionPercentage)
	}
}

func TestUpdateStudentCohort(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert student without cohort
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)",
		"student1", "student")
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}

	// Update student cohort
	newCohort := "Java11.2"
	_, err = testDB.Exec("UPDATE users SET cohort_name = ? WHERE username = ?", newCohort, "student1")
	if err != nil {
		t.Fatalf("Failed to update cohort: %v", err)
	}

	// Verify cohort was updated
	var cohortName sql.NullString
	err = testDB.QueryRow("SELECT cohort_name FROM users WHERE username = 'student1'").Scan(&cohortName)
	if err != nil {
		t.Fatalf("Failed to query cohort: %v", err)
	}

	if !cohortName.Valid || cohortName.String != newCohort {
		t.Errorf("Expected cohort '%s', got '%s'", newCohort, cohortName.String)
	}
}

func TestAssignToAllStudentsExcludesInstructors(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test users: 2 instructors and 3 students
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "instructor1", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "instructor2", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor2: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "student1", "student")
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "student2", "student")
	if err != nil {
		t.Fatalf("Failed to insert student2: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "student3", "student")
	if err != nil {
		t.Fatalf("Failed to insert student3: %v", err)
	}

	// Simulate the "assign to all" logic from addTaskHandler
	// This should only assign to students, not instructors
	rows, err := testDB.Query("SELECT id FROM users WHERE role = 'student'")
	if err != nil {
		t.Fatalf("Failed to query students: %v", err)
	}

	var studentIDs []int
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			t.Errorf("Failed to scan student ID: %v", err)
			continue
		}
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	// Create tasks for each student
	instructorID := 1
	taskTitle := "Assignment for All"
	for _, studentID := range studentIDs {
		_, err := testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
			taskTitle, studentID, instructorID)
		if err != nil {
			t.Errorf("Failed to create task for student %d: %v", studentID, err)
		}
	}

	// Verify tasks were only assigned to students
	var totalUsers, totalTasks int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'student'").Scan(&totalUsers)
	if err != nil {
		t.Fatalf("Failed to count students: %v", err)
	}

	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&totalTasks)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if totalUsers != 3 {
		t.Errorf("Expected 3 students, got %d", totalUsers)
	}

	if totalTasks != 3 {
		t.Errorf("Expected 3 tasks (one per student), got %d", totalTasks)
	}

	// Verify no tasks were assigned to instructors
	var instructorTasks int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t 
		JOIN users u ON t.owner_id = u.id 
		WHERE u.role = 'instructor'
	`).Scan(&instructorTasks)
	if err != nil {
		t.Fatalf("Failed to count instructor tasks: %v", err)
	}

	if instructorTasks != 0 {
		t.Errorf("Expected 0 tasks assigned to instructors, got %d", instructorTasks)
	}
}

// TestGroupTaskStatistics tests the calculation of group task completion statistics
func TestGroupTaskStatistics(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			cohort_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and students
	cohortName := "Java11.2"
	_, err = testDB.Exec("INSERT INTO users (id, username, role) VALUES (?, ?, ?)", 1, "instructor1", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	for i := 2; i <= 5; i++ {
		_, err = testDB.Exec("INSERT INTO users (id, username, role, cohort_name) VALUES (?, ?, ?, ?)",
			i, fmt.Sprintf("student%d", i-1), "student", cohortName)
		if err != nil {
			t.Fatalf("Failed to insert student: %v", err)
		}
	}

	// Create a group task assigned to all students in the cohort
	taskTitle := "Group Programming Project"
	instructorID := 1

	// Student 2: in_progress
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "in_progress", 2, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 3: completed
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "completed", 3, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 4: completed
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "completed", 4, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 5: pending
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "pending", 5, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Query group task statistics
	var totalCount, inProgress, completed int
	err = testDB.QueryRow(`
		SELECT 
			COUNT(*) as total_count,
			SUM(CASE WHEN status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
			SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) as completed
		FROM tasks
		WHERE title = ? AND cohort_id = ?
	`, taskTitle, cohortName).Scan(&totalCount, &inProgress, &completed)
	if err != nil {
		t.Fatalf("Failed to query group task statistics: %v", err)
	}

	// Verify statistics
	if totalCount != 4 {
		t.Errorf("Expected 4 total tasks, got %d", totalCount)
	}
	if inProgress != 1 {
		t.Errorf("Expected 1 in-progress task, got %d", inProgress)
	}
	if completed != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", completed)
	}

	// Calculate percentages
	inProgressPercentage := float64(inProgress) / float64(totalCount) * 100
	completedPercentage := float64(completed) / float64(totalCount) * 100

	expectedInProgress := 25.0
	expectedCompleted := 50.0

	if inProgressPercentage != expectedInProgress {
		t.Errorf("Expected %.1f%% in progress, got %.1f%%", expectedInProgress, inProgressPercentage)
	}
	if completedPercentage != expectedCompleted {
		t.Errorf("Expected %.1f%% completed, got %.1f%%", expectedCompleted, completedPercentage)
	}
}

// TestInstructorCanViewStudentKanban tests that instructors can view student kanban boards
func TestInstructorCanViewStudentKanban(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and student
	_, err = testDB.Exec("INSERT INTO users (id, username, role) VALUES (?, ?, ?)", 1, "instructor1", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO users (id, username, role, cohort_name) VALUES (?, ?, ?, ?)",
		2, "student1", "student", "Java11.2")
	if err != nil {
		t.Fatalf("Failed to insert student: %v", err)
	}

	// Create tasks for the student in different statuses
	statuses := []string{"pending", "in_progress", "blocked", "completed"}
	for i, status := range statuses {
		_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
			fmt.Sprintf("Task %d", i+1), status, 2, 1)
		if err != nil {
			t.Fatalf("Failed to insert task: %v", err)
		}
	}

	// Verify tasks can be retrieved for the student
	rows, err := testDB.Query(`
		SELECT title, status FROM tasks WHERE owner_id = ? ORDER BY status
	`, 2)
	if err != nil {
		t.Fatalf("Failed to query student tasks: %v", err)
	}
	defer rows.Close()

	taskCount := 0
	statusCount := make(map[string]int)
	for rows.Next() {
		var title, status string
		rows.Scan(&title, &status)
		taskCount++
		statusCount[status]++
	}

	if taskCount != 4 {
		t.Errorf("Expected 4 tasks for student, got %d", taskCount)
	}

	for _, status := range statuses {
		if statusCount[status] != 1 {
			t.Errorf("Expected 1 task with status '%s', got %d", status, statusCount[status])
		}
	}
}

func TestAssignToCohortExcludesInstructors(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			cohort_id TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test users: 1 instructor in Java11.2 cohort and 2 students in Java11.2 cohort
	cohortName := "Java11.2"
	_, err = testDB.Exec("INSERT INTO users (username, role, cohort_name) VALUES (?, ?, ?)", "instructor1", "instructor", cohortName)
	if err != nil {
		t.Fatalf("Failed to insert instructor1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role, cohort_name) VALUES (?, ?, ?)", "student1", "student", cohortName)
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role, cohort_name) VALUES (?, ?, ?)", "student2", "student", cohortName)
	if err != nil {
		t.Fatalf("Failed to insert student2: %v", err)
	}

	// Simulate the "assign to cohort" logic from addTaskHandler
	// This should only assign to students in the cohort, not instructors
	rows, err := testDB.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ?", cohortName)
	if err != nil {
		t.Fatalf("Failed to query cohort students: %v", err)
	}

	var studentIDs []int
	for rows.Next() {
		var studentID int
		if err := rows.Scan(&studentID); err != nil {
			t.Errorf("Failed to scan student ID: %v", err)
			continue
		}
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	// Create tasks for each student in cohort
	instructorID := 1
	taskTitle := "Cohort Assignment"
	for _, studentID := range studentIDs {
		_, err := testDB.Exec("INSERT INTO tasks (title, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?)",
			taskTitle, studentID, instructorID, cohortName)
		if err != nil {
			t.Errorf("Failed to create task for student %d: %v", studentID, err)
		}
	}

	// Verify tasks were only assigned to students in the cohort
	var totalStudents, totalTasks int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'student' AND cohort_name = ?", cohortName).Scan(&totalStudents)
	if err != nil {
		t.Fatalf("Failed to count cohort students: %v", err)
	}

	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE cohort_id = ?", cohortName).Scan(&totalTasks)
	if err != nil {
		t.Fatalf("Failed to count cohort tasks: %v", err)
	}

	if totalStudents != 2 {
		t.Errorf("Expected 2 students in cohort, got %d", totalStudents)
	}

	if totalTasks != 2 {
		t.Errorf("Expected 2 tasks (one per student in cohort), got %d", totalTasks)
	}

	// Verify no tasks were assigned to instructors
	var instructorTasks int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t 
		JOIN users u ON t.owner_id = u.id 
		WHERE u.role = 'instructor' AND t.cohort_id = ?
	`, cohortName).Scan(&instructorTasks)
	if err != nil {
		t.Fatalf("Failed to count instructor tasks in cohort: %v", err)
	}

	if instructorTasks != 0 {
		t.Errorf("Expected 0 tasks assigned to instructors in cohort, got %d", instructorTasks)
	}
}

func TestAssignToSpecificStudentDoesNotAllowInstructor(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test users
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "instructor1", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "instructor2", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor2: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, role) VALUES (?, ?)", "student1", "student")
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}

	// Test that we can successfully assign to a student
	instructorID := 1
	studentID := 3
	taskTitle := "Task for student"

	// First, validate the user is a student
	var role string
	err = testDB.QueryRow("SELECT role FROM users WHERE id = ?", studentID).Scan(&role)
	if err != nil || role != "student" {
		t.Fatalf("Target should be a student")
	}

	// Create task for student - this should succeed
	_, err = testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
		taskTitle, studentID, instructorID)
	if err != nil {
		t.Fatalf("Failed to create task for student: %v", err)
	}

	// Test that we cannot assign to an instructor
	targetInstructorID := 2

	// Validate the target is an instructor (not a student)
	err = testDB.QueryRow("SELECT role FROM users WHERE id = ?", targetInstructorID).Scan(&role)
	if err != nil || role != "student" {
		// This is expected - the target is an instructor, not a student
		// So we should NOT create a task for them
		t.Log("Correctly detected that target is not a student")
	} else {
		t.Fatal("Target should be an instructor, not a student")
	}

	// Verify that only one task was created (for the student, not the instructor)
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if taskCount != 1 {
		t.Errorf("Expected 1 task (for student only), got %d", taskCount)
	}

	// Verify no tasks were assigned to instructors
	var tasksAssignedToInstructor int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks t 
		JOIN users u ON t.owner_id = u.id 
		WHERE u.role = 'instructor'
	`).Scan(&tasksAssignedToInstructor)
	if err != nil {
		t.Fatalf("Failed to count instructor tasks: %v", err)
	}

	if tasksAssignedToInstructor != 0 {
		t.Errorf("Expected 0 tasks assigned to instructors, got %d", tasksAssignedToInstructor)
	}
}

func TestAddNewStudent(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Add a new student
	username := "newstudent"
	password := "password123"
	cohortName := "Java11.2"

	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'student', ?, 0)
	`, username, password, cohortName)
	if err != nil {
		t.Fatalf("Failed to add new student: %v", err)
	}

	// Verify student was added
	var retrievedUsername, retrievedPassword, retrievedRole, retrievedCohort string
	var isSuspended bool
	err = testDB.QueryRow(`
		SELECT username, password, role, cohort_name, is_suspended 
		FROM users WHERE username = ?
	`, username).Scan(&retrievedUsername, &retrievedPassword, &retrievedRole, &retrievedCohort, &isSuspended)
	if err != nil {
		t.Fatalf("Failed to query new student: %v", err)
	}

	if retrievedUsername != username {
		t.Errorf("Expected username '%s', got '%s'", username, retrievedUsername)
	}
	if retrievedPassword != password {
		t.Errorf("Expected password '%s', got '%s'", password, retrievedPassword)
	}
	if retrievedRole != "student" {
		t.Errorf("Expected role 'student', got '%s'", retrievedRole)
	}
	if retrievedCohort != cohortName {
		t.Errorf("Expected cohort '%s', got '%s'", cohortName, retrievedCohort)
	}
	if isSuspended {
		t.Errorf("Expected new student to not be suspended")
	}
}

func TestEditStudentUsername(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert a student
	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role, cohort_name)
		VALUES (?, ?, 'student', ?)
	`, "oldusername", "password123", "Java11.2")
	if err != nil {
		t.Fatalf("Failed to insert student: %v", err)
	}

	// Update the student's username
	newUsername := "newusername"
	_, err = testDB.Exec(`
		UPDATE users SET username = ? WHERE username = ? AND role = 'student'
	`, newUsername, "oldusername")
	if err != nil {
		t.Fatalf("Failed to update username: %v", err)
	}

	// Verify username was updated
	var username string
	err = testDB.QueryRow("SELECT username FROM users WHERE id = 1").Scan(&username)
	if err != nil {
		t.Fatalf("Failed to query username: %v", err)
	}

	if username != newUsername {
		t.Errorf("Expected username '%s', got '%s'", newUsername, username)
	}
}

func TestEditStudentPassword(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert a student
	oldPassword := "oldpassword"
	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES (?, ?, 'student')
	`, "student1", oldPassword)
	if err != nil {
		t.Fatalf("Failed to insert student: %v", err)
	}

	// Update the student's password
	newPassword := "newpassword123"
	_, err = testDB.Exec(`
		UPDATE users SET password = ? WHERE username = ? AND role = 'student'
	`, newPassword, "student1")
	if err != nil {
		t.Fatalf("Failed to update password: %v", err)
	}

	// Verify password was updated
	var password string
	err = testDB.QueryRow("SELECT password FROM users WHERE username = 'student1'").Scan(&password)
	if err != nil {
		t.Fatalf("Failed to query password: %v", err)
	}

	if password != newPassword {
		t.Errorf("Expected password '%s', got '%s'", newPassword, password)
	}
}

func TestEditStudentAllFields(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert a student
	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'student', ?, 0)
	`, "oldusername", "oldpassword", "Java11.2")
	if err != nil {
		t.Fatalf("Failed to insert student: %v", err)
	}

	// Update all fields
	newUsername := "newusername"
	newPassword := "newpassword"
	newCohort := "Data6.2"
	newSuspended := true

	_, err = testDB.Exec(`
		UPDATE users 
		SET username = ?, password = ?, cohort_name = ?, is_suspended = ?
		WHERE id = ? AND role = 'student'
	`, newUsername, newPassword, newCohort, newSuspended, 1)
	if err != nil {
		t.Fatalf("Failed to update student: %v", err)
	}

	// Verify all fields were updated
	var username, password, cohort string
	var suspended bool
	err = testDB.QueryRow(`
		SELECT username, password, cohort_name, is_suspended FROM users WHERE id = 1
	`).Scan(&username, &password, &cohort, &suspended)
	if err != nil {
		t.Fatalf("Failed to query student: %v", err)
	}

	if username != newUsername {
		t.Errorf("Expected username '%s', got '%s'", newUsername, username)
	}
	if password != newPassword {
		t.Errorf("Expected password '%s', got '%s'", newPassword, password)
	}
	if cohort != newCohort {
		t.Errorf("Expected cohort '%s', got '%s'", newCohort, cohort)
	}
	if suspended != newSuspended {
		t.Errorf("Expected suspended %v, got %v", newSuspended, suspended)
	}
}

// TestTaskTypeMigration verifies that the migration from old task type constraints to new ones works correctly
func TestTaskTypeMigration(t *testing.T) {
	// Create a database with the old task type constraint
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT UNIQUE NOT NULL,
password TEXT NOT NULL,
role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
cohort_name TEXT,
is_suspended BOOLEAN DEFAULT 0
)
`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table with OLD constraint (simulating an existing database)
	_, err = testDB.Exec(`
CREATE TABLE tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
description TEXT,
due_date TEXT,
status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
owner_id INTEGER NOT NULL,
created_by INTEGER NOT NULL,
assigned_to TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
task_type TEXT CHECK(task_type IN ('reading', 'coding', 'submission', 'document-creation', '')),
resource_url TEXT,
cohort_id TEXT,
in_progress_date DATETIME,
done_date DATETIME,
FOREIGN KEY (owner_id) REFERENCES users(id),
FOREIGN KEY (created_by) REFERENCES users(id)
)
`)
	if err != nil {
		t.Fatalf("Failed to create tasks table with old schema: %v", err)
	}

	// Insert test user and task with old task type
	_, err = testDB.Exec(`INSERT INTO users (id, username, password, role) VALUES (1, 'testuser', 'pass', 'student')`)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	_, err = testDB.Exec(`INSERT INTO tasks (title, owner_id, created_by, task_type) VALUES ('Old Task', 1, 1, 'reading')`)
	if err != nil {
		t.Fatalf("Failed to insert test task: %v", err)
	}

	// Verify that new task types are rejected before migration
	_, err = testDB.Exec(`INSERT INTO tasks (title, owner_id, created_by, task_type) VALUES ('New Task', 1, 1, 'group-programming-project')`)
	if err == nil {
		t.Fatal("Expected new task type to be rejected before migration, but it was accepted")
	}

	// Run the migration
	db = testDB
	if err := migrateTaskTypeConstraint(); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify old data is preserved
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = 'Old Task'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query old task: %v", err)
	}
	if count != 1 {
		t.Fatal("Old task data was lost during migration!")
	}

	// Test that we can now insert all new task types
	newTaskTypes := []string{
		"group-programming-project",
		"individual-programming-lab",
		"reading",
		"document",
		"resume",
		"prodev",
		"demo-presentation",
	}

	for _, taskType := range newTaskTypes {
		_, err = testDB.Exec(`INSERT INTO tasks (title, owner_id, created_by, task_type) VALUES (?, 1, 1, ?)`, "Task: "+taskType, taskType)
		if err != nil {
			t.Errorf("Failed to insert new task type '%s': %v", taskType, err)
		}
	}

	// Test that old task types are rejected after migration
	oldTaskTypes := []string{"coding", "submission", "document-creation"}
	for _, taskType := range oldTaskTypes {
		_, err = testDB.Exec(`INSERT INTO tasks (title, owner_id, created_by, task_type) VALUES (?, 1, 1, ?)`, "Bad Task: "+taskType, taskType)
		if err == nil {
			t.Errorf("Old task type '%s' should have been rejected after migration", taskType)
		}
	}
}

func TestAddNewInstructor(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Add a new instructor
	username := "newinstructor"
	password := "password123"

	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'instructor', NULL, 0)
	`, username, password)
	if err != nil {
		t.Fatalf("Failed to add new instructor: %v", err)
	}

	// Verify instructor was added
	var retrievedUsername, retrievedPassword, retrievedRole string
	var cohortName sql.NullString
	var isSuspended bool
	err = testDB.QueryRow(`
		SELECT username, password, role, cohort_name, is_suspended 
		FROM users WHERE username = ?
	`, username).Scan(&retrievedUsername, &retrievedPassword, &retrievedRole, &cohortName, &isSuspended)
	if err != nil {
		t.Fatalf("Failed to query new instructor: %v", err)
	}

	if retrievedUsername != username {
		t.Errorf("Expected username '%s', got '%s'", username, retrievedUsername)
	}
	if retrievedPassword != password {
		t.Errorf("Expected password '%s', got '%s'", password, retrievedPassword)
	}
	if retrievedRole != "instructor" {
		t.Errorf("Expected role 'instructor', got '%s'", retrievedRole)
	}
	if cohortName.Valid {
		t.Errorf("Expected cohort to be NULL for instructor, got '%s'", cohortName.String)
	}
	if isSuspended {
		t.Errorf("Expected new instructor to not be suspended")
	}
}

func TestInstructorCanDeleteInstructorCreatedTasks(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test users: 2 instructors and 1 student
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "instructor2", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor2: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "student1", "pass", "student")
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}

	// Instructor1 creates a task for themselves
	_, err = testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
		"Instructor Task", 1, 1)
	if err != nil {
		t.Fatalf("Failed to create instructor task: %v", err)
	}

	// Verify task was created
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("Expected 1 task, got %d", taskCount)
	}

	// Simulate instructor2 checking if they can delete the task
	// Get the task and the creator's role
	var taskID, ownerID, createdBy int
	var creatorRole string
	err = testDB.QueryRow(`
		SELECT t.id, t.owner_id, t.created_by, u.role 
		FROM tasks t 
		JOIN users u ON t.created_by = u.id 
		WHERE t.id = 1
	`).Scan(&taskID, &ownerID, &createdBy, &creatorRole)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	// Instructor2 (user ID 2) should be able to delete this task because:
	// - The task was created by instructor1 (creatorRole == "instructor")
	// - Instructor2 is also an instructor
	instructor2Role := "instructor"

	canDelete := false
	if instructor2Role == "instructor" && creatorRole == "instructor" {
		canDelete = true
	}

	if !canDelete {
		t.Errorf("Instructor2 should be able to delete tasks created by other instructors")
	}

	// Now actually delete the task
	_, err = testDB.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		t.Fatalf("Failed to delete task: %v", err)
	}

	// Verify task was deleted
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks after deletion: %v", err)
	}
	if taskCount != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", taskCount)
	}
}

func TestInstructorCannotDeleteStudentCreatedTasks(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert test users: 1 instructor and 1 student
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)", "student1", "pass", "student")
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}

	// Student creates a task for themselves
	_, err = testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
		"Student Task", 2, 2)
	if err != nil {
		t.Fatalf("Failed to create student task: %v", err)
	}

	// Verify task was created
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Fatalf("Expected 1 task, got %d", taskCount)
	}

	// Simulate instructor checking if they can delete the task
	var taskID, ownerID, createdBy int
	var creatorRole string
	err = testDB.QueryRow(`
		SELECT t.id, t.owner_id, t.created_by, u.role 
		FROM tasks t 
		JOIN users u ON t.created_by = u.id 
		WHERE t.id = 1
	`).Scan(&taskID, &ownerID, &createdBy, &creatorRole)
	if err != nil {
		t.Fatalf("Failed to query task: %v", err)
	}

	// Instructor (user ID 1) should NOT be able to delete this task because:
	// - The task was created by student1 (creatorRole == "student")
	// - Instructor does not own this task (ownerID != instructorID)
	instructorID := 1
	instructorRole := "instructor"

	canDelete := true
	if instructorRole == "instructor" && creatorRole != "instructor" && ownerID != instructorID {
		canDelete = false
	}

	if canDelete {
		t.Errorf("Instructor should NOT be able to delete tasks created by students (unless they own it)")
	}
}

func TestImportStudents(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table with cohort_name column
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Test data simulating CSV import
	tests := []struct {
		username   string
		password   string
		cohortName string
	}{
		{"AndrewYounger", "zipcode0", "Java1.0"},
		{"BillyBobThornton", "zipcode1", "Data1.0"},
		{"TestStudent", "password123", "Java11.2"},
	}

	// Simulate the import process
	for _, tt := range tests {
		_, err := testDB.Exec(`
			INSERT INTO users (username, password, role, cohort_name, is_suspended)
			VALUES (?, ?, 'student', ?, 0)
		`, tt.username, tt.password, tt.cohortName)
		if err != nil {
			t.Errorf("Failed to import student %s: %v", tt.username, err)
		}
	}

	// Verify students were imported
	var count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'student'").Scan(&count)
	if err != nil {
		t.Fatalf("Failed to count students: %v", err)
	}

	if count != len(tests) {
		t.Errorf("Expected %d students, got %d", len(tests), count)
	}

	// Verify specific student details
	var username, password, cohortName string
	err = testDB.QueryRow(`
		SELECT username, password, cohort_name 
		FROM users 
		WHERE username = 'AndrewYounger'
	`).Scan(&username, &password, &cohortName)
	if err != nil {
		t.Fatalf("Failed to query imported student: %v", err)
	}

	if username != "AndrewYounger" || password != "zipcode0" || cohortName != "Java1.0" {
		t.Errorf("Student data mismatch: got username=%s, password=%s, cohort=%s",
			username, password, cohortName)
	}

	// Test duplicate username constraint
	_, err = testDB.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'student', ?, 0)
	`, "AndrewYounger", "newpassword", "Data2.0")

	if err == nil {
		t.Errorf("Expected error when inserting duplicate username, got nil")
	}
}

// TestTaskDetailTracking tests the ability to view a specific task with all assigned students
func TestTaskDetailTracking(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table with tracking fields
	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			cohort_id TEXT,
			in_progress_date DATETIME,
			done_date DATETIME,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and students
	cohortName := "Java11.2"
	_, err = testDB.Exec("INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	for i := 2; i <= 5; i++ {
		_, err = testDB.Exec("INSERT INTO users (id, username, password, role, cohort_name) VALUES (?, ?, ?, ?, ?)",
			i, fmt.Sprintf("student%d", i-1), "pass", "student", cohortName)
		if err != nil {
			t.Fatalf("Failed to insert student: %v", err)
		}
	}

	// Create a task assigned to all students in the cohort
	taskTitle := "Group Programming Project"
	instructorID := 1

	// Student 2: pending
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "pending", 2, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 3: in_progress (with in_progress_date)
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id, in_progress_date) VALUES (?, ?, ?, ?, ?, ?)",
		taskTitle, "in_progress", 3, instructorID, cohortName, time.Now())
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 4: completed (with both dates)
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id, in_progress_date, done_date) VALUES (?, ?, ?, ?, ?, ?, ?)",
		taskTitle, "completed", 4, instructorID, cohortName, time.Now().Add(-48*time.Hour), time.Now())
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Student 5: blocked
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by, cohort_id) VALUES (?, ?, ?, ?, ?)",
		taskTitle, "blocked", 5, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Query task details with all assigned students
	rows, err := testDB.Query(`
		SELECT t.id, t.owner_id, u.username, COALESCE(u.cohort_name, ''),
		       t.status, t.created_at, t.in_progress_date, t.done_date
		FROM tasks t
		JOIN users u ON t.owner_id = u.id
		WHERE t.title = ? AND t.created_by = ? AND COALESCE(t.cohort_id, '') = ?
		ORDER BY u.username
	`, taskTitle, instructorID, cohortName)
	if err != nil {
		t.Fatalf("Failed to query task details: %v", err)
	}
	defer rows.Close()

	statusCounts := map[string]int{
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
		"blocked":     0,
	}
	totalAssignments := 0

	for rows.Next() {
		var taskID, studentID int
		var studentName, cohort, status string
		var createdAt time.Time
		var inProgressDate, doneDate sql.NullTime

		err := rows.Scan(&taskID, &studentID, &studentName, &cohort, &status, &createdAt, &inProgressDate, &doneDate)
		if err != nil {
			t.Errorf("Failed to scan row: %v", err)
			continue
		}

		statusCounts[status]++
		totalAssignments++

		// Verify tracking fields for specific statuses
		if status == "in_progress" && !inProgressDate.Valid {
			t.Errorf("Expected in_progress_date to be set for in_progress task")
		}
		if status == "completed" {
			if !doneDate.Valid {
				t.Errorf("Expected done_date to be set for completed task")
			}
		}
	}

	// Verify counts
	if totalAssignments != 4 {
		t.Errorf("Expected 4 total assignments, got %d", totalAssignments)
	}
	if statusCounts["pending"] != 1 {
		t.Errorf("Expected 1 pending task, got %d", statusCounts["pending"])
	}
	if statusCounts["in_progress"] != 1 {
		t.Errorf("Expected 1 in_progress task, got %d", statusCounts["in_progress"])
	}
	if statusCounts["completed"] != 1 {
		t.Errorf("Expected 1 completed task, got %d", statusCounts["completed"])
	}
	if statusCounts["blocked"] != 1 {
		t.Errorf("Expected 1 blocked task, got %d", statusCounts["blocked"])
	}
}

// TestImportTasks tests the task import functionality
func TestImportTasks(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT UNIQUE NOT NULL,
password TEXT NOT NULL,
role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
cohort_name TEXT,
is_suspended BOOLEAN DEFAULT 0
)
`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
CREATE TABLE tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
description TEXT,
due_date TEXT,
status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
owner_id INTEGER NOT NULL,
created_by INTEGER NOT NULL,
assigned_to TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', '')),
resource_url TEXT,
cohort_id TEXT,
in_progress_date DATETIME,
done_date DATETIME,
FOREIGN KEY (owner_id) REFERENCES users(id),
FOREIGN KEY (created_by) REFERENCES users(id)
)
`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and students
	_, err = testDB.Exec("INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	for i := 2; i <= 4; i++ {
		_, err = testDB.Exec("INSERT INTO users (id, username, password, role, cohort_name) VALUES (?, ?, ?, ?, ?)",
			i, fmt.Sprintf("student%d", i-1), "pass", "student", "Java11.2")
		if err != nil {
			t.Fatalf("Failed to insert student: %v", err)
		}
	}

	// Test data simulating CSV import (5 fields: Title, Description, DueDate, TaskType, ResourceURL)
	tests := []struct {
		title       string
		description string
		dueDate     string
		taskType    string
		resourceURL string
	}{
		{"Complete Lab 1", "Finish the Java basics lab", "2025-11-15", "individual-programming-lab", "https://example.com/lab1"},
		{"Read Chapter 5", "Read about design patterns", "2025-11-10", "reading", "https://example.com/chapter5"},
		{"Group Project", "Build a REST API", "2025-11-20", "group-programming-project", "https://github.com/example/project"},
	}

	instructorID := 1
	cohortName := "Java11.2"

	// Simulate the import process - get students in cohort
	rows, err := testDB.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ?", cohortName)
	if err != nil {
		t.Fatalf("Failed to query cohort students: %v", err)
	}

	var studentIDs []int
	for rows.Next() {
		var studentID int
		rows.Scan(&studentID)
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	if len(studentIDs) != 3 {
		t.Fatalf("Expected 3 students in cohort, got %d", len(studentIDs))
	}

	// Import tasks
	for _, tt := range tests {
		for _, studentID := range studentIDs {
			_, err := testDB.Exec(`
INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, resource_url, cohort_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, tt.title, tt.description, tt.dueDate, studentID, instructorID, tt.taskType, tt.resourceURL, cohortName)
			if err != nil {
				t.Errorf("Failed to import task '%s' for student %d: %v", tt.title, studentID, err)
			}
		}
	}

	// Verify tasks were imported
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	expectedCount := len(tests) * len(studentIDs) // 3 tasks * 3 students = 9
	if taskCount != expectedCount {
		t.Errorf("Expected %d tasks, got %d", expectedCount, taskCount)
	}

	// Verify task details for a specific task
	var title, description, dueDate, taskType, resourceURL string
	err = testDB.QueryRow(`
SELECT title, description, due_date, task_type, resource_url 
FROM tasks 
WHERE title = 'Complete Lab 1' 
LIMIT 1
`).Scan(&title, &description, &dueDate, &taskType, &resourceURL)
	if err != nil {
		t.Fatalf("Failed to query imported task: %v", err)
	}

	if title != "Complete Lab 1" {
		t.Errorf("Expected title 'Complete Lab 1', got '%s'", title)
	}
	if description != "Finish the Java basics lab" {
		t.Errorf("Expected description 'Finish the Java basics lab', got '%s'", description)
	}
	if dueDate != "2025-11-15" {
		t.Errorf("Expected due date '2025-11-15', got '%s'", dueDate)
	}
	if taskType != "individual-programming-lab" {
		t.Errorf("Expected task type 'individual-programming-lab', got '%s'", taskType)
	}
	if resourceURL != "https://example.com/lab1" {
		t.Errorf("Expected resource URL 'https://example.com/lab1', got '%s'", resourceURL)
	}

	// Verify all students got all tasks
	for _, studentID := range studentIDs {
		var studentTaskCount int
		err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = ?", studentID).Scan(&studentTaskCount)
		if err != nil {
			t.Errorf("Failed to count tasks for student %d: %v", studentID, err)
		}
		if studentTaskCount != len(tests) {
			t.Errorf("Expected %d tasks for student %d, got %d", len(tests), studentID, studentTaskCount)
		}
	}

	// Verify cohort_id was set correctly
	var cohortTaskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE cohort_id = ?", cohortName).Scan(&cohortTaskCount)
	if err != nil {
		t.Fatalf("Failed to count cohort tasks: %v", err)
	}
	if cohortTaskCount != expectedCount {
		t.Errorf("Expected %d tasks with cohort_id, got %d", expectedCount, cohortTaskCount)
	}
}

// TestImportTasksToAllStudents tests importing tasks to all students
func TestImportTasksToAllStudents(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT UNIQUE NOT NULL,
password TEXT NOT NULL,
role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
cohort_name TEXT,
is_suspended BOOLEAN DEFAULT 0
)
`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
CREATE TABLE tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
description TEXT,
due_date TEXT,
owner_id INTEGER NOT NULL,
created_by INTEGER NOT NULL,
task_type TEXT,
resource_url TEXT,
cohort_id TEXT
)
`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and students in different cohorts
	_, err = testDB.Exec("INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student1", "pass", "student", "Java11.2")
	if err != nil {
		t.Fatalf("Failed to insert student1: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student2", "pass", "student", "Java11.2")
	if err != nil {
		t.Fatalf("Failed to insert student2: %v", err)
	}
	_, err = testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student3", "pass", "student", "Data6.2")
	if err != nil {
		t.Fatalf("Failed to insert student3: %v", err)
	}

	// Get ALL students
	rows, err := testDB.Query("SELECT id FROM users WHERE role = 'student'")
	if err != nil {
		t.Fatalf("Failed to query students: %v", err)
	}

	var studentIDs []int
	for rows.Next() {
		var studentID int
		rows.Scan(&studentID)
		studentIDs = append(studentIDs, studentID)
	}
	rows.Close()

	if len(studentIDs) != 3 {
		t.Fatalf("Expected 3 students total, got %d", len(studentIDs))
	}

	// Import task to all students
	instructorID := 1
	taskTitle := "Universal Assignment"

	for _, studentID := range studentIDs {
		_, err := testDB.Exec(`
INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, resource_url, cohort_id)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)
`, taskTitle, "This is for everyone", "2025-11-30", studentID, instructorID, "reading", "https://example.com", "")
		if err != nil {
			t.Errorf("Failed to create task for student %d: %v", studentID, err)
		}
	}

	// Verify all students got the task
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ?", taskTitle).Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if taskCount != 3 {
		t.Errorf("Expected 3 tasks (one per student), got %d", taskCount)
	}
}

// TestImportTasksToSpecificStudents tests importing tasks to selected students
func TestImportTasksToSpecificStudents(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT UNIQUE NOT NULL,
password TEXT NOT NULL,
role TEXT NOT NULL,
cohort_name TEXT
)
`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
CREATE TABLE tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
owner_id INTEGER NOT NULL,
created_by INTEGER NOT NULL
)
`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructor and students
	_, err = testDB.Exec("INSERT INTO users (id, username, password, role) VALUES (?, ?, ?, ?)",
		1, "instructor1", "pass", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	for i := 2; i <= 5; i++ {
		_, err = testDB.Exec("INSERT INTO users (id, username, password, role, cohort_name) VALUES (?, ?, ?, ?, ?)",
			i, fmt.Sprintf("student%d", i-1), "pass", "student", "Java11.2")
		if err != nil {
			t.Fatalf("Failed to insert student: %v", err)
		}
	}

	// Select only students 2 and 4 (IDs 2 and 4)
	selectedStudentIDs := []int{2, 4}
	instructorID := 1
	taskTitle := "Selected Students Assignment"

	for _, studentID := range selectedStudentIDs {
		_, err := testDB.Exec("INSERT INTO tasks (title, owner_id, created_by) VALUES (?, ?, ?)",
			taskTitle, studentID, instructorID)
		if err != nil {
			t.Errorf("Failed to create task for student %d: %v", studentID, err)
		}
	}

	// Verify only selected students got the task
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ?", taskTitle).Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if taskCount != 2 {
		t.Errorf("Expected 2 tasks (for selected students only), got %d", taskCount)
	}

	// Verify specific students
	for _, studentID := range selectedStudentIDs {
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = ? AND title = ?",
			studentID, taskTitle).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check task for student %d: %v", studentID, err)
		}
		if count != 1 {
			t.Errorf("Expected 1 task for student %d, got %d", studentID, count)
		}
	}

	// Verify non-selected students did not get the task
	nonSelectedStudentIDs := []int{3, 5}
	for _, studentID := range nonSelectedStudentIDs {
		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE owner_id = ? AND title = ?",
			studentID, taskTitle).Scan(&count)
		if err != nil {
			t.Errorf("Failed to check task for student %d: %v", studentID, err)
		}
		if count != 0 {
			t.Errorf("Expected 0 tasks for non-selected student %d, got %d", studentID, count)
		}
	}
}

// TestTaskImportValidation tests validation of task import data
func TestTaskImportValidation(t *testing.T) {
	// Test valid task types
	validTaskTypes := []string{
		"group-programming-project",
		"individual-programming-lab",
		"reading",
		"document",
		"resume",
		"prodev",
		"",
	}

	taskTypeMap := map[string]bool{
		"group-programming-project":  true,
		"individual-programming-lab": true,
		"reading":                    true,
		"document":                   true,
		"resume":                     true,
		"prodev":                     true,
		"":                           true,
	}

	for _, taskType := range validTaskTypes {
		if !taskTypeMap[taskType] {
			t.Errorf("Valid task type '%s' not recognized", taskType)
		}
	}

	// Test invalid task types
	invalidTaskTypes := []string{
		"coding",
		"submission",
		"invalid",
		"test",
	}

	for _, taskType := range invalidTaskTypes {
		if taskTypeMap[taskType] {
			t.Errorf("Invalid task type '%s' was accepted", taskType)
		}
	}
}

func TestTaskStatsWithCreatorName(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create users table
	_, err = testDB.Exec(`
CREATE TABLE users (
id INTEGER PRIMARY KEY AUTOINCREMENT,
username TEXT UNIQUE NOT NULL,
role TEXT NOT NULL,
cohort_name TEXT
)
`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Create tasks table
	_, err = testDB.Exec(`
CREATE TABLE tasks (
id INTEGER PRIMARY KEY AUTOINCREMENT,
title TEXT NOT NULL,
description TEXT,
status TEXT DEFAULT 'pending',
owner_id INTEGER NOT NULL,
created_by INTEGER NOT NULL,
cohort_id TEXT,
created_at DATETIME DEFAULT CURRENT_TIMESTAMP
)
`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Insert instructors and students
	_, err = testDB.Exec("INSERT INTO users (id, username, role) VALUES (?, ?, ?)", 1, "instructor1", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO users (id, username, role) VALUES (?, ?, ?)", 2, "instructor2", "instructor")
	if err != nil {
		t.Fatalf("Failed to insert instructor: %v", err)
	}

	for i := 3; i <= 5; i++ {
		_, err = testDB.Exec("INSERT INTO users (id, username, role) VALUES (?, ?, ?)",
			i, fmt.Sprintf("student%d", i-2), "student")
		if err != nil {
			t.Fatalf("Failed to insert student: %v", err)
		}
	}

	// Create tasks assigned by different instructors
	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
		"Task by Instructor 1", "pending", 3, 1)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
		"Task by Instructor 1", "completed", 4, 1)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
		"Task by Instructor 2", "pending", 3, 2)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	_, err = testDB.Exec("INSERT INTO tasks (title, status, owner_id, created_by) VALUES (?, ?, ?, ?)",
		"Task by Instructor 2", "in_progress", 5, 2)
	if err != nil {
		t.Fatalf("Failed to insert task: %v", err)
	}

	// Query group task statistics with creator name
	rows, err := testDB.Query(`
SELECT
t.title,
t.created_by,
creator.username as creator_name,
COUNT(*) as total_count,
SUM(CASE WHEN t.status = 'in_progress' THEN 1 ELSE 0 END) as in_progress,
SUM(CASE WHEN t.status = 'completed' THEN 1 ELSE 0 END) as completed
FROM tasks t
JOIN users u ON t.owner_id = u.id
JOIN users creator ON t.created_by = creator.id
WHERE t.created_by IN (SELECT id FROM users WHERE role = 'instructor')
AND u.role = 'student'
GROUP BY t.title, t.created_by, creator.username
HAVING COUNT(*) > 1
ORDER BY t.title
`)
	if err != nil {
		t.Fatalf("Failed to query task stats: %v", err)
	}
	defer rows.Close()

	type TaskStat struct {
		Title       string
		CreatedBy   int
		CreatorName string
		TotalCount  int
		InProgress  int
		Completed   int
	}

	var taskStats []TaskStat
	for rows.Next() {
		var ts TaskStat
		err := rows.Scan(&ts.Title, &ts.CreatedBy, &ts.CreatorName, &ts.TotalCount, &ts.InProgress, &ts.Completed)
		if err != nil {
			t.Fatalf("Failed to scan task stat: %v", err)
		}
		taskStats = append(taskStats, ts)
	}

	// Verify results
	if len(taskStats) != 2 {
		t.Errorf("Expected 2 group tasks, got %d", len(taskStats))
	}

	// Check first task (instructor1)
	if taskStats[0].Title != "Task by Instructor 1" {
		t.Errorf("Expected first task title 'Task by Instructor 1', got '%s'", taskStats[0].Title)
	}
	if taskStats[0].CreatorName != "instructor1" {
		t.Errorf("Expected creator name 'instructor1', got '%s'", taskStats[0].CreatorName)
	}
	if taskStats[0].TotalCount != 2 {
		t.Errorf("Expected 2 total tasks for instructor1, got %d", taskStats[0].TotalCount)
	}
	if taskStats[0].Completed != 1 {
		t.Errorf("Expected 1 completed task for instructor1, got %d", taskStats[0].Completed)
	}

	// Check second task (instructor2)
	if taskStats[1].Title != "Task by Instructor 2" {
		t.Errorf("Expected second task title 'Task by Instructor 2', got '%s'", taskStats[1].Title)
	}
	if taskStats[1].CreatorName != "instructor2" {
		t.Errorf("Expected creator name 'instructor2', got '%s'", taskStats[1].CreatorName)
	}
	if taskStats[1].TotalCount != 2 {
		t.Errorf("Expected 2 total tasks for instructor2, got %d", taskStats[1].TotalCount)
	}
	if taskStats[1].InProgress != 1 {
		t.Errorf("Expected 1 in-progress task for instructor2, got %d", taskStats[1].InProgress)
	}
}

// TestURLNormalization tests that URLs are properly normalized with protocol
// Note: This tests the URL normalization logic used in both the Go template function
// and JavaScript function. The actual template function is defined in main() and used
// via template.FuncMap, so we test the logic here to ensure correctness.
func TestURLNormalization(t *testing.T) {
	// Test the URL normalization logic (same as used in template function)
	funcMap := map[string]interface{}{
		"normalizeURL": func(url string) string {
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
	}

	normalizeURL := funcMap["normalizeURL"].(func(string) string)

	testCases := []struct {
		input    string
		expected string
	}{
		{"example.com", "https://example.com"},
		{"www.example.com", "https://www.example.com"},
		{"https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
		{"HTTPS://Example.COM", "HTTPS://Example.COM"},
		{"HTTP://Example.COM", "HTTP://Example.COM"},
		{"", ""},
		{"  example.com  ", "https://example.com"},
		{"github.com/user/repo", "https://github.com/user/repo"},
	}

	for _, tc := range testCases {
		result := normalizeURL(tc.input)
		if result != tc.expected {
			t.Errorf("normalizeURL(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// TestAssignTaskToFilteredStudents tests the filtered student assignment feature
func TestAssignTaskToFilteredStudents(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tables
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			task_type TEXT,
			resource_url TEXT,
			cohort_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Create an instructor
	result, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"instructor1", "password", "instructor", "",
	)
	if err != nil {
		t.Fatalf("Failed to create instructor: %v", err)
	}
	instructorID, _ := result.LastInsertId()

	// Create multiple students
	student1Result, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student1", "password", "student", "Java11.2",
	)
	if err != nil {
		t.Fatalf("Failed to create student1: %v", err)
	}
	student1ID, _ := student1Result.LastInsertId()

	student2Result, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student2", "password", "student", "Java11.2",
	)
	if err != nil {
		t.Fatalf("Failed to create student2: %v", err)
	}
	student2ID, _ := student2Result.LastInsertId()

	student3Result, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student3", "password", "student", "Data6.2",
	)
	if err != nil {
		t.Fatalf("Failed to create student3: %v", err)
	}
	student3ID, _ := student3Result.LastInsertId()

	// Test Case 1: Assign task to two filtered students (student1 and student2)
	filteredStudents := fmt.Sprintf("%d,%d", student1ID, student2ID)
	studentIDStrs := strings.Split(filteredStudents, ",")
	var studentIDs []int
	var args []interface{}

	for _, idStr := range studentIDStrs {
		studentID, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			t.Fatalf("Invalid student ID: %s", idStr)
		}
		studentIDs = append(studentIDs, studentID)
		args = append(args, studentID)
	}

	// Build placeholders for IN clause
	placeholderSlice := make([]string, len(studentIDs))
	for i := range placeholderSlice {
		placeholderSlice[i] = "?"
	}
	placeholders := strings.Join(placeholderSlice, ",")

	// Validate student IDs
	query := fmt.Sprintf("SELECT id FROM users WHERE id IN (%s) AND role = 'student'", placeholders)
	rows, err := testDB.Query(query, args...)
	if err != nil {
		t.Fatalf("Failed to validate student IDs: %v", err)
	}
	defer rows.Close()

	validStudentIDs := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			t.Fatalf("Failed to scan student ID: %v", err)
		}
		validStudentIDs[id] = true
	}

	// Verify both students are valid
	if len(validStudentIDs) != 2 {
		t.Errorf("Expected 2 valid students, got %d", len(validStudentIDs))
	}

	// Create tasks for valid students
	taskTitle := "Filtered Assignment Test"
	for id := range validStudentIDs {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, taskTitle, "Test description", "2025-12-31", id, instructorID, "", "", "", "")
		if err != nil {
			t.Fatalf("Failed to create task for student %d: %v", id, err)
		}
	}

	// Verify tasks were created for the correct students
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ? AND created_by = ?", taskTitle, instructorID).Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}
	if taskCount != 2 {
		t.Errorf("Expected 2 tasks to be created, got %d", taskCount)
	}

	// Verify student1 has the task
	var student1TaskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ? AND owner_id = ?", taskTitle, student1ID).Scan(&student1TaskCount)
	if err != nil {
		t.Fatalf("Failed to count student1 tasks: %v", err)
	}
	if student1TaskCount != 1 {
		t.Errorf("Expected student1 to have 1 task, got %d", student1TaskCount)
	}

	// Verify student2 has the task
	var student2TaskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ? AND owner_id = ?", taskTitle, student2ID).Scan(&student2TaskCount)
	if err != nil {
		t.Fatalf("Failed to count student2 tasks: %v", err)
	}
	if student2TaskCount != 1 {
		t.Errorf("Expected student2 to have 1 task, got %d", student2TaskCount)
	}

	// Verify student3 does NOT have the task
	var student3TaskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ? AND owner_id = ?", taskTitle, student3ID).Scan(&student3TaskCount)
	if err != nil {
		t.Fatalf("Failed to count student3 tasks: %v", err)
	}
	if student3TaskCount != 0 {
		t.Errorf("Expected student3 to have 0 tasks, got %d", student3TaskCount)
	}

	// Test Case 2: Assign task to single filtered student
	taskTitle2 := "Single Filtered Assignment Test"
	studentIDs2 := []int{int(student3ID)}

	for _, id := range studentIDs2 {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, taskTitle2, "Test description", "2025-12-31", id, instructorID, "", "", "", "")
		if err != nil {
			t.Fatalf("Failed to create task for student %d: %v", id, err)
		}
	}

	// Verify student3 has the new task
	var student3Task2Count int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ? AND owner_id = ?", taskTitle2, student3ID).Scan(&student3Task2Count)
	if err != nil {
		t.Fatalf("Failed to count student3 tasks for second test: %v", err)
	}
	if student3Task2Count != 1 {
		t.Errorf("Expected student3 to have 1 task for second test, got %d", student3Task2Count)
	}
}

// TestFilteredAssignmentWithInvalidIDs tests that invalid student IDs are handled correctly
func TestFilteredAssignmentWithInvalidIDs(t *testing.T) {
	testDB, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}
	defer testDB.Close()

	// Create tables
	_, err = testDB.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL,
			cohort_name TEXT
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	_, err = testDB.Exec(`
		CREATE TABLE tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending',
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			task_type TEXT,
			resource_url TEXT,
			cohort_id TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		t.Fatalf("Failed to create tasks table: %v", err)
	}

	// Create an instructor
	result, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"instructor1", "password", "instructor", "",
	)
	if err != nil {
		t.Fatalf("Failed to create instructor: %v", err)
	}
	instructorID, _ := result.LastInsertId()

	// Create one valid student
	studentResult, err := testDB.Exec(
		"INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student1", "password", "student", "Java11.2",
	)
	if err != nil {
		t.Fatalf("Failed to create student: %v", err)
	}
	studentID, _ := studentResult.LastInsertId()

	// Test with mix of valid and invalid IDs
	filteredStudents := fmt.Sprintf("%d,999,1000", studentID) // Include invalid IDs
	studentIDStrs := strings.Split(filteredStudents, ",")
	var studentIDs []int
	var args []interface{}

	for _, idStr := range studentIDStrs {
		id, err := strconv.Atoi(strings.TrimSpace(idStr))
		if err != nil {
			continue // Skip invalid strings
		}
		studentIDs = append(studentIDs, id)
		args = append(args, id)
	}

	// Validate student IDs
	placeholderSlice := make([]string, len(studentIDs))
	for i := range placeholderSlice {
		placeholderSlice[i] = "?"
	}
	placeholders := strings.Join(placeholderSlice, ",")

	query := fmt.Sprintf("SELECT id FROM users WHERE id IN (%s) AND role = 'student'", placeholders)
	rows, err := testDB.Query(query, args...)
	if err != nil {
		t.Fatalf("Failed to validate student IDs: %v", err)
	}
	defer rows.Close()

	validStudentIDs := make(map[int]bool)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			continue
		}
		validStudentIDs[id] = true
	}

	// Should only have 1 valid student (the invalid IDs should be filtered out)
	if len(validStudentIDs) != 1 {
		t.Errorf("Expected 1 valid student, got %d", len(validStudentIDs))
	}

	// Create task only for valid students
	taskTitle := "Mixed Valid/Invalid IDs Test"
	for id := range validStudentIDs {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, taskTitle, "Test description", "2025-12-31", id, instructorID, "", "", "", "")
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// Verify only 1 task was created (for the valid student)
	var taskCount int
	err = testDB.QueryRow("SELECT COUNT(*) FROM tasks WHERE title = ?", taskTitle).Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}
	if taskCount != 1 {
		t.Errorf("Expected 1 task to be created, got %d", taskCount)
	}
}

// TestGroupCreation tests creating a new group
func TestGroupCreation(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create test students
	studentIDs := []int{}
	for i := 1; i <= 3; i++ {
		result, err := testDB.Exec(`
			INSERT INTO users (username, password, role, cohort_name)
			VALUES (?, ?, 'student', 'Java11.2')
		`, fmt.Sprintf("student%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create student: %v", err)
		}
		id, _ := result.LastInsertId()
		studentIDs = append(studentIDs, int(id))
	}

	// Create a group
	result, err := testDB.Exec(`
		INSERT INTO groups (name, description, project_url, is_active)
		VALUES (?, ?, ?, ?)
	`, "Test Group", "A test group for project work", "https://github.com/test/project", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get group ID: %v", err)
	}

	// Add students to group
	for _, studentID := range studentIDs {
		_, err = testDB.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			t.Fatalf("Failed to add student to group: %v", err)
		}
	}

	// Verify group was created
	var name, description, projectURL string
	var isActive bool
	err = testDB.QueryRow(`
		SELECT name, description, project_url, is_active
		FROM groups WHERE id = ?
	`, groupID).Scan(&name, &description, &projectURL, &isActive)
	if err != nil {
		t.Fatalf("Failed to retrieve group: %v", err)
	}

	if name != "Test Group" {
		t.Errorf("Expected group name 'Test Group', got '%s'", name)
	}
	if !isActive {
		t.Errorf("Expected group to be active")
	}

	// Verify group members
	var memberCount int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM group_members WHERE group_id = ?
	`, groupID).Scan(&memberCount)
	if err != nil {
		t.Fatalf("Failed to count group members: %v", err)
	}

	if memberCount != 3 {
		t.Errorf("Expected 3 group members, got %d", memberCount)
	}
}

// TestGetGroupByID tests retrieving a group with its members
func TestGetGroupByID(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create test students
	studentIDs := []int{}
	for i := 1; i <= 2; i++ {
		result, err := testDB.Exec(`
			INSERT INTO users (username, password, role)
			VALUES (?, ?, 'student')
		`, fmt.Sprintf("student%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create student: %v", err)
		}
		id, _ := result.LastInsertId()
		studentIDs = append(studentIDs, int(id))
	}

	// Create a group
	result, err := testDB.Exec(`
		INSERT INTO groups (name, description, is_active)
		VALUES (?, ?, ?)
	`, "Project Team Alpha", "Alpha team", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	groupID, _ := result.LastInsertId()

	// Add students to group
	for _, studentID := range studentIDs {
		_, err = testDB.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			t.Fatalf("Failed to add student to group: %v", err)
		}
	}

	// Test getGroupByID function
	group, err := getGroupByID(int(groupID))
	if err != nil {
		t.Fatalf("Failed to get group by ID: %v", err)
	}

	if group.Name != "Project Team Alpha" {
		t.Errorf("Expected group name 'Project Team Alpha', got '%s'", group.Name)
	}

	if len(group.Students) != 2 {
		t.Errorf("Expected 2 students in group, got %d", len(group.Students))
	}
}

// TestAssignTaskToGroup tests assigning a task to all members of a group
func TestAssignTaskToGroup(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create instructor
	result, err := testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES ('instructor1', 'password123', 'instructor')
	`)
	if err != nil {
		t.Fatalf("Failed to create instructor: %v", err)
	}
	instructorID, _ := result.LastInsertId()

	// Create test students
	studentIDs := []int{}
	for i := 1; i <= 3; i++ {
		result, err := testDB.Exec(`
			INSERT INTO users (username, password, role)
			VALUES (?, ?, 'student')
		`, fmt.Sprintf("student%d", i), "password123")
		if err != nil {
			t.Fatalf("Failed to create student: %v", err)
		}
		id, _ := result.LastInsertId()
		studentIDs = append(studentIDs, int(id))
	}

	// Create a group
	result, err = testDB.Exec(`
		INSERT INTO groups (name, is_active)
		VALUES (?, ?)
	`, "Test Group", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID, _ := result.LastInsertId()

	// Add students to group
	for _, studentID := range studentIDs {
		_, err = testDB.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			t.Fatalf("Failed to add student to group: %v", err)
		}
	}

	// Create a task for each group member
	taskTitle := "Group Project Task"
	for _, studentID := range studentIDs {
		_, err := testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, group_id)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, taskTitle, "Work on group project", "2025-12-31", studentID, instructorID, "group-programming-project", groupID)
		if err != nil {
			t.Fatalf("Failed to create task: %v", err)
		}
	}

	// Verify tasks were created for all group members
	var taskCount int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks WHERE title = ? AND group_id = ?
	`, taskTitle, groupID).Scan(&taskCount)
	if err != nil {
		t.Fatalf("Failed to count tasks: %v", err)
	}

	if taskCount != 3 {
		t.Errorf("Expected 3 tasks to be created (one per group member), got %d", taskCount)
	}

	// Verify each student has the task
	for _, studentID := range studentIDs {
		var count int
		err = testDB.QueryRow(`
			SELECT COUNT(*) FROM tasks WHERE title = ? AND owner_id = ? AND group_id = ?
		`, taskTitle, studentID, groupID).Scan(&count)
		if err != nil {
			t.Fatalf("Failed to check task for student %d: %v", studentID, err)
		}
		if count != 1 {
			t.Errorf("Expected student %d to have 1 task, got %d", studentID, count)
		}
	}
}

// TestGroupNameRetrieval tests getting group name by ID
func TestGroupNameRetrieval(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create a group
	result, err := testDB.Exec(`
		INSERT INTO groups (name, is_active)
		VALUES (?, ?)
	`, "Team Beta", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID, _ := result.LastInsertId()

	// Test getGroupNameByID function
	groupName := getGroupNameByID(int(groupID))
	if groupName != "Team Beta" {
		t.Errorf("Expected group name 'Team Beta', got '%s'", groupName)
	}

	// Test with non-existent group ID
	nonExistentName := getGroupNameByID(9999)
	if nonExistentName != "" {
		t.Errorf("Expected empty string for non-existent group, got '%s'", nonExistentName)
	}
}

// TestUpdateGroup tests updating group information
func TestUpdateGroup(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create initial group
	result, err := testDB.Exec(`
		INSERT INTO groups (name, description, is_active)
		VALUES (?, ?, ?)
	`, "Old Name", "Old Description", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID, _ := result.LastInsertId()

	// Update group
	_, err = testDB.Exec(`
		UPDATE groups
		SET name = ?, description = ?, is_active = ?
		WHERE id = ?
	`, "New Name", "New Description", false, groupID)
	if err != nil {
		t.Fatalf("Failed to update group: %v", err)
	}

	// Verify updates
	var name, description string
	var isActive bool
	err = testDB.QueryRow(`
		SELECT name, description, is_active FROM groups WHERE id = ?
	`, groupID).Scan(&name, &description, &isActive)
	if err != nil {
		t.Fatalf("Failed to retrieve updated group: %v", err)
	}

	if name != "New Name" {
		t.Errorf("Expected name 'New Name', got '%s'", name)
	}
	if description != "New Description" {
		t.Errorf("Expected description 'New Description', got '%s'", description)
	}
	if isActive {
		t.Errorf("Expected group to be inactive")
	}
}

// TestDeleteGroup tests deleting a group and cascading deletes
func TestDeleteGroup(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create a student
	result, err := testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES ('student1', 'password123', 'student')
	`)
	if err != nil {
		t.Fatalf("Failed to create student: %v", err)
	}
	studentID, _ := result.LastInsertId()

	// Create a group
	result, err = testDB.Exec(`
		INSERT INTO groups (name, is_active)
		VALUES (?, ?)
	`, "Temporary Group", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID, _ := result.LastInsertId()

	// Add student to group
	_, err = testDB.Exec(`
		INSERT INTO group_members (group_id, student_id)
		VALUES (?, ?)
	`, groupID, studentID)
	if err != nil {
		t.Fatalf("Failed to add student to group: %v", err)
	}

	// Delete the group
	_, err = testDB.Exec(`DELETE FROM groups WHERE id = ?`, groupID)
	if err != nil {
		t.Fatalf("Failed to delete group: %v", err)
	}

	// Verify group is deleted
	var count int
	err = testDB.QueryRow(`SELECT COUNT(*) FROM groups WHERE id = ?`, groupID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check group deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected group to be deleted, but it still exists")
	}

	// Verify group_members entry was cascade deleted
	err = testDB.QueryRow(`SELECT COUNT(*) FROM group_members WHERE group_id = ?`, groupID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to check group_members deletion: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected group_members to be cascade deleted, but found %d entries", count)
	}
}

// TestGetAllGroups tests retrieving all groups
func TestGetAllGroups(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create multiple groups
	groupNames := []string{"Alpha Team", "Beta Team", "Gamma Team"}
	for _, name := range groupNames {
		_, err := testDB.Exec(`
			INSERT INTO groups (name, is_active)
			VALUES (?, ?)
		`, name, true)
		if err != nil {
			t.Fatalf("Failed to create group '%s': %v", name, err)
		}
	}

	// Test getAllGroups function
	groups, err := getAllGroups()
	if err != nil {
		t.Fatalf("Failed to get all groups: %v", err)
	}

	if len(groups) != 3 {
		t.Errorf("Expected 3 groups, got %d", len(groups))
	}

	// Verify group names
	foundNames := make(map[string]bool)
	for _, group := range groups {
		foundNames[group.Name] = true
	}

	for _, expectedName := range groupNames {
		if !foundNames[expectedName] {
			t.Errorf("Expected to find group '%s' in results", expectedName)
		}
	}
}

// TestGroupDetailWithMembers tests retrieving group details with member information
func TestGroupDetailWithMembers(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create test students
	studentIDs := []int{}
	studentNames := []string{"alice", "bob", "charlie"}
	for _, name := range studentNames {
		result, err := testDB.Exec(`
			INSERT INTO users (username, password, role, cohort_name)
			VALUES (?, ?, 'student', 'Java11.2')
		`, name, "password123")
		if err != nil {
			t.Fatalf("Failed to create student: %v", err)
		}
		id, _ := result.LastInsertId()
		studentIDs = append(studentIDs, int(id))
	}

	// Create a group
	result, err := testDB.Exec(`
		INSERT INTO groups (name, description, project_url, is_active)
		VALUES (?, ?, ?, ?)
	`, "Project Alpha", "Team project for final exam", "https://github.com/team/alpha", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}

	groupID, err := result.LastInsertId()
	if err != nil {
		t.Fatalf("Failed to get group ID: %v", err)
	}

	// Add students to group
	for _, studentID := range studentIDs {
		_, err = testDB.Exec(`
			INSERT INTO group_members (group_id, student_id)
			VALUES (?, ?)
		`, groupID, studentID)
		if err != nil {
			t.Fatalf("Failed to add student to group: %v", err)
		}
	}

	// Retrieve group details using getGroupByID
	group, err := getGroupByID(int(groupID))
	if err != nil {
		t.Fatalf("Failed to retrieve group: %v", err)
	}

	// Verify group details
	if group.Name != "Project Alpha" {
		t.Errorf("Expected group name 'Project Alpha', got '%s'", group.Name)
	}
	if group.Description != "Team project for final exam" {
		t.Errorf("Expected description 'Team project for final exam', got '%s'", group.Description)
	}
	if !group.IsActive {
		t.Errorf("Expected group to be active")
	}

	// Verify member count
	if len(group.Students) != 3 {
		t.Errorf("Expected 3 students in group, got %d", len(group.Students))
	}

	// Verify all student IDs are present
	foundStudents := make(map[int]bool)
	for _, studentID := range group.Students {
		foundStudents[studentID] = true
	}

	for _, expectedID := range studentIDs {
		if !foundStudents[expectedID] {
			t.Errorf("Expected to find student ID %d in group members", expectedID)
		}
	}

	// Get student details for verification (simulating what group detail page does)
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

	rows, err := testDB.Query(query, args...)
	if err != nil {
		t.Fatalf("Failed to query student details: %v", err)
	}
	defer rows.Close()

	var students []User
	for rows.Next() {
		var s User
		if err := rows.Scan(&s.ID, &s.Username, &s.CohortName, &s.Role); err != nil {
			t.Fatalf("Failed to scan student: %v", err)
		}
		students = append(students, s)
	}

	// Verify we got all students
	if len(students) != 3 {
		t.Errorf("Expected to retrieve 3 student details, got %d", len(students))
	}

	// Verify student data
	for _, student := range students {
		if student.Role != "student" {
			t.Errorf("Expected role 'student', got '%s' for user %s", student.Role, student.Username)
		}
		if student.CohortName != "Java11.2" {
			t.Errorf("Expected cohort 'Java11.2', got '%s' for user %s", student.CohortName, student.Username)
		}
	}
}

// TestGroupTasksRetrieval tests retrieving tasks assigned to a group
func TestGroupTasksRetrieval(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()
	db = testDB

	// Create instructor
	instructorResult, err := testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES (?, ?, 'instructor')
	`, "instructor1", "password123")
	if err != nil {
		t.Fatalf("Failed to create instructor: %v", err)
	}
	instructorID, _ := instructorResult.LastInsertId()

	// Create students
	student1Result, err := testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES (?, ?, 'student')
	`, "student1", "password123")
	if err != nil {
		t.Fatalf("Failed to create student1: %v", err)
	}
	student1ID, _ := student1Result.LastInsertId()

	student2Result, err := testDB.Exec(`
		INSERT INTO users (username, password, role)
		VALUES (?, ?, 'student')
	`, "student2", "password123")
	if err != nil {
		t.Fatalf("Failed to create student2: %v", err)
	}
	student2ID, _ := student2Result.LastInsertId()

	// Create a group
	groupResult, err := testDB.Exec(`
		INSERT INTO groups (name, description, is_active)
		VALUES (?, ?, ?)
	`, "Dev Team", "Development team for project", true)
	if err != nil {
		t.Fatalf("Failed to create group: %v", err)
	}
	groupID, _ := groupResult.LastInsertId()

	// Add students to group
	_, err = testDB.Exec(`
		INSERT INTO group_members (group_id, student_id)
		VALUES (?, ?), (?, ?)
	`, groupID, student1ID, groupID, student2ID)
	if err != nil {
		t.Fatalf("Failed to add students to group: %v", err)
	}

	// Create tasks for the group (one for each member)
	taskTitle := "Complete Sprint 1"
	taskType := "group-programming-project"

	_, err = testDB.Exec(`
		INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, group_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, taskTitle, "Finish all user stories", "2025-11-15", student1ID, instructorID, taskType, groupID)
	if err != nil {
		t.Fatalf("Failed to create task for student1: %v", err)
	}

	_, err = testDB.Exec(`
		INSERT INTO tasks (title, description, due_date, owner_id, created_by, task_type, group_id)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, taskTitle, "Finish all user stories", "2025-11-15", student2ID, instructorID, taskType, groupID)
	if err != nil {
		t.Fatalf("Failed to create task for student2: %v", err)
	}

	// Retrieve tasks for this group (simulating group detail page)
	rows, err := testDB.Query(`
		SELECT id, title, due_date, COALESCE(task_type, '')
		FROM tasks
		WHERE group_id = ?
	`, groupID)
	if err != nil {
		t.Fatalf("Failed to query group tasks: %v", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.Title, &task.DueDate, &task.TaskType); err != nil {
			t.Fatalf("Failed to scan task: %v", err)
		}
		tasks = append(tasks, task)
	}

	// Should have 2 tasks (one per student)
	if len(tasks) != 2 {
		t.Errorf("Expected 2 tasks for group, got %d", len(tasks))
	}

	// Verify task details
	for _, task := range tasks {
		if task.Title != taskTitle {
			t.Errorf("Expected task title '%s', got '%s'", taskTitle, task.Title)
		}
		if task.TaskType != taskType {
			t.Errorf("Expected task type '%s', got '%s'", taskType, task.TaskType)
		}
		if task.DueDate != "2025-11-15" {
			t.Errorf("Expected due date '2025-11-15', got '%s'", task.DueDate)
		}
	}
}

func TestDeleteGroupTask(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()

	// Create instructor
	result, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"instructor1", "password123", "instructor")
	if err != nil {
		t.Fatal(err)
	}
	instructorID, _ := result.LastInsertId()

	// Create students in Java11.2 cohort
	student1Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student1", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student1ID, _ := student1Result.LastInsertId()

	student2Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student2", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student2ID, _ := student2Result.LastInsertId()

	student3Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student3", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student3ID, _ := student3Result.LastInsertId()

	// Create a task for all students in the cohort
	taskTitle := "Complete Lab 5"
	taskDescription := "Complete all exercises in Lab 5"
	dueDate := "2025-12-01"
	taskType := "individual-programming-lab"

	for _, studentID := range []int64{student1ID, student2ID, student3ID} {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, task_type, cohort_id, created_at)
			VALUES (?, ?, ?, 'pending', ?, ?, ?, 'Java11.2', datetime('now'))
		`, taskTitle, taskDescription, dueDate, studentID, instructorID, taskType)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify tasks were created
	var count int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE title = ? AND created_by = ? AND cohort_id = 'Java11.2'
	`, taskTitle, instructorID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 3 {
		t.Fatalf("Expected 3 tasks to be created, got %d", count)
	}

	// Simulate deletion using the same logic as deleteGroupTaskHandler
	deleteResult, err := testDB.Exec(`
		DELETE FROM tasks
		WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
	`, taskTitle, instructorID, "Java11.2")
	if err != nil {
		t.Fatal(err)
	}

	rowsAffected, _ := deleteResult.RowsAffected()
	if rowsAffected != 3 {
		t.Errorf("Expected 3 tasks to be deleted, got %d", rowsAffected)
	}

	// Verify all tasks were deleted
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE title = ? AND created_by = ? AND cohort_id = 'Java11.2'
	`, taskTitle, instructorID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", count)
	}
}

func TestEditGroupTask(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()

	// Create instructor
	result, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"instructor1", "password123", "instructor")
	if err != nil {
		t.Fatal(err)
	}
	instructorID, _ := result.LastInsertId()

	// Create students in Java11.2 cohort
	student1Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student1", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student1ID, _ := student1Result.LastInsertId()

	student2Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student2", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student2ID, _ := student2Result.LastInsertId()

	student3Result, err := testDB.Exec("INSERT INTO users (username, password, role, cohort_name) VALUES (?, ?, ?, ?)",
		"student3", "password123", "student", "Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	student3ID, _ := student3Result.LastInsertId()

	// Create a task for all students in the cohort
	taskTitle := "Complete Lab 5"
	taskDescription := "Complete all exercises in Lab 5"
	dueDate := "2025-12-01"
	taskType := "individual-programming-lab"
	resourceURL := "https://github.com/example/lab5"

	for _, studentID := range []int64{student1ID, student2ID, student3ID} {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, task_type, resource_url, cohort_id, created_at)
			VALUES (?, ?, ?, 'pending', ?, ?, ?, ?, 'Java11.2', datetime('now'))
		`, taskTitle, taskDescription, dueDate, studentID, instructorID, taskType, resourceURL)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Simulate edit using the same logic as editGroupTaskHandler
	newTitle := "Complete Lab 5 - Updated"
	newDescription := "Complete all exercises in Lab 5 - Updated instructions"
	newDueDate := "2025-12-15"
	newTaskType := "group-programming-project"
	newResourceURL := "https://github.com/example/lab5-updated"

	updateResult, err := testDB.Exec(`
		UPDATE tasks
		SET title = ?, description = ?, due_date = ?, task_type = ?, resource_url = ?
		WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
	`, newTitle, newDescription, newDueDate, newTaskType, newResourceURL, taskTitle, instructorID, "Java11.2")
	if err != nil {
		t.Fatal(err)
	}

	rowsAffected, _ := updateResult.RowsAffected()
	if rowsAffected != 3 {
		t.Errorf("Expected 3 tasks to be updated, got %d", rowsAffected)
	}

	// Verify all tasks were updated
	rows, err := testDB.Query(`
		SELECT title, description, due_date, task_type, resource_url
		FROM tasks
		WHERE created_by = ? AND cohort_id = 'Java11.2'
		ORDER BY owner_id
	`, instructorID)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var title, description, dueDate, taskType, resourceURL string
		err := rows.Scan(&title, &description, &dueDate, &taskType, &resourceURL)
		if err != nil {
			t.Fatal(err)
		}

		if title != newTitle {
			t.Errorf("Expected title '%s', got '%s'", newTitle, title)
		}
		if description != newDescription {
			t.Errorf("Expected description '%s', got '%s'", newDescription, description)
		}
		if dueDate != newDueDate {
			t.Errorf("Expected due date '%s', got '%s'", newDueDate, dueDate)
		}
		if taskType != newTaskType {
			t.Errorf("Expected task type '%s', got '%s'", newTaskType, taskType)
		}
		if resourceURL != newResourceURL {
			t.Errorf("Expected resource URL '%s', got '%s'", newResourceURL, resourceURL)
		}
		count++
	}

	if count != 3 {
		t.Errorf("Expected 3 tasks to be updated, got %d", count)
	}
}

func TestDeleteGroupTaskWithEmptyCohortID(t *testing.T) {
	testDB := createTestDB(t)
	defer testDB.Close()

	// Create instructor
	result, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"instructor1", "password123", "instructor")
	if err != nil {
		t.Fatal(err)
	}
	instructorID, _ := result.LastInsertId()

	// Create students with no cohort
	student1Result, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"student1", "password123", "student")
	if err != nil {
		t.Fatal(err)
	}
	student1ID, _ := student1Result.LastInsertId()

	student2Result, err := testDB.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
		"student2", "password123", "student")
	if err != nil {
		t.Fatal(err)
	}
	student2ID, _ := student2Result.LastInsertId()

	// Create a task for all students (no cohort - assigned to "all students")
	taskTitle := "General Assignment"
	taskDescription := "Complete this general assignment"
	dueDate := "2025-12-01"

	for _, studentID := range []int64{student1ID, student2ID} {
		_, err = testDB.Exec(`
			INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, created_at)
			VALUES (?, ?, ?, 'pending', ?, ?, datetime('now'))
		`, taskTitle, taskDescription, dueDate, studentID, instructorID)
		if err != nil {
			t.Fatal(err)
		}
	}

	// Verify tasks were created
	var count int
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ''
	`, taskTitle, instructorID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("Expected 2 tasks to be created, got %d", count)
	}

	// Simulate deletion using the same logic as deleteGroupTaskHandler
	deleteResult, err := testDB.Exec(`
		DELETE FROM tasks
		WHERE title = ? AND created_by = ? AND COALESCE(cohort_id, '') = ?
	`, taskTitle, instructorID, "")
	if err != nil {
		t.Fatal(err)
	}

	rowsAffected, _ := deleteResult.RowsAffected()
	if rowsAffected != 2 {
		t.Errorf("Expected 2 tasks to be deleted, got %d", rowsAffected)
	}

	// Verify all tasks were deleted
	err = testDB.QueryRow(`
		SELECT COUNT(*) FROM tasks
		WHERE title = ? AND created_by = ?
	`, taskTitle, instructorID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Errorf("Expected 0 tasks after deletion, got %d", count)
	}
}
