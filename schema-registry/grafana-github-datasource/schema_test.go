package githubdatasource

import (
	"testing"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

func TestNewSchema(t *testing.T) {
	schema, err := NewSchema()
	if err != nil {
		t.Fatalf("NewSchema: %v", err)
	}
	if schema.TargetAPIVersion != TargetAPIVersion {
		t.Fatalf("TargetAPIVersion = %q, want %q", schema.TargetAPIVersion, TargetAPIVersion)
	}
	if schema.SettingsSchema == nil || schema.SettingsSchema.Spec == nil {
		t.Fatal("SettingsSchema.Spec is nil")
	}

	if _, hasSecure := schema.SettingsSchema.Spec.Properties["secureJsonData"]; hasSecure {
		t.Fatal("secureJsonData must not be defined on the spec; use SecureValues")
	}

	secureKeys := make([]string, 0, len(schema.SettingsSchema.SecureValues))
	for _, sv := range schema.SettingsSchema.SecureValues {
		secureKeys = append(secureKeys, sv.Key)
	}
	if len(secureKeys) != len(SecureJsonDataKeys) {
		t.Fatalf("SecureValues = %v, want %v", secureKeys, SecureJsonDataKeys)
	}
	for i, key := range SecureJsonDataKeys {
		if secureKeys[i] != key {
			t.Fatalf("SecureValues = %v, want %v", secureKeys, SecureJsonDataKeys)
		}
	}

	jsonData, ok := schema.SettingsSchema.Spec.Properties["jsonData"]
	if !ok {
		t.Fatal("spec has no jsonData property")
	}
	for _, key := range []string{"githubPlan", "githubUrl", "selectedAuthType", "appId", "installationId", "cachingEnabled"} {
		if _, ok := jsonData.Properties[key]; !ok {
			t.Errorf("jsonData property %q missing from spec", key)
		}
	}

	if schema.SettingsExamples == nil || len(schema.SettingsExamples.Examples) == 0 {
		t.Fatal("SettingsExamples is empty")
	}
	if _, ok := schema.SettingsExamples.Examples["default"]; !ok {
		t.Error("SettingsExamples has no \"default\" example")
	}
	validSecureKeys := map[string]bool{}
	for _, key := range SecureJsonDataKeys {
		validSecureKeys[key] = true
	}
	for name, ex := range schema.SettingsExamples.Examples {
		value, ok := ex.Value.(map[string]any)
		if !ok {
			t.Errorf("example %q value is not an object", name)
			continue
		}
		if _, ok := value["jsonData"].(map[string]any); !ok {
			t.Errorf("example %q value has no jsonData object", name)
		}
		secure, ok := value["secureJsonData"].(map[string]any)
		if !ok || len(secure) == 0 {
			t.Errorf("example %q value has no secureJsonData object", name)
			continue
		}
		for key := range secure {
			if !validSecureKeys[key] {
				t.Errorf("example %q uses unknown secureJsonData key %q (valid: %v)", name, key, SecureJsonDataKeys)
			}
		}
	}
}

func TestLoadConfig(t *testing.T) {
	t.Run("personal access token", func(t *testing.T) {
		cfg, err := LoadConfig(backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"selectedAuthType":"personal-access-token","githubPlan":"github-basic"}`),
			DecryptedSecureJSONData: map[string]string{"accessToken": "token"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.JSONData.SelectedAuthType != AuthTypePAT {
			t.Errorf("SelectedAuthType = %q, want %q", cfg.JSONData.SelectedAuthType, AuthTypePAT)
		}
		if cfg.JSONData.GithubPlan != LicenseTypeBasic {
			t.Errorf("GithubPlan = %q, want %q", cfg.JSONData.GithubPlan, LicenseTypeBasic)
		}
		if cfg.Secrets["accessToken"] != "token" {
			t.Errorf("Secrets[accessToken] = %q, want %q", cfg.Secrets["accessToken"], "token")
		}
		if len(cfg.ConfiguredSecureKeys) != 1 || cfg.ConfiguredSecureKeys[0] != "accessToken" {
			t.Errorf("ConfiguredSecureKeys = %v, want [accessToken]", cfg.ConfiguredSecureKeys)
		}
	})

	t.Run("legacy access token without auth type defaults to PAT", func(t *testing.T) {
		cfg, err := LoadConfig(backend.DataSourceInstanceSettings{
			DecryptedSecureJSONData: map[string]string{"accessToken": "token"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if cfg.JSONData.SelectedAuthType != AuthTypePAT {
			t.Errorf("SelectedAuthType = %q, want %q", cfg.JSONData.SelectedAuthType, AuthTypePAT)
		}
	})

	t.Run("github app with string and numeric ids", func(t *testing.T) {
		cfg, err := LoadConfig(backend.DataSourceInstanceSettings{
			JSONData:                []byte(`{"selectedAuthType":"github-app","appId":"123456","installationId":12345678}`),
			DecryptedSecureJSONData: map[string]string{"privateKey": "pem"},
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		appID, err := cfg.JSONData.AppIdInt64()
		if err != nil || appID != 123456 {
			t.Errorf("AppIdInt64 = %d, %v; want 123456, nil", appID, err)
		}
		installationID, err := cfg.JSONData.InstallationIdInt64()
		if err != nil || installationID != 12345678 {
			t.Errorf("InstallationIdInt64 = %d, %v; want 12345678, nil", installationID, err)
		}
		if cfg.Secrets["privateKey"] != "pem" {
			t.Errorf("Secrets[privateKey] = %q, want %q", cfg.Secrets["privateKey"], "pem")
		}
	})

	t.Run("invalid app id errors", func(t *testing.T) {
		cfg, err := LoadConfig(backend.DataSourceInstanceSettings{
			JSONData: []byte(`{"selectedAuthType":"github-app","appId":"not-a-number"}`),
		})
		if err != nil {
			t.Fatalf("LoadConfig: %v", err)
		}
		if _, err := cfg.JSONData.AppIdInt64(); err == nil {
			t.Error("AppIdInt64 should fail for a non-numeric value")
		}
	})

	t.Run("invalid jsonData errors", func(t *testing.T) {
		if _, err := LoadConfig(backend.DataSourceInstanceSettings{JSONData: []byte(`{`)}); err == nil {
			t.Error("LoadConfig should fail for malformed jsonData")
		}
	})
}
