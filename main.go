package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB
var templates *template.Template

type User struct {
	ID          int
	Username    string
	Password    string
	Role        string // "student" or "instructor"
	CohortName  string // e.g., "Java11.2", "Data6.2"
	IsSuspended bool   // For account suspension
}

type Task struct {
	ID          int
	Title       string
	Description string
	DueDate     string
	Status      string
	OwnerID     int
	CreatedBy   int
	AssignedTo  string // For group assignments, comma-separated usernames
	CreatedAt   time.Time
	TaskType    string // "reading", "coding", "submission", "document-creation"
	ResourceURL string // URL pointing to resources for the task
	CohortID    string // Cohort name for cohort-based assignments
	
	// Additional fields for instructor view
	AssignedToSelf      bool
	TotalStudents       int
	CompletedStudents   int
	InProgressStudents  int
	PendingStudents     int
}

func initDB() error {
	var err error
	db, err = sql.Open("sqlite3", "./taskfolio.db")
	if err != nil {
		return err
	}

	// Create users table with new fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT NOT NULL CHECK(role IN ('student', 'instructor')),
			cohort_name TEXT,
			is_suspended BOOLEAN DEFAULT 0
		)
	`)
	if err != nil {
		return err
	}

	// Create tasks table with new fields
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			description TEXT,
			due_date TEXT,
			status TEXT DEFAULT 'pending' CHECK(status IN ('pending', 'in_progress', 'completed')),
			owner_id INTEGER NOT NULL,
			created_by INTEGER NOT NULL,
			assigned_to TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			task_type TEXT CHECK(task_type IN ('reading', 'coding', 'submission', 'document-creation', '')),
			resource_url TEXT,
			cohort_id TEXT,
			FOREIGN KEY (owner_id) REFERENCES users(id),
			FOREIGN KEY (created_by) REFERENCES users(id)
		)
	`)
	if err != nil {
		return err
	}

	// Migrate existing database by adding new columns if they don't exist
	if err := migrateDatabase(); err != nil {
		return err
	}

	// Insert demo users if not exists
	_, err = db.Exec(`
		INSERT OR IGNORE INTO users (id, username, password, role, cohort_name, is_suspended) VALUES 
		(1, 'instructor1', 'password123', 'instructor', NULL, 0),
		(2, 'student1', 'password123', 'student', 'Java11.2', 0),
		(3, 'student2', 'password123', 'student', 'Java11.2', 0),
		(4, 'student3', 'password123', 'student', 'Data6.2', 0)
	`)
	if err != nil {
		return err
	}

	return nil
}

// migrateDatabase adds new columns to existing tables if they don't exist
func migrateDatabase() error {
	// Check and add cohort_name column to users table
	var cohortNameExists int
	err := db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='cohort_name'
	`).Scan(&cohortNameExists)
	if err == nil && cohortNameExists == 0 {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN cohort_name TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add cohort_name column: %v", err)
		}
		log.Println("Added cohort_name column to users table")
	}

	// Check and add is_suspended column to users table
	var isSuspendedExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('users') WHERE name='is_suspended'
	`).Scan(&isSuspendedExists)
	if err == nil && isSuspendedExists == 0 {
		_, err = db.Exec(`ALTER TABLE users ADD COLUMN is_suspended BOOLEAN DEFAULT 0`)
		if err != nil {
			return fmt.Errorf("failed to add is_suspended column: %v", err)
		}
		log.Println("Added is_suspended column to users table")
	}

	// Check and add task_type column to tasks table
	var taskTypeExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='task_type'
	`).Scan(&taskTypeExists)
	if err == nil && taskTypeExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN task_type TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add task_type column: %v", err)
		}
		log.Println("Added task_type column to tasks table")
	}

	// Check and add resource_url column to tasks table
	var resourceURLExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='resource_url'
	`).Scan(&resourceURLExists)
	if err == nil && resourceURLExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN resource_url TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add resource_url column: %v", err)
		}
		log.Println("Added resource_url column to tasks table")
	}

	// Check and add cohort_id column to tasks table
	var cohortIDExists int
	err = db.QueryRow(`
		SELECT COUNT(*) FROM pragma_table_info('tasks') WHERE name='cohort_id'
	`).Scan(&cohortIDExists)
	if err == nil && cohortIDExists == 0 {
		_, err = db.Exec(`ALTER TABLE tasks ADD COLUMN cohort_id TEXT`)
		if err != nil {
			return fmt.Errorf("failed to add cohort_id column: %v", err)
		}
		log.Println("Added cohort_id column to tasks table")
	}

	return nil
}

// studentsHandler displays list of students for instructors to manage
func studentsHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Only instructors can access this page
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	// Get all students
	rows, err := db.Query(`
		SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0)
		FROM users WHERE role = 'student' ORDER BY username
	`)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var students []User
	for rows.Next() {
		var s User
		s.Role = "student"
		rows.Scan(&s.ID, &s.Username, &s.CohortName, &s.IsSuspended)
		students = append(students, s)
	}

	data := struct {
		User     User
		Students []User
	}{
		User:     user,
		Students: students,
	}

	templates.ExecuteTemplate(w, "students.html", data)
}

// editStudentHandler handles editing student cohort and suspension status
func editStudentHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Only instructors can edit students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	studentID := r.URL.Query().Get("id")

	if r.Method == "GET" {
		var student User
		var cohortName sql.NullString
		err := db.QueryRow(`
			SELECT id, username, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0)
			FROM users WHERE id = ? AND role = 'student'
		`, studentID).Scan(&student.ID, &student.Username, &cohortName, &student.IsSuspended)

		if err != nil {
			http.Error(w, "Student not found", http.StatusNotFound)
			return
		}

		if cohortName.Valid {
			student.CohortName = cohortName.String
		}

		data := struct {
			User    User
			Student User
		}{
			User:    user,
			Student: student,
		}
		templates.ExecuteTemplate(w, "edit_student.html", data)
		return
	}

	cohortName := r.FormValue("cohort_name")
	isSuspended := r.FormValue("is_suspended") == "on"
	username := r.FormValue("username")
	password := r.FormValue("password")

	// Update student information
	// If password is provided, update it; otherwise, only update other fields
	if password != "" {
		_, err = db.Exec(`
			UPDATE users 
			SET username = ?, cohort_name = ?, is_suspended = ?, password = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, password, studentID)
	} else {
		_, err = db.Exec(`
			UPDATE users 
			SET username = ?, cohort_name = ?, is_suspended = ?
			WHERE id = ? AND role = 'student'
		`, username, cohortName, isSuspended, studentID)
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// addStudentHandler handles adding new students
func addStudentHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Only instructors can add students
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	if r.Method == "GET" {
		data := struct {
			User  User
			Error string
		}{
			User: user,
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")
	cohortName := r.FormValue("cohort_name")

	// Validate required fields
	if username == "" || password == "" {
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username and password are required",
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	// Insert new student
	_, err = db.Exec(`
		INSERT INTO users (username, password, role, cohort_name, is_suspended)
		VALUES (?, ?, 'student', ?, 0)
	`, username, password, cohortName)

	if err != nil {
		// Check if it's a unique constraint violation
		data := struct {
			User  User
			Error string
		}{
			User:  user,
			Error: "Username already exists. Please choose a different username.",
		}
		templates.ExecuteTemplate(w, "add_student.html", data)
		return
	}

	http.Redirect(w, r, "/students", http.StatusSeeOther)
}

// studentProgressHandler shows task completion for a specific student
func studentProgressHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Only instructors can view student progress
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	studentID := r.URL.Query().Get("id")

	// Get student info
	var student User
	var cohortName sql.NullString
	err = db.QueryRow(`
		SELECT id, username, COALESCE(cohort_name, '')
		FROM users WHERE id = ? AND role = 'student'
	`, studentID).Scan(&student.ID, &student.Username, &cohortName)

	if err != nil {
		http.Error(w, "Student not found", http.StatusNotFound)
		return
	}

	if cohortName.Valid {
		student.CohortName = cohortName.String
	}

	// Get all tasks for this student
	rows, err := db.Query(`
		SELECT id, title, description, due_date, status, created_at,
		       COALESCE(task_type, ''), COALESCE(resource_url, '')
		FROM tasks 
		WHERE owner_id = ?
		ORDER BY created_at DESC
	`, studentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tasks []Task
	totalTasks := 0
	completedTasks := 0
	inProgressTasks := 0
	pendingTasks := 0

	for rows.Next() {
		var task Task
		rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.CreatedAt, &task.TaskType, &task.ResourceURL)
		tasks = append(tasks, task)
		totalTasks++

		switch task.Status {
		case "completed":
			completedTasks++
		case "in_progress":
			inProgressTasks++
		case "pending":
			pendingTasks++
		}
	}

	completionPercentage := 0.0
	if totalTasks > 0 {
		completionPercentage = float64(completedTasks) / float64(totalTasks) * 100
	}

	data := struct {
		User                 User
		Student              User
		Tasks                []Task
		TotalTasks           int
		CompletedTasks       int
		InProgressTasks      int
		PendingTasks         int
		CompletionPercentage float64
	}{
		User:                 user,
		Student:              student,
		Tasks:                tasks,
		TotalTasks:           totalTasks,
		CompletedTasks:       completedTasks,
		InProgressTasks:      inProgressTasks,
		PendingTasks:         pendingTasks,
		CompletionPercentage: completionPercentage,
	}

	templates.ExecuteTemplate(w, "student_progress.html", data)
}

// taskDetailHandler shows detailed statistics for a task assigned to students
func taskDetailHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Only instructors can view task details
	if user.Role != "instructor" {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	taskID := r.URL.Query().Get("id")
	if taskID == "" {
		http.Error(w, "Task ID required", http.StatusBadRequest)
		return
	}

	// Get the base task info
	var task Task
	err = db.QueryRow(`
		SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
		       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, '')
		FROM tasks WHERE id = ?
	`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
		&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt,
		&task.TaskType, &task.ResourceURL, &task.CohortID)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	// Check if instructor created this task
	if task.CreatedBy != user.ID {
		http.Error(w, "Unauthorized - you did not create this task", http.StatusForbidden)
		return
	}

	// Determine if this task is assigned to self
	assignedToSelf := task.OwnerID == user.ID

	// If it's a task assigned to students, get all related tasks with same title and created_by
	// (since cohort/all assignments create multiple task records)
	type StudentTaskInfo struct {
		StudentID   int
		Username    string
		CohortName  string
		Status      string
		TaskID      int
	}

	var studentTasks []StudentTaskInfo
	var totalCount, pendingCount, inProgressCount, completedCount int

	if !assignedToSelf {
		// Find all tasks with same title, created_by, and cohort_id (if applicable)
		query := `
			SELECT t.id, t.owner_id, t.status, u.username, COALESCE(u.cohort_name, '')
			FROM tasks t
			JOIN users u ON t.owner_id = u.id
			WHERE t.title = ? AND t.created_by = ? AND t.owner_id != t.created_by
		`
		args := []interface{}{task.Title, task.CreatedBy}

		if task.CohortID != "" {
			query += " AND t.cohort_id = ?"
			args = append(args, task.CohortID)
		}

		query += " ORDER BY u.username"

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query student tasks: %v", err), http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var st StudentTaskInfo
			err := rows.Scan(&st.TaskID, &st.StudentID, &st.Status, &st.Username, &st.CohortName)
			if err != nil {
				log.Printf("Error scanning student task: %v", err)
				continue
			}
			studentTasks = append(studentTasks, st)
			totalCount++

			switch st.Status {
			case "pending":
				pendingCount++
			case "in_progress":
				inProgressCount++
			case "completed":
				completedCount++
			}
		}
	}

	// Calculate completion percentage
	completionPercentage := 0.0
	if totalCount > 0 {
		completionPercentage = float64(completedCount) / float64(totalCount) * 100
	}

	// Group students by status for display
	var pendingStudents, inProgressStudents, completedStudents []StudentTaskInfo
	for _, st := range studentTasks {
		switch st.Status {
		case "pending":
			pendingStudents = append(pendingStudents, st)
		case "in_progress":
			inProgressStudents = append(inProgressStudents, st)
		case "completed":
			completedStudents = append(completedStudents, st)
		}
	}

	data := struct {
		User                  User
		Task                  Task
		AssignedToSelf        bool
		TotalCount            int
		PendingCount          int
		InProgressCount       int
		CompletedCount        int
		CompletionPercentage  float64
		PendingStudents       []StudentTaskInfo
		InProgressStudents    []StudentTaskInfo
		CompletedStudents     []StudentTaskInfo
	}{
		User:                  user,
		Task:                  task,
		AssignedToSelf:        assignedToSelf,
		TotalCount:            totalCount,
		PendingCount:          pendingCount,
		InProgressCount:       inProgressCount,
		CompletedCount:        completedCount,
		CompletionPercentage:  completionPercentage,
		PendingStudents:       pendingStudents,
		InProgressStudents:    inProgressStudents,
		CompletedStudents:     completedStudents,
	}

	templates.ExecuteTemplate(w, "task_detail.html", data)
}

func main() {
	if err := initDB(); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Load templates
	templates = template.Must(template.ParseGlob("templates/*.html"))

	// Setup routes
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/logout", logoutHandler)
	http.HandleFunc("/tasks", tasksHandler)
	http.HandleFunc("/tasks/add", addTaskHandler)
	http.HandleFunc("/tasks/edit", editTaskHandler)
	http.HandleFunc("/tasks/delete", deleteTaskHandler)
	http.HandleFunc("/students", studentsHandler)
	http.HandleFunc("/students/add", addStudentHandler)
	http.HandleFunc("/students/edit", editStudentHandler)
	http.HandleFunc("/students/progress", studentProgressHandler)
	http.HandleFunc("/tasks/detail", taskDetailHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("TaskFolio server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	tasks, err := getTasks(user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		User  User
		Tasks []Task
	}{
		User:  user,
		Tasks: tasks,
	}

	templates.ExecuteTemplate(w, "home.html", data)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		templates.ExecuteTemplate(w, "login.html", nil)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	var user User
	var isSuspended bool
	err := db.QueryRow("SELECT id, username, role, COALESCE(is_suspended, 0) FROM users WHERE username = ? AND password = ?",
		username, password).Scan(&user.ID, &user.Username, &user.Role, &isSuspended)

	if err != nil {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Invalid credentials"})
		return
	}

	// Check if account is suspended
	if isSuspended {
		templates.ExecuteTemplate(w, "login.html", map[string]string{"Error": "Account suspended. Please contact an instructor."})
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:  "session",
		Value: fmt.Sprintf("%d", user.ID),
		Path:  "/",
	})

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func addTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if r.Method == "GET" {
		// Get all students and cohorts if instructor
		var students []User
		cohorts := make(map[string]bool)
		if user.Role == "instructor" {
			rows, err := db.Query("SELECT id, username, COALESCE(cohort_name, '') FROM users WHERE role = 'student'")
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var s User
					rows.Scan(&s.ID, &s.Username, &s.CohortName)
					students = append(students, s)
					if s.CohortName != "" {
						cohorts[s.CohortName] = true
					}
				}
			}
		}

		// Convert cohorts map to slice
		var cohortList []string
		for cohort := range cohorts {
			cohortList = append(cohortList, cohort)
		}

		data := struct {
			User     User
			Students []User
			Cohorts  []string
		}{
			User:     user,
			Students: students,
			Cohorts:  cohortList,
		}
		templates.ExecuteTemplate(w, "add_task.html", data)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	assignedTo := r.FormValue("assigned_to")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

	// Students can only create tasks for themselves
	ownerID := user.ID
	cohortID := ""

	if user.Role == "instructor" && assignedTo != "" {
		// Check if assigning to a cohort
		if len(assignedTo) > 7 && assignedTo[:7] == "cohort:" {
			cohortID = assignedTo[7:]
			ownerID = 0 // Special case: assign to cohort
		} else if assignedTo == "all" {
			ownerID = 0 // Special case: assign to all students
		} else {
			// Parse student ID and validate it's actually a student
			ownerID, _ = strconv.Atoi(assignedTo)
			// Validate that the target user is a student, not an instructor
			var targetRole string
			err = db.QueryRow("SELECT role FROM users WHERE id = ?", ownerID).Scan(&targetRole)
			if err != nil || targetRole != "student" {
				http.Error(w, "Invalid assignment: can only assign tasks to students", http.StatusBadRequest)
				return
			}
		}
	}

	if ownerID == 0 && cohortID != "" {
		// Assign to all students in a cohort
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student' AND cohort_name = ?", cohortID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query cohort students: %v", err), http.StatusInternalServerError)
			return
		}
		// Collect all student IDs first
		var studentIDs []int
		for rows.Next() {
			var studentID int
			if err := rows.Scan(&studentID); err != nil {
				log.Printf("Failed to scan student ID: %v", err)
				continue
			}
			studentIDs = append(studentIDs, studentID)
		}
		rows.Close()

		// Now create tasks for each student
		for _, studentID := range studentIDs {
			_, err := db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, cohortID)
			if err != nil {
				log.Printf("Failed to create task for student %d: %v", studentID, err)
			}
		}
	} else if ownerID == 0 {
		// Assign to all students
		rows, err := db.Query("SELECT id FROM users WHERE role = 'student'")
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to query students: %v", err), http.StatusInternalServerError)
			return
		}
		// Collect all student IDs first
		var studentIDs []int
		for rows.Next() {
			var studentID int
			if err := rows.Scan(&studentID); err != nil {
				log.Printf("Failed to scan student ID: %v", err)
				continue
			}
			studentIDs = append(studentIDs, studentID)
		}
		rows.Close()

		// Now create tasks for each student
		for _, studentID := range studentIDs {
			_, err := db.Exec(`
				INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, title, description, dueDate, studentID, user.ID, "", taskType, resourceURL, "")
			if err != nil {
				log.Printf("Failed to create task for student %d: %v", studentID, err)
			}
		}
	} else {
		_, err = db.Exec(`
			INSERT INTO tasks (title, description, due_date, owner_id, created_by, assigned_to, task_type, resource_url, cohort_id)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, title, description, dueDate, ownerID, user.ID, assignedTo, taskType, resourceURL, cohortID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func editTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	taskID := r.URL.Query().Get("id")

	if r.Method == "GET" {
		var task Task
		err := db.QueryRow(`
			SELECT id, title, description, due_date, status, owner_id, created_by,
			       COALESCE(task_type, ''), COALESCE(resource_url, '')
			FROM tasks WHERE id = ?
		`, taskID).Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.TaskType, &task.ResourceURL)

		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		// Check permissions
		if user.Role == "student" && task.OwnerID != user.ID {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		data := struct {
			User User
			Task Task
		}{
			User: user,
			Task: task,
		}
		templates.ExecuteTemplate(w, "edit_task.html", data)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	dueDate := r.FormValue("due_date")
	status := r.FormValue("status")
	taskType := r.FormValue("task_type")
	resourceURL := r.FormValue("resource_url")

	// Verify ownership or instructor role
	var task Task
	err = db.QueryRow("SELECT owner_id, created_by FROM tasks WHERE id = ?", taskID).
		Scan(&task.OwnerID, &task.CreatedBy)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec(`
		UPDATE tasks 
		SET title = ?, description = ?, due_date = ?, status = ?, task_type = ?, resource_url = ?
		WHERE id = ?
	`, title, description, dueDate, status, taskType, resourceURL, taskID)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func deleteTaskHandler(w http.ResponseWriter, r *http.Request) {
	session, err := getSession(r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	user, err := getUserByID(session.UserID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	taskID := r.URL.Query().Get("id")

	// Verify ownership or instructor role
	var task Task
	err = db.QueryRow("SELECT owner_id, created_by FROM tasks WHERE id = ?", taskID).
		Scan(&task.OwnerID, &task.CreatedBy)

	if err != nil {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	if user.Role == "student" && task.OwnerID != user.ID {
		http.Error(w, "Unauthorized", http.StatusForbidden)
		return
	}

	_, err = db.Exec("DELETE FROM tasks WHERE id = ?", taskID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type Session struct {
	UserID int
}

func getSession(r *http.Request) (Session, error) {
	cookie, err := r.Cookie("session")
	if err != nil {
		return Session{}, err
	}

	userID, err := strconv.Atoi(cookie.Value)
	if err != nil {
		return Session{}, err
	}

	return Session{UserID: userID}, nil
}

func getUserByID(id int) (User, error) {
	var user User
	var cohortName sql.NullString
	err := db.QueryRow("SELECT id, username, role, COALESCE(cohort_name, ''), COALESCE(is_suspended, 0) FROM users WHERE id = ?", id).
		Scan(&user.ID, &user.Username, &user.Role, &cohortName, &user.IsSuspended)
	if cohortName.Valid {
		user.CohortName = cohortName.String
	}
	return user, err
}

func getTasks(user User) ([]Task, error) {
	var tasks []Task
	var rows *sql.Rows
	var err error

	if user.Role == "instructor" {
		// Instructors see all tasks they created
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, '')
			FROM tasks 
			WHERE created_by = ?
			ORDER BY created_at DESC
		`, user.ID)
	} else {
		// Students see their own tasks
		rows, err = db.Query(`
			SELECT id, title, description, due_date, status, owner_id, created_by, created_at,
			       COALESCE(task_type, ''), COALESCE(resource_url, ''), COALESCE(cohort_id, '')
			FROM tasks 
			WHERE owner_id = ?
			ORDER BY created_at DESC
		`, user.ID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Track which tasks we've already processed (for grouping cohort/all assignments)
	processedTasks := make(map[string]bool)

	for rows.Next() {
		var task Task
		err := rows.Scan(&task.ID, &task.Title, &task.Description, &task.DueDate,
			&task.Status, &task.OwnerID, &task.CreatedBy, &task.CreatedAt,
			&task.TaskType, &task.ResourceURL, &task.CohortID)
		if err != nil {
			return nil, err
		}

		// For instructors, populate additional statistics
		if user.Role == "instructor" {
			// Check if task is assigned to self
			task.AssignedToSelf = task.OwnerID == user.ID

			// Create a unique key for this task assignment
			taskKey := fmt.Sprintf("%s_%s_%s", task.Title, task.CohortID, task.CreatedAt.Format("2006-01-02 15:04:05"))

			// If not assigned to self and we haven't processed this task group yet
			if !task.AssignedToSelf && !processedTasks[taskKey] {
				// Query all related tasks (same title, created_by, and cohort_id)
				query := `
					SELECT status
					FROM tasks
					WHERE title = ? AND created_by = ? AND owner_id != created_by
				`
				args := []interface{}{task.Title, task.CreatedBy}

				if task.CohortID != "" {
					query += " AND cohort_id = ?"
					args = append(args, task.CohortID)
				}

				statusRows, err := db.Query(query, args...)
				if err == nil {
					defer statusRows.Close()
					for statusRows.Next() {
						var status string
						if statusRows.Scan(&status) == nil {
							task.TotalStudents++
							switch status {
							case "completed":
								task.CompletedStudents++
							case "in_progress":
								task.InProgressStudents++
							case "pending":
								task.PendingStudents++
							}
						}
					}
				}

				processedTasks[taskKey] = true
			}
		}

		tasks = append(tasks, task)
	}

	// For instructors, deduplicate tasks that are assigned to multiple students
	// Keep only one representative task for each group assignment
	if user.Role == "instructor" {
		uniqueTasks := make(map[string]Task)
		for _, task := range tasks {
			if task.AssignedToSelf {
				// Keep all self-assigned tasks
				key := fmt.Sprintf("self_%d", task.ID)
				uniqueTasks[key] = task
			} else {
				// For student assignments, use title+cohort as key to group them
				key := fmt.Sprintf("%s_%s", task.Title, task.CohortID)
				// Keep the first occurrence (or update if we have better stats)
				if existing, exists := uniqueTasks[key]; !exists || existing.TotalStudents == 0 {
					uniqueTasks[key] = task
				}
			}
		}

		// Convert back to slice
		tasks = make([]Task, 0, len(uniqueTasks))
		for _, task := range uniqueTasks {
			tasks = append(tasks, task)
		}
	}

	return tasks, nil
}
