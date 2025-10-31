export type TemplateId =
  | 'vscode-python'
  | 'vscode-data'
  | 'jupyter-pytorch'
  | 'jupyter-rapids'
  | 'cli-ops'
  | 'custom';

export type TemplateDefaults = {
  flavor?: string;
  image?: string;
  queue?: string;
  ports?: number[];
  env?: Record<string, string>;
};

export type WorkspaceTemplate = {
  id: TemplateId;
  name: string;
  description: string;
  highlights?: string[];
  icon: 'python' | 'data' | 'pytorch' | 'spark' | 'terminal' | 'custom';
  defaults: TemplateDefaults;
  autoShowAdvanced?: boolean;
};

export type FlavorId =
  | 'cpu-small'
  | 'cpu-medium'
  | 'cpu-large'
  | 'gpu-standard'
  | 'gpu-large';

export type WorkspaceFlavor = {
  id: FlavorId;
  name: string;
  description: string;
  flavor: string;
  resources: string;
  category: 'cpu' | 'gpu';
  gpu?: {
    model: string;
    memory?: string;
    count: number;
  };
};

export type ClusterOption = {
  id: string;
  name: string;
  description?: string;
};

type BaseFormState = {
  projectId: string;
  workloadId: string;
  clusterId: string;
  queue: string;
  flavorId: FlavorId | '';
  image: string;
  ports: string;
  env: string;
  storageGiB: number;
  runtimeHours: number;
  autoShutdown: boolean;
  gpuCount?: number;
};

export type WizardFormState = BaseFormState & {
  templateId?: TemplateId;
};

export type WizardStep = 0 | 1 | 2;
