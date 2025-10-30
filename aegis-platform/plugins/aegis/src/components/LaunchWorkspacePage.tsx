import { ChangeEvent, FC, FormEvent, useMemo, useState } from 'react';
import {
  Page,
  Header,
  Content,
  ContentHeader,
  Progress,
  WarningPanel,
  InfoCard,
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
} from '@material-ui/core';
import { makeStyles, Theme } from '@material-ui/core/styles';
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

const steps = ['Workspace Basics', 'Resources & Options', 'Review & Launch'];

const useStyles = makeStyles((theme: Theme) => ({
  page: {
    position: 'relative',
  },
  content: {
    position: 'relative',
    zIndex: 1,
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(4),
  },
  hero: {
    position: 'relative',
    overflow: 'hidden',
    padding: theme.spacing(4.5, 5),
    borderRadius: 28,
    background:
      'linear-gradient(135deg, rgba(0, 245, 255, 0.16) 0%, rgba(7, 18, 44, 0.92) 52%, rgba(18, 8, 30, 0.88) 100%)',
    boxShadow:
      '0 30px 60px rgba(3, 12, 35, 0.55), inset 0 0 45px rgba(0, 245, 255, 0.12)',
  },
  heroGlow: {
    position: 'absolute',
    inset: '-20% -30% auto auto',
    width: '60%',
    height: '160%',
    background:
      'radial-gradient(circle, rgba(0, 245, 255, 0.55) 0%, rgba(0, 245, 255, 0) 65%)',
    opacity: 0.4,
    filter: 'blur(6px)',
  },
  heroTitle: {
    fontSize: '2.5rem',
    letterSpacing: '0.18em',
    marginBottom: theme.spacing(1.5),
    textTransform: 'uppercase',
  },
  heroSubtitle: {
    color: 'rgba(197, 226, 255, 0.82)',
    maxWidth: 720,
  },
  heroMetrics: {
    display: 'grid',
    gridTemplateColumns: 'repeat(auto-fit, minmax(160px, 1fr))',
    gap: theme.spacing(2.5),
    marginTop: theme.spacing(3.5),
  },
  metricTile: {
    padding: theme.spacing(2, 2.5),
    borderRadius: 18,
    background:
      'linear-gradient(160deg, rgba(0, 245, 255, 0.18) 0%, rgba(7, 20, 44, 0.78) 60%, rgba(5, 12, 30, 0.9) 100%)',
    border: '1px solid rgba(0, 245, 255, 0.25)',
    boxShadow: '0 18px 36px rgba(0, 245, 255, 0.12)',
  },
  metricValue: {
    fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
    fontSize: '1.8rem',
    letterSpacing: '0.24em',
    color: '#00f5ff',
  },
  metricLabel: {
    marginTop: theme.spacing(1),
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
    color: 'rgba(124, 154, 196, 0.9)',
  },
  panel: {
    borderRadius: 24,
    padding: theme.spacing(4),
    position: 'relative',
    background:
      'linear-gradient(165deg, rgba(5, 14, 32, 0.95) 0%, rgba(8, 21, 46, 0.88) 50%, rgba(16, 8, 28, 0.88) 100%)',
    border: '1px solid rgba(0, 245, 255, 0.12)',
    boxShadow: '0 24px 48px rgba(3, 12, 35, 0.4)',
  },
  stepper: {
    background: 'transparent',
    padding: theme.spacing(0, 2),
  },
  sectionTitle: {
    letterSpacing: '0.16em',
    textTransform: 'uppercase',
    color: '#9cbdef',
  },
  cardGrid: {
    marginTop: theme.spacing(2.5),
  },
  selectCard: {
    borderRadius: 20,
    border: '1px solid rgba(0, 245, 255, 0.18)',
    background:
      'linear-gradient(150deg, rgba(0, 245, 255, 0.08) 0%, rgba(7, 20, 42, 0.92) 50%, rgba(16, 9, 30, 0.85) 100%)',
    transition: 'transform 180ms ease, border-color 180ms ease, box-shadow 180ms ease',
    boxShadow: '0 14px 36px rgba(0, 245, 255, 0.1)',
  },
  selectCardSelected: {
    borderColor: '#00f5ff',
    boxShadow: '0 20px 46px rgba(0, 245, 255, 0.26)',
    transform: 'translateY(-4px)',
  },
  cardAction: {
    height: '100%',
  },
  cardContent: {
    display: 'flex',
    flexDirection: 'column',
    gap: theme.spacing(1.2),
  },
  cardTitle: {
    fontFamily: "'Orbitron', 'Rajdhani', sans-serif",
    letterSpacing: '0.12em',
    textTransform: 'uppercase',
  },
  advancedToggle: {
    marginTop: theme.spacing(1),
    padding: theme.spacing(1.5, 2),
    borderRadius: 16,
    backgroundColor: 'rgba(0, 245, 255, 0.08)',
    border: '1px solid rgba(0, 245, 255, 0.16)',
  },
  advancedGrid: {
    marginTop: theme.spacing(2),
  },
  reviewCard: {
    marginTop: theme.spacing(2),
    borderRadius: 24,
    border: '1px solid rgba(0, 245, 255, 0.15)',
    background:
      'linear-gradient(170deg, rgba(8, 20, 42, 0.95) 0%, rgba(12, 26, 52, 0.92) 45%, rgba(18, 8, 30, 0.85) 100%)',
  },
  reviewLabel: {
    textTransform: 'uppercase',
    letterSpacing: '0.14em',
    color: 'rgba(124, 154, 196, 0.9)',
  },
  actionRow: {
    marginTop: theme.spacing(3),
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'space-between',
    flexWrap: 'wrap',
    gap: theme.spacing(2),
  },
  actionLeft: {
    display: 'flex',
    gap: theme.spacing(2),
  },
  inlineProgress: {
    display: 'flex',
    alignItems: 'center',
    gap: theme.spacing(2),
  },
  errorPanel: {
    marginTop: theme.spacing(2.5),
  },
}));

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

  const heroMetrics = useMemo(
    () => [
      {
        label: 'Templates',
        value: templatesForType.length.toString().padStart(2, '0'),
      },
      {
        label: 'Workspace Types',
        value: workspaceTypeCatalog.length.toString().padStart(2, '0'),
      },
      {
        label: 'GPU Flavors',
        value: flavorCatalog.length.toString().padStart(2, '0'),
      },
    ],
    [templatesForType.length],
  );

  const renderWorkspaceTypeCards = () => (
    <Grid container spacing={2}>
      {workspaceTypeCatalog.map(option => {
        const selected = option.id === workspaceTypeId;
        const className = selected
          ? `${classes.selectCard} ${classes.selectCardSelected}`
          : classes.selectCard;
        return (
          <Grid item xs={12} md={4} key={option.id}>
            <Card
              className={className}
            >
              <CardActionArea
                className={classes.cardAction}
                onClick={() => handleWorkspaceTypeSelect(option)}
              >
                <CardContent className={classes.cardContent}>
                  <Typography variant="h6" className={classes.cardTitle}>
                    {option.title}
                  </Typography>
                  <Typography variant="body2" color="textSecondary">
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
    <Grid container spacing={2}>
      {templatesForType.length === 0 ? (
        <Grid item xs={12}>
          <Typography variant="body2" color="textSecondary">
            No templates available for the selected workspace type.
          </Typography>
        </Grid>
      ) : (
        templatesForType.map(template => {
          const selected = template.id === templateId;
          const className = selected
            ? `${classes.selectCard} ${classes.selectCardSelected}`
            : classes.selectCard;
          return (
            <Grid item xs={12} md={6} key={template.id}>
              <Card
                className={className}
              >
                <CardActionArea
                  className={classes.cardAction}
                  onClick={() => handleTemplateSelect(template)}
                >
                  <CardContent className={classes.cardContent}>
                    <Typography variant="h6" className={classes.cardTitle}>
                      {template.title}
                    </Typography>
                    <Typography variant="body2" color="textSecondary">
                      {template.description}
                    </Typography>
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
    <Grid container spacing={2}>
      {flavorCatalog.map(option => {
        const selected = option.flavor === form.flavor;
        const className = selected
          ? `${classes.selectCard} ${classes.selectCardSelected}`
          : classes.selectCard;
        return (
          <Grid item xs={12} md={4} key={option.id}>
            <Card
              className={className}
            >
              <CardActionArea
                className={classes.cardAction}
                onClick={() => handleFlavorSelect(option)}
              >
                <CardContent className={classes.cardContent}>
                  <Typography variant="h6" className={classes.cardTitle}>
                    {option.title}
                  </Typography>
                  <Typography variant="subtitle2" color="textSecondary">
                    {option.resources}
                  </Typography>
                  <Typography variant="body2" color="textSecondary">
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

  return (
    <Page themeId="tool" className={classes.page}>
      <Header title="ÆGIS Launch Control" />
      <Content>
        <div className={classes.content}>
          <div className={classes.hero}>
            <div className={classes.heroGlow} />
            <Typography variant="h2" className={classes.heroTitle}>
              Deploy Mission Workspaces
            </Typography>
            <Typography variant="subtitle1" className={classes.heroSubtitle}>
              Engage the multi-cloud GPU fleet with curated mission templates,
              compliance guardrails, and real-time orchestration signals.
            </Typography>
            <div className={classes.heroMetrics}>
              {heroMetrics.map(metric => (
                <div key={metric.label} className={classes.metricTile}>
                  <Typography className={classes.metricValue}>
                    {metric.value}
                  </Typography>
                  <Typography variant="caption" className={classes.metricLabel}>
                    {metric.label}
                  </Typography>
                </div>
              ))}
            </div>
          </div>

          <div className={classes.panel}>
            <ContentHeader title="Workspace Wizard" />
            <form onSubmit={handleSubmit}>
              <Stepper
                activeStep={activeStep}
                alternativeLabel
                className={classes.stepper}
              >
                {steps.map(step => (
                  <Step key={step}>
                    <StepLabel>{step}</StepLabel>
                  </Step>
                ))}
              </Stepper>

              <Box mt={4}>
                {activeStep === 0 && (
                  <Grid container spacing={3}>
                    <Grid item xs={12} md={6}>
                      <TextField
                        label="Project ID"
                        value={form.projectId}
                        onChange={handleFormFieldChange('projectId')}
                        fullWidth
                        required
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
                        helperText="Unique identifier used to track the workspace"
                      />
                    </Grid>
                    <Grid item xs={12}>
                      <Typography variant="subtitle1" className={classes.sectionTitle}>
                        Choose a workspace type
                      </Typography>
                      <div className={classes.cardGrid}>{renderWorkspaceTypeCards()}</div>
                    </Grid>
                    <Grid item xs={12}>
                      <Typography variant="subtitle1" className={classes.sectionTitle}>
                        Pick a template
                      </Typography>
                      <div className={classes.cardGrid}>{renderTemplateCards()}</div>
                    </Grid>
                  </Grid>
                )}

                {activeStep === 1 && (
                  <Grid container spacing={3}>
                    <Grid item xs={12}>
                      <Typography variant="subtitle1" className={classes.sectionTitle}>
                        Select a resource profile
                      </Typography>
                      <div className={classes.cardGrid}>{renderFlavorCards()}</div>
                    </Grid>
                    <Grid item xs={12}>
                      <div className={classes.advancedToggle}>
                        <FormControlLabel
                          control={
                            <Switch
                              color="primary"
                              checked={forceAdvancedOpen || advancedOpen}
                              onChange={handleAdvancedToggle}
                              disabled={forceAdvancedOpen}
                            />
                          }
                          label="Advanced overrides"
                        />
                      </div>
                    </Grid>
                    <Grid item xs={12}>
                      <Collapse in={forceAdvancedOpen || advancedOpen}>
                        <Grid container spacing={3} className={classes.advancedGrid}>
                          <Grid item xs={12} md={6}>
                            <TextField
                              label="Container image"
                              value={form.image}
                              onChange={handleFormFieldChange('image')}
                              fullWidth
                              required
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              label="Queue"
                              value={form.queue}
                              onChange={handleFormFieldChange('queue')}
                              fullWidth
                              helperText="Optional scheduling queue override"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              label="Exposed ports"
                              value={form.ports}
                              onChange={handleFormFieldChange('ports')}
                              fullWidth
                              helperText="Comma or space separated list"
                            />
                          </Grid>
                          <Grid item xs={12} md={6}>
                            <TextField
                              label="Environment variables"
                              value={form.env}
                              onChange={handleFormFieldChange('env')}
                              fullWidth
                              multiline
                              minRows={3}
                              helperText="Optional KEY=VALUE pairs, one per line"
                            />
                          </Grid>
                        </Grid>
                      </Collapse>
                    </Grid>
                  </Grid>
                )}

                {activeStep === 2 && (
                  <InfoCard title="Review selection" className={classes.reviewCard}>
                    <Grid container spacing={2}>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Project
                        </Typography>
                        <Typography variant="body2">
                          {form.projectId || '—'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Workspace ID
                        </Typography>
                        <Typography variant="body2">
                          {form.workloadId || '—'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Workspace type
                        </Typography>
                        <Typography variant="body2">
                          {selectedWorkspaceType?.title ?? '—'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Template
                        </Typography>
                        <Typography variant="body2">
                          {selectedTemplate?.title ?? '—'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Flavor
                        </Typography>
                        <Typography variant="body2">
                          {form.flavor || '—'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Queue
                        </Typography>
                        <Typography variant="body2">
                          {form.queue || 'Default'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Container image
                        </Typography>
                        <Typography variant="body2">{form.image || '—'}</Typography>
                      </Grid>
                      <Grid item xs={12} md={6}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Ports
                        </Typography>
                        <Typography variant="body2">
                          {form.ports || 'Default (22, 11111)'}
                        </Typography>
                      </Grid>
                      <Grid item xs={12}>
                        <Typography variant="subtitle2" className={classes.reviewLabel}>
                          Environment variables
                        </Typography>
                        <Typography
                          variant="body2"
                          style={{ whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}
                        >
                          {form.env || 'Inherited defaults'}
                        </Typography>
                      </Grid>
                    </Grid>
                  </InfoCard>
                )}
              </Box>

              <div className={classes.actionRow}>
                <div className={classes.actionLeft}>
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
                  <div className={classes.inlineProgress}>
                    <Button
                      type="submit"
                      color="primary"
                      variant="contained"
                      disabled={isSubmitDisabled}
                    >
                      Launch Workspace
                    </Button>
                    {submitting && <Progress />}
                  </div>
                )}
              </div>
            </form>
          </div>

          {error && (
            <div className={classes.errorPanel}>
              <WarningPanel severity="error" title="Workspace submission failed">
                {error}
              </WarningPanel>
            </div>
          )}
        </div>
      </Content>
    </Page>
  );
};
