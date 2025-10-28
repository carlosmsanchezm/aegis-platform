import { keycloakSignInProvider } from './App';
import { keycloakAuthApiRef } from './apis';

describe('Keycloak sign-in configuration', () => {
  it('exposes a single Keycloak provider with the expected API ref', () => {
    expect(keycloakSignInProvider.id).toBe('keycloak');
    expect(keycloakSignInProvider.title).toContain('Keycloak');
    expect(keycloakSignInProvider.message).toMatch(/Keycloak/i);
    expect(keycloakSignInProvider.apiRef).toBe(keycloakAuthApiRef);
  });
});
