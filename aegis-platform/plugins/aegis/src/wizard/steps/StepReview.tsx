import { Card, CardContent, Grid, Typography, Button } from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import {
  ClusterOption,
  WizardFormState,
  WorkspaceFlavor,
  WorkspaceTemplate,
  WizardStep,
} from '../types';

const useStyles = makeStyles(theme => ({
  section: {
    marginBottom: theme.spacing(3),
  },
  row: {
    marginBottom: theme.spacing(1.5),
  },
  label: {
    fontWeight: 600,
    marginBottom: theme.spacing(0.5),
  },
  value: {
    color: theme.palette.text.secondary,
    whiteSpace: 'pre-wrap',
    wordBreak: 'break-word',
  },
  actions: {
    display: 'flex',
    justifyContent: 'flex-end',
    gap: theme.spacing(1),
  },
}));

export type StepReviewProps = {
  form: WizardFormState;
  selectedTemplate?: WorkspaceTemplate;
  selectedFlavor?: WorkspaceFlavor;
  selectedCluster?: ClusterOption;
  onEdit: (step: WizardStep) => void;
};

const formatPorts = (ports: string) =>
  ports
    .split(/[,\s]+/)
    .map(token => token.trim())
    .filter(Boolean)
    .join(', ');

export const StepReview = ({
  form,
  selectedTemplate,
  selectedFlavor,
  selectedCluster,
  onEdit,
}: StepReviewProps) => {
  const classes = useStyles();

  return (
    <Grid container spacing={3}>
      <Grid item xs={12} md={6}>
        <Card className={classes.section} elevation={0}>
          <CardContent>
            <div className={classes.actions}>
              <Button color="primary" size="small" onClick={() => onEdit(0)}>
                Edit templates
              </Button>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Template
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {selectedTemplate?.name ?? '—'}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Compute profile
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {selectedFlavor ? selectedFlavor.name : '—'}
                {selectedFlavor?.resources ? ` • ${selectedFlavor.resources}` : ''}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Cluster
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {selectedCluster?.name ?? (form.clusterId || '—')}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Queue
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {form.queue || 'Default'}
              </Typography>
            </div>
          </CardContent>
        </Card>
      </Grid>
      <Grid item xs={12} md={6}>
        <Card className={classes.section} elevation={0}>
          <CardContent>
            <div className={classes.actions}>
              <Button color="primary" size="small" onClick={() => onEdit(1)}>
                Edit settings
              </Button>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Project / Workspace ID
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {form.projectId} · {form.workloadId}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Container image
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {form.image}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Ports
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {formatPorts(form.ports) || 'Default'}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Environment variables
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {form.env.trim() ? form.env : 'Inherited defaults'}
              </Typography>
            </div>
            <div className={classes.row}>
              <Typography variant="subtitle2" className={classes.label}>
                Storage & TTL
              </Typography>
              <Typography variant="body2" className={classes.value}>
                {form.storageGiB} GiB •{' '}
                {form.autoShutdown
                  ? `${form.runtimeHours} hour${form.runtimeHours === 1 ? '' : 's'}`
                  : 'Auto shutdown disabled'}
              </Typography>
            </div>
          </CardContent>
        </Card>
      </Grid>
    </Grid>
  );
};
