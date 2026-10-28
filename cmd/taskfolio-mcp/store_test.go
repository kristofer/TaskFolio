package main

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// schema mirrors the CREATE TABLE statements in TaskFolio's initDB (main.go).
// Keep it in sync when main.go's schema changes.
const schema = `
CREATE TABLE users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT UNIQUE NOT NULL,
	password TEXT NOT NULL,
	role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
	cohort_name TEXT,
	is_suspended BOOLEAN DEFAULT 0,
	is_current BOOLEAN DEFAULT 1
);
CREATE TABLE tasks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	due_date TEXT,
	status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed', 'blocked')),
	owner_id INTEGER NOT NULL,
	created_by INTEGER NOT NULL,
	assigned_to TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
	task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', 'research-analysis-design', 'organize', '')),
	resource_url TEXT,
	cohort_id TEXT,
	in_progress_date DATETIME,
	done_date DATETIME,
	group_id INTEGER,
	comments TEXT,
	display_order INTEGER DEFAULT 0
);
CREATE TABLE task_templates (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT NOT NULL,
	description TEXT,
	task_type TEXT CHECK(task_type IN ('group-programming-project', 'individual-programming-lab', 'reading', 'document', 'resume', 'prodev', 'demo-presentation', 'research-analysis-design', 'organize', '')),
	resource_url TEXT,
	created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
INSERT INTO users (id, username, password, role, cohort_name, is_current) VALUES
	(1, 'instructor1', 'secret-pw', 'instructor', NULL, 1),
	(2, 'student1', 'secret-pw', 'student', 'Java11.2', 1),
	(3, 'student2', 'secret-pw', 'student', 'Java11.2', 1),
	(4, 'student3', 'secret-pw', 'student', 'Data6.2', 1),
	(5, 'alumni1', 'secret-pw', 'student', 'Java10.1', 0);
INSERT INTO tasks (id, title, status, owner_id, created_by, task_type, display_order) VALUES
	(1, 'Read chapter 1', 'completed', 2, 1, 'reading', 0),
	(2, 'Lab: loops', 'in_progress', 2, 1, 'individual-programming-lab', 0),
	(3, 'Resume draft', 'pending', 2, 2, 'resume', 0),
	(4, 'Lab: loops', 'blocked', 3, 1, 'individual-programming-lab', 0);
INSERT INTO task_templates (id, title, description, task_type, resource_url) VALUES
	(1, 'Lab: recursion', 'Write recursive functions', 'individual-programming-lab', 'https://example.com/recursion'),
	(2, 'Read: Clean Code ch. 2', 'Naming', 'reading', '');
`

// newTestStore creates a temp SQLite file with TaskFolio's schema and seed data.
// Parameters: t, the test. Returns a Store opened read-write. Fails the test on any setup error.
func newTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "taskfolio.db")
	raw, err := sql.Open("sqlite3", path)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := raw.Exec(schema); err != nil {
		t.Fatal(err)
	}
	raw.Close()
	s, err := OpenStore(path, false)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestListCohortsOnlyCurrentStudents(t *testing.T) {
	s := newTestStore(t)
	got, err := s.ListCohorts()
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Data6.2", "Java11.2"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] {
		t.Fatalf("ListCohorts = %v, want %v (past cohort Java10.1 must be excluded)", got, want)
	}
}

func TestListStudentsFiltersByCohort(t *testing.T) {
	s := newTestStore(t)
	all, err := s.ListStudents("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("ListStudents(\"\") returned %d students, want 3 current students", len(all))
	}
	java, err := s.ListStudents("Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	if len(java) != 2 || java[0].Username != "student1" || java[1].Username != "student2" {
		t.Fatalf("ListStudents(Java11.2) = %+v", java)
	}
}

func TestStudentBoardReturnsTasksInBoardOrder(t *testing.T) {
	s := newTestStore(t)
	b, err := s.StudentBoard("student1")
	if err != nil {
		t.Fatal(err)
	}
	if b.Student.Username != "student1" || len(b.Tasks) != 3 {
		t.Fatalf("StudentBoard = %+v", b)
	}
	// TaskFolio orders by status, display_order, created_at DESC.
	if b.Tasks[0].Status != "completed" || b.Tasks[1].Status != "in_progress" || b.Tasks[2].Status != "pending" {
		t.Fatalf("unexpected order: %s, %s, %s", b.Tasks[0].Status, b.Tasks[1].Status, b.Tasks[2].Status)
	}
}

func TestStudentBoardUnknownStudent(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.StudentBoard("instructor1"); err == nil {
		t.Fatal("StudentBoard(instructor1) should fail: not a student")
	}
	if _, err := s.StudentBoard("nobody"); err == nil {
		t.Fatal("StudentBoard(nobody) should fail")
	}
}

func TestCohortProgressCountsByStatus(t *testing.T) {
	s := newTestStore(t)
	p, err := s.CohortProgress("Java11.2")
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 2 {
		t.Fatalf("CohortProgress returned %d rows, want 2", len(p))
	}
	s1 := p[0]
	if s1.Username != "student1" || s1.Total != 3 || s1.Completed != 1 || s1.InProgress != 1 || s1.Pending != 1 || s1.Blocked != 0 {
		t.Fatalf("student1 progress = %+v", s1)
	}
	if p[1].Username != "student2" || p[1].Blocked != 1 {
		t.Fatalf("student2 progress = %+v", p[1])
	}
}

func TestSearchTemplates(t *testing.T) {
	s := newTestStore(t)
	got, err := s.SearchTemplates("recursi")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != 1 {
		t.Fatalf("SearchTemplates(recursi) = %+v", got)
	}
	all, err := s.SearchTemplates("")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 {
		t.Fatalf("SearchTemplates(\"\") returned %d, want 2", len(all))
	}
}

func TestResolveInstructorRejectsStudents(t *testing.T) {
	s := newTestStore(t)
	id, err := s.ResolveInstructor("instructor1")
	if err != nil || id != 1 {
		t.Fatalf("ResolveInstructor(instructor1) = %d, %v", id, err)
	}
	if _, err := s.ResolveInstructor("student1"); err == nil {
		t.Fatal("ResolveInstructor(student1) should fail")
	}
}

func TestAssignTemplateToCohort(t *testing.T) {
	s := newTestStore(t)
	res, err := s.AssignTemplate(AssignInput{InstructorID: 1, TemplateID: 1, Cohort: "Java11.2", DueDate: "2026-11-15"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Created != 2 || len(res.Skipped) != 0 {
		t.Fatalf("AssignTemplate result = %+v", res)
	}
	var n int
	var createdBy int
	var title, taskType, url, status string
	err = s.db.QueryRow(`SELECT COUNT(*), MAX(created_by), MAX(title), MAX(task_type), MAX(resource_url), MAX(status)
		FROM tasks WHERE title = 'Lab: recursion' AND due_date = '2026-11-15'`).Scan(&n, &createdBy, &title, &taskType, &url, &status)
	if err != nil {
		t.Fatal(err)
	}
	if n != 2 || createdBy != 1 || taskType != "individual-programming-lab" || url != "https://example.com/recursion" || status != "pending" {
		t.Fatalf("inserted rows: n=%d created_by=%d type=%s url=%s status=%s", n, createdBy, taskType, url, status)
	}
}

func TestAssignTemplateSkipsUnknownAndPastStudents(t *testing.T) {
	s := newTestStore(t)
	res, err := s.AssignTemplate(AssignInput{InstructorID: 1, TemplateID: 2, Usernames: []string{"student3", "alumni1", "ghost"}})
	if err != nil {
		t.Fatal(err)
	}
	if res.Created != 1 || len(res.Skipped) != 2 {
		t.Fatalf("AssignTemplate result = %+v, want 1 created and alumni1, ghost skipped", res)
	}
}

func TestAssignTemplateValidation(t *testing.T) {
	s := newTestStore(t)
	if _, err := s.AssignTemplate(AssignInput{InstructorID: 1, TemplateID: 99, Cohort: "Java11.2"}); err == nil {
		t.Fatal("unknown template should fail")
	}
	if _, err := s.AssignTemplate(AssignInput{InstructorID: 1, TemplateID: 1}); err == nil {
		t.Fatal("no cohort and no usernames should fail")
	}
	if _, err := s.AssignTemplate(AssignInput{InstructorID: 1, TemplateID: 1, Cohort: "Java11.2", Usernames: []string{"student1"}}); err == nil {
		t.Fatal("cohort and usernames together should fail")
	}
}

func TestCreateTask(t *testing.T) {
	s := newTestStore(t)
	id, err := s.CreateTask(CreateTaskInput{InstructorID: 1, Username: "student3", Title: "Demo day prep", TaskType: "demo-presentation", DueDate: "2026-12-01"})
	if err != nil {
		t.Fatal(err)
	}
	var owner, createdBy int
	if err := s.db.QueryRow(`SELECT owner_id, created_by FROM tasks WHERE id = ?`, id).Scan(&owner, &createdBy); err != nil {
		t.Fatal(err)
	}
	if owner != 4 || createdBy != 1 {
		t.Fatalf("owner=%d created_by=%d", owner, createdBy)
	}
	if _, err := s.CreateTask(CreateTaskInput{InstructorID: 1, Username: "student3", Title: "x", TaskType: "not-a-type"}); err == nil {
		t.Fatal("invalid task type should fail")
	}
	if _, err := s.CreateTask(CreateTaskInput{InstructorID: 1, Username: "student3", Title: ""}); err == nil {
		t.Fatal("empty title should fail")
	}
}

func TestUpdateTaskStatusSetsDates(t *testing.T) {
	s := newTestStore(t)
	// pending -> in_progress sets in_progress_date the first time
	if err := s.UpdateTaskStatus(3, "in_progress"); err != nil {
		t.Fatal(err)
	}
	var ip, done sql.NullString
	s.db.QueryRow(`SELECT in_progress_date, done_date FROM tasks WHERE id = 3`).Scan(&ip, &done)
	if !ip.Valid || done.Valid {
		t.Fatalf("after in_progress: in_progress_date=%v done_date=%v", ip, done)
	}
	// -> completed sets done_date
	if err := s.UpdateTaskStatus(3, "completed"); err != nil {
		t.Fatal(err)
	}
	s.db.QueryRow(`SELECT done_date FROM tasks WHERE id = 3`).Scan(&done)
	if !done.Valid {
		t.Fatal("completed should set done_date")
	}
	// completed -> pending clears done_date
	if err := s.UpdateTaskStatus(3, "pending"); err != nil {
		t.Fatal(err)
	}
	s.db.QueryRow(`SELECT done_date FROM tasks WHERE id = 3`).Scan(&done)
	if done.Valid {
		t.Fatal("leaving completed should clear done_date")
	}
	if err := s.UpdateTaskStatus(3, "archived"); err == nil {
		t.Fatal("invalid status should fail")
	}
	if err := s.UpdateTaskStatus(999, "pending"); err == nil {
		t.Fatal("unknown task should fail")
	}
}

func TestReadOnlyStoreRejectsWrites(t *testing.T) {
	s := newTestStore(t)
	path := s.path
	ro, err := OpenStore(path, true)
	if err != nil {
		t.Fatal(err)
	}
	defer ro.Close()
	if err := ro.UpdateTaskStatus(3, "completed"); err == nil {
		t.Fatal("read-only store must reject writes")
	}
}
