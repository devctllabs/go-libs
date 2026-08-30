# oidcsessionredis

`oidcsessionredis` implements `oidcsession.SessionBackend` with a caller-owned
`redis.UniversalClient` and requires Redis 6.2 or newer. It generates a 256-bit opaque session
credential and addresses the session by its SHA-256 digest, so the credential itself is never
stored. The complete provider token set is encrypted with the supplied `oidcsession.Encryptor`.

Redis TTL is the earlier of idle and absolute expiry. `Status` never extends idle lifetime;
successful `Refresh` calls do, including calls that reuse a still-valid cached access token. When a
provider refresh is required, a single-key Lua lease coordinates replicas and atomically commits
the rotated token set. Provider revocation after atomic local deletion is best effort.

The 15-second refresh lease provides best-effort single-flight behavior. A provider call that
outlives the lease may be repeated by another replica, while lease-owner fencing still prevents a
stale Redis commit.

A crash after the provider rotates a refresh token but before Redis commit can invalidate that
session. The backend deliberately fails closed and requires login again.
