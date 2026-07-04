export type GitHubLicenseType = 'github-basic' | 'github-enterprise-cloud' | 'github-enterprise-server';

export type GitHubAuthType = 'personal-access-token' | 'github-app';

export type RootConfig = Record<string, never>;

export type JsonDataConfig = {
  githubPlan?: GitHubLicenseType;
  githubUrl?: string;
  selectedAuthType?: GitHubAuthType;
  appId?: string;
  installationId?: string;
  cachingEnabled?: boolean;
};

export type SecureJsonDataConfig = Array<'accessToken' | 'privateKey'>;
