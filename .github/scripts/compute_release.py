#!/usr/bin/env python3
"""Compute the next stable tag for one module in this repository."""

from __future__ import annotations

import argparse
import re
import subprocess
import sys
from pathlib import Path
from typing import Callable


REPOSITORY_ROOT = Path(__file__).resolve().parents[2]
MODULES = (
    "codexapp",
    "config",
    "debugserver",
    "di",
    "filesystem",
    "health",
    "healthotel",
    "healthserver",
    "healthzap",
    "lifecycle",
    "log",
    "oapivalidator",
    "postgresdb",
    "sqlitedb",
    "telemetry",
    "txmanager",
)
BUMPS = ("patch", "minor", "major")
Runner = Callable[[list[str]], subprocess.CompletedProcess[str]]


def run_command(arguments: list[str]) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        arguments,
        cwd=REPOSITORY_ROOT,
        check=False,
        capture_output=True,
        text=True,
    )


def stable_tags(module: str, runner: Runner = run_command) -> list[tuple[tuple[int, int, int], str]]:
    result = runner(["git", "tag", "--list", f"{module}/v*"])
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "failed to list git tags")
    pattern = re.compile(rf"^{re.escape(module)}/v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$")
    tags: list[tuple[tuple[int, int, int], str]] = []
    for candidate in result.stdout.splitlines():
        match = pattern.fullmatch(candidate.strip())
        if match:
            tags.append((tuple(map(int, match.groups())), candidate.strip()))
    return sorted(tags)


def bump_version(version: tuple[int, int, int], bump: str) -> tuple[int, int, int]:
    major, minor, patch = version
    if bump == "major":
        return major + 1, 0, 0
    if bump == "minor":
        return major, minor + 1, 0
    if bump == "patch":
        return major, minor, patch + 1
    raise ValueError(f"unsupported bump: {bump}")


def compute_release(module: str, bump: str, runner: Runner = run_command) -> tuple[str, str]:
    if module not in MODULES:
        raise ValueError(f"unsupported module: {module}")
    if bump not in BUMPS:
        raise ValueError(f"unsupported bump: {bump}")
    tags = stable_tags(module, runner)
    previous_version, previous_tag = tags[-1] if tags else ((0, 0, 0), "")
    next_version = bump_version(previous_version, bump)
    next_tag = f"{module}/v{next_version[0]}.{next_version[1]}.{next_version[2]}"
    return previous_tag, next_tag


def require_new_commits(module: str, previous_tag: str, runner: Runner = run_command) -> None:
    revision = f"{previous_tag}..HEAD" if previous_tag else "HEAD"
    result = runner(["git", "rev-list", "--count", revision, "--", module])
    if result.returncode != 0:
        raise RuntimeError(result.stderr.strip() or "failed to inspect module commits")
    if int(result.stdout.strip()) == 0:
        raise RuntimeError(f"no unreleased commits touch {module}/")


def require_absent_tag(tag: str, runner: Runner = run_command) -> None:
    local = runner(["git", "rev-parse", "--verify", "--quiet", f"refs/tags/{tag}"])
    if local.returncode == 0:
        raise RuntimeError(f"tag already exists locally: {tag}")
    if local.returncode != 1:
        raise RuntimeError(local.stderr.strip() or "failed to inspect local tags")
    remote = runner(["git", "ls-remote", "--exit-code", "--tags", "origin", f"refs/tags/{tag}"])
    if remote.returncode == 0:
        raise RuntimeError(f"tag already exists on origin: {tag}")
    if remote.returncode not in (2,):
        raise RuntimeError(remote.stderr.strip() or "failed to inspect remote tags")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module", required=True, choices=MODULES)
    parser.add_argument("--bump", required=True, choices=BUMPS)
    parser.add_argument("--previous", action="store_true", help="print the previous stable tag")
    parser.add_argument("--require-new-commits", action="store_true")
    parser.add_argument("--require-absent-tag", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        previous_tag, next_tag = compute_release(args.module, args.bump)
        if args.require_new_commits:
            require_new_commits(args.module, previous_tag)
        if args.require_absent_tag:
            require_absent_tag(next_tag)
    except (RuntimeError, ValueError) as error:
        print(error, file=sys.stderr)
        return 1
    print(previous_tag if args.previous else next_tag)
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
