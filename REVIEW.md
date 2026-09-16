# Provider Review — 2026-09-16

Scope: `internal/client` (hand-rolled Immich API client) and `internal/provider`
(Terraform Plugin Framework resources/data sources), as of `c5d00aa`.

`go vet ./...` is clean and `go test ./...` passes, but test coverage is
schema/marshalling-only — no acceptance tests exercise actual CRUD logic,
which is how finding #1 below shipped unnoticed.

**Status: all 8 findings below are fixed on this branch** (commits
`9b425b2`..`d44344f`). Each section is left as originally written for
context; see the commit log for what actually changed. Two additional bugs
were discovered and fixed along the way, surfaced by the new acceptance
tests built for #5/#8: `user_resource.go`'s `Read()` turned an empty
`storage_label` into a known empty string instead of null (permanent
drift), and `shared_link_resource.go`'s `key` attribute had no
`UseStateForUnknown` plan modifier, so `Update()` failed outright since the
API response was discarded.

## Findings, by severity

### 1. Critical — `immich_album` silently drops updates to `users` and `asset_ids`

`internal/provider/album_resource.go` `Update()` only sends `name`,
`description`, `album_thumbnail_asset_id`, `order`, and
`is_activity_enabled` to the API. The `users` and `asset_ids` schema
attributes have no plan modifiers forcing replacement, so Terraform expects
`Update()` to apply changes to them — but it never calls
`AddUsersToAlbum`, `RemoveUserFromAlbum`, `UpdateAlbumUserRole`, or
`AddAssetsToAlbum`. Those four client methods
(`internal/client/album.go:154-238`) are never called from anywhere in the
provider.

Effect: editing `users` or `asset_ids` on an existing `immich_album` in
config has no effect on the server. Because `updateAlbumResourceModel()`
repopulates `data.Users` from the (unchanged) API response after the
no-op update, the state converges back to the *old* value, so the next
`terraform plan` shows the same diff again — a perpetual drift the user
can never apply away.

**Fix:** in `Update()`, diff `plan.Users`/`plan.AssetIds` against
`state.Users`/`state.AssetIds` and call the corresponding add/remove
endpoints before re-reading the album.

### 2. High — 11 of 16 resources don't handle "deleted outside Terraform"

`Read()` should call `resp.State.RemoveResource(ctx)` when the underlying
object is gone (404), so Terraform re-creates it on the next apply instead
of erroring forever. Only 3 resources do this — and only because their
`Read()` happens to list-and-search rather than get-by-id
(`activity_resource.go`, `face_resource.go`, `partner_resource.go`).

Missing in: `album`, `api_key`, `asset`, `library`, `memory`, `person`,
`shared_link`, `stack`, `tag`, `user`, `workflow`
(`system_config`/`admin_notification` are singleton-ish and largely exempt).

Effect: delete an album/tag/user/etc. via the Immich UI (or have it deleted
by another process) and every subsequent `terraform plan`/`apply` fails
with a generic "Client Error" instead of recreating the resource.

**Root cause / why it's hard to fix piecemeal:** see #3 — the client has no
structured way to tell a 404 apart from any other error.

### 3. High — no context propagation, no request timeout

Every one of the ~84 HTTP call sites in `internal/client` uses
`http.NewRequest(...)` instead of `http.NewRequestWithContext(ctx, ...)`,
and `client.go`'s `http.Client{}` has no `Timeout` set. Consequences:

- Terraform's own cancellation (Ctrl-C, provider-level timeouts) can never
  reach an in-flight HTTP call.
- A hung/unreachable Immich server blocks `terraform apply` indefinitely —
  there's no upper bound at all.

**Fix:** thread `context.Context` from each resource method down through
every `client.*` method into `http.NewRequestWithContext`, and set a
sane default `Timeout` on the shared `http.Client` (configurable via a
provider-schema attribute if you want it tunable).

### 4. Medium — client errors aren't structured

`doRequest()` returns `fmt.Errorf("status: %d, body: %s", ...)` — a plain
string. There's no typed error carrying the HTTP status code, so resources
can't reliably branch on "404 vs. anything else" without parsing the error
string. This is the underlying reason #2 is inconsistent instead of just
missing everywhere.

**Fix:** introduce an `APIError{StatusCode int; Body string}` returned from
`doRequest`, and add an `errors.As`-friendly helper (e.g. `client.IsNotFound(err)`)
that every `Read()` can call.

### 5. Medium — no acceptance tests

`go.mod` doesn't depend on `terraform-plugin-testing` at all. Existing tests
(`resources_test.go`, `internal/client/*_test.go`) only check schema shape
and JSON marshal/HTTP-roundtrip — nothing drives a resource through
Create→Update→Read→Delete against a mock server, so update-path bugs like
#1 have no test that would catch them.

**Fix:** add `terraform-plugin-testing`-based acceptance tests (gated by
`TF_ACC`, ideally run in CI against a disposable Immich instance via
docker-compose), and/or unit tests that call each resource's `Update()`
directly against an `httptest.Server` to assert the right endpoints are hit.

### 6. Medium — `UploadAsset` buffers the whole file in memory

`internal/client/asset.go` `UploadAsset()` builds the entire multipart body
in a `bytes.Buffer` via `io.Copy(part, file)` before sending. For an asset
manager whose core objects are photos/videos, this can spike memory
significantly on large files (RAW photos, 4K/8K video).

**Fix:** stream the multipart body with `io.Pipe` + a goroutine writing into
the pipe, so `http.NewRequest` gets a reader that's consumed incrementally.

### 7. Low — no static analysis in CI

`.github/workflows/test.yml` only runs `go test -v ./...` and a docs-sync
check. No `golangci-lint`/`staticcheck` step. Note this specifically
wouldn't have caught #1 — the dead methods are exported, so `unused`-style
linters don't flag them — but it's still worth adding for the usual class
of issues (ineffectual assignments, shadowed errors, etc.).

### 8. Low — secrets persisted as regular state attributes

`user.password`, `system_config` SMTP password, and `shared_link.password`
are plain (non-write-only) `Sensitive` string attributes, so their values
land in the Terraform state file in the clear (state encryption aside).
Terraform 1.11+ / protocol 6 write-only attributes would avoid persisting
these at all.

**Fix:** consider migrating these three fields to write-only attributes
where the provider's minimum supported Terraform version allows it.

## Suggested priority order

1. Fix #1 (album update bug) — small, high-value, user-facing correctness bug.
2. Fix #4 then #2 (structured errors → drift detection) — do together since #4 unblocks #2 cleanly across all resources.
3. #3 (context + timeout) — larger mechanical refactor, touches every client file; worth doing but plan it as its own PR.
4. #5 (acceptance tests) — ideally added alongside #1's fix so the regression is covered.
5. #6, #7, #8 — smaller, can be picked up opportunistically.
