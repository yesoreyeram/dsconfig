package githubdatasource

import "encoding/json"

type AuthType string

const (
	AuthTypePAT       AuthType = "personal-access-token"
	AuthTypeGithubApp AuthType = "github-app"
)

type LicenseType string

const (
	LicenseTypeBasic            LicenseType = "github-basic"
	LicenseTypeEnterpriseCloud  LicenseType = "github-enterprise-cloud"
	LicenseTypeEnterpriseServer LicenseType = "github-enterprise-server"
)

type RootConfig struct{}

type JsonDataConfig struct {
	GithubPlan       LicenseType     `json:"githubPlan,omitempty"`
	GitHubURL        string          `json:"githubUrl,omitempty"`
	SelectedAuthType AuthType        `json:"selectedAuthType,omitempty"`
	AppId            json.RawMessage `json:"appId,omitempty"`
	InstallationId   json.RawMessage `json:"installationId,omitempty"`
	CachingEnabled   bool            `json:"cachingEnabled,omitempty"`
}

type SecureJsonDataConfig []string

var SecureJsonDataKeys = SecureJsonDataConfig{"accessToken", "privateKey"}
