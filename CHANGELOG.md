# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [v0.1.1](https://github.com/odpf/compass/releases/tag/v0.1.1) (2021-04-12)

### Fixes

Fix /v1/types returns null on empty types
Fix search filter not working as expected
Fix error when search whitelist is empty when searching

## [v0.1.0](https://github.com/odpf/compass/releases/tag/v0.1.0) (2021-04-05)

### Features

- Add API to fetch all types
- Add API to fetch a type
- Add API to delete a type
- Add API to delete a record
- Remove /v1/entities/* APIs.
- Fix New Relic transaction name not showing detailed route pattern.
- Implement query time boosting
- Enable fuzzy search