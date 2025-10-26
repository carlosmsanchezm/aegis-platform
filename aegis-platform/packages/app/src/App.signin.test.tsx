import { keycloakSignInProviders } from './App';
import { keycloakAuthApiRef } from './apis';

describe('Keycloak sign-in configuration', () => {
  it('exposes a single Keycloak provider with the expected API ref', () => {
    expect(keycloakSignInProviders).toHaveLength(1);

    const provider = keycloakSignInProviders[0];
    expect(provider.id).toBe('keycloak');
    expect(provider.title).toContain('Keycloak');
    expect(provider.message).toMatch(/Keycloak SSO/i);
    expect(provider.apiRef).toBe(keycloakAuthApiRef);
  });
});
