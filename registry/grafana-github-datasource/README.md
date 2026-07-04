# grafana-github-datasource

Declarative configuration schema for the [GitHub datasource plugin](https://github.com/grafana/github-datasource) (`grafana-github-datasource`).

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + Secrets), `PluginID`, `SecureJsonDataConfig`, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`) |
| [`go.mod`](go.mod) | Standalone Go module (`replace`d onto the sibling `dsconfig` module; wired into the repo `go.work`) |

## Sources researched

All labels, placeholders, tooltips, options, and help text were taken verbatim from the plugin source (`main` branch, July 2026) and the exact library versions it pins:

- `src/views/ConfigEditor.tsx` — the configuration editor (labels, placeholders, section layout, conditional rendering, the "Access Token & Permissions" collapse).
- `src/types/config.ts` — `GitHubDataSourceOptions`, `GitHubAuthType`, `GitHubLicenseType`, `GitHubSecureJsonDataKeys`.
- `pkg/models/settings.go` — backend `Settings` model and `LoadSettings` (legacy defaults, `json.RawMessage` parsing).
- `pkg/github/client/client.go` — how each setting is actually consumed (API URL derivation, auth clients, proxy).
- `pkg/plugin/instance.go` — `cachingEnabled` handling.
- External editor components:
  - `DataSourceDescription` and `ConfigSection` from `@grafana/plugin-ui` `0.13.1` — intro text and the "Authentication" / "Connection" section titles.
  - `SecureSocksProxySettings` from `@grafana/ui` `12.4.2` (grafana/grafana tag `v12.4.2`, `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx`).

## Field inventory

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `virtual_selectedLicense` | — (virtual) | — | GitHub License Type | — (editor-local state) |
| `jsonData_githubPlan` | `githubPlan` | `jsonData` | — (managed by `virtual_selectedLicense`) | **No — frontend-only** |
| `jsonData_githubUrl` | `githubUrl` | `jsonData` | GitHub Enterprise Server URL | Yes |
| `jsonData_selectedAuthType` | `selectedAuthType` | `jsonData` | Authentication Type | Yes |
| `secureJsonData_accessToken` | `accessToken` | `secureJsonData` | Personal Access Token | Yes |
| `jsonData_appId` | `appId` | `jsonData` | App ID | Yes |
| `jsonData_installationId` | `installationId` | `jsonData` | Installation ID | Yes |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private Key | Yes |
| `jsonData_cachingEnabled` | `cachingEnabled` | `jsonData` | — (no UI) | Yes (backend-only) |

### Frontend-only settings

- **`githubPlan`** is written and read only by the config editor to drive the "GitHub License Type" radio. The backend never reads it — it infers Enterprise Server solely from a non-empty `githubUrl`. Selecting "Free, Pro & Team" vs "Enterprise Cloud" changes nothing in backend behavior.
- **`virtual_selectedLicense`** does not exist in storage at all: the editor's radio is backed by local React state (`selectedLicense` in `ConfigEditor.tsx`), derived from `githubPlan`/`githubUrl` on load. It is modeled here as a `kind: "virtual"` field with a `storage.computed.read` expression and `effects` describing the writes it performs.

### Backend-only settings

- **`cachingEnabled`** has no editor UI. It exists in the backend `Settings` model only (see "Potential bugs" below).

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields and base
types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `GitHubDataSourceOptions` (jsonData), `GitHubAuthType`, `GitHubLicenseType`, `GitHubSecureJsonDataKeys`, `GitHubSecureJsonData` | `src/types/config.ts` | plugin ([grafana/github-datasource](https://github.com/grafana/github-datasource)) |
| `DataSourceJsonData` (base interface `GitHubDataSourceOptions` extends: `authType`, `defaultRegion`, `profile`, `manageAlerts`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.2` (npm) |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption` | `packages/grafana-data` | `@grafana/data` `12.4.2` (npm) |
| `SecureSocksProxyConfig` / `enableSecureSocksProxy` jsonData field (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.2` (npm, grafana/grafana `v12.4.2`) |
| `ConfigSection`, `DataSourceDescription` (editor layout/intro, no storage fields) | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `0.13.1` (npm) |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + decrypted secrets), `AuthType` (`AuthTypePAT`, `AuthTypeGithubApp`), `LoadSettings`, `rawMessageToInt64` | `pkg/models/settings.go` | plugin ([grafana/github-datasource](https://github.com/grafana/github-datasource)) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`, `BasicAuthEnabled` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `httpclient.Options` (timeouts, TLS, `ProxyOptions`) consumed when building the GitHub clients | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `proxy.New(...).SecureSocksProxyEnabled()` (secure socks proxy wiring) | `backend/proxy` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| GitHub App installation transport (`ghinstallation.New`, JWT signing with `privateKey`, `itr.BaseURL`) | — | `github.com/bradleyfalzon/ghinstallation/v2` |
| REST / GraphQL clients the settings feed into (`WithEnterpriseURLs`, `NewEnterpriseClient`) | — | `github.com/google/go-github` / `github.com/shurcooL/githubv4` |
| `LicenseType` has **no backend equivalent** — `githubPlan` exists only in the frontend types | — | — |

The models in this entry flatten that spread into a single Go `Config` type (root fields + jsonData
fields + Secrets) plus `SecureJsonDataConfig`; `settings.ts` still exposes the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`). `LicenseType` constants
in `settings.go` are derived from the frontend union type since the backend defines none.

## Modeling decisions

- **Virtual license selector**: `onLicenseChange` writes `githubPlan` and clears `githubUrl` unless "Enterprise Server" is selected. This multi-field write is captured as `effects` on the virtual `virtual_selectedLicense` field; `jsonData_githubPlan` is tagged `managed-by:virtual_selectedLicense` and has no UI of its own.
- **`requiredWhen` vs the editor**: the editor renders `DataSourceDescription` with `hasRequiredFields={false}` and marks nothing required, but the backend hard-fails without credentials (`New` in `client.go` returns "access token or app token are required"). The `requiredWhen` rules encode that backend contract; an instruction records the editor discrepancy.
- **Help drawer**: the editor's top-level "Access Token & Permissions" `Collapse` is attached as the `help` drawer of `secureJsonData_accessToken`, with the markdown preserved verbatim (including upstream typos — see below).
- **Secure Socks Proxy excluded**: the editor conditionally renders `SecureSocksProxySettings` (writing `jsonData.enableSecureSocksProxy`) when the Grafana instance has `secureSocksDSProxyEnabled`, and both backend auth paths honor it. The field is deliberately omitted from this registry entry.
- **Field ID naming convention**: IDs are prefixed with their storage target for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and `virtual_` for virtual fields, which have no storage target) — followed by the camelCase storage key, e.g. `jsonData_appId`, `secureJsonData_accessToken`. The `key` property keeps the plugin's raw storage key (`appId`) — `id` is the schema reference, `key` is the storage contract.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and decrypted secrets onto a single `Config` struct (mirroring the upstream `Settings` in `pkg/models/settings.go` verbatim, json tags included). Root-level datasource fields (`url`, `basicAuth`, etc.) are not carried because the plugin does not use them. `settings.ts` keeps the three canonical TS types.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type is just the array of secret key names (`accessToken`, `privateKey`); consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the k8s-style
schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1` today) from the
embedded `dsconfig.json`: root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication type and connection variant. Each example is a full instance-settings object with the
plugin configuration nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values to be replaced with real secrets; the default example — keyed by
the empty string `""` — carries an empty `accessToken` to show what must be filled in):

| Example | Auth | Connection | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Personal Access Token (schema defaults) | GitHub.com (Free, Pro & Team) | `accessToken` (empty) |
| `personalAccessToken` | Personal Access Token | GitHub.com (Free, Pro & Team) | `accessToken` |
| `githubApp` | GitHub App | GitHub.com (Free, Pro & Team) | `privateKey` |
| `enterpriseCloud` | Personal Access Token | Enterprise Cloud (same endpoints as GitHub.com) | `accessToken` |
| `enterpriseServer` | Personal Access Token | Enterprise Server (`githubUrl`) | `accessToken` |
| `githubAppEnterpriseServer` | GitHub App | Enterprise Server (`githubUrl`) | `privateKey` |
| `legacyAccessTokenOnly` | Legacy: token with no auth type | GitHub.com | `accessToken` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)` runs
the full three-phase load flow on a datasource instance's settings and returns a fully-defaulted,
validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (the `Config.UnmarshalJSON` normalizes the legacy
   string-or-number `appId` / `installationId`), copy decrypted secrets into `Secrets`, run the
   upstream legacy fallback that promotes a lone `accessToken` to `personal-access-token`, and — under
   `github-app` auth only — parse `AppIdInt64` / `InstallationIdInt64`.
2. **`ApplyDefaults`** — fill a curated set of zero-valued discriminators with the same defaults
   the editor writes for a fresh datasource (`SelectedAuthType=AuthTypePAT`,
   `GithubPlan=LicenseTypeBasic`).
3. **`Validate`** — enforce the runtime contract (auth method + its required inputs, and
   `githubUrl` when the plan is Enterprise Server). Errors are joined so every problem surfaces at
   once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

This is the intended shape for the plugin's own upstream `LoadSettings` to sync to: a load
returns a config that is safe to use, or an error explaining why it isn't.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported for callers that
want to compose them themselves (e.g. provisioning preview, schema-example round-trip, tests
that need to distinguish parse-level from policy-level errors). Skip them by never calling
`LoadConfig` in those flows — assemble a `Config` directly.

## Potential bugs and discrepancies found upstream

1. **`githubPlan` is dead weight for the backend.** The backend decides Enterprise Server purely from `githubUrl` being non-empty (`client.go`), so a provisioned datasource with `githubPlan: "github-enterprise-server"` but no `githubUrl` silently behaves like github.com.
2. **Stale `githubUrl` overrides the stored plan in the editor.** The load derivation in `ConfigEditor.tsx` is `githubPlan === 'github-enterprise-server' || githubUrl ? 'github-enterprise-server' : githubPlan || 'github-basic'` — a datasource with `githubPlan: "github-basic"` but a leftover `githubUrl` (possible via provisioning or the API) displays and behaves as Enterprise Server.
3. **`cachingEnabled` cannot actually be disabled.** `NewDataSourceInstance` (`pkg/plugin/instance.go`) unconditionally sets `datasourceSettings.CachingEnabled = true` after `LoadSettings`, so the stored value is ignored and there is no way to turn the caching wrapper off.
4. **Misleading Private Key placeholder.** The `SecretTextArea` placeholder is `-----BEGIN CERTIFICATE-----`, but a GitHub App private key is an RSA private key PEM (`-----BEGIN RSA PRIVATE KEY-----`), not a certificate. Preserved verbatim in the schema.
5. **Trailing slash in `githubUrl` produces double-slash API URLs.** The editor placeholder suggests `http(s)://HOSTNAME/`, yet the backend builds `fmt.Sprintf("%s/api/v3", url)` for the GitHub App transport and `fmt.Sprintf("%s/api/graphql", url)` for GraphQL, yielding `HOSTNAME//api/...` when the placeholder is followed literally. Only the REST client (`WithEnterpriseURLs`) normalizes the URL.
6. **Incomplete GitHub App config hard-fails settings load.** With `selectedAuthType: "github-app"`, `LoadSettings` runs `ParseInt` on `appId`/`installationId`; missing or non-numeric values return "error parsing app id" and instance creation fails. The editor performs no validation on these inputs before save.
7. **Typos in the editor help text** (preserved verbatim in the `help` markdown to match the UI): "a access token" (should be "an access token"), "the Github documentation." (GitHub), "Meta data" (GitHub calls the permission "Metadata"), and inconsistent mid-sentence capitalization ("Select", "Ensure").
8. **`SecretInput` update quirk for the access token.** The token's `onChange` is `onSettingUpdate('accessToken', false)`, which sets `secureJsonFields.accessToken = false` while typing — intentional to keep the input editable, but it means `secureJsonFields` temporarily reports the secret as unconfigured until save.
9. **Asymmetric proxy wiring between auth paths** (works, but inconsistent): the GitHub App path gets secure-socks support implicitly via SDK `httpclient.New(opts)`, while the PAT path manually rebuilds the `oauth2` transport only when `proxy.New(opts.ProxyOptions).SecureSocksProxyEnabled()`.
10. **Legacy auth-type fallback is one-way.** `LoadSettings` defaults `selectedAuthType` to `personal-access-token` when an `accessToken` exists with no auth type, but there is no equivalent fallback for old GitHub App configs; the schema records the PAT fallback as an instruction.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values, examples, `LoadConfig` incl. legacy fallback and id parsing).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
