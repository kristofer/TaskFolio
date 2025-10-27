# Copilot Instructions for TaskFolio

## Project Overview

TaskFolio (branded as "Tasking" for end users) is a multi-tenant web-based task management system for Zip Code Wilmington coding bootcamp students and instructors. The application is built with Go and SQLite3, using server-side HTML templates for the UI.

### Key Features
- **For Students**: Create, view, edit, and delete personal tasks; track task status; view instructor-assigned tasks
- **For Instructors**: All student capabilities plus assign tasks to individuals, cohorts, or all students
- **Task Management**: Support for multiple task types (group programming projects, individual labs, reading assignments, documents, resumes, professional development)
- **Cohort-based Organization**: Students and tasks can be organized by cohorts (e.g., "Java11.2", "Data6.2")
- **Status Tracking**: Tasks can be pending, in_progress, completed, or blocked with automatic date tracking
- **Account Management**: User suspension capability and role-based access control

## ⚠️ Important Security Notice

**This is a demo/educational application.** It contains known security limitations that make it unsuitable for production use without significant security enhancements:

- **Plain text passwords** (should use bcrypt hashing)
- **No CSRF protection**
- **Basic in-memory session management** (no expiration)
- **No rate limiting**

These limitations are intentional for learning purposes but must be addressed before any production deployment. See the [Security Considerations](#security-considerations) section for details.

## Technology Stack

- **Language**: Go 1.24.7
- **Database**: SQLite3 (`github.com/mattn/go-sqlite3`)
- **Templates**: Go `html/template` package for server-side rendering
- **Session Management**: Cookie-based authentication
- **Deployment**: Docker with Docker Compose support
- **Testing**: Go standard testing package

## Architecture and Design

### Application Structure
- **Single-file architecture**: The main application logic is in `main.go` (this is intentional for simplicity)
- **Templates**: HTML templates are stored in the `templates/` directory
- **Database**: SQLite database file (default: `./taskfolio.db`, configurable via `DATABASE_PATH` env var)
- **Port**: Defaults to 8080, configurable via `PORT` env var

### Database Schema

#### Users Table
- `id`: Auto-increment primary key
- `username`: Unique username
- `password`: Plain text password ⚠️ **SECURITY WARNING: This is for demo/educational purposes only. NOT suitable for production use. Should use bcrypt or similar hashing.**
- `role`: Either "student" or "instructor"
- `cohort_name`: Optional cohort assignment (e.g., "Java11.2")
- `is_suspended`: Boolean for account suspension

#### Tasks Table
- `id`: Auto-increment primary key
- `title`: Task title
- `description`: Task description
- `due_date`: Due date string
- `status`: One of 'pending', 'in_progress', 'completed', 'blocked'
- `owner_id`: User who owns the task (foreign key to users)
- `created_by`: User who created the task (foreign key to users)
- `assigned_to`: Comma-separated usernames for group assignments
- `created_at`: Creation timestamp
- `task_type`: One of 'group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', or empty string
- `resource_url`: URL pointing to task resources
- `cohort_id`: Cohort name for cohort-based assignments
- `in_progress_date`: Timestamp when task moved to in_progress
- `done_date`: Timestamp when task moved to completed

### Key Patterns

#### Database Migration
The application includes automatic database migration logic in `initDB()` that:
1. Creates tables if they don't exist
2. Adds new columns to existing tables if needed
3. Migrates data when schema changes (e.g., updating CHECK constraints)

Always follow this pattern when adding new fields to maintain backward compatibility.

#### Session Management
- Sessions are stored in-memory using a simple map
- Session cookies are used for authentication
- The `requireAuth()` middleware checks for valid sessions
- Session data includes user ID, username, and role

#### HTML Template Rendering
- Templates are loaded once at startup via `template.ParseGlob("templates/*.html")`
- Templates receive data through typed structs passed to `templates.ExecuteTemplate()`
- Use `html.EscapeString()` for user input to prevent XSS

## Code Style and Conventions

### Go Conventions
- Follow standard Go formatting (use `gofmt`)
- Use meaningful variable names
- Keep functions focused on a single responsibility
- Use early returns to reduce nesting
- Prefer explicit error handling over panics

### SQL Best Practices
- **ALWAYS use prepared statements** with parameterized queries (already implemented)
- Never concatenate user input into SQL strings
- Use transactions for multiple related operations
- Check for `sql.ErrNoRows` when expecting single results

### Error Handling
- Log errors with context using `log.Printf()`
- Return HTTP 500 for server errors
- Return HTTP 400 for client errors (bad input)
- Return HTTP 401 for authentication failures
- Return HTTP 403 for authorization failures

### Security Considerations
- **Passwords are currently plain text** - this is a known limitation for demo purposes
- SQL injection prevention via prepared statements is implemented
- XSS prevention via `html.EscapeString()` is implemented
- CSRF protection is NOT implemented (known limitation)
- Session management is basic (in-memory, no expiration)

**When making changes:**
- Never remove the prepared statement pattern
- Always escape user input before rendering in templates
- Maintain role-based access control checks
- Don't introduce new security vulnerabilities

## Testing Practices

### Test Organization
All tests are in `main_test.go` and use the Go standard testing package.

### Test Database
Each test creates a temporary database using `setupTestDB()` which:
1. Creates an in-memory SQLite database (`:memory:`)
2. Initializes the schema
3. Returns the database connection for use in tests

Always clean up with `defer db.Close()` in tests.

### Test Patterns
- Use table-driven tests when testing multiple scenarios
- Test both success and failure cases
- Verify database state after operations
- Test role-based access control (student vs instructor permissions)
- Test edge cases (empty strings, nil values, non-existent IDs)

### Running Tests
```bash
go test -v           # Run all tests with verbose output
go test -run TestName  # Run specific test
```

### Test Coverage Areas
- Database initialization and migration
- User creation and management
- Task creation, updates, and deletion
- Role-based permissions
- Cohort-based assignments
- Task type validation
- Status transitions and date tracking

## Building and Running

### Development
```bash
go run main.go
# Access at http://localhost:8080
```

### Production Build
```bash
go build -o TaskFolio
./TaskFolio
```

### Docker
```bash
docker compose up -d          # Start with Docker
docker compose logs -f tasking # View logs
docker compose down            # Stop
```

### Environment Variables
- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: Path to SQLite database file (default: ./taskfolio.db)

## Common Development Tasks

### Adding a New Field to Users Table
1. Update the `User` struct in `main.go`
2. Add migration logic in `initDB()` to add the column if it doesn't exist
3. Update relevant queries to include the new field
4. Update templates if the field needs to be displayed
5. Add tests for the new field

### Adding a New Task Type
1. Update the `task_type` CHECK constraint in the tasks table creation
2. Add migration logic to update existing CHECK constraint
3. Update any dropdown/select elements in templates
4. Add test coverage for the new type

### Adding a New API Endpoint
1. Create the handler function following the pattern: `func handlerName(w http.ResponseWriter, r *http.Request)`
2. Use `requireAuth()` middleware if authentication is needed
3. Extract user from session: `session := sessions[sessionID]`
4. Validate inputs and permissions
5. Perform database operations with prepared statements
6. Render template or redirect as appropriate
7. Add tests for the endpoint

### Modifying Templates
- Templates are in `templates/` directory
- Main layout template typically includes header, navigation, and footer
- Use `{{.FieldName}}` to display data
- Use `{{if .Condition}}...{{end}}` for conditionals
- Use `{{range .Items}}...{{end}}` for loops
- Remember that template data is passed as a struct from the handler

## Important Notes

### Demo Data
The application includes demo accounts:
- **Instructor**: username=`instructor1`, password=`password123`
- **Students**: username=`student1`/`student2`/`student3`, password=`password123`

This data is inserted in `seedDemoData()` if not present.

### Branding
- Repository name: TaskFolio
- User-facing brand: Tasking
- Ensure user-facing elements use "Tasking" branding

### Performance Considerations
- SQLite is suitable for development and small deployments
- The application uses a single database connection pool
- Indexes are created on foreign key columns for query performance
- Session storage is in-memory (not suitable for multi-instance deployment)

## Anti-Patterns to Avoid

❌ **Don't**:
- Concatenate user input into SQL queries
- Store sensitive data in plain text (except in demo/learning context)
- Remove or bypass role-based access checks
- Create SQL queries without using prepared statements
- Add dependencies without careful consideration (keep it simple)
- Break the single-file architecture without good reason
- Remove existing tests when making changes

✅ **Do**:
- Use prepared statements for all SQL queries
- Check user roles and permissions before operations
- Maintain backward compatibility with database migrations
- Write tests for new functionality
- Follow the existing code patterns and style
- Keep the application simple and maintainable
- Document security limitations for demo purposes

## Documentation

Key documentation files:
- `README.md`: User-facing documentation and quick start
- `API_SPECIFICATION.md`: Planned API specifications (future expansion)
- `DATABASE_SCHEMA.md`: Detailed database design documentation
- `PROJECT_STRUCTURE.md`: Recommended structure for future expansion
- `TECHNICAL_REQUIREMENTS.md`: Requirements and specifications
- `TESTING.md`: Docker testing procedures
- `DOCKER.md`: Docker deployment guide
- `GETTING_STARTED.md`: Development onboarding guide

When making significant changes, update relevant documentation files.

## Future Considerations

The current implementation is a working prototype. Future enhancements being considered:
- Password hashing with bcrypt
- CSRF protection
- Proper session expiration and management
- PostgreSQL support for larger deployments
- RESTful API endpoints
- Frontend framework integration (React/Vue)
- Email notifications
- Real-time updates via WebSockets

When implementing new features, consider how they fit into this roadmap and maintain backward compatibility where possible.
