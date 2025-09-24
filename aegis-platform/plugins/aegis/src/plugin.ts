import {
	createPlugin,
	createRoutableExtension,
} from '@backstage/core-plugin-api';

import { rootRouteRef, workloadsRouteRef, workloadDetailsRouteRef } from './routes';

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

export const AegisWorkloadListPage = aegisPlugin.provide(
  createRoutableExtension({
    name: 'AegisWorkloadListPage',
    component: () =>
      import('./components/WorkloadListPage').then(
        m => m.WorkloadListPage,
      ),
    mountPoint: workloadsRouteRef,
  }),
);

export const AegisWorkloadDetailsPage = aegisPlugin.provide(
  createRoutableExtension({
    name: 'AegisWorkloadDetailsPage',
    component: () =>
      import('./components/WorkloadDetailsPage').then(
        m => m.WorkloadDetailsPage,
      ),
    mountPoint: workloadDetailsRouteRef,
  }),
);
