# S3 Replicator Tool Specification

## Overview

A one-shot, pull-based replication tool for synchronizing subsets of S3-compatible buckets between independent homelabs. Designed for disaster recovery, minimal configuration, and safe-by-default replication.

---

## High-Level Behavior

* Each peer pulls data from remote S3-compatible endpoints.
* Replication is unidirectional.
* Each peer operates independently using only S3 access.
* Status tracking is done via per-peer status objects stored in the source bucket.
* Configuration is defined via a single TOML file.

---

## Replication Rules

* Operates at the bucket level.
* Only the latest version of each object is replicated.
* Objects are selected using a `ReplicationPriority` tag:

  * Default: `0` (normal)
  * Higher values = higher priority; lower/negative = deprioritized
  * Within the same priority, objects are selected randomly
* Partial replication is allowed (e.g., limited by local quota).
* Tool favors objects that are less replicated across peers.
* Deletes are immediate when objects are removed or fall below threshold.

---

## Status Object

* Written to: `.replication-status/<peer-id>.json.zst`
* Format: zstd-compressed JSON
* Fields:

  * `peer_id`: string
  * `source_bucket`: string
  * `run_id`: UUID
  * `last_sync`: Unix timestamp (int)
  * `total_objects`: int
  * `total_bytes`: int
  * `objects`: list of `{ key, version, size, mtime }`
* Written at end of run and every 60s or 500 objects (whichever comes first)

---

## Local Storage

* Replicated objects are written to one pre-created local bucket.
* Key prefix format: `peer-id/bucket-name/...`
* Preserves `Content-Type` and object tags
* Preserves `LastModified` if possible
* Deletes local objects no longer eligible for replication

---

## Configuration

* Format: TOML
* Required flag: `--config /path/to/config.toml`
* Structure:

```toml
[[peer]]
peer_id = "alice"
endpoint = "https://s3.example.com"
access_key_id = "abc"
secret_access_key_file = "/run/secrets/alice.key"
insecure_skip_verify = false

  [[peer.bucket]]
  name = "backup"
  quota_bytes = 100_000_000_000
```

---

## Command-line Interface

* `--config`: path to TOML config (required)
* `-v`: enable debug logging
* `--dry-run`: simulate actions, write nothing
* `--version`: print tool version and exit

---

## Logging & Metrics

* Uses Go `log/slog` in plain text (key=value)
* Info level by default; `-v` enables debug
* Logs summary per bucket and at end of run
* Includes unique `run_id`
* Logs directly to stdout/stderr

---

## Error Handling

* Network/S3 errors: log and continue
* Invalid/missing peer status:

  * 404: assume no objects replicated
  * Others: warn and skip
* Missing local bucket: warn and skip
* Duplicated `peer_id`: fatal error
* Corrupt config: fatal error
* Exits with `0` on success, `1` on any error

---

## Networking

* Supports IPv4, IPv6, hostnames
* Uses HTTP or HTTPS (no special config needed)
* Validates TLS certs by default
* Optional `insecure_skip_verify` per peer

---

## Performance

* Parallel object downloads with global concurrency setting
* Simple GET requests; multipart download for files >10MiB
* Object listing is batched; downloads are per-object
* Retries not attempted during run—failures retried next run

---

## Development and Testing

* S3 access abstracted behind an interface for mocking
* No persistent database or state files
* Status and config validation on startup

---

## Not Implemented (TODOs)

* Support for `exclude_prefixes`
* Filtering by `--peer` or `--bucket`
* Granular exit codes
* Config disabling of peer–bucket pairs
* Proxy support (e.g. `HTTP_PROXY`)
* Timeout for total run
* Per-object timeout
* Detection of same endpoint used for local+remote (infinite loop prevention)

---

## Additional Questions (for future expansion)

1. Should object tag filtering support regular expressions or wildcards?
2. Do you want to support compression (e.g. zstd) of replicated object content?
3. Should the config support environment variable interpolation?
4. Should the tool validate free space before downloading each object?
5. Any plans for cross-region latency optimization or prefetching?
6. Would you want metrics integration in the future (e.g. Prometheus Pushgateway)?
7. Support for object locking / retention policies?
8. Retry policy customization (exponential backoff, max retries, etc)?
9. CLI subcommands (e.g. `replicator sync`, `replicator check`, etc)?
10. Per-bucket content-type filters?

---

Ready to hand off to a developer. Let me know if you want a stub Go module to get started.

