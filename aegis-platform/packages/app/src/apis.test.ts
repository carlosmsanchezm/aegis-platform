import { ConfigReader } from '@backstage/config';
import { OAuthRequestManager } from '@backstage/core-app-api';
import {
  apis,
  buildKeycloakProviderInfo,
  ensureOidcScopes,
  keycloakAuthApiRef,
} from './apis';

describe('Keycloak auth API wiring', () => {
  it('ensures required OIDC scopes are always present', () => {
    const scopes = ensureOidcScopes(['profile', 'custom']);
    expect(scopes).toEqual(
      expect.arrayContaining(['custom', 'openid', 'profile', 'email']),
    );
  });

  it('exposes provider metadata for Keycloak', () => {
    const info = buildKeycloakProviderInfo();
    expect(info.id).toBe('keycloak');
    expect(info.title).toMatch(/Keycloak/i);
    expect(info.icon).toBeDefined();
  });

  it('registers an OAuth2 factory for Keycloak that can be constructed', () => {
    const factory = apis.find(apiFactory => {
      // eslint-disable-next-line @typescript-eslint/no-explicit-any
      return (apiFactory as any).api?.id === keycloakAuthApiRef.id;
    });
    expect(factory).toBeDefined();

    const discoveryApi = { getBaseUrl: jest.fn(async () => 'http://example.com') };
    const oauthRequestApi = new OAuthRequestManager();
    const configApi = new ConfigReader({ auth: { environment: 'development' } });

    const instance = (factory as any).factory({ discoveryApi, oauthRequestApi, configApi });
    expect(instance).toBeDefined();
    expect(typeof instance.signIn).toBe('function');
  });
});
