package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// SnapshotRepository implements Repository for SnapshotModel using SQLite.
// Manages follower/following snapshots with entries for diff and historical comparison.
type SnapshotRepository struct {
	db *sql.DB
}

// NewSnapshotRepository creates a new snapshot repository with SQLite backend
func NewSnapshotRepository() (*SnapshotRepository, error) {
	dbPath, err := config.GetCacheDB()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &SnapshotRepository{db: db}, nil
}

// Init ensures database schema is initialized via migrations
func (r *SnapshotRepository) Init(ctx context.Context) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	return RunMigrations(r.db)
}

// Close releases database connection
func (r *SnapshotRepository) Close() error {
	return r.db.Close()
}

// Get retrieves a snapshot by ID
func (r *SnapshotRepository) Get(ctx context.Context, id string) (Model, error) {
	query := `
		SELECT id, created_at, user_did, snapshot_type, total_count, expires_at
		FROM follower_snapshots
		WHERE id = ?
	`

	var snapshot SnapshotModel
	var snapshotID string
	var createdAt, expiresAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&snapshotID,
		&createdAt,
		&snapshot.UserDid,
		&snapshot.SnapshotType,
		&snapshot.TotalCount,
		&expiresAt,
	)

	snapshot.SetID(snapshotID)
	snapshot.SetCreatedAt(createdAt)
	snapshot.ExpiresAt = expiresAt

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RepositoryError{Op: "Get", Err: errors.New("snapshot not found")}
		}
		return nil, &RepositoryError{Op: "Get", Err: err}
	}

	return &snapshot, nil
}

// List retrieves all snapshots ordered by creation date (newest first)
func (r *SnapshotRepository) List(ctx context.Context) ([]Model, error) {
	query := `
		SELECT id, created_at, user_did, snapshot_type, total_count, expires_at
		FROM follower_snapshots
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &RepositoryError{Op: "List", Err: err}
	}
	defer rows.Close()

	var snapshots []Model
	for rows.Next() {
		var snapshot SnapshotModel
		var snapshotID string
		var createdAt, expiresAt time.Time

		err := rows.Scan(
			&snapshotID,
			&createdAt,
			&snapshot.UserDid,
			&snapshot.SnapshotType,
			&snapshot.TotalCount,
			&expiresAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "List", Err: err}
		}

		snapshot.SetID(snapshotID)
		snapshot.SetCreatedAt(createdAt)
		snapshot.ExpiresAt = expiresAt
		snapshots = append(snapshots, &snapshot)
	}

	return snapshots, rows.Err()
}

// Save creates a new snapshot (snapshots are immutable, no updates)
func (r *SnapshotRepository) Save(ctx context.Context, model Model) error {
	snapshot, ok := model.(*SnapshotModel)
	if !ok {
		return &RepositoryError{Op: "Save", Err: errors.New("invalid model type: expected *SnapshotModel")}
	}

	if snapshot.ID() == "" {
		snapshot.SetID(GenerateUUID())
		snapshot.SetCreatedAt(time.Now())
	}

	if snapshot.ExpiresAt.IsZero() {
		snapshot.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	query := `
		INSERT INTO follower_snapshots (id, created_at, user_did, snapshot_type, total_count, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query,
		snapshot.ID(),
		snapshot.CreatedAt(),
		snapshot.UserDid,
		snapshot.SnapshotType,
		snapshot.TotalCount,
		snapshot.ExpiresAt,
	)

	if err != nil {
		return &RepositoryError{Op: "Save", Err: err}
	}

	return nil
}

// Delete removes a snapshot by ID (cascade deletes entries)
func (r *SnapshotRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM follower_snapshots WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	if rows == 0 {
		return &RepositoryError{Op: "Delete", Err: errors.New("snapshot not found")}
	}

	return nil
}

// FindByUserAndType retrieves the most recent fresh snapshot for a user and type.
func (r *SnapshotRepository) FindByUserAndType(ctx context.Context, userDid, snapshotType string) (*SnapshotModel, error) {
	query := `
		SELECT id, created_at, user_did, snapshot_type, total_count, expires_at
		FROM follower_snapshots
		WHERE user_did = ? AND snapshot_type = ? AND expires_at > ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var snapshot SnapshotModel
	var snapshotID string
	var createdAt, expiresAt time.Time

	err := r.db.QueryRowContext(ctx, query, userDid, snapshotType, time.Now()).Scan(
		&snapshotID,
		&createdAt,
		&snapshot.UserDid,
		&snapshot.SnapshotType,
		&snapshot.TotalCount,
		&expiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, &RepositoryError{Op: "FindByUserAndType", Err: err}
	}

	snapshot.SetID(snapshotID)
	snapshot.SetCreatedAt(createdAt)
	snapshot.ExpiresAt = expiresAt

	return &snapshot, nil
}

// FindByUserTypeAndDate retrieves a snapshot for a user, type, and specific date, closest to (but not after) the specified date.
func (r *SnapshotRepository) FindByUserTypeAndDate(ctx context.Context, userDid, snapshotType string, date time.Time) (*SnapshotModel, error) {
	query := `
		SELECT id, created_at, user_did, snapshot_type, total_count, expires_at
		FROM follower_snapshots
		WHERE user_did = ? AND snapshot_type = ? AND created_at <= ?
		ORDER BY created_at DESC
		LIMIT 1
	`

	var snapshot SnapshotModel
	var snapshotID string
	var createdAt, expiresAt time.Time

	err := r.db.QueryRowContext(ctx, query, userDid, snapshotType, date).Scan(
		&snapshotID,
		&createdAt,
		&snapshot.UserDid,
		&snapshot.SnapshotType,
		&snapshot.TotalCount,
		&expiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, &RepositoryError{Op: "FindByUserTypeAndDate", Err: err}
	}

	snapshot.SetID(snapshotID)
	snapshot.SetCreatedAt(createdAt)
	snapshot.ExpiresAt = expiresAt
	return &snapshot, nil
}

// SaveEntry saves a single snapshot entry
func (r *SnapshotRepository) SaveEntry(ctx context.Context, entry *SnapshotEntry) error {
	query := `
		INSERT INTO follower_snapshot_entries (snapshot_id, actor_did, indexed_at)
		VALUES (?, ?, ?)
	`

	_, err := r.db.ExecContext(ctx, query, entry.SnapshotID, entry.ActorDid, entry.IndexedAt)
	if err != nil {
		return &RepositoryError{Op: "SaveEntry", Err: err}
	}
	return nil
}

// SaveEntries saves multiple snapshot entries in a transaction for efficiency
func (r *SnapshotRepository) SaveEntries(ctx context.Context, entries []*SnapshotEntry) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return &RepositoryError{Op: "SaveEntries", Err: err}
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO follower_snapshot_entries (snapshot_id, actor_did, indexed_at)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return &RepositoryError{Op: "SaveEntries", Err: err}
	}
	defer stmt.Close()

	for _, entry := range entries {
		_, err := stmt.ExecContext(ctx, entry.SnapshotID, entry.ActorDid, entry.IndexedAt)
		if err != nil {
			return &RepositoryError{Op: "SaveEntries", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &RepositoryError{Op: "SaveEntries", Err: err}
	}
	return nil
}

// GetEntries retrieves all entries for a snapshot
func (r *SnapshotRepository) GetEntries(ctx context.Context, snapshotID string) ([]*SnapshotEntry, error) {
	query := `
		SELECT snapshot_id, actor_did, indexed_at
		FROM follower_snapshot_entries
		WHERE snapshot_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, &RepositoryError{Op: "GetEntries", Err: err}
	}
	defer rows.Close()

	var entries []*SnapshotEntry
	for rows.Next() {
		var entry SnapshotEntry
		err := rows.Scan(&entry.SnapshotID, &entry.ActorDid, &entry.IndexedAt)
		if err != nil {
			return nil, &RepositoryError{Op: "GetEntries", Err: err}
		}
		entries = append(entries, &entry)
	}
	return entries, rows.Err()
}

// GetActorDids retrieves just the actor DIDs for a snapshot (efficient for diffs)
func (r *SnapshotRepository) GetActorDids(ctx context.Context, snapshotID string) ([]string, error) {
	query := `
		SELECT actor_did
		FROM follower_snapshot_entries
		WHERE snapshot_id = ?
	`

	rows, err := r.db.QueryContext(ctx, query, snapshotID)
	if err != nil {
		return nil, &RepositoryError{Op: "GetActorDids", Err: err}
	}
	defer rows.Close()

	var dids []string
	for rows.Next() {
		var did string
		err := rows.Scan(&did)
		if err != nil {
			return nil, &RepositoryError{Op: "GetActorDids", Err: err}
		}
		dids = append(dids, did)
	}
	return dids, rows.Err()
}

// DeleteExpiredSnapshots removes all expired snapshots and their entries
func (r *SnapshotRepository) DeleteExpiredSnapshots(ctx context.Context) (int64, error) {
	query := "DELETE FROM follower_snapshots WHERE expires_at < ?"
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredSnapshots", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredSnapshots", Err: err}
	}
	return rows, nil
}
