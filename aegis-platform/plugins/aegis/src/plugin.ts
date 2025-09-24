import {
  createPlugin,
  createRoutableExtension,
} from '@backstage/core-plugin-api';

import { rootRouteRef } from './routes';

export const aegisPlugin = createPlugin({
  id: 'aegis',
  routes: {
    root: rootRouteRef,
  },
});

export const AegisPage = aegisPlugin.provide(
  createRoutableExtension({
    name: 'AegisPage',
    component: () =>
      import('./components/SubmitWorkloadPage').then(
        m => m.SubmitWorkloadPage,
      ),
    mountPoint: rootRouteRef,
  }),
);
