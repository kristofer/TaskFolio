package main

import (
	"database/sql"
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
