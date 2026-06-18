# CLAUDE.md

Guidance for working in this repository.

## What this is

**pgtdm** (PostgreSQL Test Data Manager) — a Go CLI that anonymizes sensitive
data in PostgreSQL plain-text dumps (`pg_dump` output) while preserving
structure. It operates purely on the dump file; it never connects to a database.

Module path: `github.com/fufild/pgtdm`. Go 1.21.

## Build, test, run

```bash
make build      # compiles ./pgtdm (go build -o pgtdm ./cmd/main.go)
make test       # go test ./... -coverprofile=cover.out
make coverage   # writes cover.html
make open       # opens cover.html
make clean      # removes binary + coverage artifacts

go test -race ./...   # run with race detector (not in Makefile; use before commits)

./pgtdm --input=dump.sql --output=anonymized.sql --rules=rules.json
```

Flags: `--input` (required), `--output` (default `./output.txt`), `--rules`
(default `./rules.json`). The input file's name/extension is irrelevant — only
its contents (PostgreSQL `COPY` format) matter.

## Architecture

Entry point [cmd/main.go](cmd/main.go) parses flags, loads rules JSON into a
`map[string][]string` (table → column rules), opens the input, and runs the
anonymizer. Three internal packages:

- [internal/anonymizer/anonymizer.go](internal/anonymizer/anonymizer.go) — the
  core. Streams the dump line-by-line with `bufio`. Detects `COPY public.table
  (cols...)` headers, builds a column-name→index map, and for each tab-separated
  data row replaces values in configured columns. `\N` (NULL) is skipped; tables
  whose name starts with `_` are skipped entirely. `AnonymizedValueCount` tracks
  replacements.
- [internal/random/random.go](internal/random/random.go) — fake-value
  generators. Embeds `data/first_names.txt`, `data/last_names.txt`,
  `data/states.txt` via `//go:embed`. Single package-level seeded
  `*rand.Rand` (seeded from `time.Now().UnixNano()` — **output is not
  deterministic across runs**).
- [internal/file/file.go](internal/file/file.go) — open/create/close helpers.
- [internal/logging/logging.go](internal/logging/logging.go) — colored logging;
  `FatalError` logs and exits.

## How rules work

`rules.json`: keys are table names (e.g. `public.users`), values are arrays of
column rules. A rule is either:

- **Implicit** — column name doubles as the strategy: `"email"`, `"phone"`.
- **Explicit** — `"column:strategy"`: `"last_4:pin"`, `"access_token:uuid"`.

Matching is done on a *normalized* name (`normalizeColumnName`: trim, strip `"`
and `\`; then in `anonymizeValue`: lowercase + strip `_`). So `ip_address` →
`ipaddress`, which matches the `ipaddress` strategy. Strategies are the `case`
labels in `anonymizeValue` ([anonymizer.go:172-214](internal/anonymizer/anonymizer.go#L172-L214));
an unknown strategy returns an error and aborts the run.

## Critical behavior to keep in mind

- **Allowlist model.** Only columns listed in `rules.json` are touched.
  Everything else — every table not in the rules, every column not listed — is
  copied through **verbatim and silently**. There is no warning for unlisted
  tables/columns, and a rule naming a column that doesn't exist in the dump is
  silently ignored. Coverage is entirely the operator's responsibility.
- Anonymization happens per value with no cross-row consistency: the same source
  value (e.g. a user email appearing in two tables) becomes two *different* fake
  values. Joins on anonymized columns will not line up.
- `DDL`, sequences, and non-`COPY` statements pass through unchanged.

## Conventions

- Follow standard Go style; `gofmt`/`goimports` are assumed. Wrap errors with
  `fmt.Errorf("...: %w", err)`.
- Table-driven tests with `testify/assert` (see
  [internal/random/random_test.go](internal/random/random_test.go),
  [internal/file/file_test.go](internal/file/file_test.go)). Run `go test -race
  ./...` before committing.
- New anonymization strategies: add a generator in `internal/random` and a
  `case` in `anonymizeValue`; cover the generator with a format-asserting test.
