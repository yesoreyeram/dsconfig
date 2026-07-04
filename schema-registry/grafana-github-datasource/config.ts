/**
 * Configuration models for the GitHub datasource plugin (`grafana-github-datasource`).
 *
 * Mirrors the plugin source of truth:
 * - `src/types/config.ts` (`GitHubLicenseType`, `GitHubAuthType`, `GitHubDataSourceOptions`,
 *   `GitHubSecureJsonDataKeys`, `GitHubSecureJsonData`)
 * - `src/views/ConfigEditor.tsx` (which additionally writes
 *   `jsonData.enableSecureSocksProxy` via the `@grafana/ui` `SecureSocksProxySettings` component)
 * - `pkg/models/settings.go` (which additionally reads `jsonData.cachingEnabled`)
 *
 * https://github.com/grafana/github-datasource
 */

export type GitHubLicenseType = 'github-basic' | 'github-enterprise-cloud' | 'github-enterprise-server';

export type GitHubAuthType = 'personal-access-token' | 'github-app';

/**
 * Fields stored in `jsonData`.
 *
 * Matches the plugin's `GitHubDataSourceOptions` (which extends `DataSourceJsonData`
 * from `@grafana/data`), plus the two `jsonData` keys written/read outside
 * `src/types/config.ts`: `enableSecureSocksProxy` (written by the `@grafana/ui`
 * `SecureSocksProxySettings` component) and `cachingEnabled` (read by
 * `pkg/models/settings.go`).
 */
export interface GitHubJsonData {
  githubPlan?: GitHubLicenseType;
  githubUrl?: string;
  selectedAuthType?: GitHubAuthType;
  appId?: string;
  installationId?: string;
  // enableSecureSocksProxy is set by the @grafana/ui SecureSocksProxySettings component
  // when the Grafana instance has secureSocksDSProxyEnabled
  enableSecureSocksProxy?: boolean;
  // cachingEnabled is not exposed in the configuration editor; the plugin backend
  // currently enables the caching wrapper for every datasource instance
  cachingEnabled?: boolean;
}

export type GitHubSecureJsonDataKeys =
  | 'accessToken' // accessToken is set if the user is using a Personal Access Token to connect to GitHub
  | 'privateKey'; // privateKey is set if the user is using a GitHub App to connect to GitHub

/**
 * Fields stored in `secureJsonData`.
 *
 * Secure fields are write-only: read existing config via `secureJsonFields`
 * to determine whether a secret is already configured.
 */
export type GitHubSecureJsonData = Partial<Record<GitHubSecureJsonDataKeys, string>>;

/**
 * Root (top-level datasource settings) fields.
 *
 * The GitHub datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused); all configuration lives in
 * `jsonData` and `secureJsonData`.
 */
export interface GitHubConfig {
  jsonData: GitHubJsonData;
  /** Write-only secrets. */
  secureJsonData?: GitHubSecureJsonData;
  /** Read-side indicator of which secrets are already configured. */
  secureJsonFields?: Partial<Record<GitHubSecureJsonDataKeys, boolean>>;
}
