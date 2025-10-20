package main

import (
	"database/sql"
	"fmt"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

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
	taskTypes := []string{"reading", "coding", "submission", "document-creation"}

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

	if taskType != "reading" {
		t.Errorf("Expected task type 'reading', got '%s'", taskType)
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
