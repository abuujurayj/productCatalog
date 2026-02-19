package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
)

func main() {
	ctx := context.Background()

	// Get emulator host from env or use default
	emulatorHost := os.Getenv("SPANNER_EMULATOR_HOST")
	if emulatorHost == "" {
		emulatorHost = "localhost:9010"
	}
	fmt.Printf("Connecting to Emulator at %s...\n", emulatorHost)

	// Database path
	projectID := "test-project"
	instanceID := "test-instance"
	databaseID := "test-database"
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", projectID, instanceID, databaseID)

	// Create admin client
	adminClient, err := database.NewDatabaseAdminClient(ctx)
	if err != nil {
		log.Fatalf("Failed to create admin client: %v", err)
	}
	defer adminClient.Close()

	// Check if database exists
	_, err = adminClient.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: dbPath,
	})

	if err != nil {
		// Database doesn't exist, create it
		fmt.Println("Database doesn't exist, creating...")
		op, err := adminClient.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
			Parent:          fmt.Sprintf("projects/%s/instances/%s", projectID, instanceID),
			CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", databaseID),
		})
		if err != nil {
			log.Fatalf("Failed to create database: %v", err)
		}
		
		// Wait for database creation
		if _, err := op.Wait(ctx); err != nil {
			log.Fatalf("Failed waiting for database creation: %v", err)
		}
		fmt.Println("Database created successfully")
	}

	// Read migration files
	files, err := ioutil.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	// Filter and sort SQL files
	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	// Apply each migration
	for _, filename := range sqlFiles {
		fmt.Printf("📂 Processing migration file: %s\n", filename)
		
		content, err := ioutil.ReadFile(filepath.Join("migrations", filename))
		if err != nil {
			log.Fatalf("Failed to read %s: %v", filename, err)
		}

		// Split statements (simple split by semicolon)
		statements := strings.Split(string(content), ";")
		
		for _, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			fmt.Printf("Applying: %s\n", stmt[:min(50, len(stmt))]+"...")
			
			op, err := adminClient.UpdateDatabaseDdl(ctx, &databasepb.UpdateDatabaseDdlRequest{
				Database:   dbPath,
				Statements: []string{stmt},
			})
			
			if err != nil {
				log.Fatalf("Failed to execute statement in %s: %v\nStatement: %s", filename, err, stmt)
			}
			
			// Wait for operation to complete
			if err := op.Wait(ctx); err != nil {
				log.Fatalf("Failed waiting for DDL completion in %s: %v", filename, err)
			}
			
			time.Sleep(1 * time.Second) // Give Spanner time to settle
		}
		
		fmt.Printf("✅ Completed %s\n", filename)
	}

	fmt.Println("🎉 All migrations completed successfully!")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}