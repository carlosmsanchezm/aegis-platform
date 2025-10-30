import { FC } from 'react';
import { Link as RouterLink } from 'react-router-dom';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  InfoCard,
} from '@backstage/core-components';
import { Box, Button, Grid, Typography } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';

const useStyles = makeStyles(theme => ({
  page: {
    position: 'relative',
    overflow: 'hidden',
    minHeight: '100vh',
  },
  backgroundMesh: {
    position: 'absolute',
    inset: '-20% -10% auto',
    height: '120%',
    background: `radial-gradient(circle at 20% 10%, ${alpha(
      theme.palette.primary.main,
      0.35,
    )} 0, transparent 60%), radial-gradient(circle at 80% 0%, ${alpha(
      theme.palette.secondary.main,
      0.28,
    )} 0, transparent 55%), radial-gradient(circle at 50% 65%, ${alpha(
      '#4b9eff',
      0.25,
    )} 0, transparent 60%)`,
    filter: 'blur(30px)',
    opacity: 0.8,
    pointerEvents: 'none',
    zIndex: 0,
  },
  header: {
    position: 'relative',
    zIndex: 1,
    paddingBottom: theme.spacing(6),
    borderBottom: `1px solid ${alpha(theme.palette.primary.main, 0.25)}`,
    background: `linear-gradient(140deg, ${alpha('#061129', 0.92)}, ${alpha(
      '#071633',
      0.88,
    )})`,
    boxShadow: '0 45px 120px rgba(16, 255, 219, 0.15)',
  },
  headerTitle: {
    fontSize: '3rem',
    letterSpacing: '0.32em',
    color: theme.palette.primary.light,
    textShadow: '0 0 32px rgba(83, 251, 224, 0.35)',
    marginBottom: theme.spacing(2),
  },
  headerSubtitle: {
    maxWidth: 740,
    color: alpha('#dfe9ff', 0.82),
    lineHeight: 1.7,
    fontFamily: `'IBM Plex Mono', monospace`,
  },
  headerActions: {
    marginTop: theme.spacing(4),
    display: 'flex',
    flexWrap: 'wrap',
    gap: theme.spacing(2),
  },
  content: {
    position: 'relative',
    zIndex: 1,
    marginTop: theme.spacing(-6),
  },
  contentHeader: {
    '& h2': {
      letterSpacing: '0.26em',
      color: theme.palette.primary.light,
      textTransform: 'uppercase',
    },
  },
  holoCard: {
    position: 'relative',
    overflow: 'hidden',
    background: `linear-gradient(160deg, ${alpha('#0d1f3b', 0.92)}, ${alpha(
      '#050b16',
      0.85,
    )})`,
    border: `1px solid ${alpha(theme.palette.primary.main, 0.35)}`,
    boxShadow: '0 40px 90px rgba(27, 255, 230, 0.14)',
    padding: theme.spacing(4),
    minHeight: 220,
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-between',
    transition: 'transform 320ms ease, box-shadow 320ms ease',
    '&:before': {
      content: '""',
      position: 'absolute',
      inset: 0,
      background: `linear-gradient(120deg, ${alpha(
        theme.palette.primary.main,
        0.12,
      )}, transparent 60%)`,
      opacity: 0,
      transition: 'opacity 320ms ease',
    },
    '&:hover': {
      transform: 'translateY(-6px)',
      boxShadow: '0 55px 120px rgba(79, 255, 232, 0.2)',
      '&:before': {
        opacity: 1,
      },
    },
  },
  stepBadge: {
    width: 36,
    height: 36,
    borderRadius: '50%',
    border: `2px solid ${alpha(theme.palette.primary.main, 0.8)}`,
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    fontFamily: `'IBM Plex Mono', monospace`,
    letterSpacing: '0.1em',
    color: theme.palette.primary.light,
    marginBottom: theme.spacing(3),
  },
  cardTitle: {
    letterSpacing: '0.2em',
    textTransform: 'uppercase',
    marginBottom: theme.spacing(2),
  },
  cardBody: {
    color: alpha('#dfe9ff', 0.85),
    lineHeight: 1.7,
  },
  telemetryCard: {
    padding: theme.spacing(4),
    border: `1px solid ${alpha('#4b9eff', 0.4)}`,
    background: `linear-gradient(160deg, ${alpha('#072140', 0.95)}, ${alpha(
      '#020810',
      0.9,
    )})`,
    boxShadow: '0 45px 120px rgba(75, 158, 255, 0.18)',
  },
  statValue: {
    fontSize: '2.5rem',
    fontWeight: 700,
    letterSpacing: '0.12em',
    color: theme.palette.primary.light,
    textShadow: '0 0 25px rgba(83, 251, 224, 0.4)',
  },
  statLabel: {
    marginTop: theme.spacing(1),
    color: alpha('#dfe9ff', 0.7),
    letterSpacing: '0.2em',
    textTransform: 'uppercase',
    fontSize: '0.75rem',
  },
  feedItem: {
    borderBottom: `1px solid ${alpha('#4b9eff', 0.2)}`,
    paddingBottom: theme.spacing(2),
    marginBottom: theme.spacing(2),
  },
  feedTitle: {
    fontFamily: `'IBM Plex Mono', monospace`,
    letterSpacing: '0.12em',
    color: theme.palette.primary.light,
  },
  feedDetail: {
    color: alpha('#dfe9ff', 0.75),
    marginTop: theme.spacing(1),
    fontSize: '0.9rem',
  },
}));

export const AegisHomePage: FC = () => {
  const classes = useStyles();

  return (
    <Page themeId="home" className={classes.page}>
      <div className={classes.backgroundMesh} />
      <Header
        title="ÆGIS COMMAND"
        subtitle="DoD/IC-grade multi-cloud GPU control plane — orchestrate mission-critical workloads, orchestrate policy and compliance overlays, and broker the fastest silicon on any cloud."
        className={classes.header}
      >
        <Typography variant="h1" className={classes.headerTitle}>
          ÆGIS COMMAND
        </Typography>
        <Typography variant="subtitle1" className={classes.headerSubtitle}>
          Deploy secure workspaces, broker GPUs across sovereign clouds, and maintain constant telemetry of mission workload posture in a cockpit built for critical infrastructure operators.
        </Typography>
        <Box className={classes.headerActions}>
          <Button
            component={RouterLink}
            color="primary"
            variant="contained"
            to="/aegis/workspaces/create"
          >
            Launch Mission Workspace
          </Button>
          <Button
            component={RouterLink}
            color="primary"
            variant="outlined"
            to="/aegis/workloads"
          >
            View Active Operations
          </Button>
          <Button
            component={RouterLink}
            color="secondary"
            variant="text"
            to="/docs"
          >
            Open Field Manual
          </Button>
        </Box>
      </Header>
      <Content className={classes.content}>
        <ContentHeader title="Mission Controls" className={classes.contentHeader} />
        <Grid container spacing={4}>
          {["Initiate", "Configure", "Launch"].map((label, index) => (
            <Grid item xs={12} md={4} key={label}>
              <InfoCard className={classes.holoCard} title={label}>
                <Box className={classes.stepBadge}>0{index + 1}</Box>
                <Typography variant="h5" className={classes.cardTitle}>
                  {label === 'Initiate'
                    ? 'Select Mission Profile'
                    : label === 'Configure'
                    ? 'Tune GPU + Workspace'
                    : 'Deploy & Monitor'}
                </Typography>
                <Typography variant="body2" className={classes.cardBody}>
                  {label === 'Initiate'
                    ? 'Choose the mission-aligned template spanning VS Code, Jupyter, or tactical CLI surfaces. Align queues, sovereign regions, and target enclave boundaries.'
                    : label === 'Configure'
                    ? 'Dial-in GPU flavor, MIG partition, storage mounts, and zero-trust identity overlays. Preview FinOps signals and enforce policy gates before launch.'
                    : 'Execute launch, stream event telemetry, and auto-route compliance and audit webhooks. Maintain oversight with live mission status and termination controls.'}
                </Typography>
              </InfoCard>
            </Grid>
          ))}
        </Grid>

        <ContentHeader title="Operational Telemetry" className={classes.contentHeader} />
        <Grid container spacing={4}>
          <Grid item xs={12} md={6}>
            <InfoCard className={classes.telemetryCard} title="Live Posture">
              <Grid container spacing={4}>
                <Grid item xs={6}>
                  <Typography className={classes.statValue}>12</Typography>
                  <Typography className={classes.statLabel}>
                    ACTIVE MISSIONS
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography className={classes.statValue}>38</Typography>
                  <Typography className={classes.statLabel}>
                    GPUs LOCKED-IN
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography className={classes.statValue}>4</Typography>
                  <Typography className={classes.statLabel}>
                    CSP THEATERS
                  </Typography>
                </Grid>
                <Grid item xs={6}>
                  <Typography className={classes.statValue}>92%</Typography>
                  <Typography className={classes.statLabel}>
                    COMPLIANCE GREEN
                  </Typography>
                </Grid>
              </Grid>
            </InfoCard>
          </Grid>
          <Grid item xs={12} md={6}>
            <InfoCard className={classes.telemetryCard} title="Signals Feed">
              {[1, 2, 3].map(item => (
                <Box key={item} className={classes.feedItem}>
                  <Typography variant="overline" className={classes.feedTitle}>
                    {item === 1
                      ? 'ATLAS GPU MARKETPLACE'
                      : item === 2
                      ? 'ZERO-TRUST NETWORK OVERLAY'
                      : 'FINOPS COMPLIANCE'}
                  </Typography>
                  <Typography variant="body2" className={classes.feedDetail}>
                    {item === 1
                      ? 'A100 multi-cloud block secured — auto-allocating MIG slices to priority missions.'
                      : item === 2
                      ? 'New workspace dialed into classified enclave with ephemeral certificates rotated.'
                      : 'Budget runway recalibrated with 17% savings vs. baseline across GPU fleets.'}
                  </Typography>
                </Box>
              ))}
              <Box display="flex" justifyContent="flex-end">
                <Button component={RouterLink} to="/notifications" color="primary">
                  View All Signals
                </Button>
              </Box>
            </InfoCard>
          </Grid>
        </Grid>
      </Content>
    </Page>
  );
};
