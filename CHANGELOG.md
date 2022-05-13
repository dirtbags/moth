# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v4.4.9] - 2022-05-12
### Changed
- Added a performance optimization for events with a large number of teams
  backed by NFS

## [v4.4.8] - 2022-05-10
### Changed
- You can now join with a team ID not appearing in `teamids.txt`, 
  as long as it is registered (in the `teams/` directory)

## [v4.4.7] - 2022-05-10
### Changed
- Initializing an instance now truncates `events.csv`

## [v4.4.6] - 2021-10-26
### Added
- State is now cached in memory, in an attempt to reduce filesystem metadata operations,
  which kill NFS.

## [v4.4.5] - 2021-10-26
### Added
- Images deploying to docker hub too. We're now at capacity for our Docker Hub team.

## [v4.4.4] - 2021-10-20
### Changed
- Trying to get CI push of built images. I expect this to fail, too. But in a way that can help me debug the issue.

## [v4.3.3] - 2021-10-20
### Fixed
- Points awarded while scoring is paused are now correctly sorted (#168)
- Writing a new mothball with the same name is now detected and the new mothball loaded (#172)
- Regression test for issue where URL path leading directories were ignored (#144)
- A few other very minor bugs were closed when I couldn't reproduce them or decided they weren't actually bugs.

### Changed
- Many error messages were changed to start with a lower-case letter, 
  in order to satisfy a new linter check.
- CI/CD moved to our Cyber Fire Gitlab instance
- I attempted to have the build thingy automatically build moth:v4 and moth:v4.3 and moth:v4.3.3 images, 
  but I can't test it without tagging a release. 
  So v4.3.4 might come out very soon after this ;)

## [v4.2.2] - 2021-09-30
### Added
- `debug.notes` front matter field

## [v4.2.1] - 2021-04-13
### Fixed
- Transpiled KSAs no longer dropped

## [v4.2.0] - 2020-03-26
### Changed
- example/5/draggable.js fix for FireFox to prevent dropping a draggable trying to load a URL
- `transpile` arguments now work the same way for the transpile binary as they do for mkpuzzle
- `transpile inventory` does what you expect: inventory of current category, not inventory of all categories

### Removed
- No longer building a `moth-devel` image,
  this is now handled by the
  [moth-devel repository](https://github.com/dirtbags/moth-devel).

### Fixed
- `transpile` will now run `mkcategory` and `mkpuzzle` when invoked without `-dir`


## [v4.1.1] - 2020-03-02
### Removed
- ppc64le and i386 builds of github, because ppc64le keep failing mysteriously, and we don't need them anyhow.


## [v4.1.0] - 2020-03-02
### Added
- `transpile` now has a `markdown` command,
  so you can use the "stock" markdown formatter

### Changed
- event.log is now events.csv, to make it easier to import to a spreadsheet
- When in devel mode, any team ID may score points. This allows more interaction with the state directory.
- When in devel mode, any team ID may be registered.
  It still works the same way if you register a team in `teamids.txt`,
  but now you can use anything and it will put you on an already existing team
  named `<devel:$ID>`.
- switched from `blackfriday` to `goldmark`, to support CommonMark
- `puzzle.json`no longer has `Pre` and `Post` sections

### Removed
- JavaScript code we didn't write is now pulled from a CDN


## [v4.0.2] - 2020-10-29
### Added
- Build multiarch Docker images
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
