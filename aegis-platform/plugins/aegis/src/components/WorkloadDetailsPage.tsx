import { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Page,
  Header,
  Content,
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
import { makeStyles, Theme } from '@material-ui/core/styles';
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

const useStyles = makeStyles((theme: Theme) => ({
  page: {
    position: 'relative',
  },
  content: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(4),
    position: 'relative',
    zIndex: 1,
  },
  hero: {
    position: 'relative',
    borderRadius: 32,
    padding: theme.spacing(4.5, 5),
    overflow: 'hidden',
    background:
      'linear-gradient(140deg, rgba(0, 245, 255, 0.16) 0%, rgba(7, 20, 44, 0.9) 55%, rgba(16, 8, 30, 0.88) 100%)',
    boxShadow: '0 36px 64px rgba(3, 12, 35, 0.55)',
  },
  heroGlow: {
    position: 'absolute',
    inset: '-20% -25% auto auto',
    width: '60%',
    height: '160%',
    background:
      'radial-gradient(circle, rgba(76, 255, 166, 0.4) 0%, rgba(76, 255, 166, 0) 65%)',
    filter: 'blur(12px)',
    opacity: 0.4,
  },
  heroHeader: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    gap: theme.spacing(2),
  },
  overline: {
    letterSpacing: '0.28em',
    textTransform: 'uppercase',
    color: 'rgba(197, 226, 255, 0.72)',
  },
  heroTitle: {
    marginTop: theme.spacing(1.5),
    letterSpacing: '0.2em',
    textTransform: 'uppercase',
  },
  statusRow: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2),
    marginTop: theme.spacing(3),
    flexWrap: 'wrap',
  },
  statusMessage: {
    color: 'rgba(197, 226, 255, 0.72)',
  },
  metaGrid: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))',
    gap: theme.spacing(2),
    marginTop: theme.spacing(3),
  },
  metaTile: {
    padding: theme.spacing(1.6, 2),
    borderRadius: 18,
    border: '1px solid rgba(0, 245, 255, 0.16)',
    backgroundColor: 'rgba(0, 245, 255, 0.05)',
  },
  metaLabel: {
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
    color: 'rgba(124, 154, 196, 0.9)',
    fontSize: '0.7rem',
  },
  heroActions: {
    marginTop: theme.spacing(3),
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(1),
    alignItems: 'flex-start',
  },
  connectButton: {
    minWidth: 220,
  },
  cards: {
    display: 'grid',
    gap: theme.spacing(3),
    gridTemplateColumns: 'repeat(auto-fit, minmax(320px, 1fr))',
  },
  infoCard: {
    borderRadius: 24,
    border: '1px solid rgba(0, 245, 255, 0.14)',
    background:
      'linear-gradient(170deg, rgba(6, 16, 38, 0.92) 0%, rgba(10, 24, 48, 0.88) 45%, rgba(16, 8, 30, 0.84) 100%)',
    boxShadow: '0 28px 52px rgba(3, 12, 35, 0.4)',
  },
  link: {
    color: '#00f5ff',
    textDecoration: 'none',
  },
  warning: {
    marginTop: theme.spacing(2),
  },
}));

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
    <Page themeId="tool" className={classes.page}>
      <Header
        title="ÆGIS — Mission Detail"
        subtitle={id ?? 'Mission Identifier'}
      />
      <Content>
        <div className={classes.content}>
          {loading && <Progress />}

          {error && (
            <div className={classes.warning}>
              <WarningPanel title="Failed to load workload" severity="error">
                {error}
              </WarningPanel>
            </div>
          )}

          {workload && (
            <>
              <div className={classes.hero}>
                <div className={classes.heroGlow} />
                <div className={classes.heroHeader}>
                  <Typography variant="overline" className={classes.overline}>
                    Project {workload.projectId ?? '—'}
                  </Typography>
                  <Button
                    variant="outlined"
                    size="small"
                    startIcon={<ArrowBackIcon />}
                    onClick={() => navigate('/aegis/workloads')}
                  >
                    Mission Log
                  </Button>
                </div>
                <Typography variant="h2" className={classes.heroTitle}>
                  {workload.id ?? 'Workspace' }
                </Typography>
                <div className={classes.statusRow}>
                  {statusChip(rawStatus)}
                  {workload.message && (
                    <Typography variant="body2" className={classes.statusMessage}>
                      {workload.message}
                    </Typography>
                  )}
                </div>
                <div className={classes.metaGrid}>
                  {[{ label: 'Flavor', value: getFlavor(workload) || '—' },
                    { label: 'Queue', value: workload.queue ?? '—' },
                    { label: 'Cluster', value: workload.clusterId ?? '—' },
                    { label: 'Endpoint', value: workload.url ?? '—' },
                  ].map(tile => (
                    <div key={tile.label} className={classes.metaTile}>
                      <Typography variant="caption" className={classes.metaLabel}>
                        {tile.label}
                      </Typography>
                      <Typography variant="body2">{tile.value}</Typography>
                    </div>
                  ))}
                </div>
                {canConnect && (
                  <div className={classes.heroActions}>
                    <Button
                      variant="contained"
                      color="primary"
                      disabled={connectButtonDisabled}
                      onClick={handleConnect}
                      className={classes.connectButton}
                    >
                      {sessionLoading ? 'Preparing secure channel…' : 'Open Secure Channel'}
                    </Button>
                    {!isRunning && (
                      <Typography variant="caption" color="textSecondary">
                        Workspace must be running before connecting.
                      </Typography>
                    )}
                  </div>
                )}
              </div>

              <div className={classes.cards}>
                <InfoCard title="Mission Metadata" className={classes.infoCard}>
                  <StructuredMetadataTable metadata={metadata} />
                </InfoCard>

                {kubectlCmd && (
                  <InfoCard title="Field Diagnostics" className={classes.infoCard}>
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
                  <InfoCard title="Launch Specification" className={classes.infoCard}>
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
              </div>

              {loc && (
                <Typography variant="body2">
                  View Kubernetes object{' '}
                  <RouterLink
                    to={`/kubernetes/overview?namespace=${loc.namespace}`}
                    className={classes.link}
                  >
                    {loc.kind} {loc.name}
                  </RouterLink>
                </Typography>
              )}
            </>
          )}
        </div>
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
