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
  getWorkload,
  getFlavor,
  mapDisplayStatus,
  parseKubernetesUrl,
  buildKubectlDescribeCommand,
} from '../api/aegisClient';

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

  const loc = parseKubernetesUrl(workload?.url);
  const kubectlCmd = buildKubectlDescribeCommand(loc);

  const metadata = workload
    ? {
        'Workload ID': workload.id ?? '—',
        Status: workload.uiStatus ?? workload.status ?? '—',
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
              <Box display="flex" alignItems="center" gridGap={16}>
                {statusChip(workload.uiStatus ?? workload.status ?? '')}
                {workload.message && (
                  <Typography variant="body2" color="textSecondary">
                    {workload.message}
                  </Typography>
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
    </Page>
  );
};
