package store

import (
	"context"
	"database/sql"
	"errors"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/stormlightlabs/skypanel/cli/internal/config"
)

// PostRepository implements Repository for PostModel using SQLite with batch operations
type PostRepository struct {
	db *sql.DB
}

// NewPostRepository creates a new post repository with SQLite backend
func NewPostRepository() (*PostRepository, error) {
	dbPath, err := config.GetCacheDB()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	return &PostRepository{db: db}, nil
}

// Init ensures database schema is initialized via migrations
func (r *PostRepository) Init(ctx context.Context) error {
	if err := config.EnsureConfigDir(); err != nil {
		return err
	}
	return RunMigrations(r.db)
}

// Close releases database connection
func (r *PostRepository) Close() error {
	return r.db.Close()
}

// Get retrieves a post by ID
func (r *PostRepository) Get(ctx context.Context, id string) (Model, error) {
	query := `
		SELECT id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at
		FROM posts
		WHERE id = ?
	`

	var post PostModel
	var postID string
	var createdAt, updatedAt time.Time

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&postID,
		&createdAt,
		&updatedAt,
		&post.URI,
		&post.AuthorDID,
		&post.Text,
		&post.FeedID,
		&post.IndexedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &RepositoryError{Op: "Get", Err: errors.New("post not found")}
		}
		return nil, &RepositoryError{Op: "Get", Err: err}
	}

	post.SetID(postID)
	post.SetCreatedAt(createdAt)
	post.SetUpdatedAt(updatedAt)

	return &post, nil
}

// List retrieves all posts ordered by indexed_at descending
func (r *PostRepository) List(ctx context.Context) ([]Model, error) {
	query := `
		SELECT id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at
		FROM posts
		ORDER BY indexed_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, &RepositoryError{Op: "List", Err: err}
	}
	defer rows.Close()

	var posts []Model
	for rows.Next() {
		var post PostModel
		var postID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&postID,
			&createdAt,
			&updatedAt,
			&post.URI,
			&post.AuthorDID,
			&post.Text,
			&post.FeedID,
			&post.IndexedAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "List", Err: err}
		}

		post.SetID(postID)
		post.SetCreatedAt(createdAt)
		post.SetUpdatedAt(updatedAt)

		posts = append(posts, &post)
	}

	return posts, rows.Err()
}

// Save creates or updates a post
func (r *PostRepository) Save(ctx context.Context, model Model) error {
	post, ok := model.(*PostModel)
	if !ok {
		return &RepositoryError{Op: "Save", Err: errors.New("invalid model type: expected *PostModel")}
	}

	if post.ID() == "" {
		post.SetID(GenerateUUID())
		post.SetCreatedAt(time.Now())
	}
	post.SetUpdatedAt(time.Now())

	query := `
		INSERT INTO posts (id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			updated_at = excluded.updated_at,
			text = excluded.text,
			feed_id = excluded.feed_id
	`

	_, err := r.db.ExecContext(ctx, query,
		post.ID(),
		post.CreatedAt(),
		post.UpdatedAt(),
		post.URI,
		post.AuthorDID,
		post.Text,
		post.FeedID,
		post.IndexedAt,
	)

	if err != nil {
		return &RepositoryError{Op: "Save", Err: err}
	}

	return nil
}

// Delete removes a post by ID
func (r *PostRepository) Delete(ctx context.Context, id string) error {
	query := "DELETE FROM posts WHERE id = ?"
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &RepositoryError{Op: "Delete", Err: err}
	}

	if rows == 0 {
		return &RepositoryError{Op: "Delete", Err: errors.New("post not found")}
	}

	return nil
}

// BatchSave efficiently saves multiple posts in a single transaction
func (r *PostRepository) BatchSave(ctx context.Context, posts []*PostModel) error {
	if len(posts) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return &RepositoryError{Op: "BatchSave", Err: err}
	}
	defer tx.Rollback()

	query := `
		INSERT INTO posts (id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(uri) DO UPDATE SET
			updated_at = excluded.updated_at,
			text = excluded.text,
			feed_id = excluded.feed_id
	`

	stmt, err := tx.PrepareContext(ctx, query)
	if err != nil {
		return &RepositoryError{Op: "BatchSave", Err: err}
	}
	defer stmt.Close()

	now := time.Now()
	for _, post := range posts {
		if post.ID() == "" {
			post.SetID(GenerateUUID())
			post.SetCreatedAt(now)
		}
		post.SetUpdatedAt(now)

		_, err := stmt.ExecContext(ctx,
			post.ID(),
			post.CreatedAt(),
			post.UpdatedAt(),
			post.URI,
			post.AuthorDID,
			post.Text,
			post.FeedID,
			post.IndexedAt,
		)
		if err != nil {
			return &RepositoryError{Op: "BatchSave", Err: err}
		}
	}

	if err := tx.Commit(); err != nil {
		return &RepositoryError{Op: "BatchSave", Err: err}
	}

	return nil
}

// QueryByFeedID retrieves posts for a specific feed with pagination
func (r *PostRepository) QueryByFeedID(ctx context.Context, feedID string, limit, offset int) ([]*PostModel, error) {
	query := `
		SELECT id, created_at, updated_at, uri, author_did, text, feed_id, indexed_at
		FROM posts
		WHERE feed_id = ?
		ORDER BY indexed_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.db.QueryContext(ctx, query, feedID, limit, offset)
	if err != nil {
		return nil, &RepositoryError{Op: "QueryByFeedID", Err: err}
	}
	defer rows.Close()

	var posts []*PostModel
	for rows.Next() {
		var post PostModel
		var postID string
		var createdAt, updatedAt time.Time

		err := rows.Scan(
			&postID,
			&createdAt,
			&updatedAt,
			&post.URI,
			&post.AuthorDID,
			&post.Text,
			&post.FeedID,
			&post.IndexedAt,
		)
		if err != nil {
			return nil, &RepositoryError{Op: "QueryByFeedID", Err: err}
		}

		post.SetID(postID)
		post.SetCreatedAt(createdAt)
		post.SetUpdatedAt(updatedAt)

		posts = append(posts, &post)
	}

	return posts, rows.Err()
}

// CountByFeedID returns the total number of posts for a feed
func (r *PostRepository) CountByFeedID(ctx context.Context, feedID string) (int, error) {
	query := "SELECT COUNT(*) FROM posts WHERE feed_id = ?"

	var count int
	err := r.db.QueryRowContext(ctx, query, feedID).Scan(&count)
	if err != nil {
		return 0, &RepositoryError{Op: "CountByFeedID", Err: err}
	}

	return count, nil
}
