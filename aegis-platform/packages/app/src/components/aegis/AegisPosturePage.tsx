import {
  Avatar,
  Box,
  Chip,
  Grid,
  List,
  ListItem,
  ListItemAvatar,
  ListItemText,
  makeStyles,
  Paper,
  Typography,
} from '@material-ui/core';
import { Content, ContentHeader, Page } from '@backstage/core-components';
import Alert from '@material-ui/lab/Alert';
import AlertTitle from '@material-ui/lab/AlertTitle';
import SecurityIcon from '@material-ui/icons/Security';
import CheckCircleIcon from '@material-ui/icons/CheckCircle';
import ErrorOutlineIcon from '@material-ui/icons/ErrorOutline';
import TimelineIcon from '@material-ui/icons/Timeline';
import { alpha } from '@material-ui/core/styles/colorManipulator';

const useStyles = makeStyles(theme => {
  const paletteMode = (theme.palette as any)?.mode ?? theme.palette.type;
  const isDark = paletteMode === 'dark';
  const panelBorder = alpha(theme.palette.text.primary, isDark ? 0.22 : 0.12);
  const panelBackground = isDark
    ? 'linear-gradient(150deg, rgba(255,255,255,0.08) 0%, rgba(0,0,0,0.55) 100%)'
    : 'linear-gradient(150deg, rgba(0,0,0,0.08) 0%, rgba(0,0,0,0.02) 100%)';
  const listBorder = alpha(theme.palette.text.primary, isDark ? 0.2 : 0.1);
  const listBackground = isDark
    ? 'rgba(255,255,255,0.05)'
    : 'rgba(0,0,0,0.04)';

  return {
    pageContent: {
      paddingBottom: theme.spacing(6),
    },
    card: {
      padding: theme.spacing(3),
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(2.5),
      border: `1px solid ${panelBorder}`,
      background: panelBackground,
    },
    header: {
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      gap: theme.spacing(2),
    },
    badgeRow: {
      display: 'flex',
      gap: theme.spacing(1.5),
      flexWrap: 'wrap',
    },
    avatar: {
      background: isDark
        ? 'linear-gradient(135deg, rgba(255,255,255,0.8), rgba(150,150,150,0.85))'
        : 'linear-gradient(135deg, rgba(0,0,0,0.75), rgba(110,110,110,0.8))',
    },
    riskList: {
      '& .MuiListItem-root': {
        borderRadius: 16,
        padding: theme.spacing(2),
        border: `1px solid ${listBorder}`,
        backgroundColor: listBackground,
        marginBottom: theme.spacing(1.5),
      },
    },
    subtle: {
      color: theme.palette.text.secondary,
    },
  };
});

const postureHighlights = [
  {
    title: 'Workspace posture',
    description: 'All mission notebooks isolated with continuous scanning',
    status: 'Hardened',
  },
  {
    title: 'Access boundaries',
    description: 'Air-gapped GPU fleets pinned with policy as code',
    status: 'Locked',
  },
  {
    title: 'Network overlays',
    description: 'Dynamic microsegmentation across clouds · zero lateral drift',
    status: 'Steady',
  },
];

const postureFindings = [
  {
    level: 'High',
    title: 'Cluster aurora-east drift corrected',
    detail: 'IAM identity boundary rolled back to approved baseline in 43s',
    icon: <SecurityIcon />,
  },
  {
    level: 'Medium',
    title: 'Notebook workspace audit trail synced',
    detail: '24 hr log export delivered to SIPR analytic vault',
    icon: <TimelineIcon />,
  },
  {
    level: 'Low',
    title: 'FinOps guardrail reminder',
    detail: 'Budget nearing 80% on azure-il6 scope, review recommended',
    icon: <ErrorOutlineIcon />,
  },
];

export const AegisPosturePage = () => {
  const classes = useStyles();

  return (
    <Page themeId="service">
      <Content className={classes.pageContent}>
        <ContentHeader title="Live Posture">
          <Chip label="Continuous" color="primary" />
          <Chip label="Last drift 43s ago" variant="outlined" />
        </ContentHeader>
        <Box px={4} pb={6}>
          <Grid container spacing={4}>
            <Grid item xs={12} md={6}>
              <Paper className={classes.card} elevation={0}>
                <div className={classes.header}>
                  <Typography variant="h5">Mission Shield</Typography>
                  <Chip label="Green" color="primary" />
                </div>
                <Typography variant="body1" className={classes.subtle}>
                  Policy, identity, and runtime telemetry fused into an adaptive
                  control plane. ÆGIS resolves drift instantly and pushes posture
                  attestations to your watchfloor in real time.
                </Typography>
                <div className={classes.badgeRow}>
                  <Chip label="Zero Trust" color="secondary" />
                  <Chip label="JIT Access" variant="default" />
                  <Chip label="Continuous ATO" variant="default" />
                </div>
                <List disablePadding className={classes.riskList}>
                  {postureHighlights.map(highlight => (
                    <ListItem key={highlight.title}>
                      <ListItemAvatar>
                        <Avatar className={classes.avatar}>
                          <CheckCircleIcon />
                        </Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={highlight.title}
                        secondary={highlight.description}
                      />
                      <Chip label={highlight.status} variant="default" />
                    </ListItem>
                  ))}
                </List>
              </Paper>
            </Grid>
            <Grid item xs={12} md={6}>
              <Paper className={classes.card} elevation={0}>
                <div className={classes.header}>
                  <Typography variant="h5">Live Findings</Typography>
                  <Chip label="Auto-remediated" color="secondary" />
                </div>
                <Typography variant="body1" className={classes.subtle}>
                  Priority signals that ÆGIS is tracking and resolving across the
                  fleet. Everything is contextualized with mission tags and cost
                  implications.
                </Typography>
                <List disablePadding className={classes.riskList}>
                  {postureFindings.map(finding => (
                    <ListItem key={finding.title}>
                      <ListItemAvatar>
                        <Avatar className={classes.avatar}>{finding.icon}</Avatar>
                      </ListItemAvatar>
                      <ListItemText
                        primary={finding.title}
                        secondary={finding.detail}
                      />
                      <Chip label={finding.level} variant="outlined" />
                    </ListItem>
                  ))}
                </List>
                <Alert severity="info">
                  <AlertTitle>Compliance streaming</AlertTitle>
                  Continuous RMF, NIST 800-53, and JSIG controls validated and
                  exported to your compliance data lake.
                </Alert>
              </Paper>
            </Grid>
          </Grid>
        </Box>
      </Content>
    </Page>
  );
};

export default AegisPosturePage;
