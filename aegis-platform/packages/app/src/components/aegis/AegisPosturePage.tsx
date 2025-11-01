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

const useStyles = makeStyles(theme => {
  const isDark = theme.palette.type === 'dark';
  const cardBorder = isDark
    ? '1px solid rgba(148, 163, 184, 0.2)'
    : '1px solid rgba(15, 23, 42, 0.08)';
  const cardBackground = isDark
    ? 'linear-gradient(150deg, rgba(15,23,42,0.9) 0%, rgba(15,23,42,0.55) 100%)'
    : 'linear-gradient(150deg, rgba(246,248,252,0.95) 0%, rgba(229,235,247,0.88) 100%)';
  const listBackground = isDark
    ? 'rgba(15, 23, 42, 0.55)'
    : 'rgba(15, 23, 42, 0.04)';
  const listBorder = isDark
    ? '1px solid rgba(148, 163, 184, 0.14)'
    : '1px solid rgba(15, 23, 42, 0.08)';
  const avatarBackground = isDark
    ? 'linear-gradient(135deg, rgba(16,185,129,0.85), rgba(14,165,233,0.85))'
    : 'linear-gradient(135deg, rgba(99,102,241,0.8), rgba(14,165,233,0.8))';

  return {
    pageContent: {
      paddingBottom: theme.spacing(6),
    },
    card: {
      padding: theme.spacing(3),
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(2.5),
      border: cardBorder,
      background: cardBackground,
      borderRadius: 24,
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
      background: avatarBackground,
    },
    riskList: {
      '& .MuiListItem-root': {
        borderRadius: 16,
        padding: theme.spacing(2),
        border: listBorder,
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
