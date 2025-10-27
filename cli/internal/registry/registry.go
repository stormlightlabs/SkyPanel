package registry

import (
	"context"
	"errors"
	"sync"

	"github.com/stormlightlabs/skypanel/cli/internal/store"
)

var (
	once     sync.Once
	instance *Registry
)

// Registry manages singleton instances of repositories and services
type Registry struct {
	service     *store.BlueskyService
	sessionRepo *store.SessionRepository
	feedRepo    *store.FeedRepository
	postRepo    *store.PostRepository
	profileRepo *store.ProfileRepository
	initialized bool
	mu          sync.RWMutex
}

// Get returns the singleton registry instance
func Get() *Registry {
	once.Do(func() {
		instance = &Registry{
			initialized: false,
		}
	})
	return instance
}

// Init initializes all repositories and runs database migrations
// Must be called before using any repository or service
func (r *Registry) Init(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.initialized {
		return nil
	}

	sessionRepo, err := store.NewSessionRepository()
	if err != nil {
		return &RegistryError{Op: "InitSessionRepo", Err: err}
	}
	if err := sessionRepo.Init(ctx); err != nil {
		return &RegistryError{Op: "InitSessionRepo", Err: err}
	}
	r.sessionRepo = sessionRepo

	feedRepo, err := store.NewFeedRepository()
	if err != nil {
		return &RegistryError{Op: "InitFeedRepo", Err: err}
	}
	if err := feedRepo.Init(ctx); err != nil {
		return &RegistryError{Op: "InitFeedRepo", Err: err}
	}
	r.feedRepo = feedRepo

	postRepo, err := store.NewPostRepository()
	if err != nil {
		return &RegistryError{Op: "InitPostRepo", Err: err}
	}
	if err := postRepo.Init(ctx); err != nil {
		return &RegistryError{Op: "InitPostRepo", Err: err}
	}
	r.postRepo = postRepo

	profileRepo, err := store.NewProfileRepository()
	if err != nil {
		return &RegistryError{Op: "InitProfileRepo", Err: err}
	}
	if err := profileRepo.Init(ctx); err != nil {
		return &RegistryError{Op: "InitProfileRepo", Err: err}
	}
	r.profileRepo = profileRepo

	r.service = store.NewBlueskyService("")

	if sessionRepo.HasValidSession(ctx) {
		accessToken, err := sessionRepo.GetAccessToken(ctx)
		if err == nil {
			refreshToken, _ := sessionRepo.GetRefreshToken(ctx)
			r.service.SetTokens(accessToken, refreshToken)
		}
	}

	r.initialized = true
	return nil
}

// Close releases all repository and service resources
func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error

	if r.service != nil {
		if err := r.service.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.sessionRepo != nil {
		if err := r.sessionRepo.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.feedRepo != nil {
		if err := r.feedRepo.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.postRepo != nil {
		if err := r.postRepo.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if r.profileRepo != nil {
		if err := r.profileRepo.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	r.initialized = false

	if len(errs) > 0 {
		return &RegistryError{Op: "Close", Err: errors.Join(errs...)}
	}

	return nil
}

// GetService returns the BlueskyService singleton
func (r *Registry) GetService() (*store.BlueskyService, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, &RegistryError{Op: "GetService", Err: errors.New("registry not initialized")}
	}

	if r.service == nil {
		return nil, &RegistryError{Op: "GetService", Err: errors.New("service not available")}
	}

	return r.service, nil
}

// GetSessionRepo returns the SessionRepository singleton
func (r *Registry) GetSessionRepo() (*store.SessionRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, &RegistryError{Op: "GetSessionRepo", Err: errors.New("registry not initialized")}
	}

	if r.sessionRepo == nil {
		return nil, &RegistryError{Op: "GetSessionRepo", Err: errors.New("session repository not available")}
	}

	return r.sessionRepo, nil
}

// GetFeedRepo returns the FeedRepository singleton
func (r *Registry) GetFeedRepo() (*store.FeedRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, &RegistryError{Op: "GetFeedRepo", Err: errors.New("registry not initialized")}
	}

	if r.feedRepo == nil {
		return nil, &RegistryError{Op: "GetFeedRepo", Err: errors.New("feed repository not available")}
	}

	return r.feedRepo, nil
}

// GetPostRepo returns the PostRepository singleton
func (r *Registry) GetPostRepo() (*store.PostRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, &RegistryError{Op: "GetPostRepo", Err: errors.New("registry not initialized")}
	}

	if r.postRepo == nil {
		return nil, &RegistryError{Op: "GetPostRepo", Err: errors.New("post repository not available")}
	}

	return r.postRepo, nil
}

// GetProfileRepo returns the ProfileRepository singleton
func (r *Registry) GetProfileRepo() (*store.ProfileRepository, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if !r.initialized {
		return nil, &RegistryError{Op: "GetProfileRepo", Err: errors.New("registry not initialized")}
	}

	if r.profileRepo == nil {
		return nil, &RegistryError{Op: "GetProfileRepo", Err: errors.New("profile repository not available")}
	}

	return r.profileRepo, nil
}

// IsInitialized returns whether the registry has been initialized
func (r *Registry) IsInitialized() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.initialized
}

// RegistryError represents an error that occurred during registry operations
type RegistryError struct {
	Op  string
	Err error
}

func (e *RegistryError) Error() string {
	return "registry." + e.Op + ": " + e.Err.Error()
}

func (e *RegistryError) Unwrap() error {
	return e.Err
}
