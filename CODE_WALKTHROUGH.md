# TaskFolio Code Walkthrough

## Introduction

This guide walks you through the TaskFolio codebase step-by-step, explaining how each part works and how they fit together. This is designed for developers who want to understand how the application is built and how to work with the code.

## Table of Contents

1. [Getting Oriented](#getting-oriented)
2. [Application Startup](#application-startup)
3. [Database Initialization](#database-initialization)
4. [Session Management](#session-management)
5. [HTTP Routing](#http-routing)
6. [Authentication](#authentication)
7. [Task Management](#task-management)
8. [Template Rendering](#template-rendering)
9. [Testing](#testing)
10. [Common Patterns](#common-patterns)

---

## Getting Oriented

### File Structure
```
main.go              # All application code (~3700 lines)
main_test.go         # All tests (~1500 lines)
templates/           # HTML templates
  ├── login.html
  ├── dashboard.html
  ├── tasks.html
  └── ...
```

### Where to Start

When reading `main.go`, follow this order:

1. **Type definitions** (lines ~25-72) - Understand the data structures
2. **Global variables** (lines ~22-23) - See what's shared across requests
3. **main() function** (near end) - See how the app starts
4. **initDB()** (lines ~73-400+) - Database schema creation
5. **HTTP handlers** (throughout) - Individual request handlers

---

## Application Startup

### The main() Function

Located at the end of `main.go`:

```go
func main() {
    // 1. Initialize database
    err := initDB()
    if err != nil {
        log.Fatal("Failed to initialize database:", err)
    }
    defer db.Close()

    // 2. Seed demo data (if needed)
    seedDemoData()

    // 3. Load HTML templates
    templates = template.Must(template.ParseGlob("templates/*.html"))

    // 4. Set up HTTP routes
    http.HandleFunc("/login", loginHandler)
    http.HandleFunc("/logout", logoutHandler)
    http.HandleFunc("/", requireAuth(indexHandler))
    http.HandleFunc("/tasks/add", requireAuth(addTaskHandler))
    // ... more routes

    // 5. Determine port
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }

    // 6. Start HTTP server
    log.Printf("Server starting on port %s", port)
    log.Fatal(http.ListenAndServe(":"+port, nil))
}
```

### Startup Flow

```
1. initDB()          → Create/migrate database schema
   ↓
2. seedDemoData()    → Add demo users if empty
   ↓
3. Load templates    → Parse all HTML files
   ↓
4. Register routes   → Map URLs to handlers
   ↓
5. Start server      → Listen on port 8080
```

### What Gets Initialized

**Global Variables**:
```go
var db *sql.DB                          // Database connection
var templates *template.Template         // Parsed HTML templates
var sessions = make(map[string]SessionData)  // Active user sessions
```

These are **global** because:
- `db`: Shared connection pool for all requests
- `templates`: Loaded once at startup, used for all responses
- `sessions`: In-memory store of logged-in users

---

## Database Initialization

### The initDB() Function

This function is critical - it:
1. Opens SQLite database connection
2. Creates tables if they don't exist
3. Migrates schema if needed (adds new columns to existing tables)

### Creating Tables

```go
func initDB() error {
    // Open database
    dbPath := os.Getenv("DATABASE_PATH")
    if dbPath == "" {
        dbPath = "./taskfolio.db"
    }
    db, err = sql.Open("sqlite3", dbPath)
    if err != nil {
        return err
    }

    // Create users table
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

    // Create tasks table
    _, err = db.Exec(`
        CREATE TABLE IF NOT EXISTS tasks (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            title TEXT NOT NULL,
            description TEXT,
            due_date TEXT,
            status TEXT DEFAULT 'pending',
            owner_id INTEGER NOT NULL,
            created_by INTEGER NOT NULL,
            -- ... more columns ...
            FOREIGN KEY (owner_id) REFERENCES users(id),
            FOREIGN KEY (created_by) REFERENCES users(id)
        )
    `)
    
    // ... more table creation ...
}
```

### Schema Migration Pattern

TaskFolio handles database evolution gracefully:

```go
// Add column if it doesn't exist
_, err = db.Exec(`
    ALTER TABLE tasks ADD COLUMN group_id INTEGER
`)
// Ignore error if column already exists
// (SQLite returns error if column exists)

// Check if column exists first (better approach)
rows, err := db.Query("PRAGMA table_info(tasks)")
// ... check if 'group_id' column is in the results ...
if !columnExists {
    db.Exec("ALTER TABLE tasks ADD COLUMN group_id INTEGER")
}
```

**Key Learning**: This pattern allows the app to work with old and new database schemas, important for production deployments.

### Demo Data Seeding

```go
func seedDemoData() {
    // Check if users exist
    var count int
    db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
    
    if count == 0 {
        // Add demo instructor
        db.Exec(`INSERT INTO users (username, password, role) 
                 VALUES (?, ?, ?)`, "instructor1", "password123", "instructor")
        
        // Add demo students
        db.Exec(`INSERT INTO users (username, password, role) 
                 VALUES (?, ?, ?)`, "student1", "password123", "student")
        // ... more demo users
    }
}
```

**Purpose**: Makes it easy to test the app without manual user creation.

---

## Session Management

### How Sessions Work

```
User logs in → Server creates session → Server sends cookie → Browser stores cookie
                                                                       ↓
Browser sends cookie with every request → Server validates session → Grants access
```

### Session Data Structure

```go
type SessionData struct {
    UserID   int
    Username string
    Role     string
}

// Global session store (in-memory)
var sessions = make(map[string]SessionData)
```

### Creating a Session (Login)

In `loginHandler()`:

```go
// 1. Validate username/password
var user User
err := db.QueryRow(
    "SELECT id, username, role FROM users WHERE username = ? AND password = ?",
    username, password,
).Scan(&user.ID, &user.Username, &user.Role)

if err != nil {
    // Invalid credentials
    return
}

// 2. Generate unique session ID
sessionID := generateRandomString(32)

// 3. Store session in memory
sessions[sessionID] = SessionData{
    UserID:   user.ID,
    Username: user.Username,
    Role:     user.Role,
}

// 4. Send session cookie to browser
http.SetCookie(w, &http.Cookie{
    Name:     "session_id",
    Value:    sessionID,
    Path:     "/",
    HttpOnly: true,  // Prevents JavaScript access
})
```

### Validating a Session (Middleware)

```go
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // 1. Get session cookie from request
        cookie, err := r.Cookie("session_id")
        if err != nil {
            // No cookie - redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // 2. Check if session exists
        session, exists := sessions[cookie.Value]
        if !exists {
            // Invalid session - redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
            return
        }

        // 3. Session valid - continue to handler
        next(w, r)
    }
}
```

### Destroying a Session (Logout)

```go
func logoutHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Get session ID from cookie
    cookie, err := r.Cookie("session_id")
    if err == nil {
        // 2. Delete session from memory
        delete(sessions, cookie.Value)
    }

    // 3. Clear cookie in browser
    http.SetCookie(w, &http.Cookie{
        Name:   "session_id",
        Value:  "",
        Path:   "/",
        MaxAge: -1,  // Tells browser to delete cookie
    })

    // 4. Redirect to login
    http.Redirect(w, r, "/login", http.StatusSeeOther)
}
```

---

## HTTP Routing

### Route Registration

In `main()`:

```go
// Public routes (no auth required)
http.HandleFunc("/login", loginHandler)
http.HandleFunc("/logout", logoutHandler)

// Protected routes (auth required)
http.HandleFunc("/", requireAuth(indexHandler))
http.HandleFunc("/tasks/add", requireAuth(addTaskHandler))
http.HandleFunc("/tasks/edit", requireAuth(editTaskHandler))
http.HandleFunc("/tasks/delete", requireAuth(deleteTaskHandler))
http.HandleFunc("/tasks/update-status", requireAuth(updateTaskStatusHandler))
// ... more routes
```

**Pattern**: 
- Public routes → Direct handler function
- Protected routes → Wrapped with `requireAuth()` middleware

### Handler Function Pattern

Every handler follows this structure:

```go
func someHandler(w http.ResponseWriter, r *http.Request) {
    // 1. Check HTTP method
    if r.Method == "GET" {
        // Handle GET request (show form/page)
    } else if r.Method == "POST" {
        // Handle POST request (process form)
    } else {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
}
```

### Getting the Current User

In a protected handler:

```go
func someProtectedHandler(w http.ResponseWriter, r *http.Request) {
    // Get session ID from cookie
    cookie, _ := r.Cookie("session_id")
    sessionID := cookie.Value

    // Look up session data
    session := sessions[sessionID]
    
    // Now you have:
    // session.UserID   - user's database ID
    // session.Username - username
    // session.Role     - "student" or "instructor"
}
```

---

## Authentication

### Login Flow

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Show login form
        templates.ExecuteTemplate(w, "login.html", nil)
        return
    }

    // POST - process login
    username := r.FormValue("username")
    password := r.FormValue("password")

    // Query database
    var user User
    err := db.QueryRow(`
        SELECT id, username, role, is_suspended 
        FROM users 
        WHERE username = ? AND password = ?
    `, username, password).Scan(&user.ID, &user.Username, &user.Role, &user.IsSuspended)

    if err == sql.ErrNoRows {
        // User not found or wrong password
        templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
            "Error": "Invalid credentials",
        })
        return
    }

    if user.IsSuspended {
        // Account suspended
        templates.ExecuteTemplate(w, "login.html", map[string]interface{}{
            "Error": "Account suspended",
        })
        return
    }

    // Create session (shown earlier)
    sessionID := generateRandomString(32)
    sessions[sessionID] = SessionData{
        UserID:   user.ID,
        Username: user.Username,
        Role:     user.Role,
    }

    http.SetCookie(w, &http.Cookie{
        Name:  "session_id",
        Value: sessionID,
        Path:  "/",
    })

    // Redirect to dashboard
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

### Authorization (Role Checking)

```go
func adminOnlyHandler(w http.ResponseWriter, r *http.Request) {
    // Get session
    cookie, _ := r.Cookie("session_id")
    session := sessions[cookie.Value]

    // Check role
    if session.Role != "instructor" {
        http.Error(w, "Forbidden - Instructors only", http.StatusForbidden)
        return
    }

    // User is instructor - proceed
    // ... handler logic ...
}
```

---

## Task Management

### Creating a Task

```go
func addTaskHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Show form
        cookie, _ := r.Cookie("session_id")
        session := sessions[cookie.Value]
        
        templates.ExecuteTemplate(w, "add_task.html", map[string]interface{}{
            "Username": session.Username,
            "Role":     session.Role,
        })
        return
    }

    // POST - create task
    cookie, _ := r.Cookie("session_id")
    session := sessions[cookie.Value]

    // Extract form data
    title := r.FormValue("title")
    description := r.FormValue("description")
    dueDate := r.FormValue("due_date")
    taskType := r.FormValue("task_type")
    resourceURL := r.FormValue("resource_url")

    // Validate
    if title == "" {
        http.Error(w, "Title required", http.StatusBadRequest)
        return
    }

    // Insert into database
    _, err := db.Exec(`
        INSERT INTO tasks 
        (title, description, due_date, owner_id, created_by, task_type, resource_url, status)
        VALUES (?, ?, ?, ?, ?, ?, ?, 'pending')
    `, title, description, dueDate, session.UserID, session.UserID, taskType, resourceURL)

    if err != nil {
        log.Printf("Error creating task: %v", err)
        http.Error(w, "Failed to create task", http.StatusInternalServerError)
        return
    }

    // Redirect to dashboard
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

### Listing Tasks

```go
func indexHandler(w http.ResponseWriter, r *http.Request) {
    cookie, _ := r.Cookie("session_id")
    session := sessions[cookie.Value]

    // Query tasks for this user
    rows, err := db.Query(`
        SELECT id, title, description, due_date, status, task_type, resource_url
        FROM tasks
        WHERE owner_id = ?
        ORDER BY created_at DESC
    `, session.UserID)
    
    if err != nil {
        log.Printf("Error querying tasks: %v", err)
        http.Error(w, "Database error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    // Scan results into slice
    var tasks []Task
    for rows.Next() {
        var t Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, 
                        &t.Status, &t.TaskType, &t.ResourceURL)
        if err != nil {
            log.Printf("Error scanning task: %v", err)
            continue
        }
        tasks = append(tasks, t)
    }

    // Calculate stats
    stats := TaskStats{}
    for _, t := range tasks {
        stats.Total++
        switch t.Status {
        case "pending":
            stats.Pending++
        case "in_progress":
            stats.InProgress++
        case "completed":
            stats.Completed++
        case "blocked":
            stats.Blocked++
        }
    }

    // Render template
    templates.ExecuteTemplate(w, "dashboard.html", map[string]interface{}{
        "Username": session.Username,
        "Role":     session.Role,
        "Tasks":    tasks,
        "Stats":    stats,
    })
}
```

### Updating Task Status

```go
func updateTaskStatusHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != "POST" {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    // Get parameters
    taskID := r.FormValue("id")
    newStatus := r.FormValue("status")

    // Validate status
    validStatuses := map[string]bool{
        "pending": true, "in_progress": true, "completed": true, "blocked": true,
    }
    if !validStatuses[newStatus] {
        http.Error(w, "Invalid status", http.StatusBadRequest)
        return
    }

    // Get session
    cookie, _ := r.Cookie("session_id")
    session := sessions[cookie.Value]

    // Verify ownership
    var ownerID int
    err := db.QueryRow("SELECT owner_id FROM tasks WHERE id = ?", taskID).Scan(&ownerID)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    if ownerID != session.UserID && session.Role != "instructor" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Update status with timestamp tracking
    if newStatus == "in_progress" {
        _, err = db.Exec(`
            UPDATE tasks 
            SET status = ?, in_progress_date = CURRENT_TIMESTAMP 
            WHERE id = ? AND in_progress_date IS NULL
        `, newStatus, taskID)
    } else if newStatus == "completed" {
        _, err = db.Exec(`
            UPDATE tasks 
            SET status = ?, done_date = CURRENT_TIMESTAMP 
            WHERE id = ?
        `, newStatus, taskID)
    } else {
        _, err = db.Exec("UPDATE tasks SET status = ? WHERE id = ?", newStatus, taskID)
    }

    if err != nil {
        log.Printf("Error updating status: %v", err)
        http.Error(w, "Failed to update", http.StatusInternalServerError)
        return
    }

    // Return success (typically as JSON for AJAX calls)
    w.Header().Set("Content-Type", "application/json")
    w.Write([]byte(`{"success": true}`))
}
```

### Deleting a Task

```go
func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
    // Get task ID from query string
    taskID := r.URL.Query().Get("id")
    if taskID == "" {
        http.Error(w, "Task ID required", http.StatusBadRequest)
        return
    }

    // Get session
    cookie, _ := r.Cookie("session_id")
    session := sessions[cookie.Value]

    // Verify ownership
    var ownerID int
    err := db.QueryRow("SELECT owner_id FROM tasks WHERE id = ?", taskID).Scan(&ownerID)
    if err != nil {
        http.Error(w, "Task not found", http.StatusNotFound)
        return
    }

    // Only owner or instructor can delete
    if ownerID != session.UserID && session.Role != "instructor" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Delete task
    _, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
    if err != nil {
        log.Printf("Error deleting task: %v", err)
        http.Error(w, "Failed to delete", http.StatusInternalServerError)
        return
    }

    // Redirect back to dashboard
    http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

---

## Template Rendering

### Loading Templates

In `main()`:

```go
templates = template.Must(template.ParseGlob("templates/*.html"))
```

This loads all `.html` files from the `templates/` directory.

### Rendering a Template

```go
// Simple render
templates.ExecuteTemplate(w, "login.html", nil)

// Render with data
data := map[string]interface{}{
    "Username": "john_doe",
    "Tasks":    tasks,
    "Message":  "Welcome!",
}
templates.ExecuteTemplate(w, "dashboard.html", data)
```

### Template Data Access

In `dashboard.html`:

```html
<h1>Welcome, {{.Username}}!</h1>

{{if .Message}}
    <div class="alert">{{.Message}}</div>
{{end}}

<ul>
{{range .Tasks}}
    <li>
        <strong>{{.Title}}</strong> - {{.Status}}
        <p>{{.Description}}</p>
    </li>
{{end}}
</ul>
```

### Security: Escaping User Input

Go templates **automatically escape HTML** to prevent XSS:

```html
<!-- User input: <script>alert('XSS')</script> -->
<!-- Rendered as: &lt;script&gt;alert('XSS')&lt;/script&gt; -->
<p>{{.UserInput}}</p>
```

For additional safety, some handlers use:

```go
safeTitle := html.EscapeString(title)
```

---

## Testing

### Test Structure

In `main_test.go`:

```go
func TestCreateUser(t *testing.T) {
    // 1. Set up test database
    db := setupTestDB()
    defer db.Close()

    // 2. Execute test
    username := "testuser"
    password := "testpass"
    role := "student"

    _, err := db.Exec(`
        INSERT INTO users (username, password, role) 
        VALUES (?, ?, ?)
    `, username, password, role)

    // 3. Verify
    if err != nil {
        t.Fatalf("Failed to create user: %v", err)
    }

    // 4. Check database state
    var count int
    db.QueryRow("SELECT COUNT(*) FROM users WHERE username = ?", username).Scan(&count)
    
    if count != 1 {
        t.Errorf("Expected 1 user, got %d", count)
    }
}
```

### Test Database Setup

```go
func setupTestDB() *sql.DB {
    // Use in-memory database
    testDB, err := sql.Open("sqlite3", ":memory:")
    if err != nil {
        panic(err)
    }

    // Create schema
    testDB.Exec(`CREATE TABLE users (...)`)
    testDB.Exec(`CREATE TABLE tasks (...)`)
    
    return testDB
}
```

**Key Point**: Each test uses `:memory:` database - fast, isolated, no cleanup needed.

### Running Tests

```bash
# Run all tests
go test -v

# Run specific test
go test -run TestCreateUser -v

# Run with coverage
go test -cover
```

---

## Common Patterns

### Pattern 1: Database Query with Error Handling

```go
// Query single row
var user User
err := db.QueryRow("SELECT id, username FROM users WHERE id = ?", userID).
    Scan(&user.ID, &user.Username)

if err == sql.ErrNoRows {
    // Not found
    http.Error(w, "User not found", http.StatusNotFound)
    return
} else if err != nil {
    // Database error
    log.Printf("Database error: %v", err)
    http.Error(w, "Internal error", http.StatusInternalServerError)
    return
}
```

### Pattern 2: Query Multiple Rows

```go
rows, err := db.Query("SELECT id, title FROM tasks WHERE owner_id = ?", userID)
if err != nil {
    return nil, err
}
defer rows.Close()  // IMPORTANT: Always close rows

var tasks []Task
for rows.Next() {
    var t Task
    if err := rows.Scan(&t.ID, &t.Title); err != nil {
        log.Printf("Scan error: %v", err)
        continue  // Skip bad rows
    }
    tasks = append(tasks, t)
}

// Check for iteration errors
if err = rows.Err(); err != nil {
    return nil, err
}
```

### Pattern 3: Form Validation

```go
title := strings.TrimSpace(r.FormValue("title"))
if title == "" {
    http.Error(w, "Title is required", http.StatusBadRequest)
    return
}

if len(title) > 200 {
    http.Error(w, "Title too long (max 200 chars)", http.StatusBadRequest)
    return
}
```

### Pattern 4: Authorization Check

```go
cookie, _ := r.Cookie("session_id")
session := sessions[cookie.Value]

// Check if user can perform action
if session.Role != "instructor" {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### Pattern 5: Redirect After POST

```go
func createSomething(w http.ResponseWriter, r *http.Request) {
    if r.Method == "POST" {
        // ... create resource ...
        
        // Redirect to prevent duplicate submission on refresh
        http.Redirect(w, r, "/success", http.StatusSeeOther)
        return
    }
}
```

---

## Debugging Tips

### 1. Add Logging

```go
log.Printf("User %s attempting to delete task %s", session.Username, taskID)
```

### 2. Check Database State

```bash
sqlite3 taskfolio.db "SELECT * FROM users;"
sqlite3 taskfolio.db "SELECT * FROM tasks WHERE owner_id = 1;"
```

### 3. View HTTP Requests

```go
log.Printf("Request: %s %s", r.Method, r.URL.Path)
log.Printf("Form data: %v", r.Form)
```

### 4. Template Debugging

In templates, print data:

```html
<pre>{{.}}</pre>  <!-- Prints entire data object -->
```

### 5. Test in Isolation

```go
// Test just database logic
func TestTaskCreation(t *testing.T) {
    db := setupTestDB()
    defer db.Close()
    
    // Test without HTTP layer
}
```

---

## Summary

Key points from this walkthrough:

1. **Single-file architecture** - All code in `main.go`
2. **Global shared state** - `db`, `templates`, `sessions`
3. **Handler pattern** - One function per route, handles GET/POST
4. **Middleware pattern** - `requireAuth()` wraps handlers
5. **Prepared statements** - Always use `?` placeholders
6. **Session-based auth** - Cookie stores session ID, map stores data
7. **Template rendering** - Server-side HTML generation
8. **In-memory testing** - Fast, isolated test databases

**Next Steps**: 
- Read [ARCHITECTURE.md](./ARCHITECTURE.md) for design decisions
- See [LEARNING_GUIDE.md](./LEARNING_GUIDE.md) for exercises
- Try modifying code and running tests
- Add a new feature following these patterns


---

## Import Tasks Feature

### Overview

The Import Tasks feature allows instructors to bulk-import tasks from a CSV file and assign them to students. This feature demonstrates several important concepts:

1. **File upload handling** - Multipart form parsing
2. **CSV parsing** - Reading and validating CSV data
3. **Current student filtering** - Ensuring only active students receive tasks
4. **Flexible assignment** - Supporting multiple assignment modes

### CSV Format

The import expects a CSV file with 5 fields per line:
```
Task Title, Description, DueDate, TaskType, ResourceURL
```

Example:
```csv
Complete Lab 1, Finish the Java basics lab, 2025-11-15, individual-programming-lab, https://example.com/lab1
Read Chapter 5, Read about design patterns, 2025-11-10, reading, https://example.com/chapter5
```

### Import Task Handler

Location: `importTasksHandler()` in `main.go` (around line 2531)

Key steps:

1. **Authentication Check**: Only instructors can import tasks
2. **Parse Multipart Form**: Handle CSV file upload
3. **Determine Assignment Type**: Three modes supported
4. **Read and Parse CSV**: Validate and process each task
5. **Create Tasks**: Insert into database for each target student

### Assignment Modes

#### 1. All Students (with selection)
When "All Students" is selected:
- Frontend loads all current students via `/api/students`
- All current students are **pre-selected** in the student picker
- Instructor can deselect specific students using Ctrl/Cmd-click
- Backend reads selected student IDs from form data

```javascript
// Frontend: Load students with pre-selection
function loadStudents(selectAll) {
    fetch('/api/students')
        .then(response => response.json())
        .then(students => {
            const select = document.getElementById('studentSelect');
            select.innerHTML = '';
            students.forEach(student => {
                const option = document.createElement('option');
                option.value = student.id;
                option.textContent = student.username + (student.cohort ? ' (' + student.cohort + ')' : '');
                // Pre-select when "All Students" is chosen
                if (selectAll) {
                    option.selected = true;
                }
                select.appendChild(option);
            });
        });
}
```

Backend handling:
```go
case "all":
    // When "All Students" is selected, read the selected student IDs from the form
    // This allows instructors to deselect specific students
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
```

#### 2. Specific Cohort
- Shows only cohorts of current students
- Queries: `SELECT DISTINCT cohort_name FROM users WHERE role = 'student' AND is_current = 1`
- Assigns to all current students in selected cohort

```go
case "cohort":
    cohortName := r.FormValue("cohort")
    if cohortName == "" {
        http.Error(w, "Please select a cohort", http.StatusBadRequest)
        return
    }
    // Only current students in the cohort
    rows, err := db.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ? AND is_current = 1", cohortName)
    // ... process rows
```

#### 3. Specific Students
- Shows only current students
- Instructor selects individual students from the list

### Current Students Filtering

**Critical Design Decision**: The system only shows and assigns tasks to **current students** (where `is_current = 1`).

This is implemented in three key places:

1. **API Students Endpoint** (`/api/students`):
```go
rows, err := db.Query(`
    SELECT id, username, COALESCE(cohort_name, '') 
    FROM users 
    WHERE role = 'student' AND is_current = 1
    ORDER BY username
`)
```

2. **API Cohorts Endpoint** (`/api/cohorts`):
```go
rows, err := db.Query(`
    SELECT DISTINCT cohort_name 
    FROM users 
    WHERE role = 'student' AND cohort_name IS NOT NULL AND cohort_name != '' AND is_current = 1
    ORDER BY cohort_name
`)
```

3. **Import Task Handler** - All assignment modes filter by `is_current = 1`

### Why Filter Current Students?

Past students should not receive new task assignments. The `is_current` field allows the system to:
- Maintain historical student records
- Track past cohorts
- Prevent accidental task assignment to alumni
- Keep student lists clean and relevant

### User Experience Flow

1. Instructor clicks "Import Tasks" button
2. Form appears with CSV upload and assignment options
3. Instructor selects assignment type:
   - **All Students**: Student picker appears with all current students selected
   - **Specific Cohort**: Cohort dropdown appears (only current cohorts)
   - **Specific Students**: Student picker appears (only current students)
4. Instructor can deselect students in "All Students" mode using Ctrl/Cmd-click
5. On submit, tasks are created for each selected student

### Frontend Implementation

Key JavaScript functions (in `templates/home.html`):

- `updateAssignOptions()`: Shows/hides appropriate selection controls
- `loadCohorts()`: Fetches and populates cohort dropdown
- `loadStudents(selectAll)`: Fetches and populates student picker, with optional pre-selection

### Error Handling

The import handler validates:
- CSV format (5 fields expected)
- Task type validity
- At least one student selected
- File upload success

Errors are collected and returned to the user with line numbers for debugging.

### Testing

Tests for Import Tasks functionality:

1. **TestImportTasksCurrentStudentsOnly**: Verifies queries filter current students
2. **TestAPICohortsCurrentStudentsOnly**: Verifies cohorts API returns only current cohorts
3. **TestAPIStudentsCurrentStudentsOnly**: Verifies students API returns only current students
4. **TestImportTasks**: Original import functionality test
5. **TestImportTasksToAllStudents**: Tests bulk assignment
6. **TestImportTasksToSpecificStudents**: Tests selective assignment

### Code Navigation Guide

For future agents working on Import Tasks:

1. **Backend Handler**: Line ~2531 in `main.go` - `importTasksHandler()`
2. **API Endpoints**: 
   - Students: Line ~2820 - `apiStudentsHandler()`
   - Cohorts: Line ~2781 - `apiCohortsHandler()`
3. **Frontend Form**: `templates/home.html` lines 449-495
4. **Frontend JavaScript**: `templates/home.html` lines 1282-1347
5. **Tests**: `main_test.go` - search for "TestImport"

### Database Schema

Relevant tables and fields:

**users table**:
- `is_current`: BOOLEAN (1 for current students, 0 for past)
- `cohort_name`: TEXT (e.g., "Java11.2", "Data6.2")
- `role`: TEXT ('student' or 'instructor')

**tasks table**:
- `owner_id`: Links to user who owns the task
- `created_by`: Links to instructor who created it
- `cohort_id`: Optional cohort name for cohort-based assignments

### Modification Guidelines

When modifying Import Tasks:

1. **Always filter by `is_current = 1`** when querying students or cohorts
2. **Test with both current and past students** to verify filtering
3. **Update tests** when changing import logic
4. **Maintain backward compatibility** with existing CSV format
5. **Preserve the three assignment modes** unless explicitly changing the UX

### Common Issues and Solutions

**Issue**: Past students appearing in selection
- **Check**: API endpoints include `AND is_current = 1` filter
- **Check**: Frontend doesn't bypass the API filtering

**Issue**: "All Students" not pre-selecting students
- **Check**: `updateAssignOptions()` calls `loadStudents(true)` for 'all' case
- **Check**: `loadStudents()` sets `option.selected = true` when `selectAll` is true

**Issue**: Import fails with "No students found"
- **Check**: At least one current student exists
- **Check**: Assignment type is correctly set in form data
- **Check**: Student IDs are being parsed correctly from form

