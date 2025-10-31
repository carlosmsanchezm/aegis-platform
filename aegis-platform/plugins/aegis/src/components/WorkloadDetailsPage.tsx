import { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Page,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
  InfoCard,
  StructuredMetadataTable,
  StatusOK,
  StatusWarning,
  StatusError,
  StatusPending,
  CopyTextButton,
} from '@backstage/core-components';
import { Box, Button, Typography } from '@material-ui/core';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
  useApi,
} from '@backstage/core-plugin-api';
import ArrowBackIcon from '@material-ui/icons/ArrowBack';
import {
  WorkloadDTO,
  ConnectionSession,
  getWorkload,
  createConnectionSession,
  renewConnectionSession,
  revokeConnectionSession,
  getFlavor,
  mapDisplayStatus,
  parseKubernetesUrl,
  buildKubectlDescribeCommand,
} from '../api/aegisClient';
import { ConnectModal } from './ConnectModal';

const statusChip = (status: string) => {
  const mapped = mapDisplayStatus(status);
  switch (mapped.color) {
    case 'ok':
      return <StatusOK>{mapped.label}</StatusOK>;
    case 'error':
      return <StatusError>{mapped.label}</StatusError>;
    case 'progress':
      return <StatusPending>{mapped.label}</StatusPending>;
    case 'warning':
    default:
      return <StatusWarning>{mapped.label}</StatusWarning>;
  }
};

const getStoredFlag = (key: string): boolean => {
  if (typeof window === 'undefined') {
    return false;
  }
  try {
    return window.localStorage.getItem(key) === 'true';
  } catch {
    return false;
  }
};

const setStoredFlag = (key: string, value: boolean) => {
  if (typeof window === 'undefined') {
    return;
  }
  try {
    window.localStorage.setItem(key, value ? 'true' : 'false');
  } catch {
    // ignore storage failures
  }
};

const HELPER_FLAG = 'aegis.helper.installed';
const SYSTEM_ACK_FLAG = 'aegis.system.use.ack';
const RULES_ACK_FLAG = 'aegis.rules.of.behavior.ack';

export const WorkloadDetailsPage: FC = () => {
  const { id } = useParams<{ id: string }>();
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);
  const navigate = useNavigate();

  const [workload, setWorkload] = useState<WorkloadDTO | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [connectOpen, setConnectOpen] = useState(false);

  const [session, setSession] = useState<ConnectionSession | null>(null);
  const [sessionLoading, setSessionLoading] = useState(false);
  const [sessionError, setSessionError] = useState<string | null>(null);
  const [pendingSession, setPendingSession] = useState(false);

  const [helperInstalled, setHelperInstalled] = useState(() =>
    getStoredFlag(HELPER_FLAG),
  );
  const [systemAcked, setSystemAcked] = useState(() =>
    getStoredFlag(SYSTEM_ACK_FLAG),
  );
  const [rulesAcked, setRulesAcked] = useState(() =>
    getStoredFlag(RULES_ACK_FLAG),
  );

  const load = useCallback(async () => {
    if (!id) {
      setError('Missing workload id');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const res = await getWorkload(fetchApi, discoveryApi, identityApi, id);
      setWorkload(res);
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setError(msg);
      alertApi.post({
        message: `Failed to load workload: ${msg}`,
        severity: 'error',
      });
    } finally {
      setLoading(false);
    }
  }, [alertApi, discoveryApi, fetchApi, identityApi, id]);

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    setSession(null);
  }, [id]);

  const requestSession = useCallback(
    async (client: 'cli' | 'vscode') => {
      if (!workload?.id) {
        alertApi.post({ message: 'Workload id is missing', severity: 'error' });
        return;
      }
      try {
        setSessionLoading(true);
        setSessionError(null);
        const created = await createConnectionSession(
          fetchApi,
          discoveryApi,
          identityApi,
          workload.id,
          client,
        );
        setSession(created);
      } catch (e: any) {
        const msg = e?.message ?? String(e);
        setSessionError(msg);
        alertApi.post({
          message: `Failed to create session: ${msg}`,
          severity: 'error',
        });
      } finally {
        setSessionLoading(false);
        setPendingSession(false);
      }
    },
    [alertApi, discoveryApi, fetchApi, identityApi, workload?.id],
  );

  useEffect(() => {
    if (
      pendingSession &&
      systemAcked &&
      rulesAcked &&
      helperInstalled &&
      !session &&
      !sessionLoading
    ) {
      requestSession('cli');
    }
  }, [
    pendingSession,
    systemAcked,
    rulesAcked,
    helperInstalled,
    session,
    sessionLoading,
    requestSession,
  ]);

  const handleConnectClose = useCallback(() => {
    setConnectOpen(false);
    setSessionError(null);
  }, []);

  const handleConnect = useCallback(() => {
    if (!workload?.id) {
      alertApi.post({ message: 'Workload id is missing', severity: 'error' });
      return;
    }
    setConnectOpen(true);
    setSessionError(null);

    if (session) {
      return;
    }

    if (!systemAcked || !rulesAcked || !helperInstalled) {
      setPendingSession(true);
      return;
    }

    requestSession('cli');
  }, [
    alertApi,
    helperInstalled,
    requestSession,
    rulesAcked,
    session,
    systemAcked,
    workload?.id,
  ]);

  const handleRenew = useCallback(async () => {
    if (!session?.sessionId) {
      return;
    }
    try {
      setSessionLoading(true);
      setSessionError(null);
      const renewed = await renewConnectionSession(
        fetchApi,
        discoveryApi,
        identityApi,
        session.sessionId,
      );
      setSession(renewed);
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setSessionError(msg);
      alertApi.post({
        message: `Failed to renew session: ${msg}`,
        severity: 'error',
      });
    } finally {
      setSessionLoading(false);
    }
  }, [alertApi, discoveryApi, fetchApi, identityApi, session?.sessionId]);

  const handleRevoke = useCallback(async () => {
    if (!session?.sessionId) {
      return;
    }
    try {
      setSessionLoading(true);
      setSessionError(null);
      await revokeConnectionSession(
        fetchApi,
        discoveryApi,
        identityApi,
        session.sessionId,
      );
      setSession(null);
      alertApi.post({ message: 'Session revoked', severity: 'info' });
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setSessionError(msg);
      alertApi.post({
        message: `Failed to revoke session: ${msg}`,
        severity: 'error',
      });
    } finally {
      setSessionLoading(false);
    }
  }, [alertApi, discoveryApi, fetchApi, identityApi, session?.sessionId]);

  const handleSystemAck = useCallback(() => {
    setSystemAcked(true);
    setStoredFlag(SYSTEM_ACK_FLAG, true);
  }, []);

  const handleRulesAck = useCallback(() => {
    setRulesAcked(true);
    setStoredFlag(RULES_ACK_FLAG, true);
  }, []);

  const handleHelperConfirmed = useCallback(() => {
    setHelperInstalled(true);
    setStoredFlag(HELPER_FLAG, true);
  }, []);

  const loc = parseKubernetesUrl(workload?.url);
  const kubectlCmd = buildKubectlDescribeCommand(loc);

  const rawStatus = workload?.uiStatus ?? workload?.status ?? '';
  const canConnect = Boolean(workload?.workspace?.interactive);
  const isRunning = rawStatus === 'RUNNING' || workload?.status === 'RUNNING';
  const connectButtonDisabled = sessionLoading || !isRunning;

  const metadata = useMemo(
    () =>
      workload
        ? {
            'Workload ID': workload.id ?? '—',
            Status: rawStatus || '—',
            Flavor: getFlavor(workload) || '—',
            Project: workload.projectId ?? '—',
            Queue: workload.queue ?? '—',
            Cluster: workload.clusterId ?? '—',
            URL: workload.url ?? '—',
          }
        : {},
    [rawStatus, workload],
  );

  return (
    <Page themeId="tool">
      <Content>
        <ContentHeader title="Workload Details">
          <Typography variant="body1" color="textSecondary">
            {id ?? '—'}
          </Typography>
        </ContentHeader>
        <ContentHeader title="Overview">
          <Button
            variant="outlined"
            size="small"
            startIcon={<ArrowBackIcon />}
            onClick={() => navigate('/aegis/workloads')}
          >
            Back to list
          </Button>
        </ContentHeader>

        {loading && <Progress />}

        {error && (
          <WarningPanel title="Failed to load workload" severity="error">
            {error}
          </WarningPanel>
        )}

        {workload && (
          <Box display="flex" flexDirection="column" gridGap={16}>
            <InfoCard title="Status">
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Box display="flex" alignItems="center" gridGap={16}>
                  {statusChip(rawStatus)}
                  {workload.message && (
                    <Typography variant="body2" color="textSecondary">
                      {workload.message}
                    </Typography>
                  )}
                </Box>
                {canConnect && (
                  <Box>
                    <Button
                      variant="contained"
                      color="primary"
                      disabled={connectButtonDisabled}
                      onClick={handleConnect}
                    >
                      {sessionLoading ? 'Preparing session…' : 'Connect'}
                    </Button>
                    {!isRunning && (
                      <Typography
                        variant="caption"
                        color="textSecondary"
                        display="block"
                      >
                        Workspace must be running before connecting.
                      </Typography>
                    )}
                  </Box>
                )}
              </Box>
            </InfoCard>

            <InfoCard title="Metadata">
              <StructuredMetadataTable metadata={metadata} />
            </InfoCard>

            {kubectlCmd && (
              <InfoCard title="Debug commands">
                <Box display="flex" alignItems="center" gridGap={8}>
                  <Typography variant="body2">{kubectlCmd}</Typography>
                  <CopyTextButton
                    text={kubectlCmd}
                    tooltip="Copy kubectl describe"
                  />
                </Box>
              </InfoCard>
            )}

            {(workload.workspace || workload.training) && (
              <InfoCard title="Specification">
                <StructuredMetadataTable
                  metadata={{
                    Type: workload.workspace ? 'Workspace' : 'Training',
                    Image:
                      workload.workspace?.image ??
                      workload.training?.image ??
                      '—',
                    Command:
                      workload.workspace?.command?.join(' ') ??
                      workload.training?.command?.join(' ') ??
                      '—',
                  }}
                />
              </InfoCard>
            )}

            {loc && (
              <Typography variant="body2">
                View Kubernetes object{' '}
                <RouterLink
                  to={`/kubernetes/overview?namespace=${loc.namespace}`}
                >
                  {loc.kind} {loc.name}
                </RouterLink>
              </Typography>
            )}
          </Box>
        )}
      </Content>
      <ConnectModal
        open={connectOpen}
        onClose={handleConnectClose}
        loading={sessionLoading}
        error={sessionError}
        session={session}
        pendingSession={pendingSession}
        helperInstalled={helperInstalled}
        onConfirmHelper={handleHelperConfirmed}
        systemAcked={systemAcked}
        onAcknowledgeSystemUse={handleSystemAck}
        rulesAcked={rulesAcked}
        onAcknowledgeRules={handleRulesAck}
        onRequestSession={requestSession}
        onRenew={handleRenew}
        onRevoke={handleRevoke}
        workloadId={workload?.id ?? ''}
      />
    </Page>
  );
};
