from __future__ import annotations

import subprocess
import unittest

from compute_release import bump_version, compute_release, require_absent_tag, require_new_commits


def completed(stdout: str = "", returncode: int = 0, stderr: str = "") -> subprocess.CompletedProcess[str]:
    return subprocess.CompletedProcess([], returncode, stdout, stderr)


class ReleaseVersionTests(unittest.TestCase):
    def test_initial_minor_release_is_v0_1_0(self) -> None:
        previous, next_tag = compute_release("health", "minor", lambda _: completed())
        self.assertEqual("", previous)
        self.assertEqual("health/v0.1.0", next_tag)

    def test_bumps_latest_stable_tag_and_ignores_nonstable_tags(self) -> None:
        tags = "health/v0.2.9\nhealth/v0.3.0-rc.1\nhealth/v0.10.1\nother/v9.0.0\n"
        previous, next_tag = compute_release("health", "patch", lambda _: completed(tags))
        self.assertEqual("health/v0.10.1", previous)
        self.assertEqual("health/v0.10.2", next_tag)

    def test_all_bump_kinds_reset_lower_components(self) -> None:
        self.assertEqual((2, 0, 0), bump_version((1, 2, 3), "major"))
        self.assertEqual((1, 3, 0), bump_version((1, 2, 3), "minor"))
        self.assertEqual((1, 2, 4), bump_version((1, 2, 3), "patch"))

    def test_rejects_unknown_module(self) -> None:
        with self.assertRaisesRegex(ValueError, "unsupported module"):
            compute_release("unknown", "patch", lambda _: completed())

    def test_requires_module_commits_after_previous_tag(self) -> None:
        calls: list[list[str]] = []

        def runner(arguments: list[str]) -> subprocess.CompletedProcess[str]:
            calls.append(arguments)
            return completed("0\n")

        with self.assertRaisesRegex(RuntimeError, "no unreleased commits"):
            require_new_commits("health", "health/v0.1.0", runner)
        self.assertEqual(
            ["git", "rev-list", "--count", "health/v0.1.0..HEAD", "--", "health"],
            calls[0],
        )

    def test_rejects_existing_remote_tag(self) -> None:
        responses = iter((completed(returncode=1), completed("tag\n")))
        with self.assertRaisesRegex(RuntimeError, "already exists on origin"):
            require_absent_tag("health/v0.1.0", lambda _: next(responses))


if __name__ == "__main__":
    unittest.main()
