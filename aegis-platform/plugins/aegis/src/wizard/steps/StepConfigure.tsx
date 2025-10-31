import {
  Box,
  Collapse,
  FormControlLabel,
  Grid,
  MenuItem,
  Switch,
  TextField,
  Typography,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import StorageIcon from '@material-ui/icons/Storage';
import ScheduleIcon from '@material-ui/icons/Schedule';
import MemoryIcon from '@material-ui/icons/Memory';
import FlashOnIcon from '@material-ui/icons/FlashOn';
import { TemplateCard } from '../components/TemplateCard';
import {
  ClusterOption,
  FlavorId,
  WizardFormState,
  WorkspaceFlavor,
  WorkspaceTemplate,
} from '../types';

const useStyles = makeStyles(theme => ({
  section: {
    marginBottom: theme.spacing(4),
  },
  sectionHeader: {
    fontWeight: 600,
    marginBottom: theme.spacing(2),
  },
  flavorGrid: {
    marginTop: theme.spacing(1),
  },
  helperRow: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(1.5),
    color: theme.palette.text.secondary,
  },
}));

export type StepConfigureErrors = {
  projectId?: string;
  workloadId?: string;
  image?: string;
  ports?: string;
  env?: string;
};

export type StepConfigureProps = {
  form: WizardFormState;
  selectedTemplate?: WorkspaceTemplate;
  flavors: WorkspaceFlavor[];
  clusterOptions: ClusterOption[];
  errors: StepConfigureErrors;
  advancedOpen: boolean;
  forceAdvancedOpen: boolean;
  onAdvancedToggle: () => void;
  onFieldChange: <K extends keyof WizardFormState>(field: K, value: WizardFormState[K]) => void;
  onFlavorSelect: (id: FlavorId) => void;
};

export const StepConfigure = ({
  form,
  selectedTemplate,
  flavors,
  clusterOptions,
  errors,
  advancedOpen,
  forceAdvancedOpen,
  onAdvancedToggle,
  onFieldChange,
  onFlavorSelect,
}: StepConfigureProps) => {
  const classes = useStyles();
  const selectedFlavor = flavors.find(flavor => flavor.id === form.flavorId);

  return (
    <div>
      <Box className={classes.section}>
        <Typography variant="h6" className={classes.sectionHeader}>
          Workspace basics
        </Typography>
        <Grid container spacing={3}>
          <Grid item xs={12} md={4}>
            <TextField
              label="Project"
              value={form.projectId}
              onChange={event => onFieldChange('projectId', event.target.value)}
              required
              fullWidth
              error={Boolean(errors.projectId)}
              helperText={errors.projectId ?? 'Project that owns this workspace'}
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              label="Workspace ID"
              value={form.workloadId}
              onChange={event => onFieldChange('workloadId', event.target.value)}
              required
              fullWidth
              error={Boolean(errors.workloadId)}
              helperText={
                errors.workloadId ?? 'Letters, numbers, and hyphen only'
              }
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              select
              label="Cluster"
              value={form.clusterId}
              onChange={event => onFieldChange('clusterId', event.target.value)}
              fullWidth
              helperText="Where the workspace will be scheduled"
            >
              {clusterOptions.map(option => (
                <MenuItem key={option.id} value={option.id}>
                  {option.name}
                </MenuItem>
              ))}
            </TextField>
          </Grid>
        </Grid>
      </Box>

      <Box className={classes.section}>
        <Typography variant="h6" className={classes.sectionHeader}>
          Compute profile
        </Typography>
        <Grid
          container
          spacing={3}
          className={classes.flavorGrid}
          role="radiogroup"
          aria-label="Compute flavors"
        >
          {flavors.map(flavor => (
            <Grid item xs={12} md={6} key={flavor.id} role="presentation">
              <TemplateCard
                title={flavor.name}
                description={flavor.description}
                icon={
                  flavor.category === 'gpu' ? (
                    <FlashOnIcon fontSize="large" />
                  ) : (
                    <MemoryIcon fontSize="large" />
                  )
                }
                meta={flavor.resources}
                selected={form.flavorId === flavor.id}
                layout="compact"
                onSelect={() => onFlavorSelect(flavor.id)}
              />
            </Grid>
          ))}
        </Grid>
        {selectedFlavor?.gpu ? (
          <Box mt={3} className={classes.helperRow}>
            <MemoryIcon fontSize="small" />
            <Typography variant="body2">
              {selectedFlavor.gpu.count}× {selectedFlavor.gpu.model}
              {selectedFlavor.gpu.memory ? ` • ${selectedFlavor.gpu.memory}` : ''}
            </Typography>
          </Box>
        ) : null}
      </Box>

      <Box className={classes.section}>
        <Typography variant="h6" className={classes.sectionHeader}>
          Runtime settings
        </Typography>
        <Grid container spacing={3}>
          <Grid item xs={12} md={4}>
            <TextField
              label="Queue"
              value={form.queue}
              onChange={event => onFieldChange('queue', event.target.value)}
              fullWidth
              helperText="Optional scheduling queue"
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              type="number"
              label="Storage (GiB)"
              value={form.storageGiB}
              inputProps={{ min: 10, step: 10 }}
              onChange={event =>
                onFieldChange('storageGiB', Number(event.target.value) || 0)
              }
              fullWidth
              helperText="Persistent volume size"
            />
          </Grid>
          <Grid item xs={12} md={4}>
            <TextField
              type="number"
              label="Auto shutdown (hours)"
              value={form.runtimeHours}
              inputProps={{ min: 1, max: 72 }}
              onChange={event =>
                onFieldChange('runtimeHours', Number(event.target.value) || 1)
              }
              fullWidth
              disabled={!form.autoShutdown}
              helperText="Workspace lifetime before termination"
            />
          </Grid>
          <Grid item xs={12}>
            <FormControlLabel
              control={
                <Switch
                  color="primary"
                  checked={form.autoShutdown}
                  onChange={(_, checked) => onFieldChange('autoShutdown', checked)}
                />
              }
              label="Automatically shut down when the TTL expires"
            />
          </Grid>
        </Grid>
      </Box>

      <Box className={classes.section}>
        <FormControlLabel
          control={
            <Switch
              color="primary"
              checked={forceAdvancedOpen || advancedOpen}
              onChange={onAdvancedToggle}
              disabled={forceAdvancedOpen}
            />
          }
          label="Advanced options"
        />
        <Collapse in={forceAdvancedOpen || advancedOpen}>
          <Box mt={2}>
            <Grid container spacing={3}>
              <Grid item xs={12} md={6}>
                <TextField
                  label="Container image"
                  value={form.image}
                  onChange={event => onFieldChange('image', event.target.value)}
                  required
                  fullWidth
                  error={Boolean(errors.image)}
                  helperText={errors.image ?? 'Fully qualified container image'}
                />
              </Grid>
              <Grid item xs={12} md={6}>
                <TextField
                  label="Exposed ports"
                  value={form.ports}
                  onChange={event => onFieldChange('ports', event.target.value)}
                  fullWidth
                  error={Boolean(errors.ports)}
                  helperText={errors.ports ?? 'Comma or space separated list'}
                />
              </Grid>
              <Grid item xs={12}>
                <TextField
                  label="Environment variables"
                  value={form.env}
                  onChange={event => onFieldChange('env', event.target.value)}
                  fullWidth
                  multiline
                  minRows={3}
                  error={Boolean(errors.env)}
                  helperText={
                    errors.env ?? 'Optional KEY=VALUE pairs, one per line'
                  }
                />
              </Grid>
            </Grid>
          </Box>
        </Collapse>
      </Box>

      <Box className={classes.section}>
        <Typography variant="body2" className={classes.helperRow}>
          <StorageIcon fontSize="small" />
          Storage and runtime controls are applied when the workspace is provisioned.
        </Typography>
        <Typography variant="body2" className={classes.helperRow}>
          <ScheduleIcon fontSize="small" />
          You can adjust TTL and storage later from the Workspaces tab.
        </Typography>
      </Box>
    </div>
  );
};
