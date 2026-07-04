// Package githubdatasource contains the configuration models for the
// GitHub datasource plugin (grafana-github-datasource).
//
// Sources of truth (https://github.com/grafana/github-datasource):
//   - pkg/models/settings.go — backend Settings and AuthType
//   - src/types/config.ts — GitHubLicenseType, GitHubAuthType,
//     GitHubDataSourceOptions, GitHubSecureJsonDataKeys
//   - src/views/ConfigEditor.tsx — the configuration editor
package githubdatasource

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

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

// RootConfig models the root (top-level datasource settings) fields.
// The GitHub datasource stores no plugin-specific fields at the root level
// (url, basicAuth, etc. are unused), so this is a blank object rather than null.
type RootConfig struct{}

// JsonDataConfig models the fields stored in the datasource's jsonData.
// It matches the jsonData subset of the plugin's Settings (pkg/models/settings.go)
// plus githubPlan, which only exists in the frontend model (src/types/config.ts).
type JsonDataConfig struct {
	// GithubPlan is frontend-only: written and read by the config editor to
	// drive the "GitHub License Type" radio; never read by the plugin backend,
	// which infers Enterprise Server solely from a non-empty GitHubURL.
	GithubPlan LicenseType `json:"githubPlan,omitempty"`
	// GitHubURL is the GitHub Enterprise Server base URL; the backend derives
	// "<url>/api/v3" (REST) and "<url>/api/graphql" (GraphQL) from it.
	GitHubURL string `json:"githubUrl,omitempty"`
	// SelectedAuthType defaults to personal-access-token in the editor; the
	// backend also defaults to it when only accessToken is set.
	SelectedAuthType AuthType `json:"selectedAuthType,omitempty"`
	// AppId is the GitHub App ID (github-app auth). The editor stores it as a
	// string; the backend accepts a JSON string or number, hence json.RawMessage.
	AppId json.RawMessage `json:"appId,omitempty"`
	// InstallationId is the GitHub App installation ID (github-app auth).
	// Like AppId, it may be stored as a JSON string or number.
	InstallationId json.RawMessage `json:"installationId,omitempty"`
	// CachingEnabled is backend-only: not exposed in the config editor; the
	// backend currently forces caching on for every instance.
	CachingEnabled bool `json:"cachingEnabled,omitempty"`
}

// SecureJsonDataConfig lists the secret key names stored in secureJsonData
// (write-only; read existing config via secureJsonFields).
type SecureJsonDataConfig []string

// SecureJsonDataKeys are the secret keys used by the plugin:
// accessToken is set if the user is using a Personal Access Token to connect
// to GitHub; privateKey is set if the user is using a GitHub App.
var SecureJsonDataKeys = SecureJsonDataConfig{"accessToken", "privateKey"}

// Config is the fully loaded configuration of a GitHub datasource instance,
// combining the root fields, the parsed jsonData, and the decrypted secrets.
type Config struct {
	// Root holds the root-level fields (none for this plugin).
	Root RootConfig
	// JSONData holds the parsed jsonData fields.
	JSONData JsonDataConfig
	// Secrets holds the decrypted secure values by key (accessToken, privateKey).
	Secrets map[string]string
	// ConfiguredSecureKeys lists which of SecureJsonDataKeys are present in
	// the instance settings, mirroring what secureJsonFields reports.
	ConfiguredSecureKeys SecureJsonDataConfig
}

// LoadConfig parses a datasource instance's settings into a Config. It mirrors
// the plugin's LoadSettings (pkg/models/settings.go), including the legacy
// fallback: instances created before selectedAuthType was introduced store
// only an accessToken, so the auth type defaults to personal-access-token
// when an accessToken is present and no auth type is set.
func LoadConfig(settings backend.DataSourceInstanceSettings) (Config, error) {
	cfg := Config{Secrets: map[string]string{}}

	if len(settings.JSONData) > 0 {
		if err := json.Unmarshal(settings.JSONData, &cfg.JSONData); err != nil {
			return cfg, fmt.Errorf("parse jsonData: %w", err)
		}
	}

	for _, key := range SecureJsonDataKeys {
		if val, ok := settings.DecryptedSecureJSONData[key]; ok {
			cfg.Secrets[key] = val
			cfg.ConfiguredSecureKeys = append(cfg.ConfiguredSecureKeys, key)
		}
	}

	if cfg.Secrets["accessToken"] != "" && cfg.JSONData.SelectedAuthType == "" {
		cfg.JSONData.SelectedAuthType = AuthTypePAT
	}

	return cfg, nil
}

// AppIdInt64 parses AppId to an int64, accepting a JSON string or number,
// using the same lenient parsing as the plugin's rawMessageToInt64.
func (j JsonDataConfig) AppIdInt64() (int64, error) {
	return rawMessageToInt64(j.AppId, "app id")
}

// InstallationIdInt64 parses InstallationId to an int64, accepting a JSON
// string or number, using the same lenient parsing as the plugin's
// rawMessageToInt64.
func (j JsonDataConfig) InstallationIdInt64() (int64, error) {
	return rawMessageToInt64(j.InstallationId, "installation id")
}

func rawMessageToInt64(r json.RawMessage, m string) (int64, error) {
	out, err := strconv.ParseInt(strings.ReplaceAll(string(r), `"`, ""), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing %s", m)
	}
	return out, nil
}
