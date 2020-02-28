# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]
### Changed
 - Endpoints `/points.json`, `/puzzles.json`, and `/messages.html` (optional theme file) combine into `/state`
 - No more `__devel__` category for dev server: this is now `state.config.devel` in the `/state` endpoint
 - Development server no longer serves a static `/` with links: it now redirects you to a randomly-generated seed URL
 - Default theme modifications to handle all this
 - Default theme now automatically "logs you in" with Team ID if it's getting state from the devel server

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
