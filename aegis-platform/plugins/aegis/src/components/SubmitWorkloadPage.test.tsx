import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { TestApiProvider, renderInTestApp } from '@backstage/test-utils';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
  FetchApi,
  AlertApi,
  IdentityApi,
} from '@backstage/core-plugin-api';
import { SubmitWorkloadPage } from './SubmitWorkloadPage';

const submittedBodies: any[] = [];

const server = setupServer(
  rest.post(
    'http://example.test/aegis/aegis.v1.AegisPlatform/SubmitWorkload',
    async (req, res, ctx) => {
      submittedBodies.push(await req.json());
      return res(
        ctx.json({
          id: 'w-ui-123',
          status: 'PLACED',
          projectId: 'p-demo',
        }),
      );
    },
  ),
  rest.post(
    'http://example.test/aegis/aegis.v1.AegisPlatform/GetWorkload',
    async (_req, res, ctx) =>
      res(
        ctx.json({
          id: 'w-ui-123',
          status: 'SUCCEEDED',
          projectId: 'p-demo',
        }),
      ),
  ),
);

jest.setTimeout(15000);

beforeAll(() => server.listen());
afterEach(() => {
  server.resetHandlers();
  submittedBodies.length = 0;
});
afterAll(() => server.close());

describe('SubmitWorkloadPage', () => {
  it('submits an interactive workspace and surfaces feedback', async () => {
    const user = userEvent.setup();

    const discoveryApi = {
      async getBaseUrl(pluginId: string) {
        if (pluginId !== 'proxy') {
          throw new Error(`unexpected pluginId ${pluginId}`);
        }
        return 'http://example.test';
      },
    };

    const alertSpy = jest.fn();
    const alertApi: AlertApi = { post: alertSpy };
    const fetchApi: FetchApi = {
      fetch: (...args) => fetch(...args),
    };

    const identityApi: IdentityApi = {
      async getBackstageIdentity() {
        return { token: 'test-token', userEntityRef: 'user:default/tester' };
      },
      async getCredentials() {
        return { token: 'test-token', userEntityRef: 'user:default/tester' };
      },
      async signOut() {
        /* not needed */
      },
    } as IdentityApi;

    await renderInTestApp(
      <TestApiProvider
        apis={[
          [discoveryApiRef, discoveryApi],
          [fetchApiRef, fetchApi],
          [alertApiRef, alertApi],
          [identityApiRef, identityApi],
        ]}
      >
        <SubmitWorkloadPage />
      </TestApiProvider>,
    );

    const projectField = await screen.findByDisplayValue('p-demo');
    await user.clear(projectField);
    await user.type(projectField, 'p-ui-e2e');

    const queueField = await screen.findByDisplayValue('default');
    await user.clear(queueField);
    await user.type(queueField, 'interactive');

    const flavorField = await screen.findByDisplayValue('a10-mig-1g');
    await user.clear(flavorField);
    await user.type(flavorField, 'a10-mig-1g');

    const imageField = await screen.findByDisplayValue('alpine:3.19');
    await user.clear(imageField);
    await user.type(imageField, 'alpine:3.19');

    const commandField = await screen.findByDisplayValue(
      'echo Hello from Aegis; sleep 2',
    );
    await user.clear(commandField);
    await user.type(commandField, 'echo from test');

    const durationField = await screen.findByDisplayValue('600');
    await user.clear(durationField);
    await user.type(durationField, '120');

    const advancedButton = await screen.findByRole('button', {
      name: /show advanced options/i,
    });
    await user.click(advancedButton);

    const portsField = await screen.findByDisplayValue('22, 11111');
    await user.clear(portsField);
    await user.type(portsField, '22, 11111, 10022');

    const envField = await screen.findByDisplayValue(/USER_NAME=aegis/);
    await user.type(envField, '\nAEGIS_SSH_USER=dev');

    const submitButton = await screen.findByRole('button', {
      name: /launch workspace/i,
    });
    await user.click(submitButton);

    await waitFor(() => expect(submittedBodies).toHaveLength(1));
    const body = submittedBodies[0];

    expect(body.workload.projectId).toBe('p-ui-e2e');
    expect(body.workload.queue).toBe('interactive');
    expect(body.workload.kind.workspace.interactive).toBe(true);
    expect(body.workload.kind.workspace.command).toEqual([
      'sh',
      '-c',
      'echo from test',
    ]);
    expect(body.workload.kind.workspace.ports).toEqual(
      expect.arrayContaining([22, 11111, 10022]),
    );
    expect(body.workload.kind.workspace.env.AEGIS_SSH_USER).toBe('dev');
    expect(body.workload.kind.workspace.env.USER_NAME).toBe('aegis');
    expect(body.workload.kind.workspace.maxDurationSeconds).toBe(120);

    await screen.findByText(/Workspace created/i);
    await screen.findByRole('button', { name: /view details/i });

    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          message: 'Interactive workspace submitted: w-ui-123',
          severity: 'success',
        }),
      );
    });
  });
});
