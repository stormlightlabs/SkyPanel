package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// ProfileRepository implements Repository for ProfileModel using SQLite with cache support
type ProfileRepository struct {
	db *sql.DB
}

// NewProfileRepository creates a new profile repository with SQLite backend
func NewProfileRepository() (*ProfileRepository, error) {
	dbPath, err := config.GetCacheDB()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &ProfileRepository{db: db}, nil
}

// Init ensures database schema is initialized via migrations
func (r *ProfileRepository) Init(ctx context.Context) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	return RunMigrations(r.db)
}

// Close releases database connection
func (r *ProfileRepository) Close() error {
	return r.db.Close()
}

// Get retrieves a profile by ID
func (r *ProfileRepository) Get(ctx context.Context, id string) (Model, error) {
	query := `
		SELECT id, created_at, updated_at, did, handle, data_json, fetched_at
		FROM profiles
		WHERE id = ?
	`

	var profile ProfileModel
	var profileID string
	var createdAt, updatedAt, fetchedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&profileID,
		&createdAt,
		&updatedAt,
		&profile.Did,
		&profile.Handle,
		&profile.DataJSON,
		&fetchedAt,
	)

	profile.SetID(profileID)
	profile.SetCreatedAt(createdAt)
	profile.SetUpdatedAt(updatedAt)
	profile.FetchedAt = fetchedAt

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RepositoryError{Op: "Get", Err: errors.New("profile not found")}
		}
		return nil, &RepositoryError{Op: "Get", Err: err}
	}

	return &profile, nil
}

// GetByDid retrieves a profile by DID (primary lookup key for profiles).
// Returns the cached profile if found, nil if not found.
func (r *ProfileRepository) GetByDid(ctx context.Context, did string) (*ProfileModel, error) {
	query := `
		SELECT id, created_at, updated_at, did, handle, data_json, fetched_at
		FROM profiles
		WHERE did = ?
	`

	var profile ProfileModel
	var profileID string
	var createdAt, updatedAt, fetchedAt time.Time

	err := r.db.QueryRowContext(ctx, query, did).Scan(
		&profileID,
		&createdAt,
		&updatedAt,
		&profile.Did,
		&profile.Handle,
		&profile.DataJSON,
		&fetchedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, &RepositoryError{Op: "GetByDid", Err: err}
	}

	profile.SetID(profileID)
	profile.SetCreatedAt(createdAt)
	profile.SetUpdatedAt(updatedAt)
	profile.FetchedAt = fetchedAt

	return &profile, nil
}

// List retrieves all cached profiles
func (r *ProfileRepository) List(ctx context.Context) ([]Model, error) {
	query := `
		SELECT id, created_at, updated_at, did, handle, data_json, fetched_at
		FROM profiles
		ORDER BY fetched_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &RepositoryError{Op: "List", Err: err}
	}
	defer rows.Close()

	var profiles []Model
	for rows.Next() {
		var profile ProfileModel
		var profileID string
		var createdAt, updatedAt, fetchedAt time.Time

		err := rows.Scan(
			&profileID,
			&createdAt,
			&updatedAt,
			&profile.Did,
			&profile.Handle,
			&profile.DataJSON,
			&fetchedAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "List", Err: err}
		}

		profile.SetID(profileID)
		profile.SetCreatedAt(createdAt)
		profile.SetUpdatedAt(updatedAt)
		profile.FetchedAt = fetchedAt

		profiles = append(profiles, &profile)
	}

	return profiles, rows.Err()
}

// Save creates or updates a profile (upsert by DID)
func (r *ProfileRepository) Save(ctx context.Context, model Model) error {
	profile, ok := model.(*ProfileModel)
	if !ok {
		return &RepositoryError{Op: "Save", Err: errors.New("invalid model type: expected *ProfileModel")}
	}

	if profile.ID() == "" {
		profile.SetID(GenerateUUID())
		profile.SetCreatedAt(time.Now())
	}
	profile.SetUpdatedAt(time.Now())

	if profile.FetchedAt.IsZero() {
		profile.FetchedAt = time.Now()
	}

	query := `
		INSERT INTO profiles (id, created_at, updated_at, did, handle, data_json, fetched_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(did) DO UPDATE SET
			updated_at = excluded.updated_at,
			handle = excluded.handle,
			data_json = excluded.data_json,
			fetched_at = excluded.fetched_at
	`

	_, err := r.db.ExecContext(ctx, query,
		profile.ID(),
		profile.CreatedAt(),
		profile.UpdatedAt(),
		profile.Did,
		profile.Handle,
		profile.DataJSON,
		profile.FetchedAt,
	)

	if err != nil {
		return &RepositoryError{Op: "Save", Err: err}
	}

	return nil
}

// Delete removes a profile by ID
func (r *ProfileRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM profiles WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	if rows == 0 {
		return &RepositoryError{Op: "Delete", Err: errors.New("profile not found")}
	}

	return nil
}

// DeleteByDid removes a profile by DID
func (r *ProfileRepository) DeleteByDid(ctx context.Context, did string) error {
	query := "DELETE FROM profiles WHERE did = ?"
	result, err := r.db.ExecContext(ctx, query, did)
	if err != nil {
		return &RepositoryError{Op: "DeleteByDid", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &RepositoryError{Op: "DeleteByDid", Err: err}
	}

	if rows == 0 {
		return &RepositoryError{Op: "DeleteByDid", Err: errors.New("profile not found")}
	}

	return nil
}
