import { createBackendModule } from '@backstage/backend-plugin-api';
import {
  authProvidersExtensionPoint,
  commonSignInResolvers,
  createOAuthProviderFactory,
} from '@backstage/plugin-auth-node';
import {
  oidcAuthenticator,
  oidcSignInResolvers,
} from '@backstage/plugin-auth-backend-module-oidc-provider';

type AuthProviderRegistrar = {
  registerProvider: (options: {
    providerId: string;
    factory: ReturnType<typeof createOAuthProviderFactory>;
  }) => void;
};

export const registerKeycloakProvider = (providers: AuthProviderRegistrar) => {
  providers.registerProvider({
    providerId: 'keycloak',
    factory: createOAuthProviderFactory({
      authenticator: oidcAuthenticator,
      signInResolverFactories: {
        ...oidcSignInResolvers,
        ...commonSignInResolvers,
      },
    }),
  });
};

/**
 * Registers a Keycloak provider that reuses the hardened OIDC authenticator.
 */
export default createBackendModule({
  pluginId: 'auth',
  moduleId: 'keycloak-provider',
  register(env) {
    env.registerInit({
      deps: {
        providers: authProvidersExtensionPoint,
      },
      async init({ providers }) {
        registerKeycloakProvider(providers);
      },
    });
  },
});
