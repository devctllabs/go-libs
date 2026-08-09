# Releasing modules

This repository contains independent public Go modules. A release is a stable GitHub release and tag named `<module>/vX.Y.Z`; release notes are generated on demand and no changelog file is committed.

## Repository settings

Protect `main`, require the `CI / gate` and `Commit checks / commitlint` checks, and allow squash merging only. Pull requests must contain one commit; both the PR title and the last commit must follow Conventional Commits.

## Bootstrap order

Release `health` and `txmanager` before the modules that depend on them:

- `health` before `healthotel`, `healthserver`, and `healthzap`.
- `txmanager` before `postgresdb` and `sqlitedb`.

Workspace replacements make local monorepo development possible, but the release workflow tests the selected module with `GOWORK=off`. Therefore a dependent module cannot be published until its declared internal dependency exists publicly.

## Preview and publish

1. Run **Preview module release** on the default branch and choose the module plus `patch`, `minor`, or `major`.
2. Review the computed tag and consumer-focused git-cliff notes in the workflow summary.
3. Run **Release module** with the same inputs.

When a module has no stable tag, its base version is `v0.0.0`; for example, an initial `minor` release becomes `v0.1.0`. Publishing refuses an existing tag and refuses a release when no new commit touches the selected module directory.

Releases are stable only: there are no release candidates and no binary artifacts for library modules.
