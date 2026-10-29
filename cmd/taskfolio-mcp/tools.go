package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Config controls which tools the server exposes and whom writes are attributed to.
type Config struct {
	AllowWrites    bool   // register the instructor write tools
	InstructorID   int    // users.id recorded as created_by on every write
	InstructorName string // shown in tool descriptions
}

// Tool inputs. The json and jsonschema tags become each tool's input schema.
type (
	noInput      struct{}
	cohortFilter struct {
		Cohort string `json:"cohort,omitempty" jsonschema:"cohort name such as Java11.2; omit for all current students"`
	}
	usernameInput struct {
		Username string `json:"username" jsonschema:"the student's TaskFolio username"`
	}
	cohortInput struct {
		Cohort string `json:"cohort" jsonschema:"cohort name such as Java11.2"`
	}
	queryInput struct {
		Query string `json:"query,omitempty" jsonschema:"text to match in template title or description; omit for the 50 newest"`
	}
	assignArgs struct {
		TemplateID  int      `json:"template_id" jsonschema:"id from search_templates"`
		Cohort      string   `json:"cohort,omitempty" jsonschema:"assign to every current student in this cohort (use this or usernames)"`
		Usernames   []string `json:"usernames,omitempty" jsonschema:"assign to these students (use this or cohort)"`
		DueDate     string   `json:"due_date,omitempty" jsonschema:"due date, YYYY-MM-DD"`
		Title       string   `json:"title,omitempty" jsonschema:"override the template title"`
		Description string   `json:"description,omitempty" jsonschema:"override the template description"`
		ResourceURL string   `json:"resource_url,omitempty" jsonschema:"override the template resource URL"`
	}
	createArgs struct {
		Username    string `json:"username" jsonschema:"the student who will own the task"`
		Title       string `json:"title"`
		Description string `json:"description,omitempty"`
		DueDate     string `json:"due_date,omitempty" jsonschema:"due date, YYYY-MM-DD"`
		TaskType    string `json:"task_type,omitempty" jsonschema:"one of group-programming-project, individual-programming-lab, reading, document, resume, prodev, demo-presentation, research-analysis-design, organize"`
		ResourceURL string `json:"resource_url,omitempty"`
	}
	statusArgs struct {
		TaskID int    `json:"task_id"`
		Status string `json:"status" jsonschema:"pending, in_progress, blocked or completed"`
	}
)

// Tool outputs. Structured output must be a JSON object, so lists are wrapped.
type (
	cohortsOut struct {
		Cohorts []string `json:"cohorts"`
	}
	studentsOut struct {
		Students []Student `json:"students"`
	}
	progressOut struct {
		Students []Progress `json:"students"`
	}
	templatesOut struct {
		Templates []Template `json:"templates"`
	}
	createdOut struct {
		TaskID int64 `json:"task_id"`
	}
	statusOut struct {
		TaskID int    `json:"task_id"`
		Status string `json:"status"`
	}
)

// NewServer builds the TaskFolio MCP server. Read tools are always registered;
// write tools only when cfg.AllowWrites is true.
// Parameters: s, the store; cfg, the configuration. Returns the server, ready to Run or Connect.
func NewServer(s *Store, cfg Config) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "taskfolio", Version: "0.1.0"}, nil)
	addReadTools(server, s)
	if cfg.AllowWrites {
		addWriteTools(server, s, cfg)
	}
	return server
}

// readOnly is the annotation shared by every read tool.
var readOnly = &mcp.ToolAnnotations{ReadOnlyHint: true}

// addReadTools registers the five read-only tools on server, backed by s.
func addReadTools(server *mcp.Server, s *Store) {
	mcp.AddTool(server, &mcp.Tool{Name: "list_cohorts", Description: "List the cohorts that have current students.", Annotations: readOnly},
		func(ctx context.Context, _ *mcp.CallToolRequest, _ noInput) (*mcp.CallToolResult, cohortsOut, error) {
			c, err := s.ListCohorts()
			return nil, cohortsOut{c}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "list_students", Description: "List current students, optionally in one cohort.", Annotations: readOnly},
		func(ctx context.Context, _ *mcp.CallToolRequest, in cohortFilter) (*mcp.CallToolResult, studentsOut, error) {
			st, err := s.ListStudents(in.Cohort)
			return nil, studentsOut{st}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "student_board", Description: "Show one student's TaskFolio board: every task with status, due date, type and comments.", Annotations: readOnly},
		func(ctx context.Context, _ *mcp.CallToolRequest, in usernameInput) (*mcp.CallToolResult, Board, error) {
			b, err := s.StudentBoard(in.Username)
			return nil, b, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "cohort_progress", Description: "Task counts by status (pending, in_progress, blocked, completed) for each current student in a cohort.", Annotations: readOnly},
		func(ctx context.Context, _ *mcp.CallToolRequest, in cohortInput) (*mcp.CallToolResult, progressOut, error) {
			p, err := s.CohortProgress(in.Cohort)
			return nil, progressOut{p}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "search_templates", Description: "Search the instructor task-template library.", Annotations: readOnly},
		func(ctx context.Context, _ *mcp.CallToolRequest, in queryInput) (*mcp.CallToolResult, templatesOut, error) {
			t, err := s.SearchTemplates(in.Query)
			return nil, templatesOut{t}, err
		})
}

// addWriteTools registers the instructor write tools; every write is attributed to cfg.InstructorID.
func addWriteTools(server *mcp.Server, s *Store, cfg Config) {
	notDestructive := false
	writes := &mcp.ToolAnnotations{DestructiveHint: &notDestructive}
	by := fmt.Sprintf(" Tasks are recorded as created by instructor %s.", cfg.InstructorName)

	mcp.AddTool(server, &mcp.Tool{Name: "assign_template", Description: "Assign a task template to a whole cohort or to listed students; creates one pending task per student." + by, Annotations: writes},
		func(ctx context.Context, _ *mcp.CallToolRequest, in assignArgs) (*mcp.CallToolResult, AssignResult, error) {
			res, err := s.AssignTemplate(AssignInput{InstructorID: cfg.InstructorID, TemplateID: in.TemplateID, Cohort: in.Cohort,
				Usernames: in.Usernames, DueDate: in.DueDate, Title: in.Title, Description: in.Description, ResourceURL: in.ResourceURL})
			return nil, res, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "create_task", Description: "Create one pending task for one current student." + by, Annotations: writes},
		func(ctx context.Context, _ *mcp.CallToolRequest, in createArgs) (*mcp.CallToolResult, createdOut, error) {
			id, err := s.CreateTask(CreateTaskInput{InstructorID: cfg.InstructorID, Username: in.Username, Title: in.Title,
				Description: in.Description, DueDate: in.DueDate, TaskType: in.TaskType, ResourceURL: in.ResourceURL})
			return nil, createdOut{id}, err
		})
	mcp.AddTool(server, &mcp.Tool{Name: "update_task_status", Description: "Move a task to pending, in_progress, blocked or completed, stamping start and done dates the way TaskFolio does.", Annotations: writes},
		func(ctx context.Context, _ *mcp.CallToolRequest, in statusArgs) (*mcp.CallToolResult, statusOut, error) {
			err := s.UpdateTaskStatus(in.TaskID, in.Status)
			return nil, statusOut(in), err
		})
}
