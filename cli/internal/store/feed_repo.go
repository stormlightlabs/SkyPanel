package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// FeedRepository implements Repository for FeedModel using SQLite
type FeedRepository struct {
	db *sql.DB
}

// NewFeedRepository creates a new feed repository with SQLite backend
func NewFeedRepository() (*FeedRepository, error) {
	dbPath, err := config.GetCacheDB()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &FeedRepository{db: db}, nil
}

// Init ensures database schema is initialized via migrations
func (r *FeedRepository) Init(ctx context.Context) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	return RunMigrations(r.db)
}

// Close releases database connection
func (r *FeedRepository) Close() error {
	return r.db.Close()
}

// Get retrieves a feed by ID
func (r *FeedRepository) Get(ctx context.Context, id string) (Model, error) {
	query := `
		SELECT id, created_at, updated_at, name, source, params, is_local
		FROM feeds
		WHERE id = ?
	`

	var feed FeedModel
	var paramsJSON string
	var feedID string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&feedID,
		&createdAt,
		&updatedAt,
		&feed.Name,
		&feed.Source,
		&paramsJSON,
		&feed.IsLocal,
	)

	feed.SetID(feedID)
	feed.SetCreatedAt(createdAt)
	feed.SetUpdatedAt(updatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RepositoryError{Op: "Get", Err: errors.New("feed not found")}
		}
		return nil, &RepositoryError{Op: "Get", Err: err}
	}

	if err := json.Unmarshal([]byte(paramsJSON), &feed.Params); err != nil {
		return nil, &RepositoryError{Op: "UnmarshalParams", Err: err}
	}

	return &feed, nil
}

// List retrieves all feeds
func (r *FeedRepository) List(ctx context.Context) ([]Model, error) {
	query := `
		SELECT id, created_at, updated_at, name, source, params, is_local
		FROM feeds
		ORDER BY created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &RepositoryError{Op: "List", Err: err}
	}
	defer rows.Close()

	var feeds []Model
	for rows.Next() {
		var feed FeedModel
		var paramsJSON string
		var feedID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&feedID,
			&createdAt,
			&updatedAt,
			&feed.Name,
			&feed.Source,
			&paramsJSON,
			&feed.IsLocal,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "List", Err: err}
		}

		feed.SetID(feedID)
		feed.SetCreatedAt(createdAt)
		feed.SetUpdatedAt(updatedAt)

		if err := json.Unmarshal([]byte(paramsJSON), &feed.Params); err != nil {
			return nil, &RepositoryError{Op: "UnmarshalParams", Err: err}
		}

		feeds = append(feeds, &feed)
	}

	return feeds, rows.Err()
}

// Save creates or updates a feed
func (r *FeedRepository) Save(ctx context.Context, model Model) error {
	feed, ok := model.(*FeedModel)
	if !ok {
		return &RepositoryError{Op: "Save", Err: errors.New("invalid model type: expected *FeedModel")}
	}

	if feed.ID() == "" {
		feed.SetID(GenerateUUID())
		feed.SetCreatedAt(time.Now())
	}
	feed.SetUpdatedAt(time.Now())

	paramsJSON, err := json.Marshal(feed.Params)
	if err != nil {
		return &RepositoryError{Op: "MarshalParams", Err: err}
	}

	query := `
		INSERT INTO feeds (id, created_at, updated_at, name, source, params, is_local)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			updated_at = excluded.updated_at,
			name = excluded.name,
			source = excluded.source,
			params = excluded.params,
			is_local = excluded.is_local
	`

	_, err = r.db.ExecContext(ctx, query,
		feed.ID(),
		feed.CreatedAt(),
		feed.UpdatedAt(),
		feed.Name,
		feed.Source,
		string(paramsJSON),
		feed.IsLocal,
	)

	if err != nil {
		return &RepositoryError{Op: "Save", Err: err}
	}

	return nil
}

// Delete removes a feed by ID
func (r *FeedRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM feeds WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	if rows == 0 {
		return &RepositoryError{Op: "Delete", Err: errors.New("feed not found")}
	}

	return nil
}

// RepositoryError represents an error that occurred during repository operations
type RepositoryError struct {
	Op  string
	Err error
}

func (e *RepositoryError) Error() string {
	return "repository." + e.Op + ": " + e.Err.Error()
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}
