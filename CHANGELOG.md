# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.2.1] - 2025-10-18

### Fixed
- Set TLS MinVersion to 1.2 for certificate retrieval (G402 security fix)
- Add proper error handling for JSON encoding in HTTP handlers (G104)
- Fix golangci-lint compatibility with Go 1.25
- Simplify CI security scanning to work on all GitHub plans

### Security
- Enhanced TLS security by enforcing minimum TLS 1.2 for external connections
- Improved error handling throughout HTTP handlers

## [0.2.0] - 2025-10-17

### Added
- Comprehensive performance testing suite with benchmarks, load tests, and stress tests
- Docker multi-platform support (linux/amd64, linux/arm64)
- Backup pins support via `include-backup-pins` query parameter
- Certificate caching with configurable TTL (`CERT_CACHE_TTL`)
- Health check endpoints (`/health` and `/readiness`)
- Structured JSON logging with configurable log levels
- Performance documentation and testing scripts
- Contributing guidelines and community standards

### Changed
- Improved documentation with detailed configuration examples
- Enhanced error messages for better debugging
- Optimized certificate retrieval with caching layer
- Updated README with production deployment examples

### Fixed
- Thread-safe certificate caching implementation
- Proper graceful shutdown handling
- Timeout configurations for all network operations

### Security
- Added IP literal blocking by default (`ALLOW_IP_LITERALS=false`)
- Implemented read header timeout protection against Slowloris attacks
- Enhanced domain validation with wildcard support
- Non-root Docker container execution (user 65532)

## [0.0.1] - 2025-01-17

### Added
- Initial release of Dynapins Server
- Core SSL pinning API with JWS (ES256) signature support
- Domain whitelist validation with wildcard patterns
- ECDSA P-256 cryptographic signing
- TLS certificate retrieval and SPKI hash generation
- Configurable timeouts and security settings
- Docker containerization with Alpine Linux base
- Comprehensive README documentation
- MIT License

### Technical Details
- Go 1.25+ compatibility
- Minimal dependencies (only lestrrat-go/jwx for JWS)
- Stateless architecture with no database requirements
- Environment variable-based configuration

[Unreleased]: https://github.com/Free-cat/dynapins-server/compare/v0.2.1...HEAD
[0.2.1]: https://github.com/Free-cat/dynapins-server/compare/v0.2.0...v0.2.1
[0.2.0]: https://github.com/Free-cat/dynapins-server/compare/v0.0.1...v0.2.0
[0.0.1]: https://github.com/Free-cat/dynapins-server/releases/tag/v0.0.1

