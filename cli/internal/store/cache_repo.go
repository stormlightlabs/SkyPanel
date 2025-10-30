package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// CacheRepository manages post rate and activity caches using SQLite.
//
// Provides methods for storing and retrieving expensive computation results stored as [PostRateCacheModel] or [ActivityCacheModel].
type CacheRepository struct {
	db *sql.DB
}

// NewCacheRepository creates a new cache repository with SQLite backend
func NewCacheRepository() (*CacheRepository, error) {
	dbPath, err := config.GetCacheDB()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &CacheRepository{db: db}, nil
}

// Init ensures database schema is initialized via migrations
func (r *CacheRepository) Init(ctx context.Context) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	return RunMigrations(r.db)
}

// Close releases database connection
func (r *CacheRepository) Close() error {
	return r.db.Close()
}

// GetPostRate retrieves cached post rate for an actor
func (r *CacheRepository) GetPostRate(ctx context.Context, actorDid string) (*PostRateCacheModel, error) {
	query := `
		SELECT actor_did, posts_per_day, last_post_date, sample_size, fetched_at, expires_at
		FROM cached_post_rates
		WHERE actor_did = ? AND expires_at > ?
	`

	var cache PostRateCacheModel
	var lastPostDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, actorDid, time.Now()).Scan(
		&cache.ActorDid,
		&cache.PostsPerDay,
		&lastPostDate,
		&cache.SampleSize,
		&cache.FetchedAt,
		&cache.ExpiresAt,
	)

	if lastPostDate.Valid {
		cache.LastPostDate = lastPostDate.Time
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, &RepositoryError{Op: "GetPostRate", Err: err}
	}

	return &cache, nil
}

// GetPostRates retrieves cached post rates for multiple actors in a single query,
// as a map of actorDid -> PostRateCacheModel for found entries.
func (r *CacheRepository) GetPostRates(ctx context.Context, actorDids []string) (map[string]*PostRateCacheModel, error) {
	if len(actorDids) == 0 {
		return make(map[string]*PostRateCacheModel), nil
	}

	query := `
		SELECT actor_did, posts_per_day, last_post_date, sample_size, fetched_at, expires_at
		FROM cached_post_rates
		WHERE actor_did IN (` + buildPlaceholders(len(actorDids)) + `) AND expires_at > ?
	`

	args := make([]interface{}, len(actorDids)+1)
	for i, did := range actorDids {
		args[i] = did
	}
	args[len(actorDids)] = time.Now()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &RepositoryError{Op: "GetPostRates", Err: err}
	}
	defer rows.Close()

	result := make(map[string]*PostRateCacheModel)
	for rows.Next() {
		var cache PostRateCacheModel
		var lastPostDate sql.NullTime

		err := rows.Scan(
			&cache.ActorDid,
			&cache.PostsPerDay,
			&lastPostDate,
			&cache.SampleSize,
			&cache.FetchedAt,
			&cache.ExpiresAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "GetPostRates", Err: err}
		}

		if lastPostDate.Valid {
			cache.LastPostDate = lastPostDate.Time
		}

		result[cache.ActorDid] = &cache
	}

	return result, rows.Err()
}

// SavePostRate saves or updates a post rate cache entry
func (r *CacheRepository) SavePostRate(ctx context.Context, cache *PostRateCacheModel) error {
	if cache.FetchedAt.IsZero() {
		cache.FetchedAt = time.Now()
	}
	if cache.ExpiresAt.IsZero() {
		cache.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	query := `
		INSERT INTO cached_post_rates (actor_did, posts_per_day, last_post_date, sample_size, fetched_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(actor_did) DO UPDATE SET
			posts_per_day = excluded.posts_per_day,
			last_post_date = excluded.last_post_date,
			sample_size = excluded.sample_size,
			fetched_at = excluded.fetched_at,
			expires_at = excluded.expires_at
	`

	var lastPostDate interface{}
	if !cache.LastPostDate.IsZero() {
		lastPostDate = cache.LastPostDate
	}

	_, err := r.db.ExecContext(ctx, query,
		cache.ActorDid,
		cache.PostsPerDay,
		lastPostDate,
		cache.SampleSize,
		cache.FetchedAt,
		cache.ExpiresAt,
	)

	if err != nil {
		return &RepositoryError{Op: "SavePostRate", Err: err}
	}

	return nil
}

// SavePostRates saves multiple post rate cache entries in a transaction
func (r *CacheRepository) SavePostRates(ctx context.Context, caches []*PostRateCacheModel) error {
	if len(caches) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return &RepositoryError{Op: "SavePostRates", Err: err}
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO cached_post_rates (actor_did, posts_per_day, last_post_date, sample_size, fetched_at, expires_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(actor_did) DO UPDATE SET
			posts_per_day = excluded.posts_per_day,
			last_post_date = excluded.last_post_date,
			sample_size = excluded.sample_size,
			fetched_at = excluded.fetched_at,
			expires_at = excluded.expires_at
	`)
	if err != nil {
		return &RepositoryError{Op: "SavePostRates", Err: err}
	}
	defer stmt.Close()

	for _, cache := range caches {
		if cache.FetchedAt.IsZero() {
			cache.FetchedAt = time.Now()
		}
		if cache.ExpiresAt.IsZero() {
			cache.ExpiresAt = time.Now().Add(24 * time.Hour)
		}

		var lastPostDate interface{}
		if !cache.LastPostDate.IsZero() {
			lastPostDate = cache.LastPostDate
		}

		_, err := stmt.ExecContext(ctx,
			cache.ActorDid,
			cache.PostsPerDay,
			lastPostDate,
			cache.SampleSize,
			cache.FetchedAt,
			cache.ExpiresAt,
		)
		if err != nil {
			return &RepositoryError{Op: "SavePostRates", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &RepositoryError{Op: "SavePostRates", Err: err}
	}

	return nil
}

// DeletePostRate removes a post rate cache entry
func (r *CacheRepository) DeletePostRate(ctx context.Context, actorDid string) error {
	query := "DELETE FROM cached_post_rates WHERE actor_did = ?"
	_, err := r.db.ExecContext(ctx, query, actorDid)
	if err != nil {
		return &RepositoryError{Op: "DeletePostRate", Err: err}
	}
	return nil
}

// GetActivity retrieves cached activity data for an actor
func (r *CacheRepository) GetActivity(ctx context.Context, actorDid string) (*ActivityCacheModel, error) {
	query := `
		SELECT actor_did, last_post_date, fetched_at, expires_at
		FROM cached_activity
		WHERE actor_did = ? AND expires_at > ?
	`

	var cache ActivityCacheModel
	var lastPostDate sql.NullTime

	err := r.db.QueryRowContext(ctx, query, actorDid, time.Now()).Scan(
		&cache.ActorDid,
		&lastPostDate,
		&cache.FetchedAt,
		&cache.ExpiresAt,
	)

	if lastPostDate.Valid {
		cache.LastPostDate = lastPostDate.Time
	}

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, &RepositoryError{Op: "GetActivity", Err: err}
	}

	return &cache, nil
}

// GetActivities retrieves cached activity data for multiple actors in a single query,
// as a map of actorDid -> ActivityCacheModel for found entries.
func (r *CacheRepository) GetActivities(ctx context.Context, actorDids []string) (map[string]*ActivityCacheModel, error) {
	if len(actorDids) == 0 {
		return make(map[string]*ActivityCacheModel), nil
	}

	query := `
		SELECT actor_did, last_post_date, fetched_at, expires_at
		FROM cached_activity
		WHERE actor_did IN (` + buildPlaceholders(len(actorDids)) + `) AND expires_at > ?
	`

	args := make([]interface{}, len(actorDids)+1)
	for i, did := range actorDids {
		args[i] = did
	}
	args[len(actorDids)] = time.Now()

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, &RepositoryError{Op: "GetActivities", Err: err}
	}
	defer rows.Close()

	result := make(map[string]*ActivityCacheModel)
	for rows.Next() {
		var cache ActivityCacheModel
		var lastPostDate sql.NullTime

		err := rows.Scan(
			&cache.ActorDid,
			&lastPostDate,
			&cache.FetchedAt,
			&cache.ExpiresAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "GetActivities", Err: err}
		}

		if lastPostDate.Valid {
			cache.LastPostDate = lastPostDate.Time
		}

		result[cache.ActorDid] = &cache
	}

	return result, rows.Err()
}

// SaveActivity saves or updates an activity cache entry
func (r *CacheRepository) SaveActivity(ctx context.Context, cache *ActivityCacheModel) error {
	if cache.FetchedAt.IsZero() {
		cache.FetchedAt = time.Now()
	}
	if cache.ExpiresAt.IsZero() {
		cache.ExpiresAt = time.Now().Add(24 * time.Hour)
	}

	query := `
		INSERT INTO cached_activity (actor_did, last_post_date, fetched_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(actor_did) DO UPDATE SET
			last_post_date = excluded.last_post_date,
			fetched_at = excluded.fetched_at,
			expires_at = excluded.expires_at
	`

	var lastPostDate interface{}
	if !cache.LastPostDate.IsZero() {
		lastPostDate = cache.LastPostDate
	}

	_, err := r.db.ExecContext(ctx, query,
		cache.ActorDid,
		lastPostDate,
		cache.FetchedAt,
		cache.ExpiresAt,
	)

	if err != nil {
		return &RepositoryError{Op: "SaveActivity", Err: err}
	}

	return nil
}

// SaveActivities saves multiple activity cache entries in a transaction
func (r *CacheRepository) SaveActivities(ctx context.Context, caches []*ActivityCacheModel) error {
	if len(caches) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return &RepositoryError{Op: "SaveActivities", Err: err}
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO cached_activity (actor_did, last_post_date, fetched_at, expires_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(actor_did) DO UPDATE SET
			last_post_date = excluded.last_post_date,
			fetched_at = excluded.fetched_at,
			expires_at = excluded.expires_at
	`)
	if err != nil {
		return &RepositoryError{Op: "SaveActivities", Err: err}
	}
	defer stmt.Close()

	for _, cache := range caches {
		if cache.FetchedAt.IsZero() {
			cache.FetchedAt = time.Now()
		}
		if cache.ExpiresAt.IsZero() {
			cache.ExpiresAt = time.Now().Add(24 * time.Hour)
		}

		var lastPostDate interface{}
		if !cache.LastPostDate.IsZero() {
			lastPostDate = cache.LastPostDate
		}

		_, err := stmt.ExecContext(ctx,
			cache.ActorDid,
			lastPostDate,
			cache.FetchedAt,
			cache.ExpiresAt,
		)
		if err != nil {
			return &RepositoryError{Op: "SaveActivities", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &RepositoryError{Op: "SaveActivities", Err: err}
	}

	return nil
}

// DeleteActivity removes an activity cache entry
func (r *CacheRepository) DeleteActivity(ctx context.Context, actorDid string) error {
	query := "DELETE FROM cached_activity WHERE actor_did = ?"
	_, err := r.db.ExecContext(ctx, query, actorDid)
	if err != nil {
		return &RepositoryError{Op: "DeleteActivity", Err: err}
	}
	return nil
}

// DeleteExpiredPostRates removes all expired post rate cache entries
func (r *CacheRepository) DeleteExpiredPostRates(ctx context.Context) (int64, error) {
	query := "DELETE FROM cached_post_rates WHERE expires_at < ?"
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredPostRates", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredPostRates", Err: err}
	}

	return rows, nil
}

// DeleteExpiredActivities removes all expired activity cache entries
func (r *CacheRepository) DeleteExpiredActivities(ctx context.Context) (int64, error) {
	query := "DELETE FROM cached_activity WHERE expires_at < ?"
	result, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredActivities", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, &RepositoryError{Op: "DeleteExpiredActivities", Err: err}
	}

	return rows, nil
}

// buildPlaceholders generates SQL placeholder string for IN queries.
//
// Example: buildPlaceholders(3) returns "?,?,?"
func buildPlaceholders(count int) string {
	if count == 0 {
		return ""
	}

	placeholders := "?"
	for i := 1; i < count; i++ {
		placeholders += ",?"
	}
	return placeholders
}
