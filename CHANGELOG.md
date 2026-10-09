# CHANGELOG

## Unreleased

## v0.1.0

FIX: `go vet` no longer reports `possible misuse of unsafe.Pointer` in `cgo.go`, so `make test` passes. The ALSA handle is held as an `unsafe.Pointer` rather than a `uintptr`; no behavior changes.

CHANGE: GitHub CI builds and tests every push, and version tags prepare draft source releases from this changelog.

CHANGE: The project adopted the practice's changelog, `docs/current/`, and journal conventions.
