# golangci-lint cleanup — remaining plan (Tiers 3 & 4)

Status: Tiers 1 and 2 are **done** for the v3 module (repo root). The v2 module
is deliberately untouched — it is in maintenance only.

Baseline: v3 went from **43 findings to 8**. All 8 remaining are the complexity
cluster described in Tier 3.

## Background: how these surface

golangci-lint runs only in CI, via the `Lint` step in
`.github/workflows/build_and_test.yml`. That step sets `continue-on-error: true`,
so findings appear as GitHub *annotations* on a green run rather than as a
failure. GitHub caps annotations at 10 per step, and lint runs on 4 matrix legs
(ubuntu/macOS × v2/v3), which is why the run summary shows exactly 40 and why
each finding appears twice.

`make test` runs `fmtcheck staticcheck vet` only, so it has never covered the
23 linters golangci-lint enables. `make lint` (added in Tier 4a) closes that gap.

### Gotcha worth knowing

golangci-lint defaults to `issues.uniq-by-line: true` — only the first finding
on any given line is reported. Silencing one can therefore *reveal* another on
the same line. This happened during Tier 2: suppressing gosec G204 on the
`exec.Command` line exposed a previously hidden `noctx` finding. Expect counts
not to decrease monotonically.

## Tier 3 — complexity refactors (the remaining 8)

These are genuine technical debt, but each is a real refactor with regression
risk. They were deliberately deferred rather than bundled into a lint-cleanup
commit, particularly with v3.0.0 in flight.

| File | Finding | Notes |
|---|---|---|
| `ahoy.go:293` | `(*appState).getCommands` — cognitive complexity **125** (limit 30) | The big one. Also the root of two of the nestif findings. |
| `config_validation.go:474` | `PrintConfigReport` — complexity **50** | Pure presentation logic; the safest of the three to split. |
| `config_validation.go:63` | `compareVersions` — complexity **33** | Small and well covered by tests; a good first candidate. |
| `ahoy.go:355` | `if cmd.Cmd != ""` — nestif 9 | Inside `getCommands`. |
| `ahoy.go:451` | `if cmd.Imports != nil` — nestif 17 | Inside `getCommands`. |
| `ahoy.go:617` | `if helpRequested` — nestif 7 | Help-dispatch branch. |
| `config_init.go:86` | nestif 9 | Config bootstrap branch. |
| `config_validation.go:129` | nestif 5 | Just over the line; may resolve for free with `compareVersions`. |

Note that the `nestif` findings substantially overlap the `gocognit` ones — five
of the eight live inside functions already flagged. Fixing the three functions
should clear most or all of the nesting findings as a side effect, so treat this
as **three units of work, not eight**.

### Suggested order

1. **`compareVersions`** — smallest, strong test coverage, likely clears the
   `config_validation.go:129` nestif too. Good calibration for the rest.
2. **`PrintConfigReport`** — presentation only, so a bug here is visible rather
   than silent. Extract per-section printers.
3. **`getCommands`** — do this alone, on its own branch, after v3.0.0 ships.
   Extract the import-resolution, entrypoint-building and env-loading phases
   into named helpers. The 132-test BATS suite is the real safety net here;
   run it before and after and diff the output.

### Prerequisite before starting

The BATS suite cannot be run from a git worktree nested under
`<repo>/.claude/worktrees/`, because ahoy searches parent directories for
`.ahoy.yml` and finds the parent repo's own config. `tests/no-ahoy-file.bats`
fails for that reason alone. Run the suite from a checkout outside the repo tree
(or from the repo root itself) when validating these refactors.

## Tier 4 — closing the loop

**4a. `make lint` target — DONE.** Added to the root `Makefile`, mirroring the
CI step. Deliberately *not* wired into `test:` while the Tier 3 findings are
outstanding, so it cannot block normal work. Requires golangci-lint on PATH and
prints an install hint if missing. Pin note: CI uses **v2.12.2**; the local
target uses whatever is installed, so check versions if results disagree.

**4b. Make lint blocking — blocked on Tier 3.** Once the 8 findings are cleared:

1. Add `lint` to the `test:` prerequisites in the root `Makefile`.
2. Remove `continue-on-error: true` from the `Lint` step in
   `build_and_test.yml`, and delete the comment above it explaining the
   non-blocking behaviour.
3. Update the `golangci-lint-ci-nonblocking` note, which documents the old
   rationale and will then be stale.

Do **not** do 4b before Tier 3 lands — it would turn a green build red.

**4c. v2 module — explicitly out of scope.** v2 still reports ~36 findings and
still emits annotations on 2 of the 4 matrix legs. Options when it comes up:
apply the same Tier 1 config exclusions (drops it to ~25 for a one-line change),
or drop the v2 leg from the lint matrix entirely on the grounds that it is
frozen. Fixing v2's Go code is hard to justify.

## Known inconsistency, left alone deliberately

`ahoy.go` logs `[warn]` in two places (circular-import detection and the
missing-flag message) and `[warning]` everywhere else. This is user-visible
output and is **asserted by the BATS suite** (`tests/no-ahoy-file.bats:17`,
`tests/simple.bats:9`), so it was not changed. Both spellings are now captured
as `logLevelWarn` and `logLevelWarning` in `constants.go`.

Unifying them is a small breaking change to output. If that is wanted, do it as
its own commit that also updates the two BATS assertions, so the behaviour
change is visible in review rather than buried in a lint cleanup.
