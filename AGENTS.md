# Agent instructions: authoring datasource schema-registry entries

This file guides AI coding agents (Claude Code, GitHub Copilot, Cursor, etc.) working in this
repository. Its primary workflow is **adding or updating a datasource configuration schema under
`schema-registry/<plugin_id>/`**. The instructions are agnostic of the datasource type — follow the
same process for any Grafana datasource plugin. A complete worked example lives in
[`schema-registry/grafana-github-datasource/`](schema-registry/grafana-github-datasource/).

A Claude Code skill wrapping this workflow is available at
[`.claude/skills/add-datasource-schema/SKILL.md`](.claude/skills/add-datasource-schema/SKILL.md).

## Repository layout

| Path | Purpose |
| --- | --- |
| `dsconfig/` | The dsconfig schema SDK: Go types (`schema.go`), validator, `baseFields` packs, dsconfig→SDK converter (`convert.go`), and the JSON Schema (`schema.json`) every `dsconfig.json` must satisfy |
| `schema/` | Conformance test suite and artifact helpers plugins import |
| `schema-registry/<plugin_id>/` | One entry per datasource plugin (see structure below) |
| `go.work` | Workspace — every registry entry module must be added here |

## Registry entry structure

Each entry is a standalone Go module:

```
schema-registry/<plugin_id>/
├── dsconfig.json    # dsconfig v1 schema — the single source of truth
├── config.ts        # TypeScript models: RootConfig, JsonDataConfig, SecureJsonDataConfig
├── config.go        # Go models (same three types) + LoadConfig utility
├── schema.go        # k8s-style SDK PluginSchema: embeds dsconfig.json + SettingsExamples
├── schema_test.go   # Guards the schema bundle shape and LoadConfig behavior
├── go.mod / go.sum  # module github.com/grafana/dsconfig/schema-registry/<plugin_id>
└── README.md        # Research notes, field inventory, discrepancies, type provenance
```

## Step 1 — Capture the inputs (research phase)

Fidelity comes from reading the real sources, never from memory:

1. **Clone the plugin repository** (e.g. `github.com/grafana/<plugin_id>`) and read:
   - the config editor component (usually `src/**/ConfigEditor.tsx`) — every label, placeholder,
     tooltip, option, section title, conditional render, and side-effecting change handler;
   - the frontend config types (usually `src/types/config.ts` or `src/types.ts`);
   - the backend settings model (usually `pkg/models/settings.go`) and its `LoadSettings` —
     legacy fallbacks, lenient parsing, defaulting;
   - how each setting is **consumed** (HTTP client construction, URL derivation, auth wiring) —
     this reveals frontend-only and backend-only fields;
   - `src/plugin.json` for the plugin ID, name, and docs URL.
2. **Resolve external components.** Config editors compose components from libraries
   (`@grafana/ui`, `@grafana/plugin-ui`, `@grafana/experimental`, SDK field packs). Pin the exact
   versions from the plugin's `package.json`/`go.mod` and read those components' sources for their
   labels, tooltips, and the storage keys they write. Never guess what a library component renders.
3. **Inventory every storage field** across three targets: `root` (top-level datasource settings),
   `jsonData`, and `secureJsonData`. Classify each field: editor-visible, frontend-only (written by
   the editor, never read by the backend), backend-only (no editor UI), or virtual (editor-local
   derived state that never hits storage).
4. **Record discrepancies** you find upstream (dead settings, misleading placeholders, typos,
   validation gaps, URL-handling quirks) — they go in the entry README, not in the schema.

## Step 2 — Author `dsconfig.json`

Validate against `dsconfig/schema.json` (`$schema` is required and must be the canonical URL).

- **Exact fidelity**: labels, placeholders, tooltips, option labels/values, section titles, and
  help text must match the config editor **verbatim — including upstream typos**. Do not invent
  tooltips: set a field `description` only where the editor actually shows one; put supplementary
  facts in `instructions` or the README.
- **Field ID naming convention**: `<target>_<camelCaseKey>` — `root_`, `jsonData_`, or
  `secureJsonData_` prefix matching the storage target (`virtual_` for virtual fields), e.g.
  `jsonData_appId`, `secureJsonData_accessToken`. No dot notation. The `key` property keeps the
  plugin's raw storage key.
- **Virtual fields**: model editor-local derived selectors (React state, not storage) as
  `kind: "virtual"` with a `storage.computed.read` expression for the load-time derivation and
  `effects` declaring the multi-field writes each selection performs. Tag driven storage fields
  `managed-by:<virtual_field_id>`.
- **Conditionals**: `dependsOn` mirrors editor visibility (may reference virtual fields);
  `requiredWhen` encodes the backend data contract (reference storage fields) even when the editor
  shows no required markers.
- **Groups**: mirror the editor's sections; the connection group comes first, then authentication,
  then the rest. Collapsible sections get `optional: true`.
- **Help drawers**: rich editor help (collapse panels, multi-step guidance) becomes the `help`
  object of the most relevant field, with the markdown preserved verbatim.
- **Roles**: apply roles from the closed vocabulary (`auth.discriminator`, `auth.bearer.token`,
  `endpoint.baseUrl`, …) wherever a field's meaning matches; skip fields with no matching role.
- **Exclusions**: do not include the Secure Socks Proxy field (`jsonData.enableSecureSocksProxy`)
  in registry entries.
- **Instructions**: maximum **6** entries, tagged (include `llm`), prioritizing crucial
  authentication guidance over implementation/interpretation details. Must cover: available auth
  methods and how to select one, a minimal JSON payload recipe per auth method (jsonData +
  secureJsonData), legacy auth interpretation, write-only secure values (`secureJsonFields` for
  read-side), and connection/URL rules with known pitfalls.

## Step 3 — Author `config.ts` and `config.go`

Both files must export exactly three config types, with doc comments citing the upstream sources:

- **`RootConfig`** — root-level (top-level datasource settings) fields only. If the plugin stores
  nothing at root, it is a **blank object** (`Record<string, never>` / `struct{}`), never null.
- **`JsonDataConfig`** — all `jsonData` fields (including frontend-only and backend-only ones),
  json-tagged in Go with the raw storage keys. Mark frontend-only/backend-only fields in comments.
- **`SecureJsonDataConfig`** — an **array of secret key names**, not a json-tagged struct
  (secure values are write-only). In Go, provide the key list as a `var` (e.g.
  `SecureJsonDataKeys`).

`config.go` additionally provides:

- **`LoadConfig(settings backend.DataSourceInstanceSettings) (Config, error)`** — parses instance
  settings into a `Config` wrapper (`Root`, `JSONData`, decrypted `Secrets` by key,
  `ConfiguredSecureKeys`). It must mirror the plugin's own `LoadSettings` behavior, including
  legacy fallbacks and lenient parsing (replicate the plugin's helper semantics, e.g.
  string-or-number ID parsing).
- Enum-like `string` types with constants for discriminator fields (auth type, license/plan …),
  mirroring the plugin's own constants where they exist.

## Step 4 — Author `schema.go` (k8s-style SDK schema)

- Embed `dsconfig.json` (`//go:embed`); expose `ConfigSchema()` (parse + resolve) and
  `NewSchema()` via `dsconfig.NewSDKSchema` — this produces the `pluginschema.PluginSchema`
  bundle (OpenAPI settings spec + `secureValues` + examples) Grafana's datasource API server
  serves as `{apiVersion}.json`.
- **`SettingsExamples()`**: one example per authentication type and connection variant, plus a
  **default example keyed by the empty string `""`** capturing the schema defaults. Every example
  value is a full instance-settings object: plugin config under `jsonData` **and the relevant
  `secureJsonData` placeholder fields** (empty string in the default example; realistic
  placeholders elsewhere — use the correct secret format, e.g. a proper PEM header for keys).
- Include a legacy example if the plugin has a legacy storage shape.

## Step 5 — Wire the module and validate

1. `go.mod`: module `github.com/grafana/dsconfig/schema-registry/<plugin_id>`, with
   `replace github.com/grafana/dsconfig/dsconfig => ../../dsconfig`; run `go mod tidy`.
2. Add `use ./schema-registry/<plugin_id>` to the repo `go.work`.
3. `schema_test.go` must assert at minimum: `NewSchema()` succeeds; `secureJsonData` is **not** in
   the settings spec; `secureValues` match the secure key list; every expected `jsonData` property
   is in the spec; the `""` default example exists; every example has `jsonData` and a non-empty
   `secureJsonData` using only known secret keys; `LoadConfig` handles each auth method, legacy
   fallback, and malformed input.
4. Full validation checklist (all must pass before committing):
   - `dsconfig.ParseAndResolveSchemaJSON` + `Validate()` on `dsconfig.json`;
   - JSON Schema validation against `dsconfig/schema.json` (draft 2020-12, strict —
     `additionalProperties: false`);
   - `go build ./... && go vet ./... && gofmt -l . && go test ./...` in the entry module;
   - `tsc --noEmit --strict` on `config.ts`;
   - the pre-existing `dsconfig` and `schema` workspace modules still build.

## Step 6 — Write the entry `README.md`

Required sections: file table; sources researched (with exact library versions); field inventory
table (schema ID, storage key, target, editor label, read-by-backend); frontend-only and
backend-only settings; modeling decisions; **where the types are defined** (frontend and backend,
including types that come from libraries/packages/SDKs rather than the plugin itself); settings
examples matrix; potential upstream bugs/discrepancies; validation performed.

## General guidelines

- Never push to a branch other than the designated working branch; never create a PR unless asked.
- Prefer small, reviewable commits with descriptive messages, one concern per commit.
- When information conflicts between the editor UI and the backend, capture **both**: the editor
  behavior in field presentation, the backend contract in validations/`requiredWhen`, and the
  conflict itself in the README's discrepancies section.
