import { useEffect, useMemo, useState } from 'react';
import { useNavigate } from 'react-router-dom';
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
import {
  Box,
  Button,
  Grid,
  Step,
  StepLabel,
  Stepper,
  Typography,
} from '@material-ui/core';
import Alert from '@material-ui/lab/Alert';
import { makeStyles } from '@material-ui/core/styles';
import {
  SubmitWorkspaceRequest,
  submitWorkspace,
  WorkloadDTO,
} from '../api/aegisClient';
import { workloadsRouteRef, workloadDetailsRouteRef } from '../routes';
import {
  parseEnvInput,
  parsePortsInput,
  validateEnvInput,
  validatePortsInput,
  formatEnvMap,
} from '../components/workspaceFormUtils';
import { StepTemplates } from './steps/StepTemplates';
import { StepConfigure, StepConfigureErrors } from './steps/StepConfigure';
import { StepReview } from './steps/StepReview';
import { WizardHeader } from './components/WizardHeader';
import {
  ClusterOption,
  FlavorId,
  TemplateId,
  WizardFormState,
  WizardStep,
  WorkspaceFlavor,
  WorkspaceTemplate,
} from './types';

const useStyles = makeStyles(theme => ({
  navControls: {
    display: 'flex',
    justifyContent: 'space-between',
    alignItems: 'center',
    marginTop: theme.spacing(4),
    gap: theme.spacing(2),
    flexWrap: 'wrap',
  },
  buttonRow: {
    display: 'flex',
    gap: theme.spacing(2),
  },
  srOnly: {
    border: 0,
    clip: 'rect(0 0 0 0)',
    height: 1,
    margin: -1,
    overflow: 'hidden',
    padding: 0,
    position: 'absolute',
    width: 1,
  },
  successBanner: {
    marginTop: theme.spacing(3),
  },
}));

const DEFAULT_PORTS = [22, 11111];

const clusterOptions: ClusterOption[] = [
  {
    id: 'secure-gpu',
    name: 'Secure GPU Fleet',
    description: 'Isolated nodes with NVIDIA accelerators',
  },
  {
    id: 'data-foundry',
    name: 'Data Foundry',
    description: 'High-memory CPU cluster for ETL workloads',
  },
  {
    id: 'edge-lab',
    name: 'Edge Lab',
    description: 'Burstable compute for experimentation',
  },
];

const templates: WorkspaceTemplate[] = [
  {
    id: 'vscode-python',
    name: 'Python Starter',
    description: 'VS Code with Python 3.11, linting, and pytest ready to go.',
    icon: 'python',
    defaults: {
      flavor: 'cpu-medium',
      image: 'ghcr.io/aegis/workspace-vscode-python:latest',
      ports: DEFAULT_PORTS,
    },
  },
  {
    id: 'vscode-data',
    name: 'Data Engineering',
    description: 'VS Code tuned for dbt, SQL workflows, and data connectors.',
    icon: 'data',
    defaults: {
      flavor: 'cpu-large',
      image: 'ghcr.io/aegis/workspace-vscode-data:latest',
      ports: DEFAULT_PORTS,
    },
  },
  {
    id: 'jupyter-pytorch',
    name: 'PyTorch GPU',
    description: 'JupyterLab with CUDA, PyTorch, and ML tooling pre-installed.',
    icon: 'pytorch',
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
    name: 'RAPIDS Accelerator',
    description: 'GPU-accelerated RAPIDS stack for large-scale analytics.',
    icon: 'spark',
    defaults: {
      flavor: 'gpu-large',
      image: 'ghcr.io/aegis/workspace-jupyter-rapids:latest',
      queue: 'gpu',
      ports: [22, 8888],
    },
  },
  {
    id: 'cli-ops',
    name: 'Operations Shell',
    description: 'Lightweight container with kubectl, helm, and cloud CLIs.',
    icon: 'terminal',
    defaults: {
      flavor: 'cpu-small',
      image: 'ghcr.io/aegis/workspace-cli:latest',
      ports: [22],
    },
  },
  {
    id: 'custom',
    name: 'Custom Image',
    description: 'Bring your own container image and network settings.',
    icon: 'custom',
    defaults: {
      ports: [22],
    },
    autoShowAdvanced: true,
  },
];

const flavors: WorkspaceFlavor[] = [
  {
    id: 'cpu-small',
    name: 'CPU Small',
    description: 'Burstable compute for quick CLI or admin work.',
    flavor: 'cpu-small',
    resources: '2 vCPU • 4 GiB RAM',
    category: 'cpu',
  },
  {
    id: 'cpu-medium',
    name: 'CPU Medium',
    description: 'Balanced compute for notebooks and IDE workflows.',
    flavor: 'cpu-medium',
    resources: '4 vCPU • 16 GiB RAM',
    category: 'cpu',
  },
  {
    id: 'cpu-large',
    name: 'CPU Large',
    description: 'Extra headroom for data preparation and build pipelines.',
    flavor: 'cpu-large',
    resources: '8 vCPU • 32 GiB RAM',
    category: 'cpu',
  },
  {
    id: 'gpu-standard',
    name: 'GPU Standard',
    description: 'Ideal for training, inference, and GPU notebooks.',
    flavor: 'gpu-standard',
    resources: '1× T4 • 4 vCPU • 32 GiB RAM',
    category: 'gpu',
    gpu: {
      model: 'NVIDIA T4',
      memory: '16 GiB',
      count: 1,
    },
  },
  {
    id: 'gpu-large',
    name: 'GPU Large',
    description: 'More powerful accelerator for heavy experimentation.',
    flavor: 'gpu-large',
    resources: '1× A10 • 8 vCPU • 64 GiB RAM',
    category: 'gpu',
    gpu: {
      model: 'NVIDIA A10',
      memory: '24 GiB',
      count: 1,
    },
  },
];

const stepsMeta = [
  {
    title: 'Pick a template',
    subtitle: 'Choose the workspace base image and tooling to start from.',
  },
  {
    title: 'Configure resources',
    subtitle:
      'Adjust compute, storage, and runtime settings to match your workload.',
  },
  {
    title: 'Review & launch',
    subtitle: 'Confirm the details before provisioning the workspace.',
  },
];

const buildRandomId = (prefix = 'ws'): string => {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return `${prefix}-${crypto.randomUUID().slice(0, 8)}`;
  }
  return `${prefix}-${Math.random().toString(36).slice(2, 10)}`;
};

const createDefaultFormState = (): WizardFormState => ({
  templateId: undefined,
  projectId: 'p-demo',
  workloadId: buildRandomId(),
  clusterId: clusterOptions[0]?.id ?? '',
  queue: 'default',
  flavorId: '',
  image: '',
  ports: DEFAULT_PORTS.join(', '),
  env: '',
  storageGiB: 50,
  runtimeHours: 4,
  autoShutdown: true,
});

const isSameErrorMap = (a: StepConfigureErrors, b: StepConfigureErrors) => {
  const keys = new Set([...Object.keys(a), ...Object.keys(b)]);
  for (const key of keys) {
    if (a[key as keyof StepConfigureErrors] !== b[key as keyof StepConfigureErrors]) {
      return false;
    }
  }
  return true;
};

const validateConfigure = (form: WizardFormState): StepConfigureErrors => {
  const errors: StepConfigureErrors = {};

  if (!form.projectId.trim()) {
    errors.projectId = 'Project is required';
  }

  if (!/^[a-z0-9-]+$/i.test(form.workloadId.trim())) {
    errors.workloadId = 'Workspace ID may include letters, numbers, and hyphen';
  }

  if (!form.image.trim()) {
    errors.image = 'Container image is required';
  }

  const portsError = validatePortsInput(form.ports);
  if (portsError) {
    errors.ports = portsError;
  }

  const envError = validateEnvInput(form.env);
  if (envError) {
    errors.env = envError;
  }

  return errors;
};

const computeCanProceedFromConfigure = (
  form: WizardFormState,
  flavor: WorkspaceFlavor | undefined,
  errors: StepConfigureErrors,
): boolean => {
  if (!flavor) {
    return false;
  }
  if (form.storageGiB < 10) {
    return false;
  }
  if (form.autoShutdown && form.runtimeHours <= 0) {
    return false;
  }
  if (!form.clusterId) {
    return false;
  }
  if (!form.workloadId.trim() || !form.projectId.trim() || !form.image.trim()) {
    return false;
  }
  return Object.keys(errors).length === 0;
};

export const WorkspaceWizard = () => {
  const classes = useStyles();
  const [activeStep, setActiveStep] = useState<WizardStep>(0);
  const [form, setForm] = useState<WizardFormState>(createDefaultFormState);
  const [advancedOpen, setAdvancedOpen] = useState(false);
  const [configureErrors, setConfigureErrors] = useState<StepConfigureErrors>({});
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [submittedWorkload, setSubmittedWorkload] = useState<WorkloadDTO | null>(
    null,
  );
  const [stepAnnouncement, setStepAnnouncement] = useState(stepsMeta[0].title);

  const alertApi = useApi(alertApiRef);
  const fetchApi = useApi(fetchApiRef);
  const discoveryApi = useApi(discoveryApiRef);
  const identityApi = useApi(identityApiRef);
  const navigate = useNavigate();
  const workloadsLink = useRouteRef(workloadsRouteRef);
  const workloadDetailsLink = useRouteRef(workloadDetailsRouteRef);

  const selectedTemplate = useMemo(
    () => templates.find(template => template.id === form.templateId),
    [form.templateId],
  );
  const selectedFlavor = useMemo(
    () => flavors.find(flavor => flavor.id === form.flavorId),
    [form.flavorId],
  );
  const selectedCluster = useMemo(
    () => clusterOptions.find(cluster => cluster.id === form.clusterId),
    [form.clusterId],
  );

  const forcedAdvanced = Boolean(selectedTemplate?.autoShowAdvanced);

  useEffect(() => {
    setStepAnnouncement(stepsMeta[activeStep].title);
  }, [activeStep]);

  useEffect(() => {
    if (activeStep === 1 && Object.keys(configureErrors).length > 0) {
      const nextErrors = validateConfigure(form);
      if (!isSameErrorMap(configureErrors, nextErrors)) {
        setConfigureErrors(nextErrors);
      }
    }
  }, [form, activeStep, configureErrors]);

  const handleTemplateSelect = (templateId: TemplateId) => {
    const template = templates.find(item => item.id === templateId);
    if (!template) {
      return;
    }

    setForm(prev => {
      const defaults = template.defaults ?? {};
      const nextPorts = defaults.ports ?? DEFAULT_PORTS;
      return {
        ...prev,
        templateId: template.id,
        flavorId: (defaults.flavor as FlavorId | undefined) ?? prev.flavorId,
        queue: defaults.queue ?? prev.queue,
        image: defaults.image ?? prev.image,
        ports: nextPorts.join(', '),
        env: defaults.env ? formatEnvMap(defaults.env) : prev.env,
      };
    });

    setAdvancedOpen(prev => (template.autoShowAdvanced ? true : prev));
    setSubmitError(null);
  };

  const handleFlavorSelect = (flavorId: FlavorId) => {
    setForm(prev => ({ ...prev, flavorId }));
    setSubmitError(null);
  };

  const updateFormField = <K extends keyof WizardFormState>(
    field: K,
    value: WizardFormState[K],
  ) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  const handleAdvancedToggle = () => {
    if (forcedAdvanced) {
      return;
    }
    setAdvancedOpen(prev => !prev);
  };

  const handleNext = () => {
    if (activeStep === 0) {
      if (!form.templateId) {
        return;
      }
      setActiveStep(1);
      return;
    }

    if (activeStep === 1) {
      const errors = validateConfigure(form);
      const nextErrors = { ...errors };
      setConfigureErrors(nextErrors);
      if (!computeCanProceedFromConfigure(form, selectedFlavor, nextErrors)) {
        return;
      }
      setActiveStep(2);
      return;
    }
  };

  const handleBack = () => {
    if (activeStep === 0) {
      return;
    }
    if (activeStep === 2) {
      setActiveStep(1);
      return;
    }
    setActiveStep(0);
  };

  const handleSubmit = async () => {
    if (!selectedFlavor) {
      return;
    }

    const errors = validateConfigure(form);
    setConfigureErrors(errors);
    if (!computeCanProceedFromConfigure(form, selectedFlavor, errors)) {
      setActiveStep(1);
      return;
    }

    const request: SubmitWorkspaceRequest = {
      id: form.workloadId.trim(),
      projectId: form.projectId.trim(),
      queue: form.queue.trim() || undefined,
      workspace: {
        flavor: selectedFlavor.flavor,
        image: form.image.trim(),
        ports: parsePortsInput(form.ports),
        env: parseEnvInput(form.env),
        interactive: true,
        maxDurationSeconds: form.autoShutdown
          ? Math.round(form.runtimeHours * 3600)
          : undefined,
      },
    };

    setSubmitting(true);
    setSubmitError(null);
    try {
      const response = await submitWorkspace(
        fetchApi,
        discoveryApi,
        identityApi,
        request,
      );
      setSubmittedWorkload(response);
      alertApi.post({
        message: `Workspace launch requested: ${response.id ?? ''}`,
        severity: 'success',
      });
    } catch (err) {
      setSubmitError(err instanceof Error ? err.message : String(err));
    } finally {
      setSubmitting(false);
    }
  };

  const resetWizard = () => {
    setForm(createDefaultFormState());
    setAdvancedOpen(false);
    setConfigureErrors({});
    setSubmitError(null);
    setSubmittedWorkload(null);
    setActiveStep(0);
  };

  const configureValidation = useMemo(
    () => validateConfigure(form),
    [form],
  );

  const canGoNext = useMemo(() => {
    if (activeStep === 0) {
      return Boolean(form.templateId);
    }
    if (activeStep === 1) {
      return computeCanProceedFromConfigure(
        form,
        selectedFlavor,
        Object.keys(configureErrors).length > 0 ? configureErrors : configureValidation,
      );
    }
    return true;
  }, [activeStep, form, selectedFlavor, configureErrors, configureValidation]);

  return (
    <Page themeId="tool">
      <Header title="Launch Interactive Workspace" />
      <Content>
        <ContentHeader title="Workspace Wizard" />
        <Stepper activeStep={activeStep} alternativeLabel>
          {stepsMeta.map(step => (
            <Step key={step.title}>
              <StepLabel>{step.title}</StepLabel>
            </Step>
          ))}
        </Stepper>

        <Box mt={4} mb={3}>
          <Typography
            variant="srOnly"
            component="p"
            className={classes.srOnly}
            aria-live="polite"
          >
            {stepAnnouncement}
          </Typography>
          <WizardHeader
            title={stepsMeta[activeStep].title}
            subtitle={stepsMeta[activeStep].subtitle}
            stepLabel={`Step ${activeStep + 1} of ${stepsMeta.length}`}
          />
        </Box>

        {activeStep === 0 ? (
          <StepTemplates
            templates={templates}
            selectedTemplateId={form.templateId}
            onSelect={handleTemplateSelect}
          />
        ) : null}

        {activeStep === 1 ? (
          <StepConfigure
            form={form}
            selectedTemplate={selectedTemplate}
            flavors={flavors}
            clusterOptions={clusterOptions}
            errors={configureErrors}
            advancedOpen={advancedOpen}
            forceAdvancedOpen={forcedAdvanced}
            onAdvancedToggle={handleAdvancedToggle}
            onFieldChange={updateFormField}
            onFlavorSelect={handleFlavorSelect}
          />
        ) : null}

        {activeStep === 2 ? (
          <StepReview
            form={form}
            selectedTemplate={selectedTemplate}
            selectedFlavor={selectedFlavor}
            selectedCluster={selectedCluster}
            onEdit={step => {
              setActiveStep(step);
            }}
          />
        ) : null}

        {submitError ? (
          <Box mt={3}>
            <WarningPanel severity="error" title="Workspace submission failed">
              {submitError}
            </WarningPanel>
          </Box>
        ) : null}

        {submittedWorkload ? (
          <Alert severity="success" className={classes.successBanner}>
            Workspace <strong>{submittedWorkload.id}</strong> was submitted. Use
            the links below to monitor provisioning or start another workspace.
          </Alert>
        ) : null}

        <div className={classes.navControls}>
          <div className={classes.buttonRow}>
            <Button
              variant="outlined"
              onClick={handleBack}
              disabled={activeStep === 0 || submitting}
            >
              Back
            </Button>
            {activeStep < 2 ? (
              <Button
                color="primary"
                variant="contained"
                onClick={handleNext}
                disabled={!canGoNext || submitting}
              >
                Next
              </Button>
            ) : (
              <Button
                color="primary"
                variant="contained"
                onClick={handleSubmit}
                disabled={submitting || Boolean(submittedWorkload)}
              >
                Launch Workspace
              </Button>
            )}
            {submitting ? <Progress /> : null}
          </div>
          <div className={classes.buttonRow}>
            {submittedWorkload ? (
              <>
                <Button
                  variant="outlined"
                  color="primary"
                  onClick={() => navigate(workloadsLink())}
                >
                  View Workspaces
                </Button>
                {submittedWorkload.id ? (
                  <Button
                    variant="outlined"
                    color="primary"
                    onClick={() =>
                      navigate(
                        workloadDetailsLink({ id: submittedWorkload.id ?? '' }),
                      )
                    }
                  >
                    View Details
                  </Button>
                ) : null}
                <Button onClick={resetWizard}>Launch another</Button>
              </>
            ) : null}
          </div>
        </div>
      </Content>
    </Page>
  );
};
