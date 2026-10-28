package main

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect starts the TaskFolio MCP server over in-memory transports and returns a client session.
// Parameters: t, the test; s, the store; cfg, the server config. Returns the connected client session.
// Fails the test if either side cannot connect.
func connect(t *testing.T, s *Store, cfg Config) *mcp.ClientSession {
	t.Helper()
	ctx := context.Background()
	server := NewServer(s, cfg)
	st, ct := mcp.NewInMemoryTransports()
	if _, err := server.Connect(ctx, st, nil); err != nil {
		t.Fatal(err)
	}
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { cs.Close() })
	return cs
}

// toolNames lists the tool names a session exposes, sorted.
// Parameters: t, the test; cs, the client session. Returns the sorted names. Fails the test on error.
func toolNames(t *testing.T, cs *mcp.ClientSession) []string {
	t.Helper()
	res, err := cs.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	var names []string
	for _, tool := range res.Tools {
		names = append(names, tool.Name)
	}
	sort.Strings(names)
	return names
}

// call invokes a tool and returns its result as JSON text.
// Parameters: t, the test; cs, the session; name, the tool; args, its arguments.
// Returns the concatenated text content and whether the call was a tool error. Fails the test on protocol errors.
func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) (string, bool) {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("CallTool(%s): %v", name, err)
	}
	var b strings.Builder
	for _, c := range res.Content {
		if tc, ok := c.(*mcp.TextContent); ok {
			b.WriteString(tc.Text)
		}
	}
	return b.String(), res.IsError
}

func TestReadOnlyServerExposesOnlyReadTools(t *testing.T) {
	cs := connect(t, newTestStore(t), Config{})
	got := strings.Join(toolNames(t, cs), ",")
	want := "cohort_progress,list_cohorts,list_students,search_templates,student_board"
	if got != want {
		t.Fatalf("tools = %s, want %s", got, want)
	}
}

func TestWriteServerAddsInstructorTools(t *testing.T) {
	cs := connect(t, newTestStore(t), Config{AllowWrites: true, InstructorID: 1, InstructorName: "instructor1"})
	got := strings.Join(toolNames(t, cs), ",")
	want := "assign_template,cohort_progress,create_task,list_cohorts,list_students,search_templates,student_board,update_task_status"
	if got != want {
		t.Fatalf("tools = %s, want %s", got, want)
	}
}

func TestReadToolsNeverExposePasswords(t *testing.T) {
	cs := connect(t, newTestStore(t), Config{})
	calls := map[string]map[string]any{
		"list_cohorts":     {},
		"list_students":    {},
		"student_board":    {"username": "student1"},
		"cohort_progress":  {"cohort": "Java11.2"},
		"search_templates": {},
	}
	for name, args := range calls {
		out, isErr := call(t, cs, name, args)
		if isErr {
			t.Fatalf("%s returned a tool error: %s", name, out)
		}
		if strings.Contains(out, "secret-pw") || strings.Contains(strings.ToLower(out), "password") {
			t.Fatalf("%s output exposes a password field: %s", name, out)
		}
	}
}

func TestStudentBoardTool(t *testing.T) {
	cs := connect(t, newTestStore(t), Config{})
	out, isErr := call(t, cs, "student_board", map[string]any{"username": "student1"})
	if isErr {
		t.Fatal(out)
	}
	var b Board
	if err := json.Unmarshal([]byte(out), &b); err != nil {
		t.Fatalf("not JSON: %v: %s", err, out)
	}
	if len(b.Tasks) != 3 {
		t.Fatalf("board has %d tasks, want 3", len(b.Tasks))
	}
	out, isErr = call(t, cs, "student_board", map[string]any{"username": "ghost"})
	if !isErr {
		t.Fatalf("unknown student should be a tool error, got %s", out)
	}
}

func TestAssignTemplateToolAttributesToInstructor(t *testing.T) {
	s := newTestStore(t)
	cs := connect(t, s, Config{AllowWrites: true, InstructorID: 1, InstructorName: "instructor1"})
	out, isErr := call(t, cs, "assign_template", map[string]any{"template_id": 1, "cohort": "Data6.2", "due_date": "2026-11-20"})
	if isErr {
		t.Fatal(out)
	}
	var res AssignResult
	if err := json.Unmarshal([]byte(out), &res); err != nil || res.Created != 1 {
		t.Fatalf("assign_template = %s (%v)", out, err)
	}
	var createdBy int
	s.db.QueryRow(`SELECT created_by FROM tasks WHERE title = 'Lab: recursion' AND owner_id = 4`).Scan(&createdBy)
	if createdBy != 1 {
		t.Fatalf("created_by = %d, want instructor 1", createdBy)
	}
}
