# TaskFolio Learning Guide

## Introduction

This guide is designed for adult learners who want to understand web application development through studying TaskFolio. Whether you're new to web development or looking to understand how a complete application works, this guide will help you learn systematically.

## Table of Contents

1. [Learning Objectives](#learning-objectives)
2. [Prerequisites](#prerequisites)
3. [Learning Path](#learning-path)
4. [Hands-On Exercises](#hands-on-exercises)
5. [Key Concepts to Master](#key-concepts-to-master)
6. [Common Challenges](#common-challenges)
7. [Projects to Build](#projects-to-build)
8. [Additional Resources](#additional-resources)

---

## Learning Objectives

By studying TaskFolio, you will learn:

### Web Development Fundamentals
- ✅ HTTP request/response cycle
- ✅ Server-side rendering vs. client-side rendering
- ✅ Session-based authentication
- ✅ Form handling and validation
- ✅ Database-backed applications

### Go Programming
- ✅ Go project structure and organization
- ✅ HTTP server creation with standard library
- ✅ Database access with `database/sql`
- ✅ Template rendering with `html/template`
- ✅ Error handling patterns
- ✅ Testing with Go's testing package

### Database Concepts
- ✅ SQL database design (tables, relationships, constraints)
- ✅ CRUD operations (Create, Read, Update, Delete)
- ✅ Prepared statements and SQL injection prevention
- ✅ Database migrations
- ✅ Foreign keys and referential integrity

### Security Practices
- ✅ Authentication vs. Authorization
- ✅ SQL injection prevention
- ✅ Cross-Site Scripting (XSS) prevention
- ✅ Session management
- ✅ Role-based access control (RBAC)

### Software Engineering
- ✅ Testing strategies
- ✅ Code organization
- ✅ Documentation practices
- ✅ Deployment with Docker
- ✅ Environment configuration

---

## Prerequisites

### Minimum Requirements

**Programming Knowledge**:
- Basic understanding of any programming language
- Familiarity with functions, variables, and control flow
- Comfort with command-line tools

**System Setup**:
- Go 1.16+ installed ([go.dev/doc/install](https://go.dev/doc/install))
- SQLite3 installed (usually included with Go SQLite driver)
- Text editor or IDE (VS Code, GoLand, or similar)
- Git for version control

### Recommended Background

Having some knowledge in these areas helps but is not required:
- HTML/CSS basics
- HTTP and web browsers
- SQL fundamentals
- Command-line usage

### Time Commitment

- **Quick Overview**: 2-4 hours (understand architecture and run the app)
- **Deep Understanding**: 20-30 hours (read all code, understand patterns)
- **Mastery**: 40-60 hours (complete exercises, build extensions)

---

## Learning Path

### Phase 1: Understanding the Big Picture (2-4 hours)

**Goal**: Understand what the app does and how it's organized.

1. **Read Documentation** (1 hour)
   - Start with [README.md](./README.md) - Overview and features
   - Read [ARCHITECTURE.md](./ARCHITECTURE.md) - System design
   - Skim [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)

2. **Run the Application** (30 minutes)
   ```bash
   # Clone repository
   git clone https://github.com/kristofer/TaskFolio.git
   cd TaskFolio

   # Run the app
   go run main.go

   # Open browser to http://localhost:8080
   ```

3. **Explore as a User** (1 hour)
   - Log in as student (username: `student1`, password: `password123`)
   - Create a few tasks
   - Edit task status (pending → in_progress → completed)
   - Log out and log in as instructor (username: `instructor1`)
   - Explore instructor features
   - Take notes on what you observe

4. **Map the User Journey** (30 minutes)
   - Draw a diagram of the user flows
   - Login → Dashboard → Create Task → Edit Task → Logout
   - Identify which pages you visited
   - List questions you have about how it works

**Checkpoint**: Can you describe what TaskFolio does and who uses it?

---

### Phase 2: Code Reading (8-12 hours)

**Goal**: Read and understand the codebase.

#### Week 1: Data and Database

**Day 1-2: Data Structures** (2 hours)
- Open `main.go`
- Find the type definitions (lines ~25-72)
- Study `User`, `Task`, and `Group` structs
- Understand what data each entity holds

**Exercise**:
```go
// Question: What's the difference between owner_id and created_by in Task?
// Answer: ___________
```

**Day 3-4: Database Schema** (3 hours)
- Read `initDB()` function
- Understand each table's CREATE TABLE statement
- Draw an Entity-Relationship Diagram on paper
- Identify foreign key relationships

**Exercise**:
```sql
-- Write a query to find all tasks owned by user ID 1
-- Your answer:

-- Write a query to find all tasks created by instructors
-- Hint: Join users table
-- Your answer:
```

**Day 5: Database Migration** (2 hours)
- Study how new columns are added to existing tables
- Understand the migration pattern
- Think: Why is this important for production apps?

**Checkpoint**: Can you explain the database schema to someone else?

---

#### Week 2: HTTP and Handlers

**Day 1: HTTP Routing** (2 hours)
- Find the `main()` function
- List all HTTP routes registered
- For each route, identify:
  - URL path
  - Handler function
  - Protected or public?

**Exercise**:
```
Route: /tasks/add
Handler: ___________
Protected? (yes/no): ___________
What does it do?: ___________
```

**Day 2-3: Handler Pattern** (3 hours)
- Pick one handler (start with `loginHandler`)
- Read through it line by line
- Identify:
  - GET vs POST handling
  - Form data extraction
  - Database queries
  - Template rendering
  - Redirects

**Exercise**:
```go
// Trace the login flow:
// 1. User submits form
// 2. ___________
// 3. ___________
// 4. ___________
// 5. User sees dashboard
```

**Day 4: Session Management** (2 hours)
- Study the `sessions` global variable
- Read `requireAuth()` middleware
- Understand cookie-based sessions
- Trace: Cookie creation → Storage → Validation → Deletion

**Exercise**:
```
Question: What happens to sessions when server restarts?
Answer: ___________

Question: Is this suitable for production? Why or why not?
Answer: ___________
```

**Day 5: Template Rendering** (2 hours)
- Open a template file (e.g., `templates/dashboard.html`)
- Understand template syntax: `{{.FieldName}}`, `{{range}}`, `{{if}}`
- Find where data is passed to template
- Modify a template and see the change

**Checkpoint**: Can you trace a complete HTTP request from browser to response?

---

### Phase 3: Hands-On Modification (10-15 hours)

**Goal**: Make small changes to understand how things work.

#### Exercise 1: Add a Task Field (3 hours)

Add a "priority" field to tasks:

**Steps**:
1. Add `Priority string` to `Task` struct
2. Update database schema to add `priority` column
3. Update `addTaskHandler` to accept priority from form
4. Update `editTaskHandler` to allow editing priority
5. Update templates to show priority field
6. Test your changes

**Learning Focus**: Full-stack feature addition

---

#### Exercise 2: Create a New Page (2 hours)

Create a "Completed Tasks" page showing only completed tasks:

**Steps**:
1. Create new handler `completedTasksHandler`
2. Query database for tasks where `status = 'completed'`
3. Create new template `completed_tasks.html`
4. Add route in `main()`
5. Add link in navigation

**Learning Focus**: MVC pattern (Model-View-Controller)

---

#### Exercise 3: Add Validation (2 hours)

Improve form validation:

**Steps**:
1. Ensure task title is between 5-200 characters
2. Validate due date is in the future
3. Show user-friendly error messages
4. Test edge cases

**Learning Focus**: Input validation and user experience

---

#### Exercise 4: Improve Security (3 hours)

Implement password hashing:

**Steps**:
1. Import `golang.org/x/crypto/bcrypt`
2. Update user creation to hash passwords
3. Update login to compare hashed passwords
4. Migrate existing users (or reset demo data)
5. Test login still works

**Learning Focus**: Security best practices

---

#### Exercise 5: Write Tests (4 hours)

Add tests for a new feature:

**Steps**:
1. Read existing tests in `main_test.go`
2. Write test for task creation
3. Write test for task validation
4. Write test for unauthorized access
5. Run tests and ensure they pass

**Learning Focus**: Test-driven development

---

### Phase 4: Deep Dive Topics (8-12 hours)

Choose areas to explore deeper:

#### Topic 1: Database Design (3 hours)

**Study**:
- Normalization (why separate tables?)
- Indexes (how do they speed up queries?)
- Transactions (why are they important?)

**Exercise**: Design a database for a different domain (e.g., library system)

---

#### Topic 2: HTTP Deep Dive (3 hours)

**Study**:
- HTTP methods (GET, POST, PUT, DELETE)
- HTTP status codes (200, 404, 500, etc.)
- Headers and cookies
- Request/Response lifecycle

**Exercise**: Use browser dev tools to inspect HTTP traffic

---

#### Topic 3: Security (4 hours)

**Study**:
- SQL injection (how does TaskFolio prevent it?)
- XSS (Cross-Site Scripting)
- CSRF (Cross-Site Request Forgery)
- Session hijacking

**Exercise**: Try to exploit TaskFolio (in a test environment!) and document vulnerabilities

---

#### Topic 4: Testing (3 hours)

**Study**:
- Unit tests vs integration tests
- Test-driven development (TDD)
- Code coverage
- Mocking and test doubles

**Exercise**: Achieve 80% code coverage on a module

---

## Hands-On Exercises

### Beginner Exercises

#### Exercise 1: Customize the UI
**Goal**: Learn template rendering

- Change the color scheme in templates
- Add your name to the footer
- Change "Tasking" branding to your own

**Time**: 30 minutes

---

#### Exercise 2: Add a Field
**Goal**: Understand data flow

- Add "estimated_hours" to tasks
- Show it on dashboard
- Allow editing

**Time**: 1-2 hours

---

#### Exercise 3: Filter Tasks
**Goal**: Practice SQL queries

- Add a page showing only pending tasks
- Add a page showing only your completed tasks
- Add a filter dropdown

**Time**: 2 hours

---

### Intermediate Exercises

#### Exercise 4: Task Categories
**Goal**: Database relationships

- Create a "categories" table
- Link tasks to categories
- Show category in task list
- Allow filtering by category

**Time**: 4-5 hours

---

#### Exercise 5: User Profile
**Goal**: CRUD operations

- Add profile page showing user info
- Allow users to change password
- Allow users to update their cohort
- Instructors can edit any user

**Time**: 3-4 hours

---

#### Exercise 6: Task Search
**Goal**: Advanced SQL

- Add search bar
- Search by title or description
- Use SQL LIKE operator
- Highlight search terms in results

**Time**: 3 hours

---

### Advanced Exercises

#### Exercise 7: Task Comments
**Goal**: Complex relationships

- Create comments table
- Allow commenting on tasks
- Show comment thread
- Only owner can delete comments

**Time**: 6-8 hours

---

#### Exercise 8: Email Notifications
**Goal**: External integration

- Send email when task assigned
- Use SMTP library
- Create email templates
- Make email sending async

**Time**: 8-10 hours

---

#### Exercise 9: API Endpoints
**Goal**: REST API design

- Create JSON API for tasks
- Support GET, POST, PUT, DELETE
- Use proper HTTP status codes
- Add API authentication

**Time**: 10-12 hours

---

## Key Concepts to Master

### 1. HTTP Request/Response Cycle

```
Client (Browser)          Server (Go App)           Database (SQLite)
      │                         │                         │
      │  1. GET /              │                         │
      ├────────────────────────>│                         │
      │                         │  2. Query tasks         │
      │                         ├────────────────────────>│
      │                         │  3. Return rows         │
      │                         │<────────────────────────┤
      │  4. HTML Response       │                         │
      │<────────────────────────┤                         │
      │                         │                         │
```

**Key Questions**:
- What is HTTP?
- What's the difference between GET and POST?
- How does the browser know where to send requests?
- What is a cookie and how does it work?

---

### 2. MVC Pattern (Model-View-Controller)

```
View (Templates)          Controller (Handlers)      Model (Database)
      │                         │                         │
      │  User clicks button     │                         │
      ├────────────────────────>│                         │
      │                         │  Query/Update data      │
      │                         ├────────────────────────>│
      │                         │  Return data            │
      │                         │<────────────────────────┤
      │  Render with data       │                         │
      │<────────────────────────┤                         │
```

**In TaskFolio**:
- **Model**: `User`, `Task` structs + database
- **View**: HTML templates in `templates/`
- **Controller**: Handler functions in `main.go`

---

### 3. Database CRUD Operations

**Create**:
```go
db.Exec("INSERT INTO tasks (title, owner_id) VALUES (?, ?)", title, userID)
```

**Read**:
```go
db.Query("SELECT * FROM tasks WHERE owner_id = ?", userID)
```

**Update**:
```go
db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, taskID)
```

**Delete**:
```go
db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
```

**Master These**:
- When to use `Exec()` vs `Query()` vs `QueryRow()`
- How to handle errors
- Why prepared statements are important
- How to iterate over multiple rows

---

### 4. Authentication vs Authorization

**Authentication**: *Who are you?*
```go
// Check username and password
user := getUserByCredentials(username, password)
if user != nil {
    createSession(user)
}
```

**Authorization**: *What can you do?*
```go
// Check if user has permission
if session.Role != "instructor" {
    http.Error(w, "Forbidden", http.StatusForbidden)
}
```

**In TaskFolio**:
- Authentication: Login with username/password
- Authorization: Students can't assign tasks to others

---

### 5. SQL Injection Prevention

**Vulnerable Code** (NEVER DO THIS):
```go
query := "SELECT * FROM users WHERE username = '" + username + "'"
db.Query(query)
```

**Attack**:
```
username = "admin' OR '1'='1"
Result: SELECT * FROM users WHERE username = 'admin' OR '1'='1'
        Returns all users!
```

**Safe Code** (ALWAYS DO THIS):
```go
db.Query("SELECT * FROM users WHERE username = ?", username)
```

The database driver escapes special characters, preventing injection.

---

### 6. Session Management

**Session Lifecycle**:

1. **Login**: Create session
   ```go
   sessionID := generateRandomString()
   sessions[sessionID] = SessionData{UserID: user.ID}
   ```

2. **Send Cookie**: Browser stores session ID
   ```go
   http.SetCookie(w, &http.Cookie{Name: "session_id", Value: sessionID})
   ```

3. **Validate**: Check session on each request
   ```go
   cookie, _ := r.Cookie("session_id")
   session, exists := sessions[cookie.Value]
   ```

4. **Logout**: Delete session
   ```go
   delete(sessions, sessionID)
   ```

---

## Common Challenges

### Challenge 1: "Where is the data coming from?"

**Symptom**: You see `{{.Username}}` in template but don't know where it comes from.

**Solution**:
1. Find which handler renders this template
2. Look for `ExecuteTemplate` call
3. Check the data parameter passed to it
4. Trace back where that data was created

---

### Challenge 2: "My changes don't appear"

**Symptom**: You edited code but don't see changes.

**Solutions**:
- **Go code**: Restart server (`Ctrl+C` then `go run main.go`)
- **Templates**: Templates are loaded at startup, must restart
- **Database**: Check if you're editing right database file
- **Browser cache**: Hard refresh (`Ctrl+Shift+R`)

---

### Challenge 3: "Database locked error"

**Symptom**: Error about database being locked.

**Solutions**:
- Close any SQLite browser tools
- Check for forgotten `defer rows.Close()`
- Ensure only one server instance is running

---

### Challenge 4: "Session not working"

**Symptom**: Logged in but immediately logged out.

**Solutions**:
- Check session is created: `log.Printf("Session: %v", sessions)`
- Verify cookie is sent: Check browser dev tools → Application → Cookies
- Ensure middleware is applied: Route should use `requireAuth()`

---

### Challenge 5: "SQL syntax error"

**Symptom**: Database query fails with syntax error.

**Solutions**:
- Print the query: `log.Printf("Query: %s", query)`
- Test in SQLite shell: `sqlite3 taskfolio.db "SELECT ..."`
- Check column names match schema
- Verify placeholders (`?`) match number of arguments

---

## Projects to Build

Apply your learning by building similar applications:

### Project 1: Blog System (Beginner)
**Features**:
- User registration and login
- Create, edit, delete blog posts
- View all posts
- Basic styling

**Skills**: CRUD operations, authentication, templates

---

### Project 2: Todo List with Sharing (Intermediate)
**Features**:
- Personal todo lists
- Share lists with other users
- Real-time updates
- Categories and tags

**Skills**: Relationships, sharing logic, advanced queries

---

### Project 3: Course Management System (Advanced)
**Features**:
- Students enroll in courses
- Instructors create courses and assignments
- Grading system
- File uploads for submissions
- Email notifications

**Skills**: Complex relationships, file handling, email integration

---

## Additional Resources

### Go Language
- [Official Go Tour](https://go.dev/tour/) - Interactive Go tutorial
- [Go by Example](https://gobyexample.com/) - Code examples
- [Effective Go](https://go.dev/doc/effective_go) - Best practices

### Web Development
- [MDN Web Docs](https://developer.mozilla.org/) - HTML, CSS, JavaScript reference
- [HTTP Status Codes](https://httpstatuses.com/) - What each code means
- [REST API Tutorial](https://restfulapi.net/) - API design principles

### Database
- [SQLite Tutorial](https://www.sqlitetutorial.net/) - SQL fundamentals
- [Database Design](https://www.vertabelo.com/blog/database-design/) - Schema design
- [SQL Style Guide](https://www.sqlstyle.guide/) - Writing clean SQL

### Security
- [OWASP Top 10](https://owasp.org/www-project-top-ten/) - Common vulnerabilities
- [Web Security Academy](https://portswigger.net/web-security) - Interactive lessons
- [Go Security](https://go.dev/security/) - Go-specific security practices

### Testing
- [Go Testing](https://go.dev/doc/tutorial/add-a-test) - Official testing guide
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests) - Testing pattern
- [Test Coverage](https://go.dev/blog/cover) - Measuring coverage

---

## Learning Plan Template

Track your progress:

```
Week 1: Understanding Architecture
[ ] Read all documentation
[ ] Run application successfully
[ ] Explore as user
[ ] Draw architecture diagram

Week 2: Code Reading - Database
[ ] Understand User, Task, Group structs
[ ] Read database schema
[ ] Draw ER diagram
[ ] Write sample SQL queries

Week 3: Code Reading - HTTP
[ ] List all routes
[ ] Read login handler
[ ] Trace request flow
[ ] Understand sessions

Week 4: First Modifications
[ ] Add a task field
[ ] Create a new page
[ ] Add validation
[ ] Write tests

Week 5-6: Deep Dive
[ ] Choose a topic (database/HTTP/security/testing)
[ ] Complete advanced exercises
[ ] Build a small project

Week 7-8: Major Project
[ ] Design a related application
[ ] Implement core features
[ ] Write tests
[ ] Deploy with Docker
```

---

## Conclusion

TaskFolio is a complete, real-world web application that demonstrates modern web development practices. By studying it systematically, you'll gain practical skills applicable to any web development project.

**Remember**:
- **Start small**: Don't try to understand everything at once
- **Hands-on**: Actually run and modify the code
- **Ask questions**: Why is it done this way? What are the trade-offs?
- **Build things**: The best way to learn is by creating
- **Be patient**: Understanding takes time; it's okay to revisit concepts

**Next Steps**:
1. Start with Phase 1 of the Learning Path
2. Work through exercises at your own pace
3. Build a project that interests you
4. Share what you've learned with others

Good luck on your learning journey! 🚀
