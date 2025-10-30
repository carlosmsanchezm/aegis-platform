import {
  ChangeEvent,
  FC,
  FormEvent,
  memo,
  RefObject,
  useCallback,
  useEffect,
  useRef,
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
import { Box, Button, Grid, TextField, Typography } from '@material-ui/core';
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
import { useFuturisticPageStyles } from './useFuturisticPageStyles';

const DEFAULT_PROJECT_ID = 'p-demo';
const DEFAULT_QUEUE = 'default';
const DEFAULT_FLAVOR = 'a10-mig-1g';
const DEFAULT_IMAGE = 'alpine:3.19';
const DEFAULT_COMMAND = 'echo Hello from Aegis; sleep 2';
const DEFAULT_DURATION = '600';
const DEFAULT_ENV_TEXT = formatDefaultEnv();
const DEFAULT_PORTS_TEXT = `${DEFAULT_SSH_PORT}, ${DEFAULT_VSCODE_PORT}`;
const TERMINAL_WORKLOAD_STATUSES = new Set(['SUCCEEDED', 'FAILED']);

type BasicFieldsProps = {
  projectIdRef: RefObject<HTMLInputElement>;
  queueRef: RefObject<HTMLInputElement>;
  flavorRef: RefObject<HTMLInputElement>;
  imageRef: RefObject<HTMLInputElement>;
  commandRef: RefObject<HTMLInputElement>;
  durationRef: RefObject<HTMLInputElement>;
  onProjectChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onFlavorChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onImageChange: (event: ChangeEvent<HTMLInputElement>) => void;
  onDurationChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  durationError: string | null;
};

const BasicFields = memo<BasicFieldsProps>(
  ({
    projectIdRef,
    queueRef,
    flavorRef,
    imageRef,
    commandRef,
    durationRef,
    onProjectChange,
    onFlavorChange,
    onImageChange,
    onDurationChange,
    durationError,
  }) => (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Project ID"
          inputRef={projectIdRef}
          defaultValue={DEFAULT_PROJECT_ID}
          onChange={onProjectChange}
          required
        />
      </Grid>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Queue"
          inputRef={queueRef}
          defaultValue={DEFAULT_QUEUE}
          helperText="Optional scheduling queue"
        />
      </Grid>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Flavor"
          inputRef={flavorRef}
          defaultValue={DEFAULT_FLAVOR}
          onChange={onFlavorChange}
          required
          helperText="GPU flavor or node profile"
        />
      </Grid>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Image"
          inputRef={imageRef}
          defaultValue={DEFAULT_IMAGE}
          onChange={onImageChange}
          required
          helperText="Container image that boots the workspace"
        />
      </Grid>
      <Grid item xs={12}>
        <TextField
          fullWidth
          label='Command (wrapped as ["sh","-c", ...])'
          inputRef={commandRef}
          defaultValue={DEFAULT_COMMAND}
          helperText='Defaults to sh -c "echo hello" when left blank'
        />
      </Grid>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Max Duration Seconds"
          type="number"
          inputRef={durationRef}
          defaultValue={DEFAULT_DURATION}
          onChange={onDurationChange}
          error={Boolean(durationError)}
          helperText={
            durationError ??
            'Optional runtime budget; queue or server defaults apply when empty'
          }
        />
      </Grid>
    </Grid>
  ),
);

BasicFields.displayName = 'BasicFields';

type AdvancedFieldsProps = {
  portsRef: RefObject<HTMLInputElement>;
  envRef: RefObject<HTMLTextAreaElement>;
  portsError: string | null;
  envError: string | null;
  onPortsChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
  onEnvChange: (
    event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>,
  ) => void;
};

const AdvancedFields = memo<AdvancedFieldsProps>(
  ({ portsRef, envRef, portsError, envError, onPortsChange, onEnvChange }) => (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6}>
        <TextField
          fullWidth
          label="Exposed ports"
          inputRef={portsRef}
          defaultValue={DEFAULT_PORTS_TEXT}
          onChange={onPortsChange}
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
          inputRef={envRef}
          defaultValue={DEFAULT_ENV_TEXT}
          onChange={onEnvChange}
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
  ),
);

AdvancedFields.displayName = 'AdvancedFields';

export const SubmitWorkloadPage: FC = () => {
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);
  const classes = useFuturisticPageStyles();

  const projectIdRef = useRef<HTMLInputElement>(null);
  const queueRef = useRef<HTMLInputElement>(null);
  const flavorRef = useRef<HTMLInputElement>(null);
  const imageRef = useRef<HTMLInputElement>(null);
  const commandRef = useRef<HTMLInputElement>(null);
  const durationRef = useRef<HTMLInputElement>(null);
  const portsRef = useRef<HTMLInputElement>(null);
  const envRef = useRef<HTMLTextAreaElement>(null);

  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<WorkloadDTO | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [portsError, setPortsError] = useState<string | null>(null);
  const [envError, setEnvError] = useState<string | null>(null);
  const [durationError, setDurationError] = useState<string | null>(null);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [isSubmitDisabled, setIsSubmitDisabled] = useState(false);
  const pollTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const computeSubmitDisabled = useCallback(
    (overrides?: {
      portsError?: string | null;
      envError?: string | null;
      durationError?: string | null;
      submitting?: boolean;
    }) => {
      const projectValue = projectIdRef.current?.value?.trim() ?? '';
      const flavorValue = flavorRef.current?.value?.trim() ?? '';
      const imageValue = imageRef.current?.value?.trim() ?? '';
      const nextPortsError = overrides?.portsError ?? portsError;
      const nextEnvError = overrides?.envError ?? envError;
      const nextDurationError = overrides?.durationError ?? durationError;
      const nextSubmitting = overrides?.submitting ?? submitting;

      return (
        nextSubmitting ||
        !projectValue ||
        !flavorValue ||
        !imageValue ||
        Boolean(nextPortsError) ||
        Boolean(nextEnvError) ||
        Boolean(nextDurationError)
      );
    },
    [portsError, envError, durationError, submitting],
  );

  const evaluateSubmitReady = useCallback(
    (overrides?: {
      portsError?: string | null;
      envError?: string | null;
      durationError?: string | null;
      submitting?: boolean;
    }) => {
      setIsSubmitDisabled(computeSubmitDisabled(overrides));
    },
    [computeSubmitDisabled],
  );

  useEffect(() => {
    evaluateSubmitReady();
  }, [evaluateSubmitReady]);

  const handleProjectChange = useCallback(() => {
    evaluateSubmitReady();
  }, [evaluateSubmitReady]);

  const handleFlavorChange = useCallback(() => {
    evaluateSubmitReady();
  }, [evaluateSubmitReady]);

  const handleImageChange = useCallback(() => {
    evaluateSubmitReady();
  }, [evaluateSubmitReady]);

  const handleDurationChange = useCallback(
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const value = event.target.value;
      if (!value.trim()) {
        setDurationError(null);
        evaluateSubmitReady({ durationError: null });
        return;
      }
      const parsed = Number.parseInt(value, 10);
      const nextError =
        !Number.isInteger(parsed) || parsed <= 0
          ? 'Duration must be a positive number'
          : null;
      setDurationError(nextError);
      evaluateSubmitReady({ durationError: nextError });
    },
    [evaluateSubmitReady],
  );

  const handlePortsChange = useCallback(
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const nextError = validatePortsInput(event.target.value);
      setPortsError(nextError);
      evaluateSubmitReady({ portsError: nextError });
    },
    [evaluateSubmitReady],
  );

  const handleEnvChange = useCallback(
    (event: ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
      const nextError = validateEnvInput(event.target.value);
      setEnvError(nextError);
      evaluateSubmitReady({ envError: nextError });
    },
    [evaluateSubmitReady],
  );

  const handleSubmit = useCallback(
    async (event: FormEvent) => {
      event.preventDefault();

      const projectId = projectIdRef.current?.value?.trim() ?? '';
      const flavor = flavorRef.current?.value?.trim() ?? '';
      const image = imageRef.current?.value?.trim() ?? '';

      if (!projectId || !flavor || !image) {
        evaluateSubmitReady();
        return;
      }

      const queueValue = queueRef.current?.value?.trim() ?? '';
      const commandText = commandRef.current?.value?.trim() ?? '';
      const durationValue = durationRef.current?.value?.trim() ?? '';
      const portsValue = portsRef.current?.value ?? DEFAULT_PORTS_TEXT;
      const envValue = envRef.current?.value ?? DEFAULT_ENV_TEXT;

      const envOverrides = parseEnvInput(envValue);
      const ports = parsePortsInput(portsValue);

      let parsedDuration: number | undefined;
      if (durationValue) {
        const parsed = Number.parseInt(durationValue, 10);
        if (!Number.isInteger(parsed) || parsed <= 0) {
          const nextError = 'Duration must be a positive number';
          setDurationError(nextError);
          evaluateSubmitReady({ durationError: nextError });
          return;
        }
        parsedDuration = parsed;
      }

      const payload: SubmitWorkspaceRequest = {
        projectId,
        queue: queueValue || undefined,
        workspace: {
          flavor,
          image,
          command: commandText.trim(),
          ports,
          env: Object.keys(envOverrides).length > 0 ? envOverrides : undefined,
          maxDurationSeconds: parsedDuration,
        },
      };

      try {
        setSubmitting(true);
        evaluateSubmitReady({ submitting: true });
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
        evaluateSubmitReady({ submitting: false });
      }
    },
    [
      alertApi,
      discoveryApi,
      evaluateSubmitReady,
      fetchApi,
      identityApi,
    ],
  );

  useEffect(() => {
    if (!result?.id || TERMINAL_WORKLOAD_STATUSES.has(result.status ?? '')) {
      if (pollTimerRef.current) {
        clearTimeout(pollTimerRef.current);
        pollTimerRef.current = null;
      }
      return undefined;
    }

    let cancelled = false;

    const pollOnce = async () => {
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

        const terminal = TERMINAL_WORKLOAD_STATUSES.has(fresh.status ?? '');
        if (terminal) {
          if (pollTimerRef.current) {
            clearTimeout(pollTimerRef.current);
            pollTimerRef.current = null;
          }
          cancelled = true;
          return;
        }

        if (!cancelled) {
          pollTimerRef.current = setTimeout(pollOnce, 450);
        }
      } catch {
        /* ignore polling errors */
      }
    };

    pollOnce();

    return () => {
      cancelled = true;
      if (pollTimerRef.current) {
        clearTimeout(pollTimerRef.current);
        pollTimerRef.current = null;
      }
    };
  }, [result?.id, result?.status, fetchApi, discoveryApi, identityApi]);

  return (
    <Page themeId="tool" className={classes.page}>
      <div className={classes.background} />
      <Header
        title="Mission Workspace Deployment"
        subtitle="Spin up zero-trust interactive surfaces across Aegis-connected clouds."
        className={classes.header}
      >
        <Typography variant="subtitle2" className={classes.headerSubtitle}>
          Launch Console · Interactive GPU workloads with mission safeguards
        </Typography>
      </Header>
      <Content className={classes.content}>
        <ContentHeader
          title="Launch Parameters"
          className={classes.sectionTitle}
        />

        <Box className={classes.callout}>
          Configure the mission package for this workspace. Defaults map to the
          demo range — adjust project, queue, and flavor to align with your
          operational theater before arming.
        </Box>

        <form onSubmit={handleSubmit} noValidate className={classes.holoPanel}>
          <BasicFields
            projectIdRef={projectIdRef}
            queueRef={queueRef}
            flavorRef={flavorRef}
            imageRef={imageRef}
            commandRef={commandRef}
            durationRef={durationRef}
            onProjectChange={handleProjectChange}
            onFlavorChange={handleFlavorChange}
            onImageChange={handleImageChange}
            onDurationChange={handleDurationChange}
            durationError={durationError}
          />

          <Box marginTop={2}>
            <Button
              variant="outlined"
              size="small"
              onClick={() => setAdvancedOpen(prev => !prev)}
            >
              {advancedOpen ? 'Hide advanced options' : 'Show advanced options'}
            </Button>
            {advancedOpen && (
              <Box marginTop={2}>
                <AdvancedFields
                  portsRef={portsRef}
                  envRef={envRef}
                  portsError={portsError}
                  envError={envError}
                  onPortsChange={handlePortsChange}
                  onEnvChange={handleEnvChange}
                />
              </Box>
            )}
          </Box>

          <Box
            marginTop={3}
            display="flex"
            alignItems="center"
            gridGap={16}
          >
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
          <Box marginTop={3} className={classes.holoPanel}>
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
            <InfoCard title="Workspace created" className={classes.holoPanel}>
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Box className={classes.inlineStat}>
                  <Typography component="span" variant="body2">
                    ID
                  </Typography>
                  <Typography component="span" variant="body2" color="inherit">
                    {result.id ?? '—'}
                  </Typography>
                </Box>
                <Box className={classes.inlineStat}>
                  <Typography component="span" variant="body2">
                    Status
                  </Typography>
                  <Typography component="span" variant="body2" color="inherit">
                    {result.status ?? '—'}
                  </Typography>
                </Box>
                {result.clusterId && (
                  <Box className={classes.inlineStat}>
                    <Typography component="span" variant="body2">
                      Cluster
                    </Typography>
                    <Typography
                      component="span"
                      variant="body2"
                      color="inherit"
                    >
                      {result.clusterId}
                    </Typography>
                  </Box>
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
