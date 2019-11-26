# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Added
- New `transpile` command to replace some functionality of devel server

### Changed
- Major rewrite/refactor of `mothd`
- `state/until` is now `state/hours` and can specify multiple begin/end hours
- `state/disabled` is now `state/enabled`
- Mothball structure has changed substantially

### Deprecated

### Removed
- Development server is gone now; use `mothd` directly with a flag to transpile on the fly

### Fixed

### Security

