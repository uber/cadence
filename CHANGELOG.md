# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

You can find a list of previous releases on the [github releases](https://github.com/uber/cadence/releases) page.

## [Unreleased]
### Added
- Added TLS support for gRPC (#4606). Use `tls` config section under service `rpc` block to enable it.
### Changed
- Default outbound between internal server components are now switched to gRPC. There is still an option to switch back to TChannel by setting dynamic config `system.enableGRPCOutbound` to `false`. However this is now considered deprecated and will be removed in the future release.

## [0.23.0] - TBD
### Added
- Added gRPC support for cross domain traffic. This can be enabled in `ClusterGroupMetadata` config section with `rpcTransport: "grpc"` option. By default, tchannel is used. (#4390)
- Fixed an Issue with workflows timing out early if the RetryPolicy timeout is less than  the workflow timeout. Now on the first attempt, a workflow will only timeout if the workflow timeout is exceeded. (#4467)
## [0.21.3] - 2021-07-17
### Added
- Added GRPC support. Cadence server will accept requests on both TChannel and GRPC. With dynamic config flag `system.enableGRPCOutbound` it will also switch to GRPC communication internally between server components.

### Fixed
- This change contains breaking change on user config. The masterClusterName config key is deprecated and is replaced with primaryClusterName key. (#4185)

### Changed
- Bump CLI version to v0.19.0
- Change `--connect-attributes` in `cadence-sql-tool` from URL encoding to the format of k1=v1,k2=v2...
- Change `--domain_data` in `cadence domain update/register` from the format of k1:v1,k2:v2... to the format of k1=v1,k2=v2...
- Deprecate local domains. All new domains should be created as global domains.
- Server will deep merge configuration files when loading. No need to copy the whole config section, specify only those fields that needs overriding. (#4165)

## [0.18.0] - 2021-01-22

## [0.16.1] - 2021-01-21

## [0.17.0] - 2021-01-13

## [0.16.0] - 2020-12-10
