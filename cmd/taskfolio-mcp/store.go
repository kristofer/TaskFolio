package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// Store is the MCP server's data layer over TaskFolio's SQLite database.
// Its queries mirror TaskFolio's handlers in the root main.go. It never selects users.password.
type Store struct {
	db   *sql.DB
	path string
}

// Student is a student as the MCP tools expose it (no password, no role).
type Student struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Cohort   string `json:"cohort"`
}

// Task is one task on a student's board.
type Task struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Status      string `json:"status"`
	TaskType    string `json:"task_type,omitempty"`
	ResourceURL string `json:"resource_url,omitempty"`
	Comments    string `json:"comments,omitempty"`
}

// Board is a student and their tasks in TaskFolio's board order.
type Board struct {
	Student Student `json:"student"`
	Tasks   []Task  `json:"tasks"`
}

// Progress is one student's task counts by status.
type Progress struct {
	Username   string `json:"username"`
	Total      int    `json:"total"`
	Pending    int    `json:"pending"`
	InProgress int    `json:"in_progress"`
	Blocked    int    `json:"blocked"`
	Completed  int    `json:"completed"`
}

// Template is a task template from the instructor library.
type Template struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	TaskType    string `json:"task_type,omitempty"`
	ResourceURL string `json:"resource_url,omitempty"`
}

// AssignInput describes a template assignment: exactly one of Cohort or Usernames must be set.
// Title, Description and ResourceURL override the template's values when non-empty.
type AssignInput struct {
	InstructorID int
	TemplateID   int
	Cohort       string
	Usernames    []string
	DueDate      string
	Title        string
	Description  string
	ResourceURL  string
}

// AssignResult reports how many tasks were created and which usernames were skipped.
type AssignResult struct {
	Created int      `json:"created"`
	Skipped []string `json:"skipped"`
}

// CreateTaskInput describes one instructor-created task for one student.
type CreateTaskInput struct {
	InstructorID int
	Username     string
	Title        string
	Description  string
	DueDate      string
	TaskType     string
	ResourceURL  string
}

// validTaskTypes matches the CHECK constraint on tasks.task_type in TaskFolio's schema.
var validTaskTypes = map[string]bool{
	"": true, "group-programming-project": true, "individual-programming-lab": true,
	"reading": true, "document": true, "resume": true, "prodev": true,
	"demo-presentation": true, "research-analysis-design": true, "organize": true,
}

// validStatuses matches the CHECK constraint on tasks.status.
var validStatuses = map[string]bool{"pending": true, "in_progress": true, "completed": true, "blocked": true}

// OpenStore opens TaskFolio's SQLite database at path.
// Parameters: path, the database file; readOnly, open with SQLite's mode=ro so writes fail.
// Returns the Store, or an error if the file is missing or not a TaskFolio database.
func OpenStore(path string, readOnly bool) (*Store, error) {
	dsn := "file:" + path + "?_busy_timeout=5000"
	if readOnly {
		dsn += "&mode=ro"
	} else {
		dsn += "&mode=rw" // never create a new empty database by accident
	}
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('users','tasks','task_templates')`).Scan(&n); err != nil {
		db.Close()
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	if n != 3 {
		db.Close()
		return nil, fmt.Errorf("%s is not a TaskFolio database (users, tasks, task_templates tables not found)", path)
	}
	return &Store{db: db, path: path}, nil
}

// Close closes the database. Returns any error from the driver.
func (s *Store) Close() error { return s.db.Close() }

// ListCohorts returns the distinct cohort names of current students, sorted.
// Mirrors apiCohortsHandler. Returns an error on query failure.
func (s *Store) ListCohorts() ([]string, error) {
	rows, err := s.db.Query(`
		SELECT DISTINCT cohort_name FROM users
		WHERE role = 'student' AND cohort_name IS NOT NULL AND cohort_name != '' AND is_current = 1
		ORDER BY cohort_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	cohorts := []string{}
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cohorts = append(cohorts, c)
	}
	return cohorts, rows.Err()
}

// ListStudents returns current students, optionally limited to one cohort, sorted by username.
// Parameters: cohort, a cohort name or "" for all. Mirrors apiStudentsHandler. Returns an error on query failure.
func (s *Store) ListStudents(cohort string) ([]Student, error) {
	rows, err := s.db.Query(`
		SELECT id, username, COALESCE(cohort_name, '') FROM users
		WHERE role = 'student' AND is_current = 1 AND (? = '' OR cohort_name = ?)
		ORDER BY username`, cohort, cohort)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	students := []Student{}
	for rows.Next() {
		var st Student
		if err := rows.Scan(&st.ID, &st.Username, &st.Cohort); err != nil {
			return nil, err
		}
		students = append(students, st)
	}
	return students, rows.Err()
}

// findStudent looks up a student (current or past) by username.
// Returns the student, or an error if no student has that username.
func (s *Store) findStudent(username string) (Student, error) {
	var st Student
	err := s.db.QueryRow(`SELECT id, username, COALESCE(cohort_name, '') FROM users WHERE username = ? AND role = 'student'`,
		username).Scan(&st.ID, &st.Username, &st.Cohort)
	if errors.Is(err, sql.ErrNoRows) {
		return st, fmt.Errorf("no student named %q", username)
	}
	return st, err
}

// StudentBoard returns a student's tasks in TaskFolio's board order (status, display_order, newest first).
// Parameters: username. Returns an error if the user is not a student. Mirrors studentProgressHandler.
func (s *Store) StudentBoard(username string) (Board, error) {
	st, err := s.findStudent(username)
	if err != nil {
		return Board{}, err
	}
	rows, err := s.db.Query(`
		SELECT id, title, COALESCE(description, ''), COALESCE(due_date, ''), status,
		       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(comments, '')
		FROM tasks WHERE owner_id = ?
		ORDER BY status, display_order, created_at DESC`, st.ID)
	if err != nil {
		return Board{}, err
	}
	defer rows.Close()
	b := Board{Student: st, Tasks: []Task{}}
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.DueDate, &t.Status, &t.TaskType, &t.ResourceURL, &t.Comments); err != nil {
			return Board{}, err
		}
		b.Tasks = append(b.Tasks, t)
	}
	return b, rows.Err()
}

// CohortProgress returns task counts by status for each current student in a cohort, sorted by username.
// Parameters: cohort, a cohort name (required). Returns an error if cohort is empty or the query fails.
func (s *Store) CohortProgress(cohort string) ([]Progress, error) {
	if cohort == "" {
		return nil, errors.New("cohort is required")
	}
	rows, err := s.db.Query(`
		SELECT u.username,
		       COUNT(t.id),
		       COALESCE(SUM(t.status = 'pending'), 0),
		       COALESCE(SUM(t.status = 'in_progress'), 0),
		       COALESCE(SUM(t.status = 'blocked'), 0),
		       COALESCE(SUM(t.status = 'completed'), 0)
		FROM users u LEFT JOIN tasks t ON t.owner_id = u.id
		WHERE u.role = 'student' AND u.is_current = 1 AND u.cohort_name = ?
		GROUP BY u.id ORDER BY u.username`, cohort)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Progress{}
	for rows.Next() {
		var p Progress
		if err := rows.Scan(&p.Username, &p.Total, &p.Pending, &p.InProgress, &p.Blocked, &p.Completed); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// SearchTemplates returns up to 50 templates whose title or description contains query, or the 50 newest if query is "".
// Mirrors apiTaskTemplatesHandler. Returns an error on query failure.
func (s *Store) SearchTemplates(query string) ([]Template, error) {
	q := `SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM task_templates ORDER BY created_at DESC, id DESC LIMIT 50`
	args := []any{}
	if query != "" {
		q = `SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM task_templates WHERE title LIKE ? OR description LIKE ? ORDER BY title LIMIT 50`
		args = []any{"%" + query + "%", "%" + query + "%"}
	}
	rows, err := s.db.Query(q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Template{}
	for rows.Next() {
		var t Template
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.TaskType, &t.ResourceURL); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// ResolveInstructor returns the user ID of an instructor account.
// Parameters: username. Returns an error if the account does not exist, is not an instructor, or is suspended.
func (s *Store) ResolveInstructor(username string) (int, error) {
	var id int
	var role string
	var suspended bool
	err := s.db.QueryRow(`SELECT id, role, COALESCE(is_suspended, 0) FROM users WHERE username = ?`, username).Scan(&id, &role, &suspended)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("no user named %q", username)
	}
	if err != nil {
		return 0, err
	}
	if role != "instructor" {
		return 0, fmt.Errorf("%q is a %s, not an instructor", username, role)
	}
	if suspended {
		return 0, fmt.Errorf("instructor %q is suspended", username)
	}
	return id, nil
}

// AssignTemplate creates one pending task per target student from a template, like assignTaskTemplateHandler.
// Targets are all current students in in.Cohort, or the current students among in.Usernames; others are skipped.
// Returns the counts, or an error for an unknown template or invalid targets. Inserts run in one transaction.
func (s *Store) AssignTemplate(in AssignInput) (AssignResult, error) {
	if (in.Cohort == "") == (len(in.Usernames) == 0) {
		return AssignResult{}, errors.New("set exactly one of cohort or usernames")
	}
	var tpl Template
	err := s.db.QueryRow(`SELECT id, title, COALESCE(description, ''), COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM task_templates WHERE id = ?`, in.TemplateID).Scan(&tpl.ID, &tpl.Title, &tpl.Description, &tpl.TaskType, &tpl.ResourceURL)
	if errors.Is(err, sql.ErrNoRows) {
		return AssignResult{}, fmt.Errorf("no template with id %d", in.TemplateID)
	}
	if err != nil {
		return AssignResult{}, err
	}
	targets, skipped, err := s.resolveTargets(in.Cohort, in.Usernames)
	if err != nil {
		return AssignResult{}, err
	}
	title := firstNonEmpty(in.Title, tpl.Title)
	desc := firstNonEmpty(in.Description, tpl.Description)
	url := firstNonEmpty(in.ResourceURL, tpl.ResourceURL)

	tx, err := s.db.Begin()
	if err != nil {
		return AssignResult{}, err
	}
	defer tx.Rollback()
	for _, st := range targets {
		if _, err := tx.Exec(`INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, task_type, resource_url)
			VALUES (?, ?, ?, 'pending', ?, ?, ?, ?)`, title, desc, in.DueDate, st.ID, in.InstructorID, tpl.TaskType, url); err != nil {
			return AssignResult{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return AssignResult{}, err
	}
	return AssignResult{Created: len(targets), Skipped: skipped}, nil
}

// resolveTargets turns a cohort or a username list into current students.
// Returns the students, the usernames that are not current students, and any query error.
func (s *Store) resolveTargets(cohort string, usernames []string) ([]Student, []string, error) {
	skipped := []string{}
	if cohort != "" {
		st, err := s.ListStudents(cohort)
		if err == nil && len(st) == 0 {
			err = fmt.Errorf("no current students in cohort %q", cohort)
		}
		return st, skipped, err
	}
	current, err := s.ListStudents("")
	if err != nil {
		return nil, nil, err
	}
	byName := map[string]Student{}
	for _, st := range current {
		byName[st.Username] = st
	}
	var targets []Student
	for _, u := range usernames {
		st, ok := byName[strings.TrimSpace(u)]
		if !ok {
			skipped = append(skipped, u)
			continue
		}
		targets = append(targets, st)
	}
	return targets, skipped, nil
}

// CreateTask creates one pending task for a current student, owned by the student and created by the instructor.
// Returns the new task ID, or an error for an empty title, invalid task type or unknown student.
func (s *Store) CreateTask(in CreateTaskInput) (int64, error) {
	if strings.TrimSpace(in.Title) == "" {
		return 0, errors.New("title is required")
	}
	if !validTaskTypes[in.TaskType] {
		return 0, fmt.Errorf("invalid task_type %q", in.TaskType)
	}
	targets, skipped, err := s.resolveTargets("", []string{in.Username})
	if err != nil {
		return 0, err
	}
	if len(skipped) > 0 {
		return 0, fmt.Errorf("%q is not a current student", in.Username)
	}
	res, err := s.db.Exec(`INSERT INTO tasks (title, description, due_date, status, owner_id, created_by, task_type, resource_url)
		VALUES (?, ?, ?, 'pending', ?, ?, ?, ?)`, in.Title, in.Description, in.DueDate, targets[0].ID, in.InstructorID, in.TaskType, in.ResourceURL)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateTaskStatus moves a task to a new status, stamping dates the way updateTaskStatusHandler does:
// first move to in_progress sets in_progress_date; completed sets done_date; leaving completed clears it.
// Returns an error for an invalid status or unknown task.
func (s *Store) UpdateTaskStatus(taskID int, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("invalid status %q", status)
	}
	var current string
	var inProgress sql.NullString
	err := s.db.QueryRow(`SELECT status, in_progress_date FROM tasks WHERE id = ?`, taskID).Scan(&current, &inProgress)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("no task with id %d", taskID)
	}
	if err != nil {
		return err
	}
	now := time.Now()
	setStarted := status == "in_progress" && !inProgress.Valid
	switch {
	case status == "completed":
		_, err = s.db.Exec(`UPDATE tasks SET status = ?, done_date = ? WHERE id = ?`, status, now, taskID)
	case current == "completed" && setStarted:
		_, err = s.db.Exec(`UPDATE tasks SET status = ?, done_date = NULL, in_progress_date = ? WHERE id = ?`, status, now, taskID)
	case current == "completed":
		_, err = s.db.Exec(`UPDATE tasks SET status = ?, done_date = NULL WHERE id = ?`, status, taskID)
	case setStarted:
		_, err = s.db.Exec(`UPDATE tasks SET status = ?, in_progress_date = ? WHERE id = ?`, status, now, taskID)
	default:
		_, err = s.db.Exec(`UPDATE tasks SET status = ? WHERE id = ?`, status, taskID)
	}
	return err
}

// firstNonEmpty returns a if it is non-empty after trimming, otherwise b.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}
