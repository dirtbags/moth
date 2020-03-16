# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
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
