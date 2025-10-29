# ROADMAP - SkyPanel Server

Implementation roadmap for SkyPanel backend service. See [README.md](./README.md) for architecture overview.

## Design Philosophy

SkyPanel Server is designed for **self-hosted deployment**, similar to Forgejo/Gitea. Users should be able to:

- Deploy on their own infrastructure (local servers, VPS, home servers)
- Run as a single binary with embedded web app
- Use SQLite by default (PostgreSQL optional for larger deployments)
- Configure via simple config files or environment variables
- Install via Docker with single docker-compose command
- Upgrade with minimal downtime
- Backup and restore data easily
- Run multi-user instances for personal use or small teams

## Phase 1: Web App MVP

Launch functional Bluesky web client first.

### Database Setup

- [ ] Design database schema for user sessions (SQLite-compatible)
- [ ] Design schema for cached timeline data
- [ ] Design schema for user preferences
- [ ] Create migration system using golang-migrate or atlas
- [ ] Write initial migration files (SQLite dialect)
- [ ] Implement database connection with SQLite as default
- [ ] Create storage package with repository pattern
- [ ] Add PostgreSQL support as optional backend
- [ ] Database abstraction layer (support both SQLite and PostgreSQL)
- [ ] Database health check endpoint
- [ ] Automatic migrations on startup

### Authentication & Sessions

- [ ] Implement OAuth 2.0 flow with Bluesky
- [ ] Session storage in PostgreSQL
- [ ] HTTP-only cookie management
- [ ] JWT token validation
- [ ] Automatic token refresh logic
- [ ] Logout and session cleanup
- [ ] Multi-device session tracking
- [ ] Session middleware for protected routes

### API Proxy Layer

- [ ] Initialize AT Protocol client (@atproto/api equivalent in Go)
- [ ] Create API client wrapper with retry logic
- [ ] Implement timeline fetching endpoint
- [ ] Implement post fetching endpoint
- [ ] Implement profile fetching endpoint
- [ ] Error handling and user-friendly messages
- [ ] Response caching with TTL
- [ ] Rate limiting per user

### Web App Routes

- [ ] Home/timeline route
- [ ] Profile view route
- [ ] Post detail/thread route
- [ ] Login/auth callback routes
- [ ] Static asset serving
- [ ] SPA fallback for client-side routing

### Post Composer

- [ ] Text-only post creation endpoint
- [ ] Rich text parsing (mentions, hashtags, links)
- [ ] Post validation
- [ ] Draft saving (optional)

### Core Features

- [ ] Follow/unfollow user endpoint
- [ ] Like/unlike post endpoint
- [ ] Repost endpoint
- [ ] Delete post endpoint
- [ ] Thread/conversation fetching

### Self-Hosted Deployment

- [ ] Single binary build with embedded web app assets
- [ ] Dockerfile for containerized deployment
- [ ] Docker Compose file with example configuration
- [ ] Configuration file support (YAML or TOML)
- [ ] Environment variable configuration
- [ ] Default SQLite database location (./data/skypanel.db)
- [ ] Data directory structure (./data/ for db, uploads, cache)
- [ ] Systemd service file for Linux
- [ ] Installation script for quick setup
- [ ] Reverse proxy documentation (nginx, Caddy, Traefik)
- [ ] SSL/TLS setup guide (Let's Encrypt)
- [ ] Docker volume management for persistence
- [ ] Health check endpoint for monitoring
- [ ] Graceful shutdown handling

### Installation Methods

- [ ] **Docker Compose** - Single command deployment
- [ ] **Binary + Systemd** - Traditional Linux service
- [ ] **Bare Binary** - Run directly on any platform
- [ ] **Kubernetes** - Helm chart for k8s deployments (optional)

### Administration

- [ ] Admin user setup on first run
- [ ] User management CLI commands
- [ ] Backup command (database + uploads)
- [ ] Restore command
- [ ] Upgrade procedure documentation
- [ ] Configuration validation on startup
- [ ] Database integrity checks

**Done Criteria:** Users can deploy their own instance via Docker Compose or binary, log in, view timeline, create text posts, and follow/unfollow users.

## Phase 2: Full Web Client Features

Complete Bluesky client functionality.

### Media Handling

- [ ] Image upload to Bluesky blob storage
- [ ] Video upload support
- [ ] Image processing (resize, compress)
- [ ] Multiple image support (carousel)
- [ ] Alt text for images
- [ ] Link card fetching and display
- [ ] Embed preview generation

### Social Features

- [ ] Followers list endpoint
- [ ] Following list endpoint
- [ ] User search endpoint
- [ ] Suggested follows algorithm

### Notifications

- [ ] Database schema for notification cache
- [ ] Fetch notifications from AT Protocol
- [ ] Notification filtering by type
- [ ] Mark as read functionality
- [ ] Unread count tracking
- [ ] Real-time notification updates (WebSocket or polling)

### Search

- [ ] Post search endpoint
- [ ] User search endpoint
- [ ] Search result caching
- [ ] Search filters (author, date, etc.)
- [ ] Pagination for search results

### Profile Management

- [ ] Profile editing endpoint
- [ ] Avatar upload
- [ ] Banner upload
- [ ] Display name and bio updates
- [ ] Profile cache invalidation

### Settings

- [ ] User preferences schema
- [ ] Privacy settings endpoints
- [ ] Content filter preferences
- [ ] Blocked users management
- [ ] Muted users management
- [ ] Session management UI

### Feed Discovery

- [ ] Browse existing feed generators
- [ ] Subscribe to feeds
- [ ] Saved feeds management
- [ ] Feed ranking/recommendations

### PWA Features

- [ ] Service worker configuration
- [ ] Offline fallback pages
- [ ] App manifest
- [ ] Push notification setup
- [ ] Install prompt handling

**Done Criteria:** Web app has feature parity with official Bluesky client.

## Phase 3: Feed Generator Service

Implement custom feed creation and publishing.

### Database Schema for Feeds

- [ ] Design feeds table (id, name, description, owner, created_at, updated_at)
- [ ] Design feed_sources table (feed_id, source_type, source_params)
- [ ] Design feed_filters table (feed_id, filter_type, filter_config)
- [ ] Design feed_metadata table (feed_id, display_name, avatar_url, published_at)
- [ ] Create migration files for feed tables
- [ ] Repository layer for feed CRUD operations

### Feed Management API

- [ ] Create feed definition endpoint
- [ ] List user's feeds endpoint
- [ ] Get feed definition endpoint
- [ ] Update feed definition endpoint
- [ ] Delete feed endpoint
- [ ] Validate feed configuration

### Feed Algorithm Engine

- [ ] Feed executor core logic
- [ ] Source aggregation (timeline, author, list)
- [ ] Post deduplication by URI
- [ ] Chronological sorting
- [ ] Pagination with cursor support

### AT Protocol Endpoints

- [ ] Implement describeFeedGenerator endpoint
- [ ] Implement getFeedSkeleton endpoint
- [ ] DID document serving
- [ ] JWT authentication for AT Protocol requests
- [ ] Feed URI generation

### Feed Testing

- [ ] Local feed testing endpoint
- [ ] Mock data generator for testing
- [ ] Feed preview before publishing
- [ ] Sample post dataset

**Done Criteria:** Users can define feeds via API, test them locally, and publish to AT Protocol.

## Phase 4: Feed Indexing & Advanced Algorithms

Real-time post indexing and advanced filtering.

### Database Schema for Indexed Posts

- [ ] Design posts table (uri, cid, author_did, author_handle, timestamp, text, labels)
- [ ] Design post_embeds table (post_uri, embed_type, embed_data)
- [ ] Design author_stats table (did, handle, post_count, avg_posts_per_day, last_post_at)
- [ ] Design engagement_metrics table (post_uri, likes, reposts, replies)
- [ ] Add indexes for common query patterns
- [ ] Create migration files for indexing tables
- [ ] Implement repository layer for indexed data

### Firehose Subscription

- [ ] WebSocket connection to Bluesky firehose
- [ ] Parse repository events
- [ ] Filter for post records (app.bsky.feed.post)
- [ ] Extract post metadata
- [ ] Store posts in database
- [ ] Reconnection logic with exponential backoff
- [ ] Event processing queue
- [ ] Handle backfill for historical posts

### Filter Implementations

- [ ] Author filter (include/exclude by DID or handle)
- [ ] Rate-based filter (posts per day threshold)
- [ ] Label filter (content labels)
- [ ] Keyword filter (text matching, case-sensitive option)
- [ ] Combine multiple filters (AND/OR logic)

### Author Rate Tracking

- [ ] Calculate posts per day for authors
- [ ] Update author stats on new posts
- [ ] Periodic recalculation job
- [ ] Rate-based feed queries

### Engagement Ranking (Optional)

- [ ] Track post engagement metrics
- [ ] Engagement-based sorting algorithm
- [ ] Decay function for time-based ranking
- [ ] Configurable ranking weights

### Feed Creation UI

- [ ] Feed builder interface in web app
- [ ] Source selection UI
- [ ] Filter configuration forms
- [ ] Live preview of feed results
- [ ] Feed metadata editor (name, description, avatar)

### CLI Integration

- [ ] `skycli feed create` command
- [ ] `skycli feed list` command
- [ ] `skycli feed edit` command
- [ ] `skycli feed delete` command
- [ ] `skycli feed test` command
- [ ] `skycli feed publish` command
- [ ] `skycli feed export/import` commands

**Done Criteria:** Real-time post indexing is working, all filter types implemented, feeds can be created via UI or CLI.

## Phase 5: Extension Enhancement

Integrate extension with web app and server.

### Shared Component Library

- [ ] Create @skypanel/ui package
- [ ] Extract common components from extension
- [ ] Port components to shared package
- [ ] Update extension to use shared components
- [ ] Update web app to use shared components

### Shared Packages

- [ ] Create @skypanel/types package
- [ ] Create @skypanel/stores package
- [ ] Create @skypanel/api-client package
- [ ] Create @skypanel/utils package
- [ ] Update extension to use shared packages
- [ ] Update web app to use shared packages

### Extension-Server Integration

- [ ] API endpoints for extension authentication
- [ ] Sync preferences between extension and server
- [ ] Deep linking from extension to web app
- [ ] Extension can open web app routes
- [ ] Shared session tokens

### Extension Features

- [ ] Context menu integration
- [ ] Keyboard shortcuts
- [ ] System notifications via extension
- [ ] Background sync workers

**Done Criteria:** Extension and web app share codebase, seamless experience between both clients.

## Phase 6: Advanced Features

Polish and advanced functionality.

### Performance

- [ ] Redis caching layer for hot data
- [ ] Query optimization and indexing review
- [ ] Database query result caching
- [ ] CDN integration for static assets
- [ ] Image optimization pipeline
- [ ] Virtual scrolling for large feeds

### Real-time Updates

- [ ] WebSocket server for real-time events
- [ ] Push new posts to connected clients
- [ ] Real-time notification delivery
- [ ] Presence tracking (online/offline)

### Analytics

- [ ] Feed performance metrics
- [ ] User engagement tracking
- [ ] Feed popularity ranking
- [ ] Usage statistics dashboard

### Multi-user Feed Collaboration

- [ ] Shared feed ownership
- [ ] Collaborative feed editing
- [ ] Feed permissions system

### Advanced Notifications

- [ ] Notification filters by type
- [ ] Custom notification rules
- [ ] Digest notifications
- [ ] Email notifications (optional)

### Monitoring & Observability

- [ ] Prometheus metrics integration
- [ ] Grafana dashboards
- [ ] Error tracking (Sentry or similar)
- [ ] Performance monitoring
- [ ] Alert configuration

**Done Criteria:** Production-ready service with monitoring, analytics, and advanced features.

## Self-Hosting Documentation & UX

Documentation and tooling for users to self-host SkyPanel.

### Installation Documentation

- [ ] **Quick Start Guide** - Get running in 5 minutes with Docker
- [ ] **Docker Compose Guide** - Detailed docker-compose setup
- [ ] **Binary Installation** - Installing from binary releases
- [ ] **Building from Source** - For developers and contributors
- [ ] **Systemd Setup** - Running as a Linux service
- [ ] **Reverse Proxy Guide** - nginx, Caddy, Traefik examples
- [ ] **SSL/TLS Setup** - Let's Encrypt with various proxy servers
- [ ] **Configuration Reference** - Complete config documentation

### Upgrade & Maintenance

- [ ] **Upgrade Guide** - How to upgrade between versions
- [ ] **Backup & Restore** - Complete backup/restore procedures
- [ ] **Migration from Cloud** - Import data from hosted instances
- [ ] **Troubleshooting** - Common issues and solutions
- [ ] **Performance Tuning** - Optimization for different scales
- [ ] **Security Hardening** - Security best practices

### User Experience

- [ ] **First-Run Setup Wizard** - Web-based initial configuration
- [ ] **Admin Dashboard** - System status, user management, settings
- [ ] **Health Check Page** - System diagnostics and status
- [ ] **Update Notifications** - Alert admin of new versions
- [ ] **Automatic Backups** - Scheduled backup functionality
- [ ] **Log Viewer** - View application logs in web UI
- [ ] **Metrics Dashboard** - Basic usage metrics and stats

### CLI Tools

- [ ] `skypanel init` - Initialize new instance
- [ ] `skypanel start` - Start server
- [ ] `skypanel stop` - Stop server gracefully
- [ ] `skypanel status` - Check server status
- [ ] `skypanel backup` - Create backup
- [ ] `skypanel restore` - Restore from backup
- [ ] `skypanel upgrade` - Upgrade to new version
- [ ] `skypanel user add` - Add new user
- [ ] `skypanel user remove` - Remove user
- [ ] `skypanel user list` - List users
- [ ] `skypanel config validate` - Validate configuration
- [ ] `skypanel migrate` - Run database migrations

### Community & Support

- [ ] Example docker-compose.yml configurations
- [ ] Community deployment scripts
- [ ] Platform-specific guides (Raspberry Pi, NAS, etc.)
- [ ] Issue templates

## Database Schema Scaffolding Tasks

Priority tasks for database design and implementation.

Schema must be compatible with both SQLite (default) and PostgreSQL (optional).

Use standard SQL types and avoid database-specific features where possible.

### Core Tables

- [ ] **users** - User accounts and profile data
    - id, did, handle, display_name, avatar_url, created_at, updated_at
- [ ] **sessions** - Active user sessions
    - id, user_id, token_hash, refresh_token_hash, expires_at, created_at
- [ ] **preferences** - User preferences and settings
    - user_id, key, value, updated_at
- [ ] **cached_profiles** - Cached profile data from AT Protocol
    - did, handle, display_name, bio, avatar_url, banner_url, cached_at, ttl
- [ ] **cached_posts** - Cached posts for timeline performance
    - uri, cid, author_did, text, created_at, cached_at, ttl

### Feed Tables

- [ ] **feeds** - Feed definitions
    - id, owner_did, name, description, created_at, updated_at, is_published
- [ ] **feed_sources** - Feed sources configuration
    - id, feed_id, source_type (timeline/author/list), source_params (JSON)
- [ ] **feed_filters** - Feed filters configuration
    - id, feed_id, filter_type (author/rate/label/keyword), filter_config (JSON), priority
- [ ] **feed_metadata** - Published feed metadata
    - feed_id, display_name, description, avatar_url, published_at, at_uri

### Indexing Tables

- [ ] **indexed_posts** - Posts indexed from firehose
    - uri, cid, author_did, author_handle, text, labels (JSON), embeds (JSON), created_at, indexed_at
- [ ] **author_stats** - Author posting statistics
    - did, handle, total_posts, posts_last_7d, posts_last_30d, avg_posts_per_day, last_post_at, updated_at
- [ ] **post_engagement** - Engagement metrics (optional)
    - post_uri, likes, reposts, replies, updated_at

### Notification Tables

- [ ] **notifications** - Cached notifications
    - id, user_did, notification_type, actor_did, post_uri, is_read, created_at
- [ ] **notification_cursors** - Pagination cursors for notifications
    - user_did, cursor, updated_at

### Indexes

- [ ] users: did (unique), handle (unique)
- [ ] sessions: user_id, token_hash (unique), expires_at
- [ ] feeds: owner_did, is_published
- [ ] feed_sources: feed_id
- [ ] feed_filters: feed_id, priority
- [ ] indexed_posts: author_did, created_at, (author_did, created_at) composite
- [ ] author_stats: did (unique), posts_last_7d, avg_posts_per_day
- [ ] notifications: user_did, is_read, created_at

### Migration System

- [ ] Set up golang-migrate or atlas
- [ ] Version control for migrations
- [ ] Migration testing scripts
- [ ] Rollback procedures
- [ ] Seed data for development

## Implementation Notes

### Database

**SQLite (Default):**

- Use WAL mode for better concurrency
- Enable foreign keys (PRAGMA foreign_keys = ON)
- Regular VACUUM for maintenance
- Backup via SQLite backup API or file copy (when db is closed)
- Connection pool size: 1 for writes, multiple for reads
- Store database in persistent volume (./data/skypanel.db)

**PostgreSQL (Optional):**

- Connection pooling with pgx or database/sql
- Regular VACUUM and ANALYZE
- Use prepared statements
- Connection pool size: 20-50 for production

**Both:**

- Use prepared statements for all queries
- Add query timeouts and context cancellation
- Use transactions for multi-table operations
- Monitor slow queries and add indexes as needed
- Database abstraction layer to support both backends

### Self-Hosting

**Build & Distribution:**

- Single static binary with embedded assets
- Cross-compile for Linux, macOS, Windows (amd64, arm64)
- Docker image with multi-stage builds
- Keep binary size reasonable (<50MB)
- Version embedded in binary (ldflags)

**Configuration:**

- Support both config file and environment variables
- Provide example configuration files
- Validate configuration on startup
- Default to secure settings
- Document all configuration options

**Data Management:**

- Default data directory: ./data/
- Subdirectories: db/, uploads/, cache/, logs/
- Easy backup (copy data directory)
- Easy upgrade (replace binary, run migrations)
- Document data directory structure

**Security:**

- Generate secure secrets on first run
- Store secrets in config file or environment
- Support reverse proxy (respect X-Forwarded-* headers)
- Rate limiting enabled by default
- Content Security Policy headers

**Multi-tenancy:**

- Each instance serves one primary user
- Additional users for family/team deployments
- User management via admin CLI or web UI
- Per-user storage quotas (optional)

**Performance:**

- SQLite is sufficient for hundreds of users
- PostgreSQL recommended for thousands of users
- Document performance characteristics
- Provide scaling guidance
