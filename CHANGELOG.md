# Changelog

## [1.0.0] - 2026-02-13

### 🎉 First Stable Release

Phloem is now production-ready! This release marks the transition from beta to stable with comprehensive testing, production hardening, and enterprise-grade reliability.

### Added
- **Redis Session Storage** - Persistent session storage with automatic failover
  - Sessions persist across Crown redeploys
  - Multi-tier storage: Redis (primary) + PostgreSQL (backup)
  - Automatic Redis detection via `REDIS_URL` environment variable
  - 30-day session TTL with refresh support

- **Production Telemetry** - Validated install tracking and usage metrics
  - Opt-in telemetry via `PHLOEM_TELEMETRY` environment variable
  - Privacy-first: only version, OS, and anonymous device ID
  - Crown API receives and stores telemetry data
  - 65+ production installs tracked and validated

- **Comprehensive Test Coverage** - 19 new tests for session storage
  - Redis integration tests with miniredis
  - HTTP proxy bypass tests (fixes Cambium LLM proxy hang)
  - OAuth state and session CRUD operations
  - TTL expiration and key prefix validation
  - Multi-provider session support

### Changed
- **HTTP Transport** - Added `Proxy: nil` to bypass local proxies
  - Fixes hang when Cambium LLM proxy is installed
  - Ensures reliable Crown API communication
  - Prevents 10-second timeout delays

- **OAuth Store Initialization** - Added nil checks to prevent crashes
  - Graceful degradation when Redis is unavailable
  - Falls back to in-memory storage for local development

### Fixed
- Session store crash on nil pointer dereference
- HTTP client hanging with Cambium proxy environment
- OAuth stores upgrade when initialization fails

### Production Readiness
- ✅ Full SaaS stack deployed to Fly.io
- ✅ E2E test suite: 14/14 tests passing
- ✅ Redis-backed session persistence
- ✅ Telemetry validated with production data
- ✅ Zero factory dependencies (fully independent)
- ✅ Privacy-first architecture confirmed

## [0.2.0] - 2026-02-08

### Added
- **OAuth Authentication** - GitHub OAuth integration for cloud sync
  - `phloem auth login` - Authenticate with GitHub via browser
  - `phloem auth logout` - Clear stored credentials
  - `phloem auth whoami` - Display current user info
  - `phloem auth scopes` - List available memory scopes
  - JWT token storage with secure file permissions (0600)

- **Scoped Memory** - Repository-aware memory storage
  - Automatic Git repository detection
  - Scope-based memory filtering (`github.com/owner/repo`)
  - Scoped duplicate detection
  - Database migration for scope column
  - `--scope` flag support for remember/recall commands

- **Chrome Extension v2** - Cloud-based browser extension
  - OAuth authentication via Crown API
  - No native messaging dependency
  - Content scripts for ChatGPT, Claude, Gemini
  - Cloud sync via Crown API endpoints

- **MCP OAuth Integration** - MCP server now uses OAuth tokens
  - Automatic JWT token loading from `~/.phloem/auth.json`
  - Bearer token authentication to Crown API
  - Backward compatible with API key authentication
  - Fixes 401 errors with new Crown OAuth/secrets service

### Changed
- Memory store now supports scope filtering
- Improved duplicate detection (scope-aware)
- Sync client prefers OAuth Bearer tokens over API keys
- MCP server automatically loads OAuth credentials

### Fixed
- SQL queries updated for scope column
- 401 authentication errors with Crown API
- Scope isolation in duplicate detection

## [0.1.0] - 2026-01-01

Initial release
