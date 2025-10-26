import { registerKeycloakProvider } from './auth-keycloak-provider';

describe('registerKeycloakProvider', () => {
  it('registers the Keycloak provider with the auth extension point', () => {
    const registerProvider = jest.fn();

    registerKeycloakProvider({ registerProvider });

    expect(registerProvider).toHaveBeenCalledTimes(1);
    const [{ providerId, factory }] = registerProvider.mock.calls[0];

    expect(providerId).toBe('keycloak');
    expect(typeof factory).toBe('function');
  });
});
