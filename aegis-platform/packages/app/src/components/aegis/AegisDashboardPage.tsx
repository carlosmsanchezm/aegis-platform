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
import { alpha } from '@material-ui/core/styles/colorManipulator';

const useStyles = makeStyles(theme => {
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const heroBorder = alpha(theme.palette.text.primary, isDark ? 0.24 : 0.12);
  const panelBorder = alpha(theme.palette.text.primary, isDark ? 0.2 : 0.1);
  const panelBackground = isDark
    ? 'linear-gradient(160deg, rgba(255,255,255,0.08) 0%, rgba(0,0,0,0.55) 100%)'
    : 'linear-gradient(160deg, rgba(0,0,0,0.08) 0%, rgba(0,0,0,0.02) 100%)';
  const highlightTone = alpha(theme.palette.text.primary, isDark ? 0.8 : 0.65);
  const progressTrack = alpha(theme.palette.text.secondary, isDark ? 0.28 : 0.12);
  const progressFill = isDark
    ? 'linear-gradient(135deg, rgba(255,255,255,0.72), rgba(206,206,206,0.9))'
    : 'linear-gradient(135deg, rgba(0,0,0,0.65), rgba(120,120,120,0.78))';

  return {
    pageContent: {
      paddingBottom: theme.spacing(6),
    },
    hero: {
      position: 'relative',
      padding: theme.spacing(4),
      borderRadius: 28,
      overflow: 'hidden',
      background: isDark
        ? 'linear-gradient(150deg, rgba(255,255,255,0.1), rgba(0,0,0,0.58))'
        : 'linear-gradient(150deg, rgba(0,0,0,0.08), rgba(0,0,0,0.02))',
      border: `1px solid ${heroBorder}`,
      boxShadow: isDark
        ? '0 28px 55px rgba(0,0,0,0.5)'
        : '0 28px 55px rgba(0,0,0,0.12)',
    },
    heroHighlight: {
      color: highlightTone,
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
      background: isDark
        ? 'radial-gradient(circle at center, rgba(255,255,255,0.22), transparent 70%)'
        : 'radial-gradient(circle at center, rgba(0,0,0,0.15), transparent 70%)',
      opacity: 0.8,
    },
    metricCard: {
      padding: theme.spacing(3),
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-between',
      background: panelBackground,
      border: `1px solid ${panelBorder}`,
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
      color: highlightTone,
      fontWeight: 500,
    },
    metricProgress: {
      marginTop: theme.spacing(2),
      height: 8,
      borderRadius: 999,
      backgroundColor: progressTrack,
      '& .MuiLinearProgress-barColorPrimary': {
        borderRadius: 999,
        background: progressFill,
      },
    },
    panel: {
      padding: theme.spacing(3),
      height: '100%',
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(3),
      background: panelBackground,
      border: `1px solid ${panelBorder}`,
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
        borderBottom: `1px solid ${alpha(theme.palette.text.primary, isDark ? 0.18 : 0.08)}`,
      },
    },
    faintAccent: {
      color: highlightTone,
    },
  };
});

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
