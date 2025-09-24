import { createRouteRef, createSubRouteRef } from '@backstage/core-plugin-api';

export const rootRouteRef = createRouteRef({
	id: 'aegis-root',
});

export const workloadsRouteRef = createSubRouteRef({
	id: 'aegis-workloads',
	parent: rootRouteRef,
	path: '/workloads',
});

export const workloadDetailsRouteRef = createSubRouteRef({
	id: 'aegis-workload-details',
	parent: rootRouteRef,
	path: '/workloads/:id',
});
