---
name: add-datasource-schema
description: Create or update a datasource configuration schema entry under schema-registry/<plugin_id>/ (dsconfig.json, config.ts, config.go, schema.go, tests, README). Use when asked to add a dsconfig schema for a Grafana datasource plugin, capture a plugin's config editor as a schema, or extend an existing schema-registry entry.
---

# Add a datasource schema-registry entry

Follow the canonical, datasource-agnostic workflow in [AGENTS.md](../../../AGENTS.md) at the repo
root. Use [schema-registry/grafana-github-datasource/](../../../schema-registry/grafana-github-datasource/)
as the worked reference example. Summary of the workflow:

## 1. Research (capture the inputs)

Clone the plugin repo and read: the config editor (`src/**/ConfigEditor.tsx`), frontend config
types, backend settings model + `LoadSettings`, and how each setting is consumed by the backend.
Resolve every external editor component (`@grafana/ui`, `@grafana/plugin-ui`, SDK packs) at the
exact version pinned in the plugin's `package.json` — read their sources for labels/tooltips and
storage keys. Inventory all fields by target (root / jsonData / secureJsonData), classifying each
as editor-visible, frontend-only, backend-only, or virtual. Record upstream discrepancies for the
README.

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

## 3. config.ts / config.go

Export exactly `RootConfig` (blank object if no root fields — never null), `JsonDataConfig`
(json-tagged storage keys), and `SecureJsonDataConfig` (array of secret key names, no tagged
struct). In Go add enum-like constants for discriminators and
`LoadConfig(backend.DataSourceInstanceSettings)` mirroring the plugin's `LoadSettings`
(legacy fallbacks, lenient parsing).

## 4. schema.go + schema_test.go

Embed dsconfig.json; `NewSchema()` via `dsconfig.NewSDKSchema`. `SettingsExamples()`: default
example keyed `""` (schema defaults, empty secret), plus one example per auth type and connection
variant — each with `jsonData` and realistic `secureJsonData` placeholders. Tests guard: no
`secureJsonData` in the spec, secure values match the key list, `""` example exists, every example
carries valid secure keys, and `LoadConfig` behavior per auth method.

## 5. Wire + validate

Per-entry `go.mod` (replace `../../dsconfig`), add to root `go.work`. All must pass:
Go validator on dsconfig.json, strict JSON Schema check against `dsconfig/schema.json`,
`go build/vet/gofmt/test` in the module, `tsc --noEmit --strict` on config.ts, and the existing
workspace modules still build.

## 6. README.md

Sources with versions, field inventory, frontend-only/backend-only settings, modeling decisions,
where the types are defined (plugin vs libraries/SDKs, frontend and backend), examples matrix,
upstream bugs/discrepancies, validation performed.
