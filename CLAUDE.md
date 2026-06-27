# CLAUDE.md

Guidance for Claude Code when working in this repository.

## Project

A Go client library for the [Chef Supermarket API](https://docs.chef.io/supermarket/supermarket_api/).
It is consumed by `cinc-cli` and others, so exported types and method
signatures are a public API — avoid breaking them without reason.

## Build & test

- There is no Makefile. Build with `go build ./...` and test with
  `go test ./...`. The suite is fully `httptest`-based — no network
  calls and no build tags — so it runs in well under a second and is
  safe to run on every change. Don't reach for a narrower scope.
- Single test: `go test -run TestName ./...`.
- Run `gofmt -w .` and `go vet ./...` before committing; both must be clean.
- Go 1.26.

## Layout

- `client.go` — the `Client`, which wires one service per resource:
  `Cookbooks`, `Search`, `Tools`, `Users`, `Universe`, `Health`. Each
  service lives in its own `<name>.go` with a sibling `<name>_test.go`.
- `transport.go`, `options.go`, `errors.go`, `pagination.go`,
  `response.go` — shared HTTP plumbing: functional options
  (`WithHTTPClient`, `WithSkipTLSVerify`, …), typed errors
  (e.g. `ErrNotFound`), and pagination.
- `internal/signing` — the Chef v1.3 signed-header protocol used by the
  write endpoints (share/delete). Read endpoints are anonymous.
- `testdata/test_key.pem` — RSA key fixture for signing tests.

## Conventions

- **TDD.** Write a failing test first, then the minimal code to pass.
- Tests stand up an `httptest.Server` and build a client through the
  `newTestClient(t, srv, signed)` helper in `testhelpers_test.go`
  (`signed=true` attaches credentials for write endpoints). Use it
  rather than hand-rolling a `Client`; use `testRSAKey(t)` for keys.
