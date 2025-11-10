-- Create an old database schema (before new task types)
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT UNIQUE NOT NULL,
    password TEXT NOT NULL,
    role TEXT NOT NULL
);

CREATE TABLE tasks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT,
    due_date TEXT,
    status TEXT DEFAULT 'pending',
    owner_id INTEGER NOT NULL,
    created_by INTEGER NOT NULL,
    assigned_to TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', '')),
    resource_url TEXT,
    cohort_id TEXT,
    FOREIGN KEY (owner_id) REFERENCES users(id),
    FOREIGN KEY (created_by) REFERENCES users(id)
);

-- Insert test data
INSERT INTO users (id, username, password, role) VALUES (1, 'instructor1', 'password123', 'instructor');
INSERT INTO users (id, username, password, role) VALUES (2, 'student1', 'password123', 'student');

INSERT INTO tasks (title, description, owner_id, created_by, task_type) 
VALUES ('Old Demo Task', 'This is a demo task', 1, 1, 'demo-presentation');

INSERT INTO tasks (title, description, owner_id, created_by, task_type) 
VALUES ('Old Reading Task', 'This is a reading task', 2, 1, 'reading');
