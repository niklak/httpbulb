# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [1.0.4] - 2024-08-16
## Changed
- Dockerfile: updated golang image version to 1.23

## [1.0.3] - 2024-07-30

### Fixed
- Fixed Dockerfile (catch SIGINT and SIGTERM).


## [1.0.2] - 2024-07-18

### Added
- Added a simple CORS middleware for the server application.


## [1.0.1] - 2024-07-04

### Fixed
- Fixed responsive design for the server application.



## [1.0.0] - 2024-07-03

### Added

- A router to work with `http.Server` and `httptest.Server`
- An executable to run this tool as standalone server.
- Docker image to run the httpbulb server inside a container.