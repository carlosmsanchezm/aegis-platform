import { FC, useCallback, useEffect, useState } from 'react';
import { useParams, useNavigate, Link as RouterLink } from 'react-router-dom';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
  InfoCard,
  CopyTextButton,
  StructuredMetadataTable,
  StatusOK,
  StatusWarning,
  StatusError,
  StatusPending,
} from '@backstage/core-components';
import { Box, Button, Typography } from '@material-ui/core';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  useApi,
} from '@backstage/core-plugin-api';
import ArrowBackIcon from '@material-ui/icons/ArrowBack';
import {
  WorkloadDTO,
  ConnectionDetails,
  getWorkload,
  getWorkspaceConnectionDetails,
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

export const WorkloadDetailsPage: FC = () => {
  const { id } = useParams<{ id: string }>();
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const alertApi = useApi(alertApiRef);
  const navigate = useNavigate();

  const [workload, setWorkload] = useState<WorkloadDTO | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [connectOpen, setConnectOpen] = useState(false);
  const [connectLoading, setConnectLoading] = useState(false);
  const [connectDetails, setConnectDetails] = useState<ConnectionDetails | null>(null);
  const [connectError, setConnectError] = useState<string | null>(null);

  const load = useCallback(async () => {
    if (!id) {
      setError('Missing workload id');
      return;
    }
    try {
      setLoading(true);
      setError(null);
      const res = await getWorkload(fetchApi, discoveryApi, id);
      setWorkload(res);
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setError(msg);
      alertApi.post({ message: `Failed to load workload: ${msg}`, severity: 'error' });
    } finally {
      setLoading(false);
    }
  }, [alertApi, discoveryApi, fetchApi, id]);

  useEffect(() => {
    load();
  }, [load]);

  const handleConnectClose = useCallback(() => {
    setConnectOpen(false);
    setConnectLoading(false);
    setConnectError(null);
    setConnectDetails(null);
  }, []);

  const handleConnect = useCallback(async () => {
    if (!workload?.id) {
      alertApi.post({ message: 'Workload id is missing', severity: 'error' });
      return;
    }

    setConnectOpen(true);
    setConnectLoading(true);
    setConnectError(null);
    setConnectDetails(null);

    try {
      const details = await getWorkspaceConnectionDetails(fetchApi, discoveryApi, workload.id);
      setConnectDetails(details);
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setConnectError(msg);
      alertApi.post({ message: `Failed to fetch connection details: ${msg}`, severity: 'error' });
    } finally {
      setConnectLoading(false);
    }
  }, [alertApi, discoveryApi, fetchApi, workload?.id]);

  const loc = parseKubernetesUrl(workload?.url);
  const kubectlCmd = buildKubectlDescribeCommand(loc);

  const rawStatus = workload?.uiStatus ?? workload?.status ?? '';
  const canConnect = Boolean(workload?.workspace?.interactive);
  const isRunning = rawStatus === 'RUNNING' || workload?.status === 'RUNNING';
  const connectButtonDisabled = connectLoading || !isRunning;

  const metadata = workload
    ? {
        'Workload ID': workload.id ?? '—',
        Status: rawStatus || '—',
        Flavor: getFlavor(workload) || '—',
        Project: workload.projectId ?? '—',
        Queue: workload.queue ?? '—',
        Cluster: workload.clusterId ?? '—',
        URL: workload.url ?? '—',
      }
    : {};

  return (
    <Page themeId="tool">
      <Header title="Workload Details" subtitle={id} />
      <Content>
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
                      {connectLoading ? 'Requesting token…' : 'Connect'}
                    </Button>
                    {!isRunning && (
                      <Typography variant="caption" color="textSecondary" display="block">
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
                  <CopyTextButton text={kubectlCmd} tooltip="Copy kubectl describe" />
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
                    Command: workload.workspace?.command?.join(' ') ??
                      workload.training?.command?.join(' ') ??
                      '—',
                  }}
                />
              </InfoCard>
            )}

            {loc && (
              <Typography variant="body2">
                View Kubernetes object:{' '}
                <RouterLink to={`/kubernetes/overview?namespace=${loc.namespace}`}>
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
        loading={connectLoading}
        error={connectError}
        details={connectDetails}
        workloadId={workload?.id ?? ''}
      />
    </Page>
  );
};
