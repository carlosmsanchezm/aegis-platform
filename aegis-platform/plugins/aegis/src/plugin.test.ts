import {
  AegisCreateWorkspacePage,
  aegisPlugin,
  createWorkspaceRouteRef,
} from './index';

describe('aegis', () => {
  it('should export plugin', () => {
    expect(aegisPlugin).toBeDefined();
  });

  it('should expose create workspace route and page', () => {
    expect(createWorkspaceRouteRef).toBeDefined();
    expect(AegisCreateWorkspacePage).toBeDefined();
  });
});
