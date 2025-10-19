import {
  ChangeEvent,
  FC,
  FormEvent,
  useEffect,
  useMemo,
  useState,
} from 'react';
import {
  Page,
  Content,
  Header,
  ContentHeader,
  Progress,
  WarningPanel,
  InfoCard,
} from '@backstage/core-components';
import {
  useApi,
  fetchApiRef,
  alertApiRef,
  discoveryApiRef,
  identityApiRef,
} from '@backstage/core-plugin-api';
import {
  Box,
  Button,
  Collapse,
  Grid,
  TextField,
  Typography,
} from '@material-ui/core';
import { Link as RouterLink } from 'react-router-dom';
import {
  DEFAULT_SSH_PORT,
  DEFAULT_VSCODE_PORT,
  SubmitWorkspaceRequest,
  WorkloadDTO,
  getWorkload,
  submitWorkspace,
} from '../api/aegisClient';
import {
  formatDefaultEnv,
  parseEnvInput,
  parsePortsInput,
  validateEnvInput,
  validatePortsInput,
} from './workspaceFormUtils';

const DEFAULT_ENV_TEXT = formatDefaultEnv();
const DEFAULT_PORTS_TEXT = `${DEFAULT_SSH_PORT}, ${DEFAULT_VSCODE_PORT}`;

type FormState = {
  projectId: string;
  queue: string;
  flavor: string;
  image: string;
  command: string;
  maxDurationSeconds: string;
  ports: string;
  env: string;
};

export const SubmitWorkloadPage: FC = () => {
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);

  const [form, setForm] = useState<FormState>({
    projectId: 'p-demo',
    queue: 'default',
    flavor: 'a10-mig-1g',
    image: 'alpine:3.19',
    command: 'echo Hello from Aegis; sleep 2',
    maxDurationSeconds: '600',
    ports: DEFAULT_PORTS_TEXT,
    env: DEFAULT_ENV_TEXT,
  });
  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<WorkloadDTO | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [portsError, setPortsError] = useState<string | null>(null);
  const [envError, setEnvError] = useState<string | null>(null);
  const [durationError, setDurationError] = useState<string | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState(false);

  const isSubmitDisabled = useMemo(
    () =>
      submitting ||
      !form.projectId.trim() ||
      !form.flavor.trim() ||
      !form.image.trim() ||
      Boolean(portsError) ||
      Boolean(envError) ||
      Boolean(durationError),
    [
      submitting,
      form.projectId,
      form.flavor,
      form.image,
      portsError,
      envError,
      durationError,
    ],
  );

  const handleFieldChange =
    (field: keyof FormState) =>
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = event.target.value;
      setForm(prev => ({ ...prev, [field]: value }));
    };

  const handlePortsChange = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const value = event.target.value;
    setForm(prev => ({ ...prev, ports: value }));
    setPortsError(validatePortsInput(value));
  };

  const handleEnvChange = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const value = event.target.value;
    setForm(prev => ({ ...prev, env: value }));
    setEnvError(validateEnvInput(value));
  };

  const handleDurationChange = (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => {
    const value = event.target.value;
    setForm(prev => ({ ...prev, maxDurationSeconds: value }));
    if (!value.trim()) {
      setDurationError(null);
      return;
    }
    const parsed = Number.parseInt(value, 10);
    if (!Number.isInteger(parsed) || parsed <= 0) {
      setDurationError('Duration must be a positive number');
    } else {
      setDurationError(null);
    }
  };

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();

    const currentPortsError = validatePortsInput(form.ports);
    const currentEnvError = validateEnvInput(form.env);

    const trimmedDuration = form.maxDurationSeconds.trim();
    let parsedDuration: number | undefined;
    let currentDurationError: string | null = null;
    if (trimmedDuration) {
      const parsed = Number.parseInt(trimmedDuration, 10);
      if (!Number.isInteger(parsed) || parsed <= 0) {
        currentDurationError = 'Duration must be a positive number';
      } else {
        parsedDuration = parsed;
      }
    }

    setPortsError(currentPortsError);
    setEnvError(currentEnvError);
    setDurationError(currentDurationError);

    if (currentPortsError || currentEnvError || currentDurationError) {
      return;
    }

    const envOverrides = parseEnvInput(form.env);
    const ports = parsePortsInput(form.ports);
    const commandText = form.command.trim();

    const payload: SubmitWorkspaceRequest = {
      projectId: form.projectId.trim(),
      queue: form.queue.trim() || undefined,
      workspace: {
        flavor: form.flavor.trim(),
        image: form.image.trim(),
        command: commandText,
        ports,
        env: Object.keys(envOverrides).length > 0 ? envOverrides : undefined,
        maxDurationSeconds: parsedDuration,
      },
    };

    try {
      setSubmitting(true);
      setError(null);
      setResult(null);
      const created = await submitWorkspace(
        fetchApi,
        discoveryApi,
        identityApi,
        payload,
      );
      setResult(created);
      alertApi.post({
        message: `Interactive workspace submitted: ${created.id ?? '(no id)'}`,
        severity: 'success',
      });
    } catch (err: any) {
      const msg = err?.message ?? String(err);
      setError(msg);
      alertApi.post({ message: `Submit failed: ${msg}`, severity: 'error' });
    } finally {
      setSubmitting(false);
    }
  };

  useEffect(() => {
    const terminalStatuses = new Set(['SUCCEEDED', 'FAILED']);
    if (!result?.id || terminalStatuses.has(result.status ?? '')) {
      return undefined;
    }

    let cancelled = false;
    const timer = setInterval(async () => {
      try {
        const fresh = await getWorkload(
          fetchApi,
          discoveryApi,
          identityApi,
          result.id!,
        );
        if (cancelled) {
          return;
        }
        setResult(prev => {
          if (!prev) {
            return fresh;
          }
          if (
            prev.status === fresh.status &&
            prev.clusterId === fresh.clusterId &&
            prev.url === fresh.url &&
            prev.message === fresh.message
          ) {
            return prev;
          }
          return fresh;
        });
      } catch {
        /* ignore polling errors */
      }
    }, 2000);

    return () => {
      cancelled = true;
      clearInterval(timer);
    };
  }, [result?.id, result?.status, fetchApi, discoveryApi, identityApi]);

  return (
    <Page themeId="tool">
      <Header
        title="Aegis — Launch Workspace"
        subtitle="Create an interactive workspace that supports connection sessions"
      />
      <Content>
        <ContentHeader title="Workspace Form" />

        <Box marginBottom={2}>
          <Typography variant="body2" color="textSecondary">
            Defaults target the local demo environment. Update project, queue,
            and flavor to match real resources before launch.
          </Typography>
        </Box>

        <form onSubmit={handleSubmit} noValidate>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Project ID"
                value={form.projectId}
                onChange={handleFieldChange('projectId')}
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Queue"
                value={form.queue}
                onChange={handleFieldChange('queue')}
                helperText="Optional scheduling queue"
              />
            </Grid>

            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Flavor"
                value={form.flavor}
                onChange={handleFieldChange('flavor')}
                required
                helperText="GPU flavor or node profile"
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Image"
                value={form.image}
                onChange={handleFieldChange('image')}
                required
                helperText="Container image that boots the workspace"
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                label='Command (wrapped as ["sh","-c", ...])'
                value={form.command}
                onChange={handleFieldChange('command')}
                helperText='Defaults to sh -c "echo hello" when left blank'
              />
            </Grid>

            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Max Duration Seconds"
                type="number"
                value={form.maxDurationSeconds}
                onChange={handleDurationChange}
                error={Boolean(durationError)}
                helperText={
                  durationError ??
                  'Optional runtime budget; queue or server defaults apply when empty'
                }
              />
            </Grid>
          </Grid>

          <Box marginTop={2}>
            <Button
              variant="outlined"
              size="small"
              onClick={() => setAdvancedOpen(prev => !prev)}
            >
              {advancedOpen ? 'Hide advanced options' : 'Show advanced options'}
            </Button>
            <Collapse in={advancedOpen}>
              <Box marginTop={2}>
                <Grid container spacing={3}>
                  <Grid item xs={12} md={6}>
                    <TextField
                      fullWidth
                      label="Exposed ports"
                      value={form.ports}
                      onChange={handlePortsChange}
                      error={Boolean(portsError)}
                      helperText={
                        portsError ??
                        `Comma or space separated list. Defaults include SSH (${DEFAULT_SSH_PORT}) and VS Code (${DEFAULT_VSCODE_PORT}).`
                      }
                    />
                  </Grid>
                  <Grid item xs={12}>
                    <TextField
                      fullWidth
                      label="Environment variables"
                      value={form.env}
                      onChange={handleEnvChange}
                      error={Boolean(envError)}
                      helperText={
                        envError ??
                        'Optional KEY=VALUE pairs, one per line. Keys such as AEGIS_SSH_USER, USER_NAME, and PASSWORD_ACCESS affect connection helpers.'
                      }
                      multiline
                      minRows={4}
                    />
                  </Grid>
                </Grid>
              </Box>
            </Collapse>
          </Box>

          <Box marginTop={3} display="flex" alignItems="center" gridGap={16}>
            <Button
              type="submit"
              color="primary"
              variant="contained"
              disabled={isSubmitDisabled}
            >
              Launch Workspace
            </Button>
            {submitting && <Progress />}
          </Box>
        </form>

        {error && (
          <Box marginTop={3}>
            <WarningPanel title="Submission Error" severity="error">
              <Typography
                variant="body2"
                component="pre"
                style={{ whiteSpace: 'pre-wrap' }}
              >
                {error}
              </Typography>
            </WarningPanel>
          </Box>
        )}

        {result && (
          <Box marginTop={3}>
            <InfoCard title="Workspace created">
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Typography variant="body2">
                  <strong>ID:</strong> {result.id ?? '—'}
                </Typography>
                <Typography variant="body2">
                  <strong>Status:</strong> {result.status ?? '—'}
                </Typography>
                {result.clusterId && (
                  <Typography variant="body2">
                    <strong>Cluster:</strong> {result.clusterId}
                  </Typography>
                )}
                <Typography variant="body2" color="textSecondary">
                  The workspace must reach RUNNING before a connection session
                  can mint successfully.
                </Typography>
                {result.id && (
                  <Box>
                    <Button
                      component={RouterLink}
                      to={`/aegis/workloads/${result.id}`}
                      color="primary"
                      variant="outlined"
                      size="small"
                    >
                      View Details
                    </Button>
                  </Box>
                )}
              </Box>
            </InfoCard>
          </Box>
        )}
      </Content>
    </Page>
  );
};
