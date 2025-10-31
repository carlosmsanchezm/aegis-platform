import { renderInTestApp, TestApiProvider } from '@backstage/test-utils';
import { screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
} from '@backstage/core-plugin-api';
import { WorkspaceWizard } from './WorkspaceWizard';
import {
  rootRouteRef,
  workloadsRouteRef,
  workloadDetailsRouteRef,
} from '../routes';

const createApis = (): Array<[any, unknown]> => {
  const alertApi = { post: jest.fn() };
  const fetchApi = { fetch: jest.fn() };
  const discoveryApi = {
    getBaseUrl: jest.fn().mockResolvedValue('http://example.test'),
  };
  const identityApi = {
    getBackstageIdentity: jest.fn().mockResolvedValue({ token: 'test' }),
    getCredentials: jest.fn().mockResolvedValue({ token: 'test' }),
  };

  return [
    [alertApiRef, alertApi],
    [fetchApiRef, fetchApi],
    [discoveryApiRef, discoveryApi],
    [identityApiRef, identityApi],
  ];
};

const mountedRoutes = {
  '/aegis': rootRouteRef,
  '/aegis/workloads': workloadsRouteRef,
  '/aegis/workloads/:id': workloadDetailsRouteRef,
};

describe('WorkspaceWizard', () => {
  const originalCrypto = global.crypto;

  beforeAll(() => {
    (global as any).crypto = {
      randomUUID: jest.fn().mockReturnValue('1234567890abcdef'),
    };
  });

  afterAll(() => {
    (global as any).crypto = originalCrypto;
  });

  it('requires selecting a template before continuing', async () => {
    await renderInTestApp(
      <TestApiProvider apis={createApis()}>
        <WorkspaceWizard />
      </TestApiProvider>,
      { mountedRoutes },
    );

    const nextButton = await screen.findByRole('button', { name: /next/i });
    expect(nextButton).toBeDisabled();

    const templateButton = await screen.findByRole('button', {
      name: /python starter/i,
    });
    await userEvent.click(templateButton);

    expect(nextButton).toBeEnabled();
  });

  it('performs validation on workspace id in configure step', async () => {
    await renderInTestApp(
      <TestApiProvider apis={createApis()}>
        <WorkspaceWizard />
      </TestApiProvider>,
      { mountedRoutes },
    );

    const templateButton = await screen.findByRole('button', {
      name: /python starter/i,
    });
    await userEvent.click(templateButton);

    const nextButton = await screen.findByRole('button', { name: /next/i });
    await userEvent.click(nextButton);

    const workspaceIdField = await screen.findByDisplayValue(/ws-/i);
    await userEvent.clear(workspaceIdField);
    await userEvent.type(workspaceIdField, 'INVALID ID!');

    expect(nextButton).toBeDisabled();
  });
});
