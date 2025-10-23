# Tasking

**task management for ZipCode students**

A multi-tenant web-based task management system for Zip Code Wilmington coding bootcamp students and instructors, built with Go and SQLite3.

> **Note**: While the repository is named TaskFolio, the application is branded as **Tasking** for end users.

## Features

### For Students
- Create, view, edit, and delete personal tasks
- Track task status (pending, in_progress, completed)
- View tasks assigned by instructors
- Organize tasks with titles, descriptions, and due dates

### For Instructors
- All student capabilities
- Assign tasks to individual students
- Assign tasks to all students at once
- Manage and track tasks created for students

## Technology Stack

- **Backend**: Go (Golang)
- **Database**: SQLite3
- **Frontend**: HTML5, CSS3 (embedded templates)
- **Session Management**: Cookie-based authentication

## Quick Start

### With Docker (Recommended)

1. Clone the repository:
```bash
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio
```

2. Start with Docker Compose:
```bash
docker compose up -d
```

3. Open your browser and navigate to:
```
http://localhost:8080
```

See [DOCKER.md](./DOCKER.md) for detailed Docker deployment instructions and production setup.

### Without Docker

### Prerequisites
- Go 1.16 or higher
- SQLite3

### Installation

1. Clone the repository:
```bash
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio
```

2. Install dependencies:
```bash
go mod download
```

3. Build the application:
```bash
go build -o tasking
```

4. Run the application:
```bash
./tasking
```

5. Open your browser and navigate to:
```
http://localhost:8080
```

## Demo Accounts

The application comes pre-populated with demo accounts:

**Instructor Account:**
- Username: `instructor1`
- Password: `password123`

**Student Accounts:**
- Username: `student1`, `student2`, or `student3`
- Password: `password123`

## Database Schema

### Users Table
- `id`: Primary key
- `username`: Unique username
- `password`: User password (plain text for demo - should be hashed in production)
- `role`: User role (`student` or `instructor`)

### Tasks Table
- `id`: Primary key
- `title`: Task title
- `description`: Task description
- `due_date`: Due date for the task
- `status`: Task status (`pending`, `in_progress`, or `completed`)
- `owner_id`: ID of the user who owns the task
- `created_by`: ID of the user who created the task
- `assigned_to`: For group assignments
- `created_at`: Timestamp of task creation

## API Endpoints

- `GET /` - Home page (requires authentication)
- `GET /login` - Login page
- `POST /login` - Login submission
- `GET /logout` - Logout
- `GET /tasks/add` - Add task form
- `POST /tasks/add` - Create new task
- `GET /tasks/edit?id={id}` - Edit task form
- `POST /tasks/edit?id={id}` - Update task
- `GET /tasks/delete?id={id}` - Delete task

## Configuration

The application uses the following environment variables:

- `PORT`: Server port (default: 8080)
- `DATABASE_PATH`: Path to SQLite database file (default: ./taskfolio.db)

## Deployment

For production deployment instructions, including:
- Docker setup with persistent database
- HTTPS configuration for `https://tasking.zipcode.rocks`
- Automated backups
- Resource limits

See [DOCKER.md](./DOCKER.md) for complete deployment guide.

## Development

### Running in Development Mode

```bash
go run main.go
```

### Building for Production

```bash
go build -o tasking
```

## Security Notes

⚠️ **Important**: This is a demo application. For production use:

1. Implement proper password hashing (e.g., bcrypt)
2. Use secure session management
3. Add CSRF protection
4. Implement HTTPS
5. Add input validation and sanitization
6. Use prepared statements (already implemented)
7. Add rate limiting
8. Implement proper error handling

## Future Enhancements

- [ ] Password hashing with bcrypt
- [ ] Email notifications for task assignments
- [ ] Task categories/tags
- [ ] Task priorities
- [ ] File attachments
- [ ] Comments on tasks
- [ ] Task search and filtering
- [ ] Export tasks to CSV/PDF
- [ ] Mobile responsive improvements
- [ ] Dark mode

## License

MIT License

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Support

For questions or issues, please open an issue on GitHub.
A comprehensive student task management system designed for Zip Code Wilmington.

## Overview

TaskFolio is a modern, web-based task management application that helps students organize, track, and complete their assignments and projects efficiently. Built with a focus on simplicity, performance, and collaboration.

## Key Features

- **Task Management**: Create, organize, and track tasks with priorities, due dates, and status tracking
- **Collaboration**: Comment on tasks, @mention team members, and work together
- **Teams**: Organize users into teams for group projects and assignments
- **Real-time Updates**: Get instant notifications when tasks are updated
- **Analytics**: Track productivity and completion metrics
- **Multiple Views**: List, Kanban board, and calendar views
- **Mobile Responsive**: Works seamlessly on desktop, tablet, and mobile devices

## Technology Stack

### Backend
- **Language**: Go 1.21+
- **Database**: PostgreSQL 15+
- **Authentication**: JWT-based authentication
- **API**: RESTful API with WebSocket support

### Frontend
- **Framework**: React 18+ / Vue 3+
- **Build Tool**: Vite
- **State Management**: Redux/Zustand or Pinia
- **UI Components**: Material-UI or Ant Design

### DevOps
- **Containerization**: Docker & Docker Compose
- **CI/CD**: GitHub Actions
- **Monitoring**: Prometheus & Grafana
- **Deployment**: Cloud platform (AWS/GCP/Azure/DigitalOcean)

## Project Documentation

This repository contains comprehensive planning documents for the TaskFolio application.

### 🚀 Start Here
- **[GETTING_STARTED.md](./GETTING_STARTED.md)** - New to the project? Start here!
- **[PLANNING_SUMMARY.md](./PLANNING_SUMMARY.md)** - Executive summary of all planning documents

### 📚 Complete Documentation
- **[IMPLEMENTATION_PLAN.md](./IMPLEMENTATION_PLAN.md)** - Complete multi-phase implementation plan with tasks and sub-tasks
- **[DATABASE_SCHEMA.md](./DATABASE_SCHEMA.md)** - Detailed database design and schema documentation
- **[API_SPECIFICATION.md](./API_SPECIFICATION.md)** - RESTful API endpoint specifications
- **[PROJECT_STRUCTURE.md](./PROJECT_STRUCTURE.md)** - Recommended project directory structure and conventions
- **[TECHNICAL_REQUIREMENTS.md](./TECHNICAL_REQUIREMENTS.md)** - Functional and non-functional requirements
- **[ROADMAP.md](./ROADMAP.md)** - Development timeline and milestones
- **[SPRINT_PLANNING.md](./SPRINT_PLANNING.md)** - Sprint planning template and examples

**Total**: 9 comprehensive documents with 85,000+ words of specifications, designs, and guidelines.

## Getting Started

### Prerequisites

- Go 1.21 or higher
- PostgreSQL 15 or higher
- Node.js 18 or higher (for frontend)
- Docker and Docker Compose (optional, for containerized development)

### Quick Start (Coming Soon)

```bash
# Clone the repository
git clone https://github.com/kristofer/TaskFolio.git
cd TaskFolio

# Set up environment variables
cp .env.example .env

# Start with Docker Compose
docker-compose up -d

# Or run locally
# Start the backend
cd cmd/api
go run main.go

# Start the frontend (in another terminal)
cd web
npm install
npm run dev
```

## Development Timeline

- **Phase 1**: Project Setup & Infrastructure (1-2 weeks)
- **Phase 2**: Core Backend Development (3-4 weeks)
- **Phase 3**: Frontend Development (4-5 weeks)
- **Phase 4**: Integration & Features (3-4 weeks)
- **Phase 5**: Testing & QA (2-3 weeks)
- **Phase 6**: Deployment & Launch (2-3 weeks)

**Total**: Approximately 4-5 months

## Contributing

We welcome contributions! Please read our contributing guidelines (coming soon) before submitting pull requests.

### Development Workflow

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Write or update tests
5. Commit your changes (`git commit -m 'Add some amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

## Testing

```bash
# Run backend tests
go test -v ./...

# Run frontend tests
cd web
npm test

# Run E2E tests
npm run test:e2e
```

## Project Status

🚧 **Currently in Planning Phase** 🚧

This project is currently in the planning and design phase. We are actively working on:

- [x] Project structure and architecture
- [x] Database schema design
- [x] API specification
- [x] Technical requirements
- [ ] Initial project setup
- [ ] Core backend implementation
- [ ] Frontend development
- [ ] Testing and deployment

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contact

For questions or feedback about TaskFolio, please open an issue on GitHub.

## Acknowledgments

- Zip Code Wilmington for project inspiration
- The Go and React/Vue communities for excellent tools and libraries
- All contributors who help make TaskFolio better

---

**Note**: This is an educational project designed for learning and development purposes.
