import {
  Box,
  Chip,
  makeStyles,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableRow,
  Typography,
} from '@material-ui/core';
import {
  Content,
  ContentHeader,
  Page,
} from '@backstage/core-components';

const useStyles = makeStyles(theme => {
  const isDark = theme.palette.type === 'dark';
  const rowBackground = isDark
    ? 'rgba(15, 23, 42, 0.6)'
    : 'rgba(15, 23, 42, 0.06)';
  const rowBorder = isDark
    ? '1px solid rgba(148, 163, 184, 0.14)'
    : '1px solid rgba(15, 23, 42, 0.08)';
  const summaryBorder = isDark
    ? '1px solid rgba(148, 163, 184, 0.2)'
    : '1px solid rgba(15, 23, 42, 0.1)';
  const summaryBackground = isDark
    ? 'rgba(15, 23, 42, 0.75)'
    : 'linear-gradient(135deg, rgba(246, 248, 252, 0.96), rgba(231, 235, 247, 0.9))';

  return {
    pageContent: {
      paddingBottom: theme.spacing(6),
    },
    tableWrapper: {
      padding: theme.spacing(4),
      backgroundColor: isDark ? 'rgba(9, 13, 23, 0.88)' : '#FFFFFF',
      borderRadius: 24,
      border: rowBorder,
    },
    table: {
      borderCollapse: 'separate',
      borderSpacing: '0 12px',
      '& th': {
        color: theme.palette.text.secondary,
        fontWeight: 500,
        textTransform: 'uppercase',
        fontSize: '0.75rem',
        letterSpacing: '0.14em',
        borderBottom: 'none',
      },
      '& td': {
        borderBottom: 'none',
        backgroundColor: rowBackground,
        border: rowBorder,
        borderRight: 'none',
        borderLeft: 'none',
        padding: theme.spacing(2),
      },
      '& tr': {
        borderRadius: 20,
      },
    },
    firstCell: {
      borderTopLeftRadius: 20,
      borderBottomLeftRadius: 20,
      borderLeft: rowBorder,
    },
    lastCell: {
      borderTopRightRadius: 20,
      borderBottomRightRadius: 20,
      borderRight: rowBorder,
    },
    summaryCard: {
      padding: theme.spacing(3),
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'space-between',
      border: summaryBorder,
      marginBottom: theme.spacing(3),
      background: summaryBackground,
    },
    summaryLabel: {
      color: theme.palette.text.secondary,
      textTransform: 'uppercase',
      letterSpacing: '0.12em',
      fontSize: '0.75rem',
    },
    summaryValue: {
      fontSize: '2.1rem',
      fontWeight: 600,
      letterSpacing: '-0.03em',
    },
  };
});

const clusters = [
  {
    name: 'aurora-east',
    cloud: 'AWS GovCloud',
    gpu: '64x NVIDIA H100',
    posture: 'Hardened',
    cost: '$12.4k / day',
    latency: '87 ms',
  },
  {
    name: 'atlas-mi300x',
    cloud: 'Azure IL6',
    gpu: '48x AMD MI300X',
    posture: 'Steady',
    cost: '$9.8k / day',
    latency: '103 ms',
  },
  {
    name: 'neptune-a100',
    cloud: 'GCP Secure',
    gpu: '40x NVIDIA A100',
    posture: 'Investigate',
    cost: '$7.1k / day',
    latency: '95 ms',
  },
  {
    name: 'sentinel-edge',
    cloud: 'Edge Classified',
    gpu: '24x NVIDIA L40S',
    posture: 'Hardened',
    cost: '$4.6k / day',
    latency: '42 ms',
  },
];

export const AegisClustersPage = () => {
  const classes = useStyles();

  return (
    <Page themeId="apis">
      <Content className={classes.pageContent}>
        <ContentHeader title="Cluster Atlas">
          <Chip label="Synced" color="primary" />
          <Chip label="4 clouds" variant="outlined" />
        </ContentHeader>
        <Box px={4} pb={6}>
          <Paper className={classes.summaryCard} elevation={0}>
            <div>
              <Typography className={classes.summaryLabel}>Total GPUs</Typography>
              <Typography className={classes.summaryValue}>176</Typography>
            </div>
            <div>
              <Typography className={classes.summaryLabel}>Aggregate posture</Typography>
              <Typography variant="h6">Mission-ready · 99.2% compliant</Typography>
            </div>
            <div>
              <Typography className={classes.summaryLabel}>Spend velocity</Typography>
              <Typography variant="h6">-4.2% vs last 24h</Typography>
            </div>
          </Paper>
          <Paper className={classes.tableWrapper} elevation={0}>
            <Table className={classes.table}>
              <TableHead>
                <TableRow>
                  <TableCell>Cluster</TableCell>
                  <TableCell>Cloud</TableCell>
                  <TableCell>GPU Profile</TableCell>
                  <TableCell>Posture</TableCell>
                  <TableCell>Run Rate</TableCell>
                  <TableCell>Latency</TableCell>
                </TableRow>
              </TableHead>
              <TableBody>
                {clusters.map(cluster => (
                  <TableRow key={cluster.name}>
                    <TableCell className={classes.firstCell}>
                      <Typography variant="subtitle1">{cluster.name}</Typography>
                      <Typography variant="body2" color="textSecondary">
                        Mission tag · {cluster.cloud}
                      </Typography>
                    </TableCell>
                    <TableCell>{cluster.cloud}</TableCell>
                    <TableCell>{cluster.gpu}</TableCell>
                    <TableCell>
                      <Chip
                        label={cluster.posture}
                        color={cluster.posture === 'Investigate' ? 'secondary' : 'primary'}
                        variant={cluster.posture === 'Hardened' ? 'default' : 'outlined'}
                      />
                    </TableCell>
                    <TableCell>{cluster.cost}</TableCell>
                    <TableCell className={classes.lastCell}>{cluster.latency}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </Paper>
        </Box>
      </Content>
    </Page>
  );
};

export default AegisClustersPage;
