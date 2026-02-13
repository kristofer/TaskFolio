# TaskFolio - Getting Started Guide

Welcome to TaskFolio! This guide will help you get started with the project, whether you're a new developer, student, or contributor.

## Table of Contents

1. [Understanding the Project](#understanding-the-project)
2. [Documentation Overview](#documentation-overview)
3. [Quick Start](#quick-start)
4. [Development Setup](#development-setup)
5. [Learning Path](#learning-path)
6. [Contributing](#contributing)

---

## Understanding the Project

### What is TaskFolio?

TaskFolio (branded as "Tasking") is a web-based task management system built for Zip Code Wilmington coding bootcamp. It's designed as both a **functional application** and an **educational tool** for learning web development.

**Current Features**:
- User authentication with role-based access (students and instructors)
- Task creation, editing, deletion, and status tracking
- Cohort-based organization
- Group collaboration support
- Instructor task assignment capabilities

### Project Goals

- **Educational Value**: Demonstrate real-world web development patterns in digestible format
- **Simplicity**: Single-file architecture for easier understanding
- **Completeness**: Full working application from database to UI
- **Best Practices**: Show proper security, testing, and documentation

---

## Documentation Overview

TaskFolio has comprehensive documentation for different audiences:

### 📚 For Learners and Students

| Document | Purpose | Time Required |
|----------|---------|---------------|
| **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** | Structured learning path with exercises | 40-60 hours total |
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | Understand how the app is designed | 2-3 hours |
| **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** | Line-by-line code explanation | 8-12 hours |

**Start here**: [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)

### 👨‍💻 For Developers

| Document | Purpose | When to Use |
|----------|---------|-------------|
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | System design and decisions | Before making changes |
| **[CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)** | Understanding the code | Daily reference |
| **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** | Database structure | Working with data |
| **[TECHNICAL_REQUIREMENTS.md](./TECHNICAL_REQUIREMENTS.md)** | What the system does | Feature planning |

**Start here**: [ARCHITECTURE.md](./ARCHITECTURE.md)

### 🏫 For Instructors

| Document | Purpose | When to Use |
|----------|---------|-------------|
| **[LEARNING_GUIDE.md](./LEARNING_GUIDE.md)** | Exercises and projects for students | Teaching web dev |
| **[ARCHITECTURE.md](./ARCHITECTURE.md)** | Concepts to teach | Lesson planning |
| Demo accounts | Pre-configured users | Classroom demos |

**Start here**: [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)

### 🗂️ Planning Documents

Original design documents are in **[docs/planning/](./docs/planning/)**:
- Implementation plan for future features
- Sprint planning templates
- API specifications for RESTful architecture
- Multi-package project structure recommendations

These represent the **vision** for future enhancements, not the current implementation.

---

## Quick Start

### Running the Application

**Prerequisites**: Go 1.16+, SQLite3

```bash
# Clone repository
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio

# Install dependencies
go mod download

# Run the application
go run main.go

# Open browser
open http://localhost:8080
```

### Demo Accounts

| Username | Password | Role |
|----------|----------|------|
| `instructor1` | `password123` | Instructor |
| `student1` | `password123` | Student |
| `student2` | `password123` | Student |
| `student3` | `password123` | Student |

### With Docker

```bash
docker compose up -d
open http://localhost:8080
```

See [DOCKER.md](./DOCKER.md) for production deployment.

---

## Development Setup

### Prerequisites

**Required**:
- Go 1.16 or higher ([install](https://go.dev/doc/install))
- SQLite3 (usually included with Go SQLite driver)
- Git

**Recommended**:
- VS Code with Go extension
- Docker Desktop (for deployment testing)
- SQLite browser (for database inspection)

### Initial Setup

```bash
# 1. Clone repository
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio

# 2. Install dependencies
go mod download

# 3. Run tests
go test -v

# 4. Run application
go run main.go

# 5. Verify it works
curl http://localhost:8080/login
```

### Development Workflow

```bash
# Make code changes
# Edit main.go or templates/

# Run tests
go test -v

# Run application
go run main.go

# Test in browser
open http://localhost:8080

# Format code
go fmt

# Build binary
go build -o tasking
```

### Project Structure

```
TaskFolio/
├── main.go              # All application code (~3700 lines)
├── main_test.go         # All tests (~1500 lines)
├── templates/           # HTML templates
│   ├── login.html
│   ├── dashboard.html
│   └── ...
├── go.mod               # Dependencies
├── taskfolio.db         # Database (created at runtime)
├── Dockerfile
├── docker-compose.yml
└── docs/
    └── planning/        # Future enhancement plans
```

**Why single file?** Educational choice - easier to understand complete flow. See [ARCHITECTURE.md](./ARCHITECTURE.md#single-file-architecture).

---

## Learning Path

### For Complete Beginners (40-60 hours)

**Week 1: Understand What It Does**
1. Run the application
2. Use it as a student and instructor
3. Read [ARCHITECTURE.md](./ARCHITECTURE.md)
4. Understand the high-level design

**Week 2-3: Understand How It Works**
1. Read [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)
2. Follow one request from browser to database
3. Understand sessions and authentication
4. Study the database schema

**Week 4-6: Make Changes**
1. Complete exercises in [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)
2. Add a simple field to tasks
3. Create a new page
4. Write tests for your changes

**Week 7-8: Build Something**
1. Build a related application
2. Apply patterns you learned
3. Share your work

### For Experienced Developers (8-12 hours)

**Day 1: Architecture**
- Read [ARCHITECTURE.md](./ARCHITECTURE.md) (2 hours)
- Understand design decisions and trade-offs

**Day 2: Code**
- Read [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md) (3 hours)
- Trace request flows
- Understand patterns

**Day 3: Hands-On**
- Make a small change (2 hours)
- Write a test
- Run the application

**Day 4+: Extend**
- Add a feature from your domain knowledge
- Refactor something
- Contribute back

---

## Contributing

### Making Changes

1. **Understand the code**: Read relevant documentation
2. **Make minimal changes**: Follow existing patterns
3. **Write tests**: Match existing test style
4. **Run tests**: `go test -v`
5. **Test manually**: Run the app and verify
6. **Format code**: `go fmt`
7. **Commit**: Clear commit message

### Code Style

- Follow existing Go conventions
- Use `gofmt` for formatting
- Add comments for complex logic
- Write tests for new features
- Keep functions focused

### Testing

```bash
# Run all tests
go test -v

# Run specific test
go test -run TestTaskCreation -v

# With coverage
go test -cover
```

See [TESTING.md](./TESTING.md) for details.

---

## Next Steps
### Based on Your Role

**Students/Learners**: 
- Start with [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)
- Follow the structured learning path
- Complete hands-on exercises

**Developers**:
- Read [ARCHITECTURE.md](./ARCHITECTURE.md) and [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)
- Set up development environment (above)
- Make a small change to understand the workflow

**Instructors**:
- Review [LEARNING_GUIDE.md](./LEARNING_GUIDE.md) for teaching materials
- Use demo accounts in classroom
- Assign exercises to students

---

## Resources

### Documentation
- [ARCHITECTURE.md](./ARCHITECTURE.md) - System design
- [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md) - Code explanation
- [LEARNING_GUIDE.md](./LEARNING_GUIDE.md) - Learning path
- [DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md) - Data model
- [DOCKER.md](./DOCKER.md) - Deployment
- [TESTING.md](./TESTING.md) - Testing guide

### Learning Resources

**Go Programming**:
- [Official Go Tour](https://go.dev/tour/)
- [Go by Example](https://gobyexample.com/)
- [Effective Go](https://go.dev/doc/effective_go)

**Web Development**:
- [MDN Web Docs](https://developer.mozilla.org/)
- [HTTP Status Codes](https://httpstatuses.com/)

**Database**:
- [SQLite Tutorial](https://www.sqlitetutorial.net/)
- [SQL Style Guide](https://www.sqlstyle.guide/)

**Security**:
- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [Web Security Academy](https://portswigger.net/web-security)

---

## Quick Commands Reference

```bash
# Development
go run main.go              # Run application
go test -v                  # Run tests
go fmt                      # Format code
go build -o tasking         # Build binary

# Docker
docker compose up -d        # Start with Docker
docker compose down         # Stop containers
docker compose logs -f      # View logs

# Database
sqlite3 taskfolio.db        # Open database
.schema users               # View table schema
SELECT * FROM users;        # Query data
```

---

## Getting Help

- **Documentation**: Check the relevant .md files above
- **Code Questions**: See [CODE_WALKTHROUGH.md](./CODE_WALKTHROUGH.md)
- **Learning**: See [LEARNING_GUIDE.md](./LEARNING_GUIDE.md)
- **Issues**: Open a GitHub issue

---

## Welcome!

Whether you're learning web development, contributing code, or using this for teaching, we're glad you're here. TaskFolio is designed to be both useful and educational.

**Get started**: Pick your path above and dive in! 🚀

---

**Last Updated**: February 2026  
**Version**: 2.0  
**Status**: Active educational project
