// TODO: Session Repository (stores tokens in ~/.skycli/.meta.json, record in sqlite ~/.skycli/db.sqlite or appdata)
// TODO: Feed Repository
// TODO: Post Repository
package store

import "context"

// Repository defines a generic persistence contract.
type Repository interface {
	Init(ctx context.Context) error
	Close() error
	Get(ctx context.Context, id string) (Model, error)
	List(ctx context.Context) ([]Model, error)
	Save(ctx context.Context, feed Model) error
	Delete(ctx context.Context, id string) error
}
