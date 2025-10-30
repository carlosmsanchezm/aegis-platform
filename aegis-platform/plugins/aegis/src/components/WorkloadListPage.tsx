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
  identityApiRef,
  useApi,
  useRouteRef,
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
import { makeStyles, Theme } from '@material-ui/core/styles';
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
import { createWorkspaceRouteRef } from '../routes';

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
    padding: theme.spacing(4.5, 5),
    borderRadius: 30,
    overflow: 'hidden',
    background:
      'linear-gradient(120deg, rgba(0, 245, 255, 0.18) 0%, rgba(6, 18, 42, 0.92) 55%, rgba(16, 7, 28, 0.88) 100%)',
    boxShadow: '0 28px 56px rgba(3, 12, 35, 0.55)',
  },
  heroGlow: {
    position: 'absolute',
    inset: '-25% -35% auto auto',
    width: '60%',
    height: '160%',
    background:
      'radial-gradient(circle, rgba(255, 45, 149, 0.45) 0%, rgba(255, 45, 149, 0) 65%)',
    filter: 'blur(12px)',
    opacity: 0.4,
  },
  heroTitle: {
    letterSpacing: '0.18em',
    textTransform: 'uppercase',
    marginBottom: theme.spacing(1.5),
  },
  heroSubtitle: {
    color: 'rgba(197, 226, 255, 0.8)',
    maxWidth: 680,
  },
  heroStats: {
    marginTop: theme.spacing(3.5),
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(180px, 1fr))',
    gap: theme.spacing(2),
  },
  statCard: {
    padding: theme.spacing(2.2, 2.6),
    borderRadius: 20,
    border: '1px solid rgba(0, 245, 255, 0.18)',
    background:
      'linear-gradient(160deg, rgba(0, 245, 255, 0.08) 0%, rgba(7, 18, 44, 0.82) 60%, rgba(16, 8, 30, 0.85) 100%)',
    boxShadow: '0 16px 36px rgba(0, 245, 255, 0.12)',
  },
  statValue: {
    fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
    fontSize: '2.2rem',
    letterSpacing: '0.22em',
    color: '#00f5ff',
  },
  statLabel: {
    marginTop: theme.spacing(1),
    letterSpacing: '0.12em',
    color: 'rgba(124, 154, 196, 0.9)',
    textTransform: 'uppercase',
  },
  filtersPanel: {
    borderRadius: 26,
    padding: theme.spacing(3, 4),
    background:
      'linear-gradient(150deg, rgba(5, 12, 28, 0.96) 0%, rgba(8, 20, 40, 0.9) 45%, rgba(16, 8, 30, 0.86) 100%)',
    border: '1px solid rgba(0, 245, 255, 0.12)',
    boxShadow: '0 22px 40px rgba(3, 12, 35, 0.45)',
  },
  sectionTitle: {
    letterSpacing: '0.16em',
    textTransform: 'uppercase',
    color: '#9cbdef',
  },
  filtersGrid: {
    marginTop: theme.spacing(1),
  },
  filterLabel: {
    letterSpacing: '0.14em',
    textTransform: 'uppercase',
    color: '#9cbdef',
  },
  filterField: {
    '& .MuiInputBase-root': {
      borderRadius: 16,
      backgroundColor: 'rgba(0, 245, 255, 0.06)',
    },
  },
  actionButtons: {
    display: 'flex',
    gap: theme.spacing(1.5),
    flexWrap: 'wrap',
  },
  tablePanel: {
    borderRadius: 28,
    padding: theme.spacing(3, 3, 1),
    border: '1px solid rgba(0, 245, 255, 0.15)',
    background:
      'linear-gradient(160deg, rgba(6, 18, 42, 0.92) 0%, rgba(10, 26, 52, 0.9) 45%, rgba(16, 8, 30, 0.84) 100%)',
    boxShadow: '0 26px 46px rgba(3, 12, 35, 0.42)',
  },
  tableTitle: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginBottom: theme.spacing(2),
  },
  tableCaption: {
    color: 'rgba(197, 226, 255, 0.7)',
    letterSpacing: '0.08em',
  },
  summaryBadges: {
    display: 'flex',
    gap: theme.spacing(3),
    flexWrap: 'wrap',
    color: 'rgba(197, 226, 255, 0.8)',
  },
  warningPanel: {
    marginTop: theme.spacing(2.5),
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

type WorkloadRow = WorkloadDTO & { displayStatus: string };

type StatusFilter = 'all' | 'active' | 'terminal';

export const WorkloadListPage: FC = () => {
  const classes = useStyles();
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);
  const navigate = useNavigate();
  const createWorkspaceLink = useRouteRef(createWorkspaceRouteRef);
  const createWorkspacePath = createWorkspaceLink();

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
        const items = await listWorkloads(
          fetchApi,
          discoveryApi,
          identityApi,
          projectId,
        );
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
        alertApi.post({
          message: `Failed to load workloads: ${msg}`,
          severity: 'error',
        });
      } finally {
        if (!silent) {
          setLoading(false);
        }
      }
    },
    [alertApi, discoveryApi, fetchApi, identityApi, projectId],
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
              <RouterLink
                to={`/aegis/workloads/${row.id}`}
                style={{ textDecoration: 'none' }}
              >
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
        render: row => (
          <Typography variant="body2">{getFlavor(row)}</Typography>
        ),
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
            return (
              <Typography variant="body2" color="textSecondary">
                N/A
              </Typography>
            );
          }
          const loc = parseKubernetesUrl(row.url);
          const cmd = buildKubectlDescribeCommand(loc);
          return (
            <Box display="flex" alignItems="center" gridGap={8}>
              <Typography variant="body2">{row.url}</Typography>
              {cmd ? (
                <CopyTextButton text={cmd} tooltip="Copy kubectl describe" />
              ) : null}
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
    <Page themeId="tool" className={classes.page}>
      <Header
        title="ÆGIS — Mission Operations"
        subtitle="Observe GPU sorties, status telemetry, and cross-cloud placement"
      />
      <Content>
        <div className={classes.content}>
          <div className={classes.hero}>
            <div className={classes.heroGlow} />
            <Typography variant="h2" className={classes.heroTitle}>
              Operational Dashboard
            </Typography>
            <Typography variant="subtitle1" className={classes.heroSubtitle}>
              Track every mission workspace, enforce compliance overlays, and
              surface anomalies before they escalate.
            </Typography>
            <div className={classes.heroStats}>
              <div className={classes.statCard}>
                <Typography className={classes.statValue}>
                  {rows.length.toString().padStart(2, '0')}
                </Typography>
                <Typography variant="caption" className={classes.statLabel}>
                  Total Missions
                </Typography>
              </div>
              <div className={classes.statCard}>
                <Typography className={classes.statValue}>
                  {activeCount.toString().padStart(2, '0')}
                </Typography>
                <Typography variant="caption" className={classes.statLabel}>
                  Active Sorties
                </Typography>
              </div>
              <div className={classes.statCard}>
                <Typography className={classes.statValue}>
                  {completedCount.toString().padStart(2, '0')}
                </Typography>
                <Typography variant="caption" className={classes.statLabel}>
                  Completed Missions
                </Typography>
              </div>
            </div>
          </div>

          <div className={classes.filtersPanel}>
            <ContentHeader title="Mission Filters">
              <div className={classes.actionButtons}>
                <Button
                  variant="outlined"
                  size="small"
                  startIcon={<RefreshIcon />}
                  onClick={() => load()}
                >
                  Refresh
                </Button>
                <Button
                  variant="contained"
                  color="primary"
                  component={RouterLink}
                  to={createWorkspacePath}
                >
                  Launch Workspace
                </Button>
              </div>
            </ContentHeader>
            <Grid container spacing={3} className={classes.filtersGrid}>
              <Grid item xs={12} md={4}>
                <Typography variant="caption" className={classes.filterLabel}>
                  Project
                </Typography>
                <TextField
                  fullWidth
                  value={projectId}
                  onChange={handleProjectChange}
                  className={classes.filterField}
                />
              </Grid>
              <Grid item xs={12} md={4}>
                <Typography variant="caption" className={classes.filterLabel}>
                  Workload ID
                </Typography>
                <TextField
                  fullWidth
                  value={search}
                  onChange={event => setSearch(event.target.value)}
                  className={classes.filterField}
                />
              </Grid>
              <Grid item xs={12} md={4}>
                <Typography variant="caption" className={classes.filterLabel}>
                  Status
                </Typography>
                <Select
                  fullWidth
                  value={statusFilter}
                  onChange={handleStatusFilter}
                  displayEmpty
                  inputProps={{ 'aria-label': 'Status filter' }}
                  className={classes.filterField}
                >
                  <MenuItem value="all">All</MenuItem>
                  <MenuItem value="active">Active</MenuItem>
                  <MenuItem value="terminal">Terminal</MenuItem>
                </Select>
              </Grid>
            </Grid>
            {loading && <Progress />}
          </div>

          {error && (
            <div className={classes.warningPanel}>
              <WarningPanel title="Failed to load workloads" severity="error">
                {error}
              </WarningPanel>
            </div>
          )}

          <div className={classes.tablePanel}>
            <div className={classes.tableTitle}>
              <Typography variant="h5" className={classes.sectionTitle}>
                Mission Logs
              </Typography>
              <div className={classes.summaryBadges}>
                <Typography variant="body2">
                  Active: {activeCount}
                </Typography>
                <Typography variant="body2">
                  Terminal: {completedCount}
                </Typography>
              </div>
            </div>
            <Typography variant="caption" className={classes.tableCaption}>
              Tap a mission to inspect GPU placement, compliance posture, and
              connection tooling.
            </Typography>
            <Table
              options={{
                paging: false,
                search: false,
                sorting: true,
                padding: 'dense',
                rowStyle: row => ({
                  cursor: 'pointer',
                  backgroundColor: isTerminalStatus((row as WorkloadRow).status)
                    ? 'rgba(79, 23, 255, 0.08)'
                    : 'rgba(0, 245, 255, 0.04)',
                  borderBottom: '1px solid rgba(0, 245, 255, 0.08)',
                }),
              }}
              data={filteredRows}
              columns={columns}
              title=""
              onRowClick={(_, row) => {
                if (row?.id) {
                  navigate(`/aegis/workloads/${row.id}`);
                }
              }}
            />
          </div>
        </div>
      </Content>
    </Page>
  );
};
