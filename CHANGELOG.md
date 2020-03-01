# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [4.0.0] - Unreleased
### Added
- New `transpile` command to replace some functionality of devel server

### Changed
- Major rewrite/refactor of `mothd`
  - There are now providers for State, Puzzles, and Theme. Sqlite, Redis, or S3 should fit in easily.
  - Server no longer provides unlocked content
  - Puzzle URLs are now just `/content/${cat}/${points}/`
- `state/until` is now `state/hours` and can specify multiple begin/end hours
- `state/disabled` is now `state/enabled`
- Mothball structure has changed substantially
  - Mothballs no longer contain `map.txt`
  - Clients now expect unlocked puzzles to just be `map[string][]int`

### Deprecated

### Removed
- Development server is gone now; use `mothd` directly with a flag to transpile on the fly

### Fixed

### Security

