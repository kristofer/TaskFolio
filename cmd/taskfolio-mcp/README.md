# taskfolio-mcp

A [Model Context Protocol](https://modelcontextprotocol.io) server that lets an AI assistant read and manage a TaskFolio database: students' boards, cohort progress, and the task-template library.

It runs over stdio and reads TaskFolio's SQLite file directly. Point it at a **copy** of the database, not the live production file.

## Tools

| Tool | Mode | What it does |
|---|---|---|
| `list_cohorts` | read | Cohorts that have current students |
| `list_students` | read | Current students, optionally one cohort |
| `student_board` | read | One student's tasks with status, due date, type and comments |
| `cohort_progress` | read | Task counts by status for each student in a cohort |
| `search_templates` | read | Search the instructor template library |
| `assign_template` | write | Assign a template to a cohort or listed students |
| `create_task` | write | Create one task for one student |
| `update_task_status` | write | Move a task between pending, in_progress, blocked and completed |

Only current students (`is_current = 1`) are listed or assigned work, matching TaskFolio's own rules. No tool ever returns passwords.

## Safety

- **Read-only by default.** Without `--allow-writes`, the database is opened with SQLite's `mode=ro` and the write tools are not registered.
- **Writes are attributed.** `--allow-writes` requires `--as <username>`, which must be an active instructor account. Every task created is recorded with that instructor as `created_by`, just like the web UI.
- **No accidental databases.** The server refuses a path that is not an existing TaskFolio database.

## Build

```sh
go build -o taskfolio-mcp ./cmd/taskfolio-mcp
```

## Use with Claude Code

Read-only:

```sh
claude mcp add taskfolio -- /path/to/taskfolio-mcp --db /path/to/taskfolio-copy.db
```

With instructor writes:

```sh
claude mcp add taskfolio-admin -- /path/to/taskfolio-mcp --db /path/to/taskfolio-copy.db --allow-writes --as instructor1
```

Then ask things like "Which Java11.2 students have blocked tasks?" or "Assign the recursion lab template to Data6.2, due next Friday."

## Tests

```sh
go test ./cmd/taskfolio-mcp/
```

The tests build a temporary database from TaskFolio's schema (a copy lives in `store_test.go`; keep it in sync with `initDB` in the root `main.go`) and exercise both the store and the MCP tools through an in-memory client.
