# Tasking

**Task Management for ZipCode Students**

A multi-tenant web-based task management system for Zip Code Wilmington coding bootcamp students and instructors, built with Go and SQLite3.

> **Note**: While the repository is named TaskFolio, the application is branded as **Tasking** for end users.

---

## 🎓 Learning Resources

**New to this codebase?** Start here:

| Document | Purpose | Audience |
|----------|---------|----------|
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | Understand the system design and architecture | Developers, Architects |
| **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** | Step-by-step code explanation | Developers |
| **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** | Pedagogical guide with exercises | Students, Learners |
| **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** | Database design and schema | Database developers |
| **[GETTING_STARTED.md](./GETTING_STARTED.md)** | Developer onboarding | New contributors |
| **[DOCKER.md](./DOCKER.md)** | Deployment with Docker | DevOps, System admins |

---

## Features

### For Students
- Create, view, edit, and delete personal tasks
- Track task status (pending, in_progress, completed, blocked)
- View tasks assigned by instructors
- Organize tasks by type (labs, projects, reading, etc.)
- Group collaboration support
- Cohort-based organization

### For Instructors
- All student capabilities
- Assign tasks to individual students, cohorts, or all students
- Manage student groups
- Track student progress
- Suspend/unsuspend accounts
- **Task Templates Library**: Create, import, export, and manage reusable task templates
- Quickly assign task templates to multiple students with custom due dates

---

## Technology Stack

- **Language**: Go 1.24.7
- **Database**: SQLite3
- **Templates**: Go `html/template` (server-side rendering)
- **Session Management**: Cookie-based authentication
- **Deployment**: Docker with Docker Compose

**Why these technologies?** See [ARCHITECTURE.md](./ARCHITECTURE.md#technology-stack) for detailed rationale.

---

## Quick Start

### With Docker (Recommended)

```bash
# 1. Clone the repository
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio

# 2. Start with Docker Compose
docker compose up -d

# 3. Open browser
open http://localhost:8080
```

See [DOCKER.md](./DOCKER.md) for production deployment.

### Without Docker

**Prerequisites**: Go 1.16+, SQLite3

```bash
# 1. Clone and enter directory
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio

# 2. Install dependencies
go mod download

# 3. Run the application
go run main.go

# 4. Open browser
open http://localhost:8080
```

---

## Demo Accounts

The application comes pre-populated with demo accounts for testing:

| Role | Username | Password |
|------|----------|----------|
| Instructor | `instructor1` | `password123` |
| Student | `student1`, `student2`, `student3` | `password123` |

---

## Application Structure

```
TaskFolio/
├── main.go              # Application code (~3700 lines)
├── main_test.go         # Tests (~1500 lines)
├── templates/           # HTML templates
│   ├── login.html
│   ├── dashboard.html
│   └── ...
├── taskfolio.db         # SQLite database (created at runtime)
├── Dockerfile           # Container definition
├── docker-compose.yml   # Docker deployment
└── docs/                # Documentation
    ├── planning/        # Original planning documents
    └── ...
```

**Why single-file architecture?** This is an intentional educational choice. See [ARCHITECTURE.md](./ARCHITECTURE.md#single-file-architecture) for details.

---

## Understanding the Application

### For Students and Learners

TaskFolio is designed to be a learning tool. Start with:

1. **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** - Structured learning path with exercises
2. **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Understand how it's built
3. **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** - Line-by-line code explanation

### For Developers

Contributing or modifying? Read:

1. **[GETTING_STARTED.md](./GETTING_STARTED.md)** - Development setup
2. **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** - Understanding the codebase
3. **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** - Data model

### For Instructors

Using this for teaching? Check out:

1. **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** - Student exercises and projects
2. **[ARCHITECTURE.md](./ARCHITECTURE.md)** - Architectural concepts to teach
3. **Demo accounts** - Pre-configured for classroom use

---

## Core Concepts Demonstrated

This application teaches:

- ✅ **Web Development**: HTTP request/response, server-side rendering, sessions
- ✅ **Database**: SQL, CRUD operations, migrations, relationships
- ✅ **Security**: Authentication, authorization, SQL injection prevention, XSS protection
- ✅ **Go Programming**: Standard library, testing, error handling
- ✅ **Software Engineering**: Testing, documentation, deployment

See [LEARNING_GUIDE.md](./LEARNING_GUIDE.md#key-concepts-to-master) for detailed explanations.

---

## Database Schema
See [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) for complete schema documentation, including:
- Users table (authentication and roles)
- Tasks table (task management with status tracking)
- Groups table (team collaboration)
- Group members (many-to-many relationships)
- Task templates (reusable task library for instructors)

---

## Using Task Templates (Instructor Feature)

Task templates allow instructors to create a library of reusable tasks that can be quickly assigned to students.

### Creating Templates

1. Navigate to "Task Templates" from the home page
2. Click "Add Template" to create a new template
3. Or click "Import CSV" to bulk import templates

**CSV Import Format:**
```csv
Title, Description, TaskType, ResourceURL
Lab 1, Complete Java basics, individual-programming-lab, https://github.com/...
Reading Chapter 5, Design patterns, reading, https://docs.google.com/...
```

### Assigning Templates to Students

1. Go to "Manage Students"
2. Use the search/filter to find students
3. Select students using checkboxes
4. Click "Assign Template"
5. Search for and select a template
6. Set a due date (optional)
7. Click "Assign to Students"

The template will create individual tasks for each selected student with the template's content.

### Managing Templates

- **Edit**: Update template details from the Task Templates page
- **Delete**: Remove templates that are no longer needed
- **Export**: Select templates and export to CSV for backup or sharing

---

## Configuration

Environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `DATABASE_PATH` | `./taskfolio.db` | Path to SQLite database file |

---

## Development

### Running Tests

```bash
# Run all tests
go test -v

# Run with coverage
go test -cover

# Run specific test
go test -run TestTaskCreation -v
```

### Building

```bash
# Development
go run main.go

# Production build
go build -o tasking
./tasking
```

See [GETTING_STARTED.md](./GETTING_STARTED.md) for detailed development setup.

---

## Security Notes

⚠️ **Educational/Demo Application**: This application has intentional security limitations for learning purposes:

| Limitation | Status | Production Fix |
|------------|--------|----------------|
| Plain text passwords | ⚠️ Intentional | Use `bcrypt` hashing |
| No CSRF protection | ⚠️ Intentional | Add CSRF tokens |
| In-memory sessions | ⚠️ Intentional | Use Redis/database sessions |
| No rate limiting | ⚠️ Intentional | Add rate limiting middleware |

**What IS secure**: ✅ SQL injection prevention via prepared statements, ✅ XSS prevention via template escaping

See [ARCHITECTURE.md](./ARCHITECTURE.md#security-model) for detailed security discussion.

---

## Deployment

See [DOCKER.md](./DOCKER.md) for production deployment including:
- Docker setup with persistent database
- HTTPS configuration
- Automated backups
- Resource limits

---

## Testing

Run tests:
```bash
go test -v        # All tests with output
go test -cover    # With coverage report
```

See [TESTING.md](./TESTING.md) for Docker-based testing procedures.

---

## Documentation Index

### Learning Resources
- **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** - Pedagogical guide with exercises for students
- **[ARCHITECTURE.md](./ARCHITECTURE.md)** - System architecture and design decisions
- **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** - Line-by-line code explanation

### Development Guides
- **[GETTING_STARTED.md](./GETTING_STARTED.md)** - Developer onboarding and setup
- **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** - Complete database schema documentation
- **[TECHNICAL_REQUIREMENTS.md](./TECHNICAL_REQUIREMENTS.md)** - System requirements

### Deployment & Operations
- **[DOCKER.md](./DOCKER.md)** - Docker deployment guide
- **[TESTING.md](./TESTING.md)** - Testing procedures
- **[DEPLOY.md](./DEPLOY.md)** - Deployment instructions

### Planning Documents
- **[docs/planning/](./docs/planning/)** - Original project planning documents

---

## Contributing

Contributions welcome! See [GETTING_STARTED.md](./GETTING_STARTED.md) for development setup.

For questions or issues, please open an issue on GitHub.

---

## License

MIT License

---

## Acknowledgments

- Built for Zip Code Wilmington coding bootcamp
- Designed as an educational tool for learning web development
- Demonstrates real-world application architecture in a digestible format

---

**Project Status**: ✅ Functional application with comprehensive educational documentation

**Next Steps for Learners**: Start with [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)
