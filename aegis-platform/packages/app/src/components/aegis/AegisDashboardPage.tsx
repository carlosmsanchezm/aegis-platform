import { Link as RouterLink } from 'react-router-dom';
import {
  Box,
  Button,
  Chip,
  Divider,
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
  HeaderLabel,
  Page,
} from '@backstage/core-components';

const useStyles = makeStyles(theme => ({
  pageContent: {
    paddingBottom: theme.spacing(6),
  },
  hero: {
    position: 'relative',
    padding: theme.spacing(4),
    borderRadius: 28,
    overflow: 'hidden',
    background:
      'linear-gradient(135deg, rgba(16,185,129,0.12), rgba(14,165,233,0.12))',
    border: '1px solid rgba(148, 163, 184, 0.24)',
    boxShadow: '0 28px 55px rgba(8, 15, 31, 0.45)',
  },
  heroHighlight: {
    color: theme.palette.primary.main,
  },
  heroSubtitle: {
    marginTop: theme.spacing(1.5),
    maxWidth: 640,
    color: theme.palette.text.secondary,
  },
  heroActions: {
    marginTop: theme.spacing(3),
    display: 'flex',
    gap: theme.spacing(2),
    flexWrap: 'wrap',
  },
  gradientOrb: {
    position: 'absolute',
    width: 320,
    height: 320,
    borderRadius: '50%',
    filter: 'blur(120px)',
    right: -120,
    top: -120,
    background:
      'radial-gradient(circle at center, rgba(56,189,248,0.55), transparent 70%)',
    opacity: 0.8,
  },
  metricCard: {
    padding: theme.spacing(3),
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    background:
      'linear-gradient(160deg, rgba(15,23,42,0.85) 0%, rgba(15,23,42,0.5) 100%)',
  },
  metricLabel: {
    color: theme.palette.text.secondary,
    textTransform: 'uppercase',
    letterSpacing: '0.12em',
    fontSize: '0.75rem',
    marginBottom: theme.spacing(1),
  },
  metricValue: {
    fontSize: '2.25rem',
    fontWeight: 600,
    letterSpacing: '-0.03em',
  },
  metricDelta: {
    marginTop: theme.spacing(1),
    color: theme.palette.success.main,
    fontWeight: 500,
  },
  metricProgress: {
    marginTop: theme.spacing(2),
    height: 8,
    borderRadius: 999,
    backgroundColor: 'rgba(148, 163, 184, 0.15)',
    '& .MuiLinearProgress-barColorPrimary': {
      borderRadius: 999,
      background:
        'linear-gradient(135deg, rgba(16,185,129,0.8), rgba(14,165,233,0.9))',
    },
  },
  panel: {
    padding: theme.spacing(3),
    height: '100%',
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(3),
  },
  panelHeader: {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: theme.spacing(2),
  },
  listItem: {
    padding: theme.spacing(1.5, 0),
    '&:not(:last-child)': {
      borderBottom: '1px solid rgba(148, 163, 184, 0.1)',
    },
  },
  faintAccent: {
    color: theme.palette.secondary.main,
  },
}));

const gpuMetrics = [
  {
    label: 'GPU Fleet Online',
    value: '128',
    delta: '+12 new nodes',
    progress: 72,
  },
  {
    label: 'Workspace Sessions',
    value: '342',
    delta: 'Live in last 24h',
    progress: 64,
  },
  {
    label: 'Policy Compliance',
    value: '99.2%',
    delta: 'Auto-remediated 4 drifts',
    progress: 92,
  },
];

const upcomingLaunches = [
  {
    title: 'Trident Recon / Mission Batch',
    subtitle: 'us-gov-west-2 · 64x H100',
    chip: 'Scheduled',
  },
  {
    title: 'Atlas Notebook Fleet Expansion',
    subtitle: 'Azure IL6 · 24x MI300X',
    chip: 'Queued',
  },
  {
    title: 'Sentinel FinOps Sync',
    subtitle: 'Cross-cloud cost recalc in flight',
    chip: 'Running',
  },
];

const securitySignals = [
  {
    title: 'Policy drift auto-corrected',
    description: 'IAM boundary tightened for workspace atlas-notebook-47',
    tone: 'Resolved',
  },
  {
    title: 'Elevated GPU spend projection',
    description: 'Projected +8% over baseline for NGA cluster aurora-east',
    tone: 'Review',
  },
  {
    title: 'KMS rotation completed',
    description: 'DoD cloud keyring rotated across IL4 tenants',
    tone: 'Healthy',
  },
];

export const AegisDashboardPage = () => {
  const classes = useStyles();

  return (
    <Page themeId="home">
      <Content className={classes.pageContent}>
        <ContentHeader title="ÆGIS Control Center">
          <HeaderLabel label="Posture" value="Live" />
          <HeaderLabel label="Clouds" value="AWS · Azure · GCP" />
          <HeaderLabel label="GPU Pools" value="H100 · A100 · MI300X" />
          <Button
            variant="contained"
            color="primary"
            component={RouterLink}
            to="/aegis/workspaces/create"
          >
            Launch Secure Workspace
          </Button>
        </ContentHeader>
        <Box px={4} pb={6}>
          <Paper className={classes.hero} elevation={0}>
            <div className={classes.gradientOrb} />
            <Typography variant="h2">
              Mission-grade multi-cloud GPUs, orchestrated with{' '}
              <span className={classes.heroHighlight}>zero drag</span>.
            </Typography>
            <Typography variant="body1" className={classes.heroSubtitle}>
              ÆGIS brokers GPU capacity, hardens workspaces, and fuses policy,
              identity, and telemetry so operators can move from intent to
              execution instantly.
            </Typography>
            <div className={classes.heroActions}>
              <Button
                variant="contained"
                color="primary"
                component={RouterLink}
                to="/aegis/telemetry"
              >
                View Telemetry Pulse
              </Button>
              <Button
                variant="outlined"
                color="default"
                component={RouterLink}
                to="/aegis/posture"
              >
                Review Live Posture
              </Button>
            </div>
          </Paper>
        </Box>
        <Box px={4}>
          <Grid container spacing={4}>
            {gpuMetrics.map(metric => (
              <Grid item xs={12} md={4} key={metric.label}>
                <Paper className={classes.metricCard} elevation={0}>
                  <Typography className={classes.metricLabel}>
                    {metric.label}
                  </Typography>
                  <div>
                    <Typography className={classes.metricValue}>
                      {metric.value}
                    </Typography>
                    <Typography className={classes.metricDelta}>
                      {metric.delta}
                    </Typography>
                  </div>
                  <LinearProgress
                    variant="determinate"
                    value={metric.progress}
                    className={classes.metricProgress}
                  />
                </Paper>
              </Grid>
            ))}
            <Grid item xs={12} md={6}>
              <Paper className={classes.panel} elevation={0}>
                <div className={classes.panelHeader}>
                  <Typography variant="h5">Launch Timeline</Typography>
                  <Chip label="Next 24 hours" color="primary" size="small" />
                </div>
                <Divider light />
                <List disablePadding>
                  {upcomingLaunches.map(item => (
                    <ListItem key={item.title} className={classes.listItem}>
                      <ListItemText
                        primary={item.title}
                        secondary={item.subtitle}
                      />
                      <Chip label={item.chip} variant="default" />
                    </ListItem>
                  ))}
                </List>
              </Paper>
            </Grid>
            <Grid item xs={12} md={6}>
              <Paper className={classes.panel} elevation={0}>
                <div className={classes.panelHeader}>
                  <Typography variant="h5">Signals</Typography>
                  <Chip label="Realtime" color="secondary" size="small" />
                </div>
                <Divider light />
                <List disablePadding>
                  {securitySignals.map(signal => (
                    <ListItem key={signal.title} className={classes.listItem}>
                      <ListItemText
                        primary={signal.title}
                        secondary={signal.description}
                      />
                      <Typography variant="body2" className={classes.faintAccent}>
                        {signal.tone}
                      </Typography>
                    </ListItem>
                  ))}
                </List>
              </Paper>
            </Grid>
          </Grid>
        </Box>
      </Content>
    </Page>
  );
};

export default AegisDashboardPage;
