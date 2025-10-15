import { ChangeEvent, FC, FormEvent, useMemo, useState } from 'react';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
  InfoCard,
} from '@backstage/core-components';
import { Box, Button, Grid, TextField, Typography } from '@material-ui/core';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
  useApi,
} from '@backstage/core-plugin-api';
import {
  SubmitWorkspaceRequest,
  submitWorkspace,
  WorkloadDTO,
} from '../api/aegisClient';
import { parseEnvInput, parsePortsInput } from './workspaceFormUtils';

const randomId = () => {
  if (typeof crypto?.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `workspace-${Math.random().toString(16).slice(2, 10)}`;
};

export const LaunchWorkspacePage: FC = () => {
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);

  const [form, setForm] = useState({
    workloadId: randomId(),
    projectId: '',
    queue: '',
    flavor: '',
    image: '',
    ports: '22',
    env: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [submitted, setSubmitted] = useState<WorkloadDTO | null>(null);

  const isSubmitDisabled = useMemo(
    () =>
      !form.projectId.trim() ||
      !form.flavor.trim() ||
      !form.image.trim() ||
      !form.workloadId.trim(),
    [form],
  );

  const handleChange =
    (field: keyof typeof form) => (event: ChangeEvent<HTMLInputElement>) => {
      setForm(prev => ({ ...prev, [field]: event.target.value }));
    };

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (isSubmitDisabled) {
      return;
    }

    const ports = parsePortsInput(form.ports);
    const env = parseEnvInput(form.env);

    const payload: SubmitWorkspaceRequest = {
      id: form.workloadId.trim(),
      projectId: form.projectId.trim(),
      queue: form.queue.trim() || undefined,
      workspace: {
        flavor: form.flavor.trim(),
        image: form.image.trim(),
        interactive: true,
        ports: ports.length > 0 ? ports : undefined,
        env: Object.keys(env).length > 0 ? env : undefined,
      },
    };

    try {
      setSubmitting(true);
      setError(null);
      const created = await submitWorkspace(
        fetchApi,
        discoveryApi,
        identityApi,
        payload,
      );
      setSubmitted(created);
      alertApi.post({
        message: `Submitted interactive workspace ${created.id ?? payload.id}`,
        severity: 'success',
      });
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setError(msg);
      alertApi.post({
        message: `Failed to submit workspace: ${msg}`,
        severity: 'error',
      });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Page themeId="tool">
      <Header title="Launch Interactive Workspace" />
      <Content>
        <ContentHeader title="Workspace Parameters" />
        <form onSubmit={handleSubmit}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <TextField
                label="Workload ID"
                value={form.workloadId}
                onChange={handleChange('workloadId')}
                fullWidth
                required
                helperText="Unique identifier for this workspace request"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                label="Project ID"
                value={form.projectId}
                onChange={handleChange('projectId')}
                fullWidth
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                label="Queue"
                value={form.queue}
                onChange={handleChange('queue')}
                fullWidth
                helperText="Optional scheduling queue"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                label="Flavor"
                value={form.flavor}
                onChange={handleChange('flavor')}
                fullWidth
                required
                helperText="GPU flavor or node profile"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                label="Container image"
                value={form.image}
                onChange={handleChange('image')}
                fullWidth
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                label="Exposed ports"
                value={form.ports}
                onChange={handleChange('ports')}
                fullWidth
                helperText="Comma or space separated list. Defaults to 22"
              />
            </Grid>
            <Grid item xs={12}>
              <TextField
                label="Environment variables"
                value={form.env}
                onChange={handleChange('env')}
                fullWidth
                multiline
                minRows={3}
                helperText="Optional KEY=VALUE pairs, one per line"
              />
            </Grid>
          </Grid>

          <Box marginTop={3} display="flex" alignItems="center" gridGap={16}>
            <Button
              type="submit"
              color="primary"
              variant="contained"
              disabled={isSubmitDisabled || submitting}
            >
              Launch Workspace
            </Button>
            {submitting && <Progress />}
          </Box>
        </form>

        {error && (
          <Box marginTop={3}>
            <WarningPanel severity="error" title="Workspace submission failed">
              {error}
            </WarningPanel>
          </Box>
        )}

        {submitted && (
          <Box marginTop={3}>
            <InfoCard title="Workspace submitted">
              <Typography variant="body2">
                Workload <strong>{submitted.id ?? form.workloadId}</strong> is
                queued. Track status from the workloads list or open the details
                page once it appears.
              </Typography>
            </InfoCard>
          </Box>
        )}
      </Content>
    </Page>
  );
};
