# TaskFolio Architecture Guide

## Overview

This document explains the architecture of TaskFolio (branded as "Tasking"), a web-based task management system built with Go and SQLite3. This guide is designed to help you understand how the application is structured, how its components work together, and the design decisions behind its implementation.

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Technology Stack](#technology-stack)
3. [Application Structure](#application-structure)
4. [Data Model](#data-model)
5. [Request Flow](#request-flow)
6. [Design Patterns](#design-patterns)
7. [Security Model](#security-model)
8. [Key Architectural Decisions](#key-architectural-decisions)

---

## High-Level Architecture

TaskFolio follows a **monolithic server-side rendered web application** architecture:

```
┌─────────────┐
│   Browser   │
│  (Client)   │
└──────┬──────┘
       │ HTTP Requests
       ▼
┌─────────────────────────────────┐
│      Go HTTP Server             │
│  ┌──────────────────────────┐   │
│  │  HTTP Handlers           │   │
│  │  (Login, Tasks, etc.)    │   │
│  └───────────┬──────────────┘   │
│              │                  │
│  ┌───────────▼──────────────┐   │
│  │  Business Logic          │   │
│  │  (Auth, Validation)      │   │
│  └───────────┬──────────────┘   │
│              │                  │
│  ┌───────────▼──────────────┐   │
│  │  Database Layer          │   │
│  │  (SQL Queries)           │   │
│  └───────────┬──────────────┘   │
└──────────────┼──────────────────┘
               │
        ┌──────▼──────┐
        │   SQLite3   │
        │  Database   │
        └─────────────┘
```

### Key Characteristics:

- **Single Binary**: The entire application compiles to a single executable
- **Server-Side Rendering**: HTML is generated on the server using Go templates
- **Session-Based Auth**: Users authenticate via cookies storing session IDs
- **Embedded Database**: SQLite3 database runs in the same process as the web server

---

## Technology Stack

### Backend
- **Language**: Go (Golang) 1.24.7
- **Web Framework**: Go standard library `net/http`
- **Database**: SQLite3 via `github.com/mattn/go-sqlite3` driver
- **Templating**: Go's `html/template` package

### Why These Technologies?

1. **Go**: 
   - Simple, readable syntax ideal for learning
   - Excellent standard library for web development
   - Strong typing helps catch errors at compile time
   - Fast compilation and execution

2. **SQLite3**:
   - Zero-configuration database
   - Perfect for small to medium applications
   - File-based (no separate database server needed)
   - Full SQL support

3. **Server-Side Templates**:
   - Simpler than separate frontend framework
   - Easier to understand request/response flow
   - No JavaScript build tools required
   - Good for learning web fundamentals

---

## Application Structure

### File Organization

```
TaskFolio/
├── main.go              # All application code (intentional monolith)
├── main_test.go         # All tests
├── templates/           # HTML templates
│   ├── login.html
│   ├── dashboard.html
│   ├── tasks.html
│   └── ...
├── taskfolio.db         # SQLite database file (created at runtime)
├── go.mod               # Go module definition
├── go.sum               # Dependency checksums
├── Dockerfile           # Docker container definition
└── docker-compose.yml   # Docker deployment config
```

### Single-File Architecture

TaskFolio intentionally uses a **monolithic single-file structure** (`main.go`). This is an educational choice:

**Benefits for Learning**:
- All code is in one place - easy to search and understand flow
- No need to navigate multiple packages
- Clear cause-and-effect relationships
- Great for understanding web application fundamentals

**Trade-offs**:
- Not suitable for large production applications
- Would become unwieldy beyond ~5,000 lines
- Limited code reusability across projects

**When to Refactor**: If the application grows beyond 10,000 lines or needs to be split into microservices, consider adopting a multi-package structure (see `PROJECT_STRUCTURE.md` for recommendations).

---

## Data Model

### Core Entities

TaskFolio has three main entities:

#### 1. Users
```go
type User struct {
    ID          int
    Username    string
    Password    string  // ⚠️ Plain text (demo only!)
    Role        string  // "student" or "instructor"
    CohortName  string  // e.g., "Java11.2"
    IsSuspended bool
    IsCurrent   bool    // Active vs. alumni
}
```

#### 2. Tasks
```go
type Task struct {
    ID             int
    Title          string
    Description    string
    DueDate        string
    Status         string  // "pending", "in_progress", "completed", "blocked"
    OwnerID        int     // Who owns this task
    CreatedBy      int     // Who created it (may be instructor)
    AssignedTo     string  // Comma-separated usernames for groups
    TaskType       string  // Type of assignment
    ResourceURL    string  // Link to materials
    CohortID       string  // Cohort assignment
    InProgressDate *time.Time
    DoneDate       *time.Time
    GroupID        *int
    Comments       string
    DisplayOrder   int
}
```

#### 3. Groups
```go
type Group struct {
    ID          int
    Name        string
    Description string
    ProjectURL  string
    IsActive    bool
    CreatedAt   time.Time
    Students    []int  // Member IDs
}
```

### Database Schema

TaskFolio uses SQLite3 with a simple schema:

```sql
-- Users table
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
    cohort_name TEXT,
    is_suspended BOOLEAN DEFAULT 0,
    is_current BOOLEAN DEFAULT 1
);

-- Tasks table
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    due_date TEXT,
    status TEXT DEFAULT 'pending' 
        CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
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
    comments TEXT,
    display_order INTEGER DEFAULT 0,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);

-- Groups table  
CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    project_url TEXT,
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Group members junction table
CREATE TABLE group_members (
    group_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    PRIMARY KEY (group_id, user_id),
    FOREIGN KEY (group_id) REFERENCES groups(id),
    FOREIGN KEY (user_id) REFERENCES users(id)
);
```

### Data Relationships

```
User (1) ──owns──> (many) Task
User (1) ──creates──> (many) Task
User (many) ──belongs to──> (many) Group
Task (many) ──assigned to──> (1) Group
```

---

## Request Flow

Understanding how a web request flows through the application is key to understanding the architecture.

### Example: Viewing the Dashboard

```
1. Browser sends GET request to http://localhost:8080/
   ↓
2. Go's http.ServeMux routes to indexHandler()
   ↓
3. requireAuth() middleware checks for valid session cookie
   ├─ No session? Redirect to /login
   └─ Valid session? Continue to handler
   ↓
4. indexHandler retrieves tasks from database
   ├─ Queries: SELECT * FROM tasks WHERE owner_id = ?
   └─ Creates []Task slice
   ↓
5. Handler prepares template data struct
   ↓
6. templates.ExecuteTemplate("dashboard.html", data)
   ├─ Inserts data into HTML template
   └─ Generates final HTML
   ↓
7. HTTP response sent to browser with HTML content
   ↓
8. Browser renders HTML page
```

### Example: Creating a Task

```
1. User fills out form and submits (POST /tasks/add)
   ↓
2. Router sends to addTaskHandler()
   ↓
3. requireAuth() middleware validates session
   ↓
4. Handler parses form data: r.FormValue("title"), etc.
   ↓
5. Handler validates input
   ├─ Title not empty?
   ├─ Valid task type?
   └─ Valid due date?
   ↓
6. Handler checks authorization
   └─ Is user an instructor? (if assigning to others)
   ↓
7. Handler executes SQL INSERT with prepared statement
   ├─ db.Exec("INSERT INTO tasks (...) VALUES (?, ?, ...)", title, desc, ...)
   └─ Database assigns auto-increment ID
   ↓
8. Handler redirects to dashboard: http.Redirect("/")
   ↓
9. Browser follows redirect and loads updated dashboard
```

### Key Observations:

- **Stateless Handlers**: Each request is independent
- **Database as State**: All persistent data in SQLite
- **Sessions in Memory**: User authentication state in Go map
- **Server-Side Rendering**: HTML generated for each request

---

## Design Patterns

### 1. Handler Pattern

Each route has a dedicated handler function:

```go
func loginHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // Show login form
        templates.ExecuteTemplate(w, "login.html", nil)
    } else if r.Method == "POST" {
        // Process login
        username := r.FormValue("username")
        password := r.FormValue("password")
        // ... validate and create session
    }
}
```

**Pattern**: One function handles both GET (show form) and POST (process form)

### 2. Middleware Pattern

Cross-cutting concerns like authentication use middleware:

```go
func requireAuth(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Check if user is authenticated
        sessionID := getSessionID(r)
        if session, exists := sessions[sessionID]; exists {
            // Authenticated - continue to handler
            next(w, r)
        } else {
            // Not authenticated - redirect to login
            http.Redirect(w, r, "/login", http.StatusSeeOther)
        }
    }
}

// Usage:
http.HandleFunc("/", requireAuth(indexHandler))
```

**Pattern**: Wrap handlers with authentication check

### 3. Repository Pattern (Lightweight)

Database operations use prepared statements:

```go
func getTasksByUser(userID int) ([]Task, error) {
    rows, err := db.Query(
        "SELECT id, title, description, status FROM tasks WHERE owner_id = ?",
        userID,
    )
    if err != nil {
        return nil, err
    }
    defer rows.Close()
    
    var tasks []Task
    for rows.Next() {
        var t Task
        err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Status)
        if err != nil {
            return nil, err
        }
        tasks = append(tasks, t)
    }
    return tasks, nil
}
```

**Pattern**: Encapsulate SQL queries in functions

### 4. Template Data Pattern

Data is passed to templates via structs:

```go
type DashboardData struct {
    Username   string
    Role       string
    Tasks      []Task
    Stats      TaskStats
    CohortName string
}

data := DashboardData{
    Username: session.Username,
    Tasks:    tasks,
    // ...
}
templates.ExecuteTemplate(w, "dashboard.html", data)
```

**Pattern**: Strongly-typed data contracts between handlers and templates

### 5. Session Management Pattern

Simple in-memory session storage:

```go
// Global session store
var sessions = make(map[string]SessionData)

type SessionData struct {
    UserID   int
    Username string
    Role     string
}

// Create session
sessionID := generateRandomString()
sessions[sessionID] = SessionData{
    UserID:   user.ID,
    Username: user.Username,
    Role:     user.Role,
}

// Store in cookie
http.SetCookie(w, &http.Cookie{
    Name:  "session_id",
    Value: sessionID,
})
```

**Pattern**: Map-based session storage with cookie-based session ID

---

## Security Model

### Authentication Flow

```
1. User submits username/password
   ↓
2. Server queries database for user
   ↓
3. Compare submitted password with stored password
   ├─ Match? Create session
   └─ No match? Show error
   ↓
4. Store session in global sessions map
   ↓
5. Send session_id cookie to browser
   ↓
6. Browser includes cookie in all subsequent requests
   ↓
7. requireAuth() middleware validates session_id
```

### Authorization Model

**Role-Based Access Control (RBAC)**:

```
Student can:
- Create own tasks
- View own tasks
- Edit own tasks
- Delete own tasks
- View tasks assigned by instructor

Instructor can:
- All student capabilities
- Create tasks for others
- Assign tasks to cohorts
- View all student tasks
- Manage groups
- Suspend user accounts
```

Implementation:
```go
// Check in handler
session := sessions[sessionID]
if session.Role != "instructor" {
    http.Error(w, "Forbidden", http.StatusForbidden)
    return
}
```

### Security Limitations (Educational Context)

⚠️ **This application has intentional security limitations**:

1. **Plain Text Passwords**: Stored unencrypted in database
   - **Why**: Simplifies learning, reduces dependencies
   - **Fix for Production**: Use `bcrypt.GenerateFromPassword()`

2. **No CSRF Protection**: Forms don't use CSRF tokens
   - **Why**: Reduces complexity for learners
   - **Fix for Production**: Add CSRF middleware

3. **In-Memory Sessions**: Lost on server restart
   - **Why**: Simple to understand
   - **Fix for Production**: Use Redis or database-backed sessions

4. **No Rate Limiting**: Unlimited login attempts
   - **Why**: Simpler code
   - **Fix for Production**: Add rate limiting middleware

5. **No HTTPS Enforcement**: HTTP allowed
   - **Why**: Local development friendly
   - **Fix for Production**: Force HTTPS in production

### What IS Secure

✅ **SQL Injection Prevention**: All queries use prepared statements
✅ **XSS Prevention**: User input is escaped via `html.EscapeString()`
✅ **Role-Based Access**: Handlers check user roles

---

## Key Architectural Decisions

### Why Monolithic Single-File?

**Decision**: Keep all code in `main.go`

**Rationale**:
- Educational tool - easy to understand complete flow
- Small enough application (~3700 lines)
- Reduces cognitive overhead of package navigation
- Inline code comments stay close to implementation

**Trade-off**: Won't scale to large teams or complex features

### Why SQLite Instead of PostgreSQL?

**Decision**: Use SQLite for database

**Rationale**:
- Zero configuration - just run the binary
- Perfect for single-server deployments
- File-based backup is simple
- Full SQL support for learning
- Handles up to ~100 concurrent users well

**Trade-off**: Not suitable for high-concurrency or distributed systems

### Why Server-Side Rendering?

**Decision**: Use Go templates instead of React/Vue

**Rationale**:
- Simpler mental model for learners
- No JavaScript build pipeline
- Easier to debug (view source shows actual HTML)
- Less code to understand
- Good enough for interactive features needed

**Trade-off**: Less interactive than SPA, more server load

### Why In-Memory Sessions?

**Decision**: Store sessions in Go map

**Rationale**:
- Simple to implement and understand
- No external dependencies
- Fast lookups
- Good for learning session concepts

**Trade-off**: Sessions lost on restart, won't work with multiple servers

### Why Cookie-Based Auth Instead of JWT?

**Decision**: Use session cookies

**Rationale**:
- Simpler to implement
- Server controls session lifecycle
- Easy to invalidate sessions
- Good for server-rendered apps

**Trade-off**: Requires server-side state, less suitable for microservices

---

## Deployment Architecture

### Development
```
┌──────────────────┐
│  Developer's     │
│    Computer      │
│                  │
│  go run main.go  │
│  Port: 8080      │
└──────────────────┘
```

### Production (Docker)
```
┌────────────────────────────────┐
│         Docker Host            │
│  ┌──────────────────────────┐  │
│  │   TaskFolio Container    │  │
│  │   Port: 8080 (internal)  │  │
│  │                          │  │
│  │   /data/taskfolio.db     │  │
│  │   (volume mount)         │  │
│  └──────────────────────────┘  │
│                                │
│  Port 8080 → 8080 (host)       │
└────────────────────────────────┘
```

### Key Deployment Characteristics:

- **Stateful**: Database file must persist
- **Single Instance**: SQLite limits concurrent writes
- **Volume Mounts**: Database stored outside container
- **Port Mapping**: Container port 8080 mapped to host

---

## Summary

TaskFolio demonstrates a **simple, monolithic web application architecture** ideal for learning:

✅ **Single binary** - easy to deploy
✅ **Self-contained** - includes database
✅ **Server-side rendered** - simple request/response model
✅ **Straightforward patterns** - handlers, middleware, templates
✅ **Educational security** - clear examples of vulnerabilities and fixes

This architecture is **perfect for**:
- Learning web development fundamentals
- Understanding request/response cycles
- Teaching database interactions
- Demonstrating authentication/authorization
- Small to medium deployments (< 1000 users)

This architecture is **not suitable for**:
- Large-scale production systems
- High-concurrency applications
- Distributed/microservice architectures
- Applications requiring real-time features

---

**Next Steps**: See [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md) to understand the code implementation in detail.
