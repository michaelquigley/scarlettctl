---
title: go vet flags unsafe.Pointer use in cgo.go
state: evaluating
created: 2026-09-21
tags: [defect]
log:
  - stamp: 2026-10-09
    note: fixed by holding the handle as unsafe.Pointer; five sites, not three — docs/journal/2026-10-09.md
---

Make `make test` pass: `go vet ./...` reports "possible misuse of unsafe.Pointer" at `cgo.go:215`, `cgo.go:245`, and `cgo.go:278`. Either restructure those conversions so the pointer arithmetic happens in a single expression vet accepts, or, if the ALSA usage is sound as written, document why and scope the vet check accordingly.

## why

The Makefile convention's gate (`go test ./... -count=1` plus `go vet ./...`) arrived in this repo on 2026-09-21 and fails on these three lines; nothing ran vet here before. A redundant-newline `fmt.Println` in `routing.go` that the same gate caught was fixed in that pass.
