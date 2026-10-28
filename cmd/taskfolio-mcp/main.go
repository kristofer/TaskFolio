// Command taskfolio-mcp is a Model Context Protocol server over a TaskFolio SQLite database.
//
// It runs on stdio. By default it opens the database read-only and exposes read tools only.
// With --allow-writes and --as <instructor>, it also exposes instructor write tools, and every
// task it creates is recorded as created by that instructor.
//
// Point it at a copy of the database, not the live production file.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// main parses flags, opens the store and serves MCP over stdio until the client disconnects.
// Exits with status 2 on bad flags and 1 on startup or server errors. Logs go to stderr,
// because stdout carries the MCP protocol.
func main() {
	dbPath := flag.String("db", os.Getenv("DATABASE_PATH"), "path to the TaskFolio SQLite database (default $DATABASE_PATH)")
	allowWrites := flag.Bool("allow-writes", false, "register instructor write tools and open the database read-write")
	as := flag.String("as", "", "instructor username that writes are attributed to (required with --allow-writes)")
	flag.Parse()

	log.SetOutput(os.Stderr)
	if *dbPath == "" {
		fmt.Fprintln(os.Stderr, "taskfolio-mcp: --db is required")
		os.Exit(2)
	}
	if *allowWrites && *as == "" {
		fmt.Fprintln(os.Stderr, "taskfolio-mcp: --allow-writes requires --as <instructor username>")
		os.Exit(2)
	}

	store, err := OpenStore(*dbPath, !*allowWrites)
	if err != nil {
		log.Fatalf("taskfolio-mcp: %v", err)
	}
	defer store.Close()

	cfg := Config{AllowWrites: *allowWrites, InstructorName: *as}
	if *allowWrites {
		if cfg.InstructorID, err = store.ResolveInstructor(*as); err != nil {
			log.Fatalf("taskfolio-mcp: %v", err)
		}
	}

	if err := NewServer(store, cfg).Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatalf("taskfolio-mcp: %v", err)
	}
}
