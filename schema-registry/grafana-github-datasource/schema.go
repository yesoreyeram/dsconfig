package githubdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source of
// truth for the GitHub datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the GitHub datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the GitHub
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the GitHub
// datasource, covering each authentication type and connection (license)
// variant the config editor supports. Each example value is a full instance
// settings object with the plugin configuration nested under jsonData.
// Secure values (accessToken, privateKey) are write-only and never part of
// the settings spec; they are called out in each example's description.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"personalAccessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub.com)",
					Description: "Authenticate against GitHub.com (Free, Pro & Team) with a fine grained personal access token. The accessToken secure value must be provided separately.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeBasic),
						},
					},
				},
			},
			"githubApp": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GitHub App (GitHub.com)",
					Description: "Authenticate against GitHub.com as a GitHub App installation. appId and installationId may be JSON strings or numbers; the privateKey secure value must be provided separately.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypeGithubApp),
							"githubPlan":       string(LicenseTypeBasic),
							"appId":            "123456",
							"installationId":   "12345678",
						},
					},
				},
			},
			"enterpriseCloud": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub Enterprise Cloud)",
					Description: "GitHub Enterprise Cloud uses the same API endpoints as GitHub.com, so no URL is configured; githubPlan only drives the config editor. The accessToken secure value must be provided separately.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeEnterpriseCloud),
						},
					},
				},
			},
			"enterpriseServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub Enterprise Server)",
					Description: "On-prem GitHub Enterprise Server: the backend derives <githubUrl>/api/v3 (REST) and <githubUrl>/api/graphql (GraphQL). The accessToken secure value must be provided separately.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeEnterpriseServer),
							"githubUrl":        "https://github.example.com",
						},
					},
				},
			},
			"githubAppEnterpriseServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GitHub App (GitHub Enterprise Server)",
					Description: "GitHub App installation on an on-prem GitHub Enterprise Server. The privateKey secure value must be provided separately.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypeGithubApp),
							"githubPlan":       string(LicenseTypeEnterpriseServer),
							"githubUrl":        "https://github.example.com",
							"appId":            "123456",
							"installationId":   "12345678",
						},
					},
				},
			},
			"legacyAccessTokenOnly": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: access token without an auth type",
					Description: "Datasources created before selectedAuthType existed store only the accessToken secure value; the backend defaults them to personal-access-token.",
					Value: map[string]any{
						"jsonData": map[string]any{},
					},
				},
			},
		},
	}
}
