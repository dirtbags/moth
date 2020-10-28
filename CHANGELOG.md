# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v4.1.0-pre] - Unreleased
### Changed
- Stop building devel server from this codebase; this is moving to a new repo

## [v4.0.1] - 2020-10-27
### Fixed
- Clear Debug.summary field when making mothballs

### Changed
- Regulated category/puzzle provider API: now everything returns a JSON dictionary (or octet stream for files)

### Added
- More log events
- [Log channels document](docs/logs.md)
- More detailed [API documntation](docs/api.md)
- Transpiler warning if `mkpuzzle` exists but is not executable

## [v4.0.0] - 2020-10-14
### Fixed
- Multiple bugs preventing production server from working properly
- CI builds should be working now
- Team registration now correctly writes names to files
- Anonymized team names now only computed once per team
- Don't output "self" team for unauthenticated state exports

### Added
- Documented the HTTP API
- Added a drawing of how things fit together

## [v4.0-rc1] - 2020-10-13
### Changed
- Major rewrite/refactor of `mothd`
  - Clear separation of roles: State, Puzzles, and Theme
    - Sqlite, Redis, or S3 should fit in easily now
    - Will allow "dynamic" puzzles now, we just need a flag to enable it
  - Server no longer provides unlocked content
    - Puzzle URLs are now just `/content/${cat}/${points}/`
  - Changes to `state` directory
    - Most files now have a bit of (English) documentation at the beginning
    - `state/until` is now `state/hours` and can specify multiple begin/end hours
    - `state/disabled` is now `state/enabled`
- Mothball structure has changed
  - Mothballs no longer contain `map.txt`
  - Mothballs no longer obfuscate content paths
  - Clients now expect unlocked puzzles to just be `map[string][]int`
- New `/state` API endpoint
  - Provides *all* server state: event log, team mapping, messages, configuration

### Added
- New `transpile` CLI command
  - Provides `mothball` action to create mothballs
  - Lets you test a few development server things, if you want

### Deprecated

### Removed
- Development server is gone now; use `mothd` directly with a flag to transpile on the fly

### Fixed

### Security

## [v3.5.1] - 2020-03-16
### Fixed
- Support insta-checking for legacy puzzles

## [v3.5.0] - 2020-03-13
### Changed
- We are now using SHA256 instead of djb2hash
### Added
- URL parameter to points.json to allow returning only the JSON for a single
  team by its team id (e.g., points.json?id=abc123).
- A CONTRIBUTING.md to describe expectations when contributing to MOTH
- Include basic metadata in mothballs
- add_script_stream convenience function allows easy script addition to puzzle
- Autobuild Docker images to test buildability
- Extract and use X-Forwarded-For headers in mothd logging
- Mothballs can now specify `X-Answer-Pattern` header fields, which allow `*`
  at the beginning, end, or both, of an answer. This is `X-` because we
  are hoping to change how this works in the future.
### Fixed
- Handle cases where non-legacy puzzles don't have an `author` attribute
- Handle YAML-formatted file and script lists as expected
- YAML-formatted example puzzle actually works as expected
- points.log will now always be sorted chronologically

## [3.4.3] - 2019-11-20
### Fixed
- Made top-scoring teams full-width

## [3.4.2] - 2019-11-18
### Fixed
- Issue with multiple answers in devel server and YAML-format .moth

## [3.4.1] - 2019-11-17
### Fixed
- Scoreboard was double-counting points

## [3.4] - 2019-11-13
### Added
- A changelog
- Support for embedding Python libraries at the category or puzzle level
- Minimal PWA support to permit caching of currently-unlocked content
- Embedded graph in scoreboard
- Optional tracking of participant IDs
- New `notices.html` file for sending broadcast messages to players
### Changed
- Use native JS URL objects instead of wrangling everything by hand
