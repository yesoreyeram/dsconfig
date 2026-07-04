---
name: add-datasource-schema
description: Create or update a datasource configuration schema entry under registry/<plugin_id>/ (dsconfig.json, settings.ts, settings.go, schema.go, tests, README). Use when asked to add a dsconfig schema for a Grafana datasource plugin, capture a plugin's config editor as a schema, or extend an existing registry entry.
---

# Add a datasource registry entry

Follow the canonical, datasource-agnostic workflow in [AGENTS.md](../../../AGENTS.md) at the repo
root. Use [registry/grafana-github-datasource/](../../../registry/grafana-github-datasource/)
as the worked reference example. Summary of the workflow:

## 1. Research (capture the inputs)

**Proof-driven authoring is mandatory.** Every value that lands in `dsconfig.json` (label,
placeholder, tooltip, description, option label/value, section title, help text, value type,
default, validation, required marker, visibility condition, storage key, storage target) must
be traceable to a specific `file:line` in the upstream repo at the HEAD of `main`. No memory,
no guessing, no cached context — read the real sources.

**Plugin ID lookup**: the authoritative plugin ID is the `id` field of `src/plugin.json` in the
upstream repo (not the repo name, npm name, or Go module path). Read it first and use it verbatim
as the registry entry directory and as `pluginType` in `dsconfig.json`; also capture `name`
(→ `pluginName`) and `info.links[]` (→ `docURL`).

**Always fetch the latest `main` before reading anything** — `git clone` fresh or
`git -C <clone> fetch origin && git checkout main && git pull --ff-only`. Record the researched
commit SHA in the entry README so reviewers can reproduce the work.

Then read at the pinned HEAD: `src/plugin.json`, the config editor (`src/**/ConfigEditor.tsx`),
frontend config types, backend settings model + `LoadSettings`, and how each setting is
consumed by the backend. Resolve every external editor component (`@grafana/ui`,
`@grafana/plugin-ui`, SDK packs) at the exact version pinned in the plugin's `package.json` —
read their sources for labels/tooltips and storage keys. Inventory all fields by target
(root / jsonData / secureJsonData) with `file:line` provenance, classifying each as
editor-visible, frontend-only, backend-only, or virtual. Record upstream discrepancies for
the README.

## 2. dsconfig.json

- Verbatim fidelity to the editor (labels, placeholders, tooltips, options, help text — typos
  included). No invented descriptions.
- Field IDs: `<target>_<camelCaseKey>` (`root_`, `jsonData_`, `secureJsonData_`, `virtual_`);
  `key` keeps the raw storage key.
- Editor-local derived selectors → `kind: "virtual"` + `storage.computed.read` + `effects`;
  tag driven fields `managed-by:<virtual_id>`.
- `dependsOn` = editor visibility; `requiredWhen` = backend contract.
- Groups mirror editor sections: connection first, then authentication, then the rest.
- Exclude the Secure Socks Proxy field.
- `instructions`: max 6, tagged `llm`, auth guidance first — auth methods, minimal JSON payload
  per method, legacy interpretation, write-only secrets, connection/URL pitfalls.

## 3. settings.ts / settings.go

`settings.ts` exports exactly `RootConfig` (blank object if no root fields — never null),
`JsonDataConfig` (all jsonData fields keyed by raw storage names), and `SecureJsonDataConfig`
(array of secret key names).

`settings.go` exports a **flat `Config` struct** mirroring the plugin's upstream backend `Settings`
(`pkg/models/settings.go`) verbatim — same fields, same json tags, same custom `UnmarshalJSON` if
upstream has one — plus a `DecryptedSecureJSONData map[SecureJsonDataKey]string`. **Only carry root-level fields
(`URL`, `BasicAuth`, `User`, …) on `Config` when the plugin's backend actually reads them**; most
datasources ignore root fields and should omit them. If included, tag them `json:"-"`. Also add
`SecureJsonDataKey` (strict string alias) with typed constants for each secret key, enum-like
constants for discriminator fields, and `LoadConfig(ctx context.Context, backend.DataSourceInstanceSettings)`
mirroring `LoadSettings` (legacy fallbacks, lenient parsing, conditional int64 conversions) with
contextual logging via `backend.Logger.FromContext(ctx)`. `LoadConfig` internally runs the full parse → `ApplyDefaults` → `Validate` sequence and returns a
fully-defaulted, validated `Config`. Keep `(*Config).ApplyDefaults()` (curated editor-parity
defaults on zero-valued discriminators) and `(Config).Validate() error` (runtime contract check)
exported as separate methods so callers that assemble a `Config` outside of `LoadConfig` can still
invoke them individually. Document the internal three-phase sequence in the entry README.

## 4. schema.go + schema_test.go

Embed dsconfig.json; `NewSchema()` via `dsconfig.NewSDKSchema`. `SettingsExamples()`: default
example keyed `""` (schema defaults, empty secret), plus one example per auth type and connection
variant — each with `jsonData` and realistic `secureJsonData` placeholders. Tests guard: no
`secureJsonData` in the spec, secure values match the key list, `""` example exists, every example
carries valid secure keys, and `LoadConfig` behavior per auth method.

## 5. Wire + validate

Registry entries share a single `registry/go.mod` (module `github.com/grafana/dsconfig/registry`,
with `replace ../dsconfig` and `replace ../schema`); a new entry is just a subpackage — no new
`go.mod`, no `go.work` edit. Run `go mod tidy` inside `registry/` if new imports were added.
All must pass:
Go validator on dsconfig.json, strict JSON Schema check against `dsconfig/schema.json`,
`go build/vet/gofmt/test` in the module, `tsc --noEmit --strict` on settings.ts, and the existing
workspace modules still build.

## 6. README.md

Sources with versions, field inventory, frontend-only/backend-only settings, modeling decisions,
where the types are defined (plugin vs libraries/SDKs, frontend and backend — types/fields only,
no UI components or functions), examples matrix, upstream bugs/discrepancies, validation performed.
