import { FC, useCallback, useEffect, useMemo, useState } from 'react';
import { Link as RouterLink, useNavigate } from 'react-router-dom';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
  CopyTextButton,
  StatusOK,
  StatusWarning,
  StatusError,
  StatusPending,
  Table,
  TableColumn,
} from '@backstage/core-components';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  useApi,
} from '@backstage/core-plugin-api';
import {
  Box,
  Button,
  Grid,
  MenuItem,
  Select,
  SelectChangeEvent,
  TextField,
  Typography,
} from '@material-ui/core';
import RefreshIcon from '@material-ui/icons/Refresh';
import {
  WorkloadDTO,
  getFlavor,
  isTerminalStatus,
  listWorkloads,
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

type WorkloadRow = WorkloadDTO & { displayStatus: string };

type StatusFilter = 'all' | 'active' | 'terminal';

export const WorkloadListPage: FC = () => {
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const alertApi = useApi(alertApiRef);
  const navigate = useNavigate();

  const [projectId, setProjectId] = useState('p-demo');
  const [rows, setRows] = useState<WorkloadRow[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all');
  const [search, setSearch] = useState('');
  const [shouldPoll, setShouldPoll] = useState(true);

  const load = useCallback(
    async (opts?: { silent?: boolean }) => {
      const silent = opts?.silent ?? false;
      try {
        if (!silent) {
          setLoading(true);
        }
        setError(null);
        const items = await listWorkloads(fetchApi, discoveryApi, projectId);
        const mapped: WorkloadRow[] = items.map(w => ({
          ...w,
          displayStatus: w.uiStatus ?? w.status ?? 'PLACED',
        }));
        setRows(mapped);
        const anyActive = mapped.some(w => !isTerminalStatus(w.status));
        setShouldPoll(anyActive);
      } catch (e: any) {
        const msg = e?.message ?? String(e);
        setError(msg);
        alertApi.post({ message: `Failed to load workloads: ${msg}`, severity: 'error' });
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [alertApi, discoveryApi, fetchApi, projectId],
  );

  useEffect(() => {
    load();
  }, [load]);

  useEffect(() => {
    if (!shouldPoll) {
      return () => {};
    }
    const timer = setInterval(() => {
      load({ silent: true });
    }, 4000);
    return () => clearInterval(timer);
  }, [shouldPoll, load]);

  const handleProjectChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setProjectId(event.target.value);
  };

  const handleStatusFilter = (event: SelectChangeEvent<StatusFilter>) => {
    setStatusFilter(event.target.value as StatusFilter);
  };

  const filteredRows = useMemo(() => {
    return rows
      .filter(row => {
        if (!search) {
          return true;
        }
        return row.id?.toLowerCase().includes(search.toLowerCase()) ?? false;
      })
      .filter(row => {
        if (statusFilter === 'terminal') {
          return isTerminalStatus(row.status);
        }
        if (statusFilter === 'active') {
          return !isTerminalStatus(row.status);
        }
        return true;
      });
  }, [rows, search, statusFilter]);

  const columns = useMemo<TableColumn<WorkloadRow>[]>(
    () => [
      {
        title: 'Workload ID',
        field: 'id',
        render: row => (
          <Box display="flex" alignItems="center" gridGap={8}>
            {row.id ? (
              <RouterLink to={`/aegis/workloads/${row.id}`} style={{ textDecoration: 'none' }}>
                <Typography variant="body2" color="primary">
                  {row.id}
                </Typography>
              </RouterLink>
            ) : (
              <Typography variant="body2">—</Typography>
            )}
            {row.id ? <CopyTextButton text={row.id} /> : null}
          </Box>
        ),
      },
      {
        title: 'Status',
        field: 'displayStatus',
        render: row => statusChip(row.displayStatus),
      },
      {
        title: 'Flavor',
        field: 'flavor',
        render: row => <Typography variant="body2">{getFlavor(row)}</Typography>,
      },
      {
        title: 'Project',
        field: 'projectId',
      },
      {
        title: 'Link',
        field: 'url',
        render: row => {
          if (!row.url) {
            return <Typography variant="body2" color="textSecondary">N/A</Typography>;
          }
          const loc = parseKubernetesUrl(row.url);
          const cmd = buildKubectlDescribeCommand(loc);
          return (
            <Box display="flex" alignItems="center" gridGap={8}>
              <Typography variant="body2">{row.url}</Typography>
              {cmd ? <CopyTextButton text={cmd} tooltip="Copy kubectl describe" /> : null}
            </Box>
          );
        },
      },
      {
        title: 'Actions',
        field: 'actions',
        sorting: false,
        render: row => (
          <Button
            variant="outlined"
            size="small"
            component={RouterLink}
            to={row.id ? `/aegis/workloads/${row.id}` : '#'}
            disabled={!row.id}
          >
            Details
          </Button>
        ),
      },
    ],
    [],
  );

  const activeCount = rows.filter(r => !isTerminalStatus(r.status)).length;
  const completedCount = rows.filter(r => isTerminalStatus(r.status)).length;

  return (
    <Page themeId="tool">
      <Header title="Aegis — Workload Status" subtitle="Monitor submitted workloads" />
      <Content>
        <ContentHeader title="Filters">
          <Button
            variant="outlined"
            size="small"
            startIcon={<RefreshIcon />}
            onClick={() => load()}
          >
            Refresh
          </Button>
        </ContentHeader>

        <Grid container spacing={2} alignItems="flex-end">
          <Grid item xs={12} md={4}>
            <TextField
              label="Project ID"
              fullWidth
              value={projectId}
              onChange={handleProjectChange}
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              label="Search by Workload ID"
              fullWidth
              value={search}
              onChange={event => setSearch(event.target.value)}
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <Select
              fullWidth
              value={statusFilter}
              onChange={handleStatusFilter}
              displayEmpty
              inputProps={{ 'aria-label': 'Status filter' }}
            >
              <MenuItem value="all">All</MenuItem>
              <MenuItem value="active">Active</MenuItem>
              <MenuItem value="terminal">Terminal</MenuItem>
            </Select>
          </Grid>
        </Grid>

        <Box mt={2} display="flex" gridGap={16}>
          <Typography variant="body2">Active: {activeCount}</Typography>
          <Typography variant="body2">Completed: {completedCount}</Typography>
        </Box>

        {loading && <Progress />}

        {error && (
          <Box mt={2}>
            <WarningPanel title="Failed to load workloads" severity="error">
              {error}
            </WarningPanel>
          </Box>
        )}

        <Box mt={2}>
          <Table
            options={{
              paging: false,
              search: false,
              sorting: true,
              padding: 'dense',
              rowStyle: {
                cursor: 'pointer',
              },
            }}
            data={filteredRows}
            columns={columns}
            title="Workloads"
            onRowClick={(_, row) => {
              if (row?.id) {
                navigate(`/aegis/workloads/${row.id}`);
              }
            }}
          />
        </Box>
      </Content>
    </Page>
  );
};
