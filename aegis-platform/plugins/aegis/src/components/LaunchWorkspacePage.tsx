import { ChangeEvent, FC, FormEvent, useMemo, useState } from 'react';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
} from '@backstage/core-components';
import {
  alertApiRef,
  discoveryApiRef,
  fetchApiRef,
  identityApiRef,
  useApi,
  useRouteRef,
} from '@backstage/core-plugin-api';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Button,
  Card,
  CardActionArea,
  CardContent,
  Collapse,
  FormControlLabel,
  Grid,
  Step,
  StepLabel,
  Stepper,
  Switch,
  TextField,
  Typography,
  Chip,
} from '@material-ui/core';
import { makeStyles } from '@material-ui/core/styles';
import { alpha } from '@material-ui/core/styles/colorManipulator';
import clsx from 'clsx';
import LaptopMacIcon from '@material-ui/icons/LaptopMac';
import ScatterPlotIcon from '@material-ui/icons/ScatterPlot';
import TerminalIcon from '@material-ui/icons/Terminal';
import CodeIcon from '@material-ui/icons/Code';
import StorageIcon from '@material-ui/icons/Storage';
import FunctionsIcon from '@material-ui/icons/Functions';
import BuildIcon from '@material-ui/icons/Build';
import LayersIcon from '@material-ui/icons/Layers';
import type { SvgIconComponent } from '@material-ui/icons';
import { SubmitWorkspaceRequest, submitWorkspace } from '../api/aegisClient';
import { parseEnvInput, parsePortsInput } from './workspaceFormUtils';
import { workloadsRouteRef } from '../routes';

type WorkspaceTypeId = 'vscode' | 'jupyter' | 'cli';

type WorkspaceTypeOption = {
  id: WorkspaceTypeId;
  title: string;
  description: string;
};

type TemplateOption = {
  id: string;
  title: string;
  description: string;
  workspaceTypes: WorkspaceTypeId[];
  defaults: {
    flavor?: string;
    image?: string;
    queue?: string;
    ports?: number[];
    env?: Record<string, string>;
  };
  autoShowAdvanced?: boolean;
};

type FlavorOption = {
  id: string;
  title: string;
  description: string;
  flavor: string;
  resources: string;
};

const useStyles = makeStyles(theme => {
  const isDark = theme.palette.type === 'dark';
  const border = `1px solid ${theme.palette.divider}`;

  return {
    content: {
      maxWidth: 1280,
      margin: '0 auto',
      width: '100%',
    },
    form: {
      marginTop: theme.spacing(2),
    },
    stepper: {
      marginBottom: theme.spacing(3),
    },
    section: {
      marginBottom: theme.spacing(4),
    },
    sectionTitle: {
      marginBottom: theme.spacing(0.5),
      fontWeight: 600,
    },
    sectionSubtitle: {
      color: theme.palette.text.secondary,
      marginBottom: theme.spacing(2),
    },
    lead: {
      color: theme.palette.text.secondary,
      maxWidth: 640,
    },
    selectionGrid: {
      marginTop: theme.spacing(1),
    },
    selectionCard: {
      borderRadius: 'var(--aegis-card-radius)',
      border,
      backgroundColor: theme.palette.background.paper,
      height: '100%',
      transition:
        'border-color 200ms ease, box-shadow 200ms ease, transform 200ms ease',
    },
    selectionCardActive: {
      borderColor: theme.palette.primary.main,
      boxShadow: `0 0 0 1px ${alpha(theme.palette.primary.main, 0.35)}`,
      transform: 'translateY(-2px)',
    },
    cardAction: {
      height: '100%',
      padding: theme.spacing(2.5, 2.5, 2.75),
      display: 'flex',
      alignItems: 'stretch',
    },
    cardContent: {
      padding: 0,
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'flex-start',
      gap: theme.spacing(1.25),
      width: '100%',
    },
    cardIcon: {
      width: 40,
      height: 40,
      borderRadius: 14,
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      backgroundColor: alpha(theme.palette.primary.main, isDark ? 0.2 : 0.12),
      color: theme.palette.primary.main,
    },
    cardTitle: {
      fontWeight: 600,
    },
    cardDescription: {
      color: theme.palette.text.secondary,
    },
    cardMeta: {
      color: theme.palette.text.secondary,
      fontSize: '0.78rem',
    },
    templateChip: {
      marginTop: theme.spacing(0.5),
      backgroundColor: alpha(theme.palette.secondary.main, isDark ? 0.2 : 0.12),
      color: theme.palette.secondary.main,
      fontWeight: 600,
    },
    flavorBadge: {
      marginTop: theme.spacing(0.5),
      backgroundColor: alpha(theme.palette.success.main, isDark ? 0.18 : 0.14),
      color: theme.palette.success.main,
      fontWeight: 600,
    },
    flavorResources: {
      color: theme.palette.text.secondary,
      fontWeight: 500,
    },
    formCard: {
      borderRadius: 'var(--aegis-card-radius)',
      border,
      backgroundColor: theme.palette.background.paper,
      padding: theme.spacing(3),
    },
    fieldGrid: {
      marginTop: theme.spacing(1),
      marginBottom: theme.spacing(1),
    },
    advancedToggle: {
      marginTop: theme.spacing(2),
    },
    advancedHelper: {
      color: theme.palette.text.secondary,
      marginBottom: theme.spacing(2),
    },
    summaryCard: {
      borderRadius: 'var(--aegis-card-radius)',
      border,
      backgroundColor: theme.palette.background.paper,
      padding: theme.spacing(3),
    },
    summaryGrid: {
      display: 'grid',
      gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))',
      gap: theme.spacing(2),
      marginTop: theme.spacing(2),
    },
    summaryRow: {
      display: 'flex',
      flexDirection: 'column',
      gap: theme.spacing(0.5),
    },
    summaryLabel: {
      textTransform: 'uppercase',
      letterSpacing: '0.08em',
      fontSize: '0.7rem',
      color: theme.palette.text.secondary,
      fontWeight: 600,
    },
    summaryValue: {
      fontWeight: 600,
    },
    summaryBlock: {
      marginTop: theme.spacing(2.5),
      padding: theme.spacing(1.5),
      borderRadius: 12,
      backgroundColor: alpha(theme.palette.primary.main, isDark ? 0.1 : 0.05),
    },
    summaryBlockLabel: {
      textTransform: 'uppercase',
      letterSpacing: '0.08em',
      fontSize: '0.7rem',
      color: theme.palette.text.secondary,
      fontWeight: 600,
      marginBottom: theme.spacing(0.75),
    },
    actionRow: {
      marginTop: theme.spacing(4),
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      gap: theme.spacing(2),
      flexWrap: 'wrap',
    },
    navButtons: {
      display: 'flex',
      gap: theme.spacing(1.5),
    },
    launchGroup: {
      display: 'flex',
      alignItems: 'center',
      gap: theme.spacing(2),
    },
    progress: {
      color: theme.palette.primary.main,
    },
    errorPanel: {
      marginTop: theme.spacing(3),
    },
    templateEmpty: {
      color: theme.palette.text.secondary,
    },
  };
});

const workspaceTypeCatalog: WorkspaceTypeOption[] = [
  {
    id: 'vscode',
    title: 'VS Code',
    description: 'Full-featured IDE with terminal and debugging support.',
  },
  {
    id: 'jupyter',
    title: 'JupyterLab',
    description: 'Notebook-centric environment for data exploration.',
  },
  {
    id: 'cli',
    title: 'CLI Workspace',
    description: 'Lightweight shell session for quick administration tasks.',
  },
];

const templateCatalog: TemplateOption[] = [
  {
    id: 'vscode-python',
    title: 'Python Starter',
    description: 'VS Code image tuned for Python, linting, and testing.',
    workspaceTypes: ['vscode'],
    defaults: {
      flavor: 'cpu-medium',
      image: 'carlosmsanchez/aegis-workspace-vscode:latest',
      ports: [22, 11111],
    },
  },
  {
    id: 'vscode-data',
    title: 'Data Engineering',
    description: 'VS Code with dbt, SQL utilities, and data connectors.',
    workspaceTypes: ['vscode'],
    defaults: {
      flavor: 'cpu-large',
      image: 'ghcr.io/aegis/workspace-vscode-data:latest',
      ports: [22, 11111],
    },
  },
  {
    id: 'jupyter-pytorch',
    title: 'PyTorch GPU',
    description: 'JupyterLab with CUDA, PyTorch, and ML tooling pre-installed.',
    workspaceTypes: ['jupyter'],
    defaults: {
      flavor: 'gpu-standard',
      image: 'ghcr.io/aegis/workspace-jupyter-pytorch:latest',
      queue: 'gpu',
      ports: [22, 8888],
      env: {
        NOTEBOOK_TOKEN: 'aegis',
      },
    },
  },
  {
    id: 'jupyter-rapids',
    title: 'RAPIDS Accelerator',
    description: 'GPU-accelerated RAPIDS stack for large-scale analytics.',
    workspaceTypes: ['jupyter'],
    defaults: {
      flavor: 'gpu-large',
      image: 'ghcr.io/aegis/workspace-jupyter-rapids:latest',
      queue: 'gpu',
      ports: [22, 8888],
    },
  },
  {
    id: 'cli-ops',
    title: 'Operations Shell',
    description: 'Lean container with kubectl, helm, and cloud CLIs.',
    workspaceTypes: ['cli'],
    defaults: {
      flavor: 'cpu-small',
      image: 'ghcr.io/aegis/workspace-cli:latest',
      ports: [22],
    },
  },
  {
    id: 'custom',
    title: 'Custom Image',
    description: 'Bring your own container image and connection settings.',
    workspaceTypes: ['vscode', 'jupyter', 'cli'],
    defaults: {
      ports: [22],
    },
    autoShowAdvanced: true,
  },
];

const flavorCatalog: FlavorOption[] = [
  {
    id: 'cpu-small',
    title: 'Small',
    description: '2 vCPU, 4 GiB RAM — great for quick CLI sessions.',
    flavor: 'cpu-small',
    resources: '2 vCPU • 4 GiB RAM',
  },
  {
    id: 'cpu-medium',
    title: 'Medium',
    description: '4 vCPU, 16 GiB RAM — balanced choice for most notebooks.',
    flavor: 'cpu-medium',
    resources: '4 vCPU • 16 GiB RAM',
  },
  {
    id: 'cpu-large',
    title: 'Large',
    description: '8 vCPU, 32 GiB RAM — heavier IDE workloads and data prep.',
    flavor: 'cpu-large',
    resources: '8 vCPU • 32 GiB RAM',
  },
  {
    id: 'gpu-standard',
    title: 'GPU Standard',
    description: '1× NVIDIA T4, 4 vCPU, 32 GiB RAM — training and inference.',
    flavor: 'gpu-standard',
    resources: '1× T4 • 4 vCPU • 32 GiB RAM',
  },
  {
    id: 'gpu-large',
    title: 'GPU Large',
    description: '1× NVIDIA A10, 8 vCPU, 64 GiB RAM — larger GPU workloads.',
    flavor: 'gpu-large',
    resources: '1× A10 • 8 vCPU • 64 GiB RAM',
  },
];

const workspaceTypeIconMap: Record<WorkspaceTypeId, SvgIconComponent> = {
  vscode: LaptopMacIcon,
  jupyter: ScatterPlotIcon,
  cli: TerminalIcon,
};

const templateIconMap: Record<string, SvgIconComponent> = {
  'vscode-python': CodeIcon,
  'vscode-data': StorageIcon,
  'jupyter-pytorch': FunctionsIcon,
  'jupyter-rapids': ScatterPlotIcon,
  'cli-ops': BuildIcon,
  custom: LayersIcon,
};

const steps = ['Workspace Basics', 'Resources & Options', 'Review & Launch'];

const randomId = () => {
  if (typeof crypto?.randomUUID === 'function') {
    return crypto.randomUUID();
  }
  return `workspace-${Math.random().toString(16).slice(2, 10)}`;
};

const templateEnvToText = (env?: Record<string, string>): string => {
  if (!env) {
    return '';
  }
  return Object.entries(env)
    .map(([key, value]) => `${key}=${value}`)
    .join('\n');
};

const portsToText = (ports?: number[]): string => {
  if (!ports || ports.length === 0) {
    return '';
  }
  return ports.join(', ');
};

export const LaunchWorkspacePage: FC = () => {
  const classes = useStyles();
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const alertApi = useApi(alertApiRef);
  const workloadsLink = useRouteRef(workloadsRouteRef);
  const navigate = useNavigate();

  const [activeStep, setActiveStep] = useState(0);
  const [workspaceTypeId, setWorkspaceTypeId] =
    useState<WorkspaceTypeId | null>(null);
  const [templateId, setTemplateId] = useState<string | null>(null);
  const [forceAdvancedOpen, setForceAdvancedOpen] = useState(false);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [form, setForm] = useState({
    workloadId: randomId(),
    projectId: '',
    queue: '',
    flavor: '',
    image: '',
    ports: '22',
    env: '',
  });
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const templatesForType = useMemo(() => {
    if (!workspaceTypeId) {
      return templateCatalog;
    }
    return templateCatalog.filter(template =>
      template.workspaceTypes.includes(workspaceTypeId),
    );
  }, [workspaceTypeId]);

  const selectedTemplate = useMemo(
    () => templateCatalog.find(template => template.id === templateId) ?? null,
    [templateId],
  );

  const selectedWorkspaceType = useMemo(
    () =>
      workspaceTypeCatalog.find(option => option.id === workspaceTypeId) ??
      null,
    [workspaceTypeId],
  );

  const handleFormFieldChange =
    (field: keyof typeof form) => (event: ChangeEvent<HTMLInputElement>) => {
      setForm(prev => ({ ...prev, [field]: event.target.value }));
    };

  const applyTemplate = (template: TemplateOption) => {
    setTemplateId(template.id);
    setForceAdvancedOpen(Boolean(template.autoShowAdvanced));
    setAdvancedOpen(prev => prev || Boolean(template.autoShowAdvanced));

    setForm(prev => ({
      ...prev,
      flavor: template.defaults.flavor ?? prev.flavor,
      image:
        template.defaults.image !== undefined
          ? template.defaults.image
          : prev.image,
      queue:
        template.defaults.queue !== undefined
          ? template.defaults.queue
          : prev.queue,
      ports:
        template.defaults.ports !== undefined
          ? portsToText(template.defaults.ports)
          : prev.ports,
      env:
        template.defaults.env !== undefined
          ? templateEnvToText(template.defaults.env)
          : prev.env,
    }));
  };

  const handleWorkspaceTypeSelect = (option: WorkspaceTypeOption) => {
    const nextTypeId = option.id;
    setWorkspaceTypeId(nextTypeId);
    const matchingTemplates = templateCatalog.filter(template =>
      template.workspaceTypes.includes(nextTypeId),
    );
    if (matchingTemplates.length > 0) {
      applyTemplate(matchingTemplates[0]);
    } else {
      setTemplateId(null);
    }
  };

  const handleTemplateSelect = (template: TemplateOption) => {
    applyTemplate(template);
  };

  const handleFlavorSelect = (flavor: FlavorOption) => {
    setForm(prev => ({ ...prev, flavor: flavor.flavor }));
  };

  const handleAdvancedToggle = (event: ChangeEvent<HTMLInputElement>) => {
    if (forceAdvancedOpen) {
      return;
    }
    setAdvancedOpen(event.target.checked);
  };

  const goNextStep = () => {
    setActiveStep(prev => Math.min(prev + 1, steps.length - 1));
  };

  const goPreviousStep = () => {
    setActiveStep(prev => Math.max(prev - 1, 0));
  };

  const canProceedFromBasics =
    Boolean(workspaceTypeId) &&
    Boolean(templateId) &&
    Boolean(form.projectId.trim()) &&
    Boolean(form.workloadId.trim());

  const canProceedFromResources =
    Boolean(form.flavor.trim()) && Boolean(form.image.trim());

  const isSubmitDisabled =
    submitting ||
    !form.projectId.trim() ||
    !form.flavor.trim() ||
    !form.image.trim() ||
    !form.workloadId.trim();

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    if (isSubmitDisabled) {
      return;
    }

    const ports = parsePortsInput(form.ports);
    const env = parseEnvInput(form.env);

    const payload: SubmitWorkspaceRequest = {
      id: form.workloadId.trim(),
      projectId: form.projectId.trim(),
      queue: form.queue.trim() || undefined,
      workspace: {
        flavor: form.flavor.trim() || undefined,
        image: form.image.trim() || undefined,
        interactive: true,
        ports: ports.length > 0 ? ports : undefined,
        env: Object.keys(env).length > 0 ? env : undefined,
      },
    };

    try {
      setSubmitting(true);
      setError(null);
      await submitWorkspace(fetchApi, discoveryApi, identityApi, payload);
      alertApi.post({
        message: `Submitted interactive workspace ${payload.id}`,
        severity: 'success',
      });
      if (workloadsLink) {
        navigate(workloadsLink());
      }
    } catch (e: any) {
      const msg = e?.message ?? String(e);
      setError(msg);
      alertApi.post({
        message: `Failed to submit workspace: ${msg}`,
        severity: 'error',
      });
    } finally {
      setSubmitting(false);
    }
  };

  const renderWorkspaceTypeCards = () => (
    <Grid container spacing={2} className={classes.selectionGrid}>
      {workspaceTypeCatalog.map(option => {
        const Icon = workspaceTypeIconMap[option.id];
        const selected = option.id === workspaceTypeId;
        return (
          <Grid item xs={12} md={4} key={option.id}>
            <Card
              elevation={0}
              className={clsx(classes.selectionCard, {
                [classes.selectionCardActive]: selected,
              })}
            >
              <CardActionArea
                className={classes.cardAction}
                onClick={() => handleWorkspaceTypeSelect(option)}
              >
                <CardContent className={classes.cardContent}>
                  <div className={classes.cardIcon}>
                    <Icon fontSize="small" />
                  </div>
                  <Typography variant="subtitle1" className={classes.cardTitle}>
                    {option.title}
                  </Typography>
                  <Typography
                    variant="body2"
                    className={classes.cardDescription}
                  >
                    {option.description}
                  </Typography>
                </CardContent>
              </CardActionArea>
            </Card>
          </Grid>
        );
      })}
    </Grid>
  );

  const renderTemplateCards = () => (
    <Grid container spacing={2} className={classes.selectionGrid}>
      {templatesForType.length === 0 ? (
        <Grid item xs={12}>
          <Typography variant="body2" className={classes.templateEmpty}>
            No templates available for the selected workspace type.
          </Typography>
        </Grid>
      ) : (
        templatesForType.map(template => {
          const selected = template.id === templateId;
          const TemplateIcon = templateIconMap[template.id] ?? LayersIcon;
          const recommendedFlavor = template.defaults.flavor
            ? flavorCatalog.find(option => option.flavor === template.defaults.flavor)
            : undefined;
          return (
            <Grid item xs={12} md={6} key={template.id}>
              <Card
                elevation={0}
                className={clsx(classes.selectionCard, {
                  [classes.selectionCardActive]: selected,
                })}
              >
                <CardActionArea
                  className={classes.cardAction}
                  onClick={() => handleTemplateSelect(template)}
                >
                  <CardContent className={classes.cardContent}>
                    <div className={classes.cardIcon}>
                      <TemplateIcon fontSize="small" />
                    </div>
                    <Typography
                      variant="subtitle1"
                      className={classes.cardTitle}
                    >
                      {template.title}
                    </Typography>
                    <Typography
                      variant="body2"
                      className={classes.cardDescription}
                    >
                      {template.description}
                    </Typography>
                    {template.defaults.queue && (
                      <Chip
                        size="small"
                        label={`${template.defaults.queue.toUpperCase()} queue`}
                        className={classes.templateChip}
                      />
                    )}
                    {recommendedFlavor && (
                      <Typography variant="caption" className={classes.cardMeta}>
                        Suggested flavor: {recommendedFlavor.title}
                      </Typography>
                    )}
                  </CardContent>
                </CardActionArea>
              </Card>
            </Grid>
          );
        })
      )}
    </Grid>
  );

  const renderFlavorCards = () => (
    <Grid container spacing={2} className={classes.selectionGrid}>
      {flavorCatalog.map(option => {
        const selected = option.flavor === form.flavor;
        const isGpuFlavor = option.flavor.startsWith('gpu');
        return (
          <Grid item xs={12} md={4} key={option.id}>
            <Card
              elevation={0}
              className={clsx(classes.selectionCard, {
                [classes.selectionCardActive]: selected,
              })}
            >
              <CardActionArea
                className={classes.cardAction}
                onClick={() => handleFlavorSelect(option)}
              >
                <CardContent className={classes.cardContent}>
                  <div className={classes.cardIcon}>
                    <LayersIcon fontSize="small" />
                  </div>
                  <Typography variant="subtitle1" className={classes.cardTitle}>
                    {option.title}
                  </Typography>
                  <Typography
                    variant="subtitle2"
                    className={classes.flavorResources}
                  >
                    {option.resources}
                  </Typography>
                  <Typography
                    variant="body2"
                    className={classes.cardDescription}
                  >
                    {option.description}
                  </Typography>
                  {isGpuFlavor && (
                    <Chip size="small" label="GPU" className={classes.flavorBadge} />
                  )}
                </CardContent>
              </CardActionArea>
            </Card>
          </Grid>
        );
      })}
    </Grid>
  );

  const summaryItems = [
    { label: 'Project', value: form.projectId || '—' },
    { label: 'Workspace ID', value: form.workloadId || '—' },
    { label: 'Workspace type', value: selectedWorkspaceType?.title ?? '—' },
    { label: 'Template', value: selectedTemplate?.title ?? '—' },
    { label: 'Flavor', value: form.flavor || '—' },
    { label: 'Queue', value: form.queue || 'Default' },
    { label: 'Container image', value: form.image || '—' },
    {
      label: 'Ports',
      value: form.ports.trim() ? form.ports : 'Default (22, 11111)',
    },
  ];

  return (
    <Page themeId="tool">
      <Header
        title="Launch Interactive Workspace"
        subtitle="Curate a secure VS Code, Jupyter, or CLI surface across the ÆGIS multi-cloud mesh."
      />
      <Content className={classes.content}>
        <ContentHeader title="Workspace Wizard">
          <Typography variant="body1" className={classes.lead}>
            Provision developer workstations with GPU-aware defaults, cost
            guardrails, and policy overlays before handing them to mission
            teams.
          </Typography>
        </ContentHeader>
        <form className={classes.form} onSubmit={handleSubmit}>
          <Stepper activeStep={activeStep} alternativeLabel className={classes.stepper}>
            {steps.map(step => (
              <Step key={step}>
                <StepLabel>{step}</StepLabel>
              </Step>
            ))}
          </Stepper>

          <Box>
            {activeStep === 0 && (
              <>
                <Box className={classes.section}>
                  <Typography variant="h6" className={classes.sectionTitle}>
                    Workspace basics
                  </Typography>
                  <Typography variant="body2" className={classes.sectionSubtitle}>
                    Name the workspace and select the experience best suited
                    for the mission.
                  </Typography>
                  <Grid container spacing={2}>
                    <Grid item xs={12} md={6}>
                      <TextField
                        label="Project ID"
                        value={form.projectId}
                        onChange={handleFormFieldChange('projectId')}
                        fullWidth
                        required
                        variant="outlined"
                        helperText="Project that will own this workspace"
                      />
                    </Grid>
                    <Grid item xs={12} md={6}>
                      <TextField
                        label="Workspace ID"
                        value={form.workloadId}
                        onChange={handleFormFieldChange('workloadId')}
                        fullWidth
                        required
                        variant="outlined"
                        helperText="Unique identifier for this launch request"
                      />
                    </Grid>
                  </Grid>
                </Box>
                <Box className={classes.section}>
                  <Typography variant="h6" className={classes.sectionTitle}>
                    Choose workspace type
                  </Typography>
                  <Typography variant="body2" className={classes.sectionSubtitle}>
                    Select the environment that aligns with how the operator
                    works.
                  </Typography>
                  {renderWorkspaceTypeCards()}
                </Box>
                <Box className={classes.section}>
                  <Typography variant="h6" className={classes.sectionTitle}>
                    Pick a template
                  </Typography>
                  <Typography variant="body2" className={classes.sectionSubtitle}>
                    Templates hydrate curated images, GPU defaults, and
                    connection details.
                  </Typography>
                  {renderTemplateCards()}
                </Box>
              </>
            )}

            {activeStep === 1 && (
              <>
                <Box className={classes.section}>
                  <Typography variant="h6" className={classes.sectionTitle}>
                    Size the workspace
                  </Typography>
                  <Typography variant="body2" className={classes.sectionSubtitle}>
                    Align compute resources to mission demand, from CPU prep to
                    GPU-intensive training.
                  </Typography>
                  {renderFlavorCards()}
                </Box>
                <Card elevation={0} className={classes.formCard}>
                  <Typography variant="subtitle1" className={classes.cardTitle}>
                    Runtime options
                  </Typography>
                  <Typography variant="body2" className={classes.cardDescription}>
                    Tune queue routing and container details before launch.
                  </Typography>
                  <Grid container spacing={2} className={classes.fieldGrid}>
                    <Grid item xs={12} md={4}>
                      <TextField
                        label="Queue"
                        value={form.queue}
                        onChange={handleFormFieldChange('queue')}
                        fullWidth
                        variant="outlined"
                        helperText="Defaults to control plane routing if empty"
                      />
                    </Grid>
                    <Grid item xs={12} md={8}>
                      <TextField
                        label="Container image"
                        value={form.image}
                        onChange={handleFormFieldChange('image')}
                        fullWidth
                        required
                        variant="outlined"
                        helperText="OCI image reference (registry/repo:tag)"
                      />
                    </Grid>
                  </Grid>
                  <FormControlLabel
                    className={classes.advancedToggle}
                    control={
                      <Switch
                        color="primary"
                        checked={forceAdvancedOpen || advancedOpen}
                        onChange={handleAdvancedToggle}
                        disabled={forceAdvancedOpen}
                      />
                    }
                    label="Show advanced overrides"
                  />
                  <Collapse in={advancedOpen || forceAdvancedOpen}>
                    <Typography variant="body2" className={classes.advancedHelper}>
                      Override exposed ports or inject environment variables as
                      KEY=VALUE pairs.
                    </Typography>
                    <Grid container spacing={2}>
                      <Grid item xs={12} md={6}>
                        <TextField
                          label="Exposed ports"
                          value={form.ports}
                          onChange={handleFormFieldChange('ports')}
                          fullWidth
                          variant="outlined"
                          helperText="Comma-separated list (e.g. 22, 8888)"
                        />
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <TextField
                          label="Environment variables"
                          value={form.env}
                          onChange={handleFormFieldChange('env')}
                          fullWidth
                          multiline
                          rows={4}
                          variant="outlined"
                          helperText="One KEY=VALUE per line"
                        />
                      </Grid>
                    </Grid>
                  </Collapse>
                </Card>
              </>
            )}

            {activeStep === 2 && (
              <Card elevation={0} className={classes.summaryCard}>
                <Typography variant="h6" className={classes.sectionTitle}>
                  Review launch plan
                </Typography>
                <Typography variant="body2" className={classes.sectionSubtitle}>
                  Validate the workspace profile before dispatching to the ÆGIS
                  control plane.
                </Typography>
                <div className={classes.summaryGrid}>
                  {summaryItems.map(item => (
                    <div key={item.label} className={classes.summaryRow}>
                      <span className={classes.summaryLabel}>{item.label}</span>
                      <Typography variant="body1" className={classes.summaryValue}>
                        {item.value}
                      </Typography>
                    </div>
                  ))}
                </div>
                <div className={classes.summaryBlock}>
                  <div className={classes.summaryBlockLabel}>Environment variables</div>
                  <Typography variant="body2">
                    {form.env || 'Inherited defaults'}
                  </Typography>
                </div>
                {/* TODO: Attach usage and billing telemetry once backend APIs land. */}
              </Card>
            )}
          </Box>

          <Box className={classes.actionRow}>
            <div className={classes.navButtons}>
              <Button
                type="button"
                variant="outlined"
                disabled={activeStep === 0 || submitting}
                onClick={goPreviousStep}
              >
                Back
              </Button>
              {activeStep < steps.length - 1 && (
                <Button
                  type="button"
                  color="primary"
                  variant="contained"
                  disabled={
                    submitting ||
                    (activeStep === 0 && !canProceedFromBasics) ||
                    (activeStep === 1 && !canProceedFromResources)
                  }
                  onClick={goNextStep}
                >
                  Next
                </Button>
              )}
            </div>
            {activeStep === steps.length - 1 && (
              <div className={classes.launchGroup}>
                <Button
                  type="submit"
                  color="primary"
                  variant="contained"
                  disabled={isSubmitDisabled}
                >
                  Launch Workspace
                </Button>
                {submitting && <Progress className={classes.progress} />}
              </div>
            )}
          </Box>
        </form>

        {error && (
          <div className={classes.errorPanel}>
            <WarningPanel severity="error" title="Workspace submission failed">
              {error}
            </WarningPanel>
          </div>
        )}
      </Content>
    </Page>
  );
};
