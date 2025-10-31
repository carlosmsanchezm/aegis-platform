import {
  Box,
  Chip,
  Grid,
  LinearProgress,
  List,
  ListItem,
  ListItemText,
  makeStyles,
  Paper,
  Typography,
} from '@material-ui/core';
import {
  Content,
  ContentHeader,
  Page,
} from '@backstage/core-components';

const useStyles = makeStyles(theme => ({
  pageContent: {
    paddingBottom: theme.spacing(6),
  },
  metricCard: {
    padding: theme.spacing(3),
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(2),
    border: '1px solid rgba(148, 163, 184, 0.18)',
    background:
      'linear-gradient(160deg, rgba(15,23,42,0.88) 0%, rgba(15,23,42,0.55) 100%)',
  },
  metricHeader: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  metricValue: {
    fontSize: '2rem',
    fontWeight: 600,
    letterSpacing: '-0.02em',
  },
  metricTrend: {
    color: theme.palette.success.main,
  },
  graph: {
    height: 8,
    borderRadius: 999,
    backgroundColor: 'rgba(148, 163, 184, 0.15)',
    '& .MuiLinearProgress-barColorPrimary': {
      borderRadius: 999,
      background:
        'linear-gradient(135deg, rgba(14,165,233,0.85), rgba(99,102,241,0.95))',
    },
  },
  listItem: {
    padding: theme.spacing(2, 0),
    '&:not(:last-child)': {
      borderBottom: '1px solid rgba(148, 163, 184, 0.08)',
    },
  },
  subtle: {
    color: theme.palette.text.secondary,
  },
  sparklines: {
    display: 'flex',
    gap: 4,
    alignItems: 'flex-end',
    height: 36,
  },
  sparkBar: {
    width: 6,
    borderRadius: 8,
    background:
      'linear-gradient(180deg, rgba(16,185,129,0.9) 0%, rgba(56,189,248,0.9) 100%)',
  },
}));

const analytics = [
  {
    label: 'GPU Utilization',
    value: '72%',
    trend: '+5.4%',
    progress: 72,
    status: 'Healthy',
  },
  {
    label: 'Workspace Latency (p95)',
    value: '142 ms',
    trend: '-18 ms',
    progress: 62,
    status: 'Improving',
  },
  {
    label: 'GPU Queue Depth',
    value: '0.86',
    trend: '-0.12',
    progress: 44,
    status: 'Balanced',
  },
];

const observability = [
  {
    name: 'aws-gov-west-2 · titan-h100',
    summary: 'GPU saturation 68% · Pod disruption budget intact',
    status: 'Operational',
  },
  {
    name: 'azure-il6 · atlas-mi300x',
    summary: 'Scaling event underway · 6 nodes joining in 45s',
    status: 'Scaling',
  },
  {
    name: 'gcp-us-central-secure · neptune-a100',
    summary: 'Two nodes in maintenance window · capacity rerouted',
    status: 'Maintaining',
  },
];

const sparklineHeights = [18, 12, 28, 24, 34, 26, 30, 20, 32, 24];

export const AegisTelemetryPage = () => {
  const classes = useStyles();

  return (
    <Page themeId="tool">
      <Content className={classes.pageContent}>
        <ContentHeader title="Telemetry Pulse">
          <Chip label="Realtime" color="primary" />
          <Chip label="Streaming from all regions" variant="outlined" />
        </ContentHeader>
        <Box px={4} pb={6}>
          <Grid container spacing={4}>
            {analytics.map(metric => (
              <Grid item xs={12} md={4} key={metric.label}>
                <Paper className={classes.metricCard} elevation={0}>
                  <div className={classes.metricHeader}>
                    <Typography variant="subtitle1">{metric.label}</Typography>
                    <Chip label={metric.status} color="secondary" size="small" />
                  </div>
                  <Typography className={classes.metricValue}>
                    {metric.value}
                  </Typography>
                  <Typography variant="body2" className={classes.metricTrend}>
                    {metric.trend} vs baseline
                  </Typography>
                  <div className={classes.sparklines}>
                    {sparklineHeights.map((height, index) => (
                      <span
                        key={`${metric.label}-spark-${index}`}
                        className={classes.sparkBar}
                        style={{ height }}
                      />
                    ))}
                  </div>
                  <LinearProgress
                    variant="determinate"
                    value={metric.progress}
                    className={classes.graph}
                  />
                </Paper>
              </Grid>
            ))}
            <Grid item xs={12} md={5}>
              <Paper className={classes.metricCard} elevation={0}>
                <Typography variant="h5">Signal Stream</Typography>
                <Typography variant="body2" className={classes.subtle}>
                  Federated traces summarized across clouds, filtered for mission
                  activity and GPU hotspots.
                </Typography>
                <List disablePadding>
                  {observability.map(item => (
                    <ListItem key={item.name} className={classes.listItem}>
                      <ListItemText
                        primary={item.name}
                        secondary={item.summary}
                      />
                      <Chip label={item.status} variant="default" />
                    </ListItem>
                  ))}
                </List>
              </Paper>
            </Grid>
            <Grid item xs={12} md={7}>
              <Paper className={classes.metricCard} elevation={0}>
                <Typography variant="h5">Streaming Analytics</Typography>
                <Typography variant="body2" className={classes.subtle}>
                  Live GPU health, workspace performance, and policy events.
                </Typography>
                <Box mt={2}>
                  <Grid container spacing={2}>
                    {[72, 64, 58, 76, 67, 71].map((value, index) => (
                      <Grid item xs={12} md={4} key={`segment-${index}`}>
                        <Paper elevation={0} className={classes.metricCard}>
                          <Typography variant="subtitle2">
                            Segment {index + 1}
                          </Typography>
                          <Typography className={classes.metricValue}>
                            {value}%
                          </Typography>
                          <LinearProgress
                            variant="determinate"
                            value={value}
                            className={classes.graph}
                          />
                        </Paper>
                      </Grid>
                    ))}
                  </Grid>
                </Box>
              </Paper>
            </Grid>
          </Grid>
        </Box>
      </Content>
    </Page>
  );
};

export default AegisTelemetryPage;
