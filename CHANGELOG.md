# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

You can find a list of previous releases on the [github releases](https://github.com/uber/cadence/releases) page.

## [Unreleased]
### Added
- Added GRPC support. Cadence server will accept requests on both TChannel and GRPC. With dynamic config flag `system.enableGRPCOutbound` it will also switch to GRPC communication internally between server components.

### Fixed
- This change contains breaking change on user config. The masterClusterName config key is deprecated and is replaced with primaryClusterName key. (#4185)

### Changed
- Bump CLI version to v0.19.0
- Change `--connect-attributes` in `cadence-sql-tool` from URL encoding to the format of k1=v1,k2=v2...
- Change `--domain_data` in `cadence domain update/register` from the format of k1:v1,k2:v2... to the format of k1=v1,k2=v2...

## [0.18.0] - 2021-01-22

## [0.16.1] - 2021-01-21

## [0.17.0] - 2021-01-13

## [0.16.0] - 2020-12-10
