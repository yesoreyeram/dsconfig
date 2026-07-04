// Package githubdatasource contains the configuration models for the
// GitHub datasource plugin (grafana-github-datasource).
//
// It mirrors the plugin source of truth:
//   - pkg/models/settings.go (Settings, AuthType)
//   - src/types/config.ts (GitHubLicenseType, GitHubAuthType,
//     GitHubDataSourceOptions, GitHubSecureJsonData)
//   - src/views/ConfigEditor.tsx (which additionally writes
//     jsonData.enableSecureSocksProxy via the @grafana/ui
//     SecureSocksProxySettings component)
//
// https://github.com/grafana/github-datasource
package githubdatasource

import "encoding/json"

// AuthType is the authentication type selected in the configuration editor
// ("Authentication Type"). Stored in jsonData.selectedAuthType.
type AuthType string

const (
	// AuthTypePAT authenticates with a GitHub fine grained personal access token.
	AuthTypePAT AuthType = "personal-access-token"
	// AuthTypeGithubApp authenticates as a GitHub App installation.
	AuthTypeGithubApp AuthType = "github-app"
)

// LicenseType is the GitHub license type selected in the configuration editor
// ("GitHub License Type"). Stored in jsonData.githubPlan.
type LicenseType string

const (
	// LicenseTypeBasic is the "Free, Pro & Team" GitHub plan.
	LicenseTypeBasic LicenseType = "github-basic"
	// LicenseTypeEnterpriseCloud is the "Enterprise Cloud" GitHub plan.
	LicenseTypeEnterpriseCloud LicenseType = "github-enterprise-cloud"
	// LicenseTypeEnterpriseServer is the "Enterprise Server" (on-prem) GitHub plan.
	LicenseTypeEnterpriseServer LicenseType = "github-enterprise-server"
)

// JSONData models the fields stored in the datasource's jsonData.
type JSONData struct {
	// GithubPlan is the GitHub license type. Datasources created before this
	// field was introduced may have GitHubURL set without a plan; consumers
	// should treat a non-empty GitHubURL as github-enterprise-server and fall
	// back to github-basic when both are empty.
	GithubPlan LicenseType `json:"githubPlan,omitempty"`
	// GitHubURL is the GitHub Enterprise Server URL. Only used when GithubPlan
	// is github-enterprise-server; the backend derives "<url>/api/v3" (REST)
	// and "<url>/api/graphql" (GraphQL) from it.
	GitHubURL string `json:"githubUrl,omitempty"`
	// SelectedAuthType selects between personal-access-token and github-app
	// authentication. Datasources created before this field was introduced
	// have an accessToken but no auth type; the backend defaults them to
	// personal-access-token.
	SelectedAuthType AuthType `json:"selectedAuthType,omitempty"`
	// AppId is the GitHub App ID (github-app auth only). The configuration
	// editor stores it as a string, but the backend accepts a JSON string or
	// number, hence json.RawMessage.
	AppId json.RawMessage `json:"appId,omitempty"`
	// InstallationId is the GitHub App installation ID (github-app auth only).
	// Like AppId, it may be stored as a JSON string or number.
	InstallationId json.RawMessage `json:"installationId,omitempty"`
	// CachingEnabled enables the query caching wrapper in the plugin backend.
	// Not exposed in the configuration editor; the backend currently enables
	// caching for every datasource instance.
	CachingEnabled bool `json:"cachingEnabled,omitempty"`
	// EnableSecureSocksProxy connects to the datasource via the Grafana secure
	// socks proxy. Written by the @grafana/ui SecureSocksProxySettings
	// component when the Grafana instance has secureSocksDSProxyEnabled.
	EnableSecureSocksProxy bool `json:"enableSecureSocksProxy,omitempty"`
}

// SecureJSONData models the fields stored in the datasource's secureJsonData.
// Secure fields are write-only: read existing config via Config.SecureJSONFields
// to determine whether a secret is already configured.
type SecureJSONData struct {
	// AccessToken is set if the user is using a Personal Access Token to
	// connect to GitHub.
	AccessToken string `json:"accessToken,omitempty"`
	// PrivateKey is set if the user is using a GitHub App to connect to GitHub.
	PrivateKey string `json:"privateKey,omitempty"`
}

// Config models the root (top-level datasource settings) fields.
//
// The GitHub datasource stores no plugin-specific fields at the root level
// (url, basicAuth, etc. are unused); all configuration lives in JSONData and
// SecureJSONData.
type Config struct {
	JSONData JSONData `json:"jsonData"`
	// SecureJSONData is write-only.
	SecureJSONData *SecureJSONData `json:"secureJsonData,omitempty"`
	// SecureJSONFields indicates which secrets are already configured.
	SecureJSONFields map[string]bool `json:"secureJsonFields,omitempty"`
}
