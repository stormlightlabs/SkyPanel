package store

import (
	"database/sql"
	"testing"

	"github.com/stormlightlabs/skypanel/cli/internal/utils"
)

func TestRunMigrations(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("schema_migrations table not found: %v", err)
	}

	if count != 4 {
		t.Errorf("expected 4 migrations applied, got %d", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&count)
	if err != nil {
		t.Errorf("feeds table not created: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err != nil {
		t.Errorf("posts table not created: %v", err)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM profiles").Scan(&count)
	if err != nil {
		t.Errorf("profiles table not created: %v", err)
	}
}

func TestRunMigrations_Idempotent(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("first RunMigrations failed: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("second RunMigrations failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}

	if count != 4 {
		t.Errorf("expected 4 migrations, got %d", count)
	}
}

func TestRollback(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if err := Rollback(db, 1); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 migration after rollback, got %d", count)
	}

	err = db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&count)
	if err == nil {
		t.Error("posts table should not exist after rollback")
	}

	err = db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&count)
	if err != nil {
		t.Errorf("feeds table should still exist: %v", err)
	}
}

func TestRollback_Complete(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	if err := Rollback(db, 0); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}

	if count != 0 {
		t.Errorf("expected 0 migrations after complete rollback, got %d", count)
	}

	var feedCount int
	err = db.QueryRow("SELECT COUNT(*) FROM feeds").Scan(&feedCount)
	if err == nil {
		t.Error("feeds table should not exist after complete rollback")
	}

	var postCount int
	err = db.QueryRow("SELECT COUNT(*) FROM posts").Scan(&postCount)
	if err == nil {
		t.Error("posts table should not exist after complete rollback")
	}
}

func TestMigrationOrdering(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	rows, err := db.Query("SELECT version FROM schema_migrations ORDER BY version")
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}
	defer rows.Close()

	expectedVersions := []int{1, 2, 3, 4}
	var actualVersions []int

	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			t.Fatalf("failed to scan version: %v", err)
		}
		actualVersions = append(actualVersions, version)
	}

	if len(actualVersions) != len(expectedVersions) {
		t.Errorf("expected %d versions, got %d", len(expectedVersions), len(actualVersions))
	}

	for i, expected := range expectedVersions {
		if i >= len(actualVersions) || actualVersions[i] != expected {
			t.Errorf("migration %d: expected version %d, got %d", i, expected, actualVersions[i])
		}
	}
}

func TestGetAppliedMigrations(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := createMigrationsTable(db); err != nil {
		t.Fatalf("failed to create migrations table: %v", err)
	}

	_, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?), (?)", 1, 2)
	if err != nil {
		t.Fatalf("failed to insert test migrations: %v", err)
	}

	applied, err := getAppliedMigrations(db)
	if err != nil {
		t.Fatalf("getAppliedMigrations failed: %v", err)
	}

	if !applied[1] {
		t.Error("migration 1 should be applied")
	}
	if !applied[2] {
		t.Error("migration 2 should be applied")
	}
	if applied[3] {
		t.Error("migration 3 should not be applied")
	}
}

func TestLoadMigrations(t *testing.T) {
	upMigrations, err := loadMigrations("up")
	if err != nil {
		t.Fatalf("failed to load up migrations: %v", err)
	}

	if len(upMigrations) != 4 {
		t.Errorf("expected 4 up migrations, got %d", len(upMigrations))
	}

	for i := 1; i < len(upMigrations); i++ {
		if upMigrations[i-1].Version >= upMigrations[i].Version {
			t.Errorf("migrations not sorted: %d >= %d", upMigrations[i-1].Version, upMigrations[i].Version)
		}
	}

	downMigrations, err := loadMigrations("down")
	if err != nil {
		t.Fatalf("failed to load down migrations: %v", err)
	}

	if len(downMigrations) != 4 {
		t.Errorf("expected 4 down migrations, got %d", len(downMigrations))
	}
}

func TestExecuteMigration(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	m := migration{
		Version: 1,
		Name:    "test_migration",
		SQL:     "CREATE TABLE test (id INTEGER PRIMARY KEY, name TEXT)",
	}

	if err := executeMigration(db, m); err != nil {
		t.Fatalf("executeMigration failed: %v", err)
	}

	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM test").Scan(&count)
	if err != nil {
		t.Errorf("test table not created: %v", err)
	}
}

func TestRecordAndRemoveMigration(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := createMigrationsTable(db); err != nil {
		t.Fatalf("failed to create migrations table: %v", err)
	}

	if err := recordMigration(db, 42); err != nil {
		t.Fatalf("recordMigration failed: %v", err)
	}

	var exists bool
	err := db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", 42).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check migration: %v", err)
	}
	if !exists {
		t.Error("migration 42 should be recorded")
	}

	if err := removeMigration(db, 42); err != nil {
		t.Fatalf("removeMigration failed: %v", err)
	}

	err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = ?)", 42).Scan(&exists)
	if err != nil {
		t.Fatalf("failed to check migration: %v", err)
	}
	if exists {
		t.Error("migration 42 should be removed")
	}
}

func TestMigrationWithForeignKey(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	if err := RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	_, err := db.Exec(`
		INSERT INTO feeds (id, created_at, updated_at, name, source, is_local)
		VALUES ('feed1', datetime('now'), datetime('now'), 'Test Feed', 'timeline', 1)
	`)
	if err != nil {
		t.Fatalf("failed to insert feed: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO posts (id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at)
		VALUES ('post1', datetime('now'), datetime('now'), 'at://test', 'did:test', 'Hello', 'feed1', datetime('now'))
	`)
	if err != nil {
		t.Fatalf("failed to insert post: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO posts (id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at)
		VALUES ('post2', datetime('now'), datetime('now'), 'at://test2', 'did:test', 'Hello', 'nonexistent', datetime('now'))
	`)
	if err == nil {
		t.Error("expected foreign key constraint error, got nil")
	}
}

func TestCreateMigrationsTable(t *testing.T) {
	db, cleanup := utils.NewTestDB(t)
	defer cleanup()

	if err := createMigrationsTable(db); err != nil {
		t.Fatalf("createMigrationsTable failed: %v", err)
	}

	rows, err := db.Query("PRAGMA table_info(schema_migrations)")
	if err != nil {
		t.Fatalf("failed to get table info: %v", err)
	}
	defer rows.Close()

	columns := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dfltValue sql.NullString

		if err := rows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			t.Fatalf("failed to scan column info: %v", err)
		}
		columns[name] = true
	}

	if !columns["version"] {
		t.Error("version column missing")
	}
	if !columns["applied_at"] {
		t.Error("applied_at column missing")
	}
}
