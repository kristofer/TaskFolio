# TaskFolio Database Schema

## Overview
This document describes the **current implementation** of the TaskFolio database schema using SQLite3. The application uses a simple, educational database design suitable for a coding bootcamp task management system.

> **Note**: This documentation reflects the actual implemented schema. For future expansion plans, see ROADMAP.md.

## Entity Relationship Diagram (ERD)

```
Users ──┬── Tasks (owner_id)
        ├── Tasks (created_by)
        ├── Groups (via group_members)
        └── GroupMembers

Tasks ──┬── Groups (group_id)
        ├── Users (owner_id)
        └── Users (created_by)

Groups ──┬── GroupMembers
         └── Tasks (group_id)
```

## Tables

### users
Stores student and instructor account information.

**Security Note**: This is a demo/educational application. Passwords are stored in plain text, which is **NOT suitable for production use**. Production systems should use bcrypt or similar hashing.

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,  -- ⚠️ Plain text for demo purposes only
    role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
    cohort_name TEXT,  -- e.g., 'Java11.2', 'Data6.2'
    is_suspended BOOLEAN DEFAULT 0
);
```

**Fields:**
- `id`: Auto-incrementing primary key
- `username`: Unique username for login
- `password`: Plain text password (demo only)
- `role`: Either 'student' or 'instructor'
- `cohort_name`: Optional cohort assignment (e.g., "Java11.2", "Data6.2")
- `is_suspended`: Account suspension flag (0=active, 1=suspended)

**Demo Data:**
The application seeds demo accounts on first run:
- Instructor: `username=instructor1`, `password=password123`
- Students: `username=student1/student2/student3`, `password=password123`

### tasks
Core table for task management, supporting both student-created and instructor-assigned tasks.

```sql
CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    due_date TEXT,
    status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
    owner_id INTEGER NOT NULL,  -- User who owns/is responsible for the task
    created_by INTEGER NOT NULL,  -- User who created the task
    assigned_to TEXT,  -- Comma-separated usernames for group assignments
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    task_type TEXT CHECK(task_type IN (
        'group-programming-project',
        'individual-programming-lab',
        'reading',
        'document',
        'resume',
        'prodev',
        'demo-presentation',
        ''
    )),
    resource_url TEXT,  -- URL to task resources (repos, docs, etc.)
    cohort_id TEXT,  -- Cohort name for cohort-wide assignments
    in_progress_date DATETIME,  -- When task status changed to 'in_progress'
    done_date DATETIME,  -- When task status changed to 'completed'
    group_id INTEGER,  -- Reference to groups table for group tasks
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id),
    FOREIGN KEY (group_id) REFERENCES groups(id)
);
```

**Fields:**
- `id`: Auto-incrementing primary key
- `title`: Task title/name
- `description`: Detailed description of the task
- `due_date`: Due date as text string
- `status`: Current status (pending, in_progress, completed, blocked)
- `owner_id`: User responsible for completing the task
- `created_by`: User who created the task (may differ from owner for instructor assignments)
- `assigned_to`: For group assignments, comma-separated list of usernames
- `created_at`: Creation timestamp
- `task_type`: Type of task (project, lab, reading, document, resume, professional development, demo)
- `resource_url`: URL to external resources (GitHub repos, Google Docs, etc.)
- `cohort_id`: For cohort-wide assignments, the cohort name
- `in_progress_date`: Timestamp when task was first moved to in_progress status
- `done_date`: Timestamp when task was moved to completed status
- `group_id`: Foreign key to groups table for group-based tasks

### groups
Student groups for collaborative work and group programming projects.

```sql
CREATE TABLE groups (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    description TEXT,
    project_url TEXT,  -- URL to group project repository
    is_active BOOLEAN DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
```

**Fields:**
- `id`: Auto-incrementing primary key
- `name`: Group name
- `description`: Description of the group's purpose
- `project_url`: URL to the group's project repository or resources
- `is_active`: Whether the group is currently active (1=active, 0=inactive)
- `created_at`: Creation timestamp

### group_members
Junction table for many-to-many relationship between groups and students.

```sql
CREATE TABLE group_members (
    group_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    PRIMARY KEY (group_id, student_id),
    FOREIGN KEY (group_id) REFERENCES groups(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE CASCADE
);
```

**Fields:**
- `group_id`: Foreign key to groups table
- `student_id`: Foreign key to users table (students only)
- Composite primary key ensures each student can only be in a group once
- CASCADE delete ensures cleanup when groups or students are removed

## Key Relationships

### Task Ownership vs Creation
- `owner_id`: The user responsible for completing the task
- `created_by`: The user who created the task
- For student-created tasks: `owner_id` = `created_by`
- For instructor-assigned tasks: `owner_id` = student, `created_by` = instructor

### Cohort-based Organization
- Users can belong to cohorts via the `cohort_name` field
- Tasks can be assigned to entire cohorts via the `cohort_id` field
- Instructors can filter views and assignments by cohort

### Group Assignments
- Groups are created with students as members
- Tasks can be assigned to groups via `group_id`
- Students in a group can collaborate on group tasks
- The `assigned_to` field can also contain comma-separated usernames for ad-hoc grouping

## Common Query Patterns

### Get All Tasks for a User
```sql
SELECT * FROM tasks 
WHERE owner_id = ? 
ORDER BY created_at DESC;
```

### Get Tasks by Status for a User
```sql
SELECT * FROM tasks 
WHERE owner_id = ? AND status = ? 
ORDER BY due_date ASC;
```

### Get All Students in a Cohort
```sql
SELECT * FROM users 
WHERE role = 'student' AND cohort_name = ?;
```

### Get Group Members
```sql
SELECT u.* 
FROM users u
INNER JOIN group_members gm ON u.id = gm.student_id
WHERE gm.group_id = ?;
```

### Get Tasks for a Group
```sql
SELECT * FROM tasks 
WHERE group_id = ? 
ORDER BY created_at DESC;
```

## Database Migrations

The application uses automatic schema migration on startup via the `migrateDatabase()` function. This approach:

1. **Checks for missing columns** using SQLite's `pragma_table_info()`
2. **Adds columns** using `ALTER TABLE` if they don't exist
3. **Preserves existing data** during migrations
4. **Handles CHECK constraint updates** by recreating tables when needed

### Migration Process for Adding New Fields

When adding new fields to the schema:

1. Update the struct definition in main.go
2. Add column creation in `initDB()` for new databases
3. Add column migration in `migrateDatabase()` for existing databases
4. Update any relevant queries to use the new field
5. Write tests for the new functionality

Example migration code:
```go
var columnExists int
err := db.QueryRow(`
    SELECT COUNT(*) FROM pragma_table_info('table_name') WHERE name='column_name'
`).Scan(&columnExists)
if err == nil && columnExists == 0 {
    _, err = db.Exec(`ALTER TABLE table_name ADD COLUMN column_name TYPE`)
    if err != nil {
        return fmt.Errorf("failed to add column: %v", err)
    }
}
```

## Implementation Notes

### SQLite Characteristics
- **Single file database**: Default location is `./taskfolio.db`
- **Auto-increment IDs**: Uses INTEGER PRIMARY KEY AUTOINCREMENT
- **Foreign key support**: Enabled via `PRAGMA foreign_keys = ON`
- **Simple types**: Uses TEXT, INTEGER, BOOLEAN, DATETIME
- **No UUID support**: Uses integer IDs instead

### Session Management
Currently implemented as in-memory map (not persisted to database):
- Simple cookie-based authentication
- Sessions stored in Go map in application memory
- No expiration mechanism (known limitation)
- Not suitable for multi-instance deployments

### Security Considerations

⚠️ **This is a demo/educational application** with known security limitations:

1. **Plain text passwords**: Should use bcrypt in production
2. **No CSRF protection**: Vulnerable to cross-site request forgery
3. **Basic session management**: No expiration, in-memory only
4. **No rate limiting**: Vulnerable to brute force attacks

See the main README.md for complete security disclaimers.

## Future Expansion Considerations

Features that may be added in future versions:
- Password hashing (bcrypt)
- Proper session management with expiration
- Email notifications
- Task comments/discussion
- File attachments
- Task history/audit trail
- Tags for task organization
- Task dependencies
- PostgreSQL support for larger deployments
- RESTful API endpoints

For roadmap details, see ROADMAP.md.
