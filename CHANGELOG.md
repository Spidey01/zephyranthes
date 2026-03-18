# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

<!-- markdownlint-disable MD024 -->

## [Unreleased] - WIP

### Added

- Add `-C DIR` / `--directory DIR` option to change to `DIR` before opening and running backup specs.
  - Since `contents` are resolved relative to the current directory, this can be
    used to control how files are resolved or avoid storing the full path.

## [v1.1.0] - 2026-03-15

This release mainly introduces support for environment variables as part of
backup specs. Additionally, symbolic links should now work when using a zip archive. Note however that this has only been tested under Unix-based systems and with Info-Zip (unzip) for extraction.

### Changed

- Logging now goes to standard output if `--log-file -` is specified.
- Using `--dry-run` now exercises far more code, and no longer leaves behind an empty archive.

### Added

- Environment variable expansion for the path and contents list.
- Support for symbolic links in ZIP archives.
- Unit tests covering command line options, archives, and backup specs.
- Added a new `--version` flag to print app version.
- There is now a change log :).

### Fixed

## [v1.0.0] - 2026-03-13

Initial release. This represents the first version to see regular use on my
machines during 2025. It's worked well enough that I finally declared it
released :).

---

- \[unreleased\]: [changes](https://github.com/Spidey01/zephyranthes/compare/v1.1.0...HEAD)
- [[1.1.0](https://github.com/Spidey01/zephyranthes/releases/tag/v1.1.0)]: [commits](https://github.com/Spidey01/zephyranthes/commits/v1.1.0/)
- [[1.0.0](https://github.com/Spidey01/zephyranthes/releases/tag/v1.0.0)]: [commits](https://github.com/Spidey01/zephyranthes/commits/v1.0.0/)
