# TaskFolio

A multi-tenant web-based task management system for Zip Code Wilmington coding bootcamp students and instructors, built with Go and SQLite3.

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
go build -o taskfolio
```

4. Run the application:
```bash
./taskfolio
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

## Development

### Running in Development Mode

```bash
go run main.go
```

### Building for Production

```bash
go build -o taskfolio
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
