import * as PluginExports from './index';

describe('aegis', () => {
  it('should export plugin', () => {
    expect(PluginExports.aegisPlugin).toBeDefined();
  });

  it('should expose all routable pages and routeRefs', () => {
    const expectedPages = [
      'AegisPage',
      'AegisWorkloadListPage',
      'AegisWorkloadDetailsPage',
      'AegisLaunchWorkspacePage',
      'AegisCreateWorkspacePage',
    ] as const;

    const expectedRouteRefs = [
      'rootRouteRef',
      'workloadsRouteRef',
      'workloadDetailsRouteRef',
      'launchWorkspaceRouteRef',
      'createWorkspaceRouteRef',
    ] as const;

    expectedPages.forEach(name => {
      expect(PluginExports[name]).toBeDefined();
    });
    expectedRouteRefs.forEach(name => {
      expect(PluginExports[name]).toBeDefined();
    });
  });
});
