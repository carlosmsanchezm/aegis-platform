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
      init({ providers }) {
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
      },
    });
  },
});
