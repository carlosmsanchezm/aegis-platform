import { screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { TestApiProvider, renderInTestApp } from '@backstage/test-utils';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  FetchApi,
  AlertApi,
} from '@backstage/core-plugin-api';
import { SubmitWorkloadPage } from './SubmitWorkloadPage';

const server = setupServer(
  rest.post(
    'http://example.test/aegis/aegis.v1.AegisPlatform/SubmitWorkload',
    (_req, res, ctx) =>
      res(
        ctx.json({
          id: 'w-ui-123',
          status: 'PLACED',
          projectId: 'p-demo',
        }),
      ),
  ),
  rest.post(
    'http://example.test/aegis/aegis.v1.AegisPlatform/GetWorkload',
    (_req, res, ctx) =>
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
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('SubmitWorkloadPage', () => {
  it('submits a workload and shows success feedback', async () => {
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

    await renderInTestApp(
      <TestApiProvider
        apis={[
          [discoveryApiRef, discoveryApi],
          [fetchApiRef, fetchApi],
          [alertApiRef, alertApi],
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
    await user.type(queueField, 'default');

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

    const submitButton = await screen.findByRole('button', { name: /submit/i });
    await user.click(submitButton);

    await screen.findByText(/Workload Created/i);
    expect(await screen.findByText('w-ui-123')).toBeInTheDocument();

    await waitFor(() => {
      expect(alertSpy).toHaveBeenCalledWith(
        expect.objectContaining({
          message: 'Workload submitted: w-ui-123',
          severity: 'success',
        }),
      );
    });
  });
});
