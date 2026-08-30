# Releasing modules

This repository contains independent public Go modules. A release is a stable GitHub release and tag named `<module>/vX.Y.Z`; release notes are generated on demand and no changelog file is committed.

## Repository settings

Protect `main`, require the `CI / gate` and `Commit checks / commitlint` checks, and allow squash merging only. Pull requests must contain one commit; both the PR title and the last commit must follow Conventional Commits.

## Bootstrap order

Release shared modules before the modules that depend on them:

- `grpcserver` before `grpczap`.
- `health` before `healthgrpc`, `healthotel`, `healthserver`, and `healthzap`.
- `txmanager` before `postgresdb` and `sqlitedb`.
- `oapivalidator` before `oapivalidatorjwt`.
- `retry` before `oidcsession`, and `oidcsession` before `oidcsessionredis`.
- `retry` before `kafka`, and `kafka` before `kafkaproto` and `kafkazap`.
- `kafka`, `postgresdb`, and `retry` before `kafkaoutbox`, then `kafkaoutbox` before `kafkaoutboxzap`.

The workspace `use` directives provide local package sources. Versioned internal requirements that do not have public tags yet are centralized as `go.work` replacements; publishable `go.mod` files do not contain local paths. Preview and release both run `go mod tidy -diff`, download dependencies, and run race-enabled tests with `GOWORK=off`, then verify that `go.mod` and `go.sum` stayed unchanged.

Changes to a base module and its dependants may land in one pull request because normal CI uses the workspace. Release them in the order above. A dependent preview or release intentionally fails until the base tag exists and its checksum updates have been committed.

## Preview and publish

1. Run **Preview module release** on the default branch and choose the module plus `patch`, `minor`, or `major`.
2. Review the computed tag and consumer-focused git-cliff notes in the workflow summary.
3. Run **Release module** with the same inputs.

When a module has no stable tag, its base version is `v0.0.0`; for example, an initial `minor` release becomes `v0.1.0`. Publishing refuses an existing tag and refuses a release when no new commit touches the selected module directory.

Releases are stable only: there are no release candidates and no binary artifacts for library modules.
