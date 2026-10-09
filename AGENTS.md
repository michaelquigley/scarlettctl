# scarlettctl

A Go library and command-line tool for the mixer, routing, and preamp controls of Focusrite Scarlett, Vocaster, and Clarett USB interfaces on Linux, through the ALSA control interface. The library is the product; `scarlettctl` is its command line. Published as open source; tested on 4th-generation Scarletts only.

## Status

The library and CLI are in use: card discovery, control enumeration and read/write, the routing matrix, mixer levels, preamp settings, and event monitoring. sessionmixer and patchbay consume the library. No release has been tagged; consumers pin a commit, and `CHANGELOG.md` accumulates under `## Unreleased`. `docs/current/` is the record of what exists; what is ahead is on the roadmap.

## Stack

- Go, single module `github.com/michaelquigley/scarlettctl`.
- cgo against libasound. Every machine that builds it needs `libasound2-dev`, and the binary links the machine's own ALSA.
- `github.com/spf13/cobra` for the CLI, `golang.org/x/sys` for polling. No logging library: the library emits nothing, and the CLI prints to stdout and reports errors on stderr.

## Building

- `make` installs, `make test` runs tests plus `go vet`, `make clean` clears the project GOBIN.
- Nothing in `go test` touches hardware. Validation is by hand with a device attached, on the studio hosts, cross-checked against `alsamixer` or `amixer`.
- `scarlettctl list` names the cards; `scarlettctl controls <card> --verbose` is the fastest look at what a device exposes.

## Layout

- The root package is the library. `types.go` holds the types (`Card`, `Control`, `ControlType`, `RoutingSource`, `RoutingSink`, `PortCategory`, and the private ALSA handle); `card.go` discovers and opens cards; `control.go` enumerates, finds, reads, and writes controls; `routing.go`, `mixer.go`, and `preamp.go` are the typed views over the controls; `events.go` is the `EventMonitor`.
- `cgo.go` holds every libasound call. It is the only cgo file; the rest of the package is pure Go over the handle it returns.
- `cmd/scarlettctl/main.go` — the cobra CLI, one file: `list`, `controls`, `get`, `set`, `watch`, `routing`, `route`, `mixer`, `preamp`, `gain`, `phantom`.
- `docs/current/` — built behavior. `docs/future/roadmap/` — the roadmap. `docs/journal/` — project memory.

## Key conventions

- **libasound stays in `cgo.go`.** Other files never import `"C"`. The handle crosses the boundary as an `unsafe.Pointer` to C memory (`types.go` is not a cgo file, so it cannot name `*C.snd_ctl_t`); `cgo.go` applies the C type. It is not a `uintptr`: that hides the pointer from the toolchain and `go vet` rejects the round trip.
- **Generations differ in control names.** Gen 1 and Gen 2/3/4 devices name the same controls differently; lookups go by prefix or substring, and new behavior must hold for both namings.
- **Values are typed and validated.** Ranges and enum items are checked before a write, and string values (`on`/`off`, enum item names) are resolved by the library, not the caller.
- **The exported surface is shared.** sessionmixer and patchbay import this package at pinned commits. A change to an exported type or function is a change to both; say so in the changelog entry.
- Comments start lowercase unless the first word is an uppercase Go identifier. Output to the user is lowercase, with dynamic data in single quotes: `the user selected the 'value' setting`. Go files are named `dashManager.go`, tests `dashManager_test.go`. Never use emoji.
- **Changelog.** `CHANGELOG.md` follows the in-house convention: newest-first releases, prose entries each led by one of `FEATURE`/`CHANGE`/`FIX`, and an `## Unreleased` slot at the top that agents write into. The full spec is the grimoire's `software/conventions/changelog-convention.md`.

## Review

A change is gated by terminus over the working tree, against scarlettctl's canon (`terminus-canon/projects/scarlettctl`): run it before handing off, present each finding for Michael's decision (apply, adjust, or veto), and bring it to clean. A vetoed finding may be raised again; report it as vetoed. Full-repo reviews are maintenance, run only on request. Their findings are filed as roadmap cards in `inbox`, not fixed in passing. Until the canon project exists, say so in the handoff rather than skipping silently.

## Roadmap

This repo's roadmap lives in `docs/future/roadmap/` — one frontmatter-markdown item per file, per the roadmap convention in the grimoire (software/conventions/roadmap-convention.md). You may add items freely: write the file directly with required `title`, `state: inbox`, and `created:` (today, YYYY-MM-DD), optional `tags`/`source`/`log`, and a body that is a small, clear prompt -- the problem or solution to execute, not documentation of it; trust the code and the day's journal entry for what's discoverable, and point a `log:` stamp at the specific journal entry when a card leans on hard-won context. Everything above the first `##` heading is the prompt; supporting material that isn't the prompt goes in named sections below it (`## why` for justification, `## background` for a longer description), which are conventional, never required, and never validated. The filename is the slug of the title (lowercase ASCII, hyphens; discard every other character); never overwrite an existing file. Read sibling items for the shape.

Hard rules: never touch `order.yaml` (priority is the operator's judgment, set at triage); never commit roadmap changes unless directed — the uncommitted diff is the review queue; never delete items; edits change only the lines that express them. Label the kind from the house set when one fits: defect, documentation, enhancement, epic, feature, story; add `spike` alongside it when the work carries unknowns that need discovery.

## For agents arriving cold

This is `scarlettctl`, part of Michael Quigley's broader software practice. Its consumers (`sessionmixer`, `patchbay`) and related projects (`dfx`, `mercurius`, `terminus`) live as siblings in `~/Repos/q/products/`. The library was written by reading `alsa-scarlett-gui` (see `CREDITS.md`), which remains the reference for what the kernel driver exposes. The grimoire at `~/Repos/q/writing/grimoire` holds cross-project conventions; for scarlettctl, the in-repo `docs/` is authoritative. If you've been assigned a role, read the grimoire's `agents/agent-roles.md`.

## Project memory

Durable knowledge about this project lives in `docs/journal/`, dated files `docs/journal/YYYY-MM-DD.md`. This is project memory; it does not go in harness-local storage (`.claude/` or equivalent), where it's invisible to every other harness and collaborator and dies with the host. Concretely: do not write to your harness's memory directory or memory tool for this project — even when the harness presents it as the default place for durable knowledge. That tool is the silo this convention exists to replace; the journal is the only durable home.

On arrival, read the most recent entries to pick up where the last session left off, before you start changing things. Treat them as prior-session context, not verified truth — if an entry conflicts with the code or a `docs/current/` doc, the code wins.

Write the smallest entry that carries the session's durable insight, and nothing more. The test for every line: *would a competent agent get this wrong, or waste time rediscovering it, working from the tree alone?* If it's recoverable by reading the code, the diff, `docs/current/`, or git history, leave it out.

That filter keeps four kinds of thing and discards the rest:

- **Decisions whose rationale isn't visible in the result** — why a value was chosen, what a line guards against, why something that looks like dead code or a no-op is load-bearing.
- **Deliberate non-actions** — a change you considered and chose not to make, so the next agent doesn't "fix" it. An unchanged file leaves no trace in a diff.
- **Couplings that span files** — two places that must move together, an ordering that matters, an assumption one file makes about another.
- **Live state** — what's unverified, unfinished, or waiting on something external.

Skip change inventories, restatements of the diff, and play-by-play of how you worked. There's no write-time approval gate; Michael reviews on commit. Append to the day's file if it exists, and write the few lines you'd want the next agent to read — honest and self-contained.

## Commits

The operator commits; agents don't. Never run `git commit` or `git push` in this repo. Finish the edit, leave the change in the working tree (staged is fine), report what changed, and hand off — the uncommitted diff is the review queue and the commit is the operator's act of acceptance. Approval of a change is not direction to commit; only an explicit instruction to commit is, and only for that commit.
