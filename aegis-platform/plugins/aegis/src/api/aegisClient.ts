import {
  DiscoveryApi,
  FetchApi,
  IdentityApi,
} from '@backstage/core-plugin-api';

const base64UrlDecode = (input: string): string => {
  const normalized = input.replace(/-/g, '+').replace(/_/g, '/');
  if (typeof window !== 'undefined' && typeof window.atob === 'function') {
    return window.atob(normalized);
  }
  if (
    typeof globalThis !== 'undefined' &&
    typeof (globalThis as any).atob === 'function'
  ) {
    return (globalThis as any).atob(normalized);
  }
  if (typeof Buffer !== 'undefined') {
    return Buffer.from(normalized, 'base64').toString('utf-8');
  }
  throw new Error('Base64 decoding is not supported in this environment');
};

export type WorkspaceSpec = {
  flavor?: string;
  image?: string;
  command?: string[];
  interactive?: boolean;
  ports?: number[];
  env?: Record<string, string>;
  maxDurationSeconds?: number;
};

export type TrainingSpec = {
  flavor?: string;
  image?: string;
  command?: string[];
};

export type WorkloadDTO = {
  id?: string;
  projectId?: string;
  queue?: string;
  clusterId?: string;
  status?: string;
  uiStatus?: string;
  message?: string;
  url?: string;
  workspace?: WorkspaceSpec;
  training?: TrainingSpec;
};

export type ConnectionSession = {
  sessionId: string;
  token: string;
  sshUser?: string;
  sshHostAlias: string;
  vscodeUri?: string;
  sshConfig?: string;
  proxyUrl: string;
  expiresAtUtc: string;
  oneTime?: boolean;
};

export type SubmitWorkspaceRequest = {
  id?: string;
  projectId: string;
  queue?: string;
  workspace: {
    flavor?: string;
    image?: string;
    command?: string | string[];
    interactive?: boolean;
    ports?: number[];
    env?: Record<string, string>;
    maxDurationSeconds?: number;
  };
};

export type ListWorkloadsResponse = {
  items: WorkloadDTO[];
};

export const DEFAULT_SSH_PORT = 22;
export const DEFAULT_VSCODE_PORT = 11111;

export const WORKSPACE_DEFAULT_ENV: Record<string, string> = {
  VSCODE_QUALITY: 'stable',
  PUID: '1000',
  PGID: '1000',
  PASSWORD_ACCESS: 'true',
  USER_NAME: 'aegis',
  USER_PASSWORD: 'aegis123',
};

const normalizeCommand = (command: string | string[] | undefined): string[] => {
  if (Array.isArray(command)) {
    const sanitized = command
      .map(part => (typeof part === 'string' ? part.trim() : ''))
      .filter(part => part.length > 0);
    if (sanitized.length > 0) {
      return sanitized;
    }
  }

  const text = typeof command === 'string' ? command.trim() : '';
  if (!text) {
    return [];
  }
  return ['sh', '-c', text];
};

export const ensureWorkspacePorts = (ports?: number[]): number[] => {
  const portSet = new Set<number>();
  (ports ?? []).forEach(port => {
    if (Number.isFinite(port) && port > 0) {
      portSet.add(Math.trunc(port));
    }
  });

  if (portSet.size === 0) {
    portSet.add(DEFAULT_SSH_PORT);
  }
  portSet.add(DEFAULT_VSCODE_PORT);

  return Array.from(portSet).sort((a, b) => a - b);
};

export const mergeWorkspaceEnv = (
  user?: Record<string, string>,
): Record<string, string> | undefined => {
  const merged: Record<string, string> = {};

  Object.entries(WORKSPACE_DEFAULT_ENV).forEach(([key, value]) => {
    if (value) {
      merged[key] = value;
    }
  });

  if (user) {
    Object.entries(user).forEach(([key, value]) => {
      const trimmedKey = key?.trim();
      if (!trimmedKey) {
        return;
      }
      const normalizedValue =
        typeof value === 'string' ? value.trim() : String(value ?? '');
      if (!normalizedValue) {
        return;
      }
      merged[trimmedKey] = normalizedValue;
    });
  }

  return Object.keys(merged).length > 0 ? merged : undefined;
};

const buildProxyUrl = async (
  discoveryApi: DiscoveryApi,
  method: string,
): Promise<string> => {
  const baseUrl = await discoveryApi.getBaseUrl('proxy');
  return `${baseUrl}/aegis/${method}`;
};

const postJson = async <TReq extends object, TRes>(
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  method: string,
  body: TReq,
  options?: { requireAuth?: boolean },
): Promise<TRes> => {
  const url = await buildProxyUrl(discoveryApi, method);

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
  };

  let identityToken: string | undefined;
  let userEntityRef: string | undefined;

  try {
    const identity = await identityApi.getBackstageIdentity({ optional: true });
    identityToken = identity?.token;
    userEntityRef = identity?.userEntityRef;
  } catch (err) {
    if (process.env.NODE_ENV === 'development') {
      // eslint-disable-next-line no-console
      console.warn('Aegis: unable to obtain Backstage identity', err);
    }
  }

  if (!identityToken) {
    try {
      const credentials = await identityApi.getCredentials();
      identityToken = credentials.token;
      userEntityRef = userEntityRef ?? credentials.userEntityRef;
    } catch (err) {
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.warn('Aegis: failed to resolve credentials', err);
      }
    }
  }

  let subject = userEntityRef;

  if (!subject && identityToken) {
    try {
      const [, payload] = identityToken.split('.');
      if (payload) {
        const json = JSON.parse(base64UrlDecode(payload));
        subject = json.sub || json.email || json.preferred_username;
      }
    } catch (err) {
      if (process.env.NODE_ENV === 'development') {
        // eslint-disable-next-line no-console
        console.warn('Aegis: failed to parse identity token', err);
      }
    }
  }

  if (identityToken) {
    headers.Authorization = `Bearer ${identityToken}`;
    headers['Grpc-Metadata-Authorization'] = `Bearer ${identityToken}`;
    headers['grpc-metadata-authorization'] = `Bearer ${identityToken}`;
  } else if (options?.requireAuth) {
    throw new Error(
      'Authentication is required to mint a connection session. Refresh the page and sign in.',
    );
  }

  if (!subject) {
    subject = 'user:default/guest';
  }

  if (subject) {
    headers['x-aegis-user'] = subject;
    headers['Grpc-Metadata-X-Aegis-User'] = subject;
    headers['grpc-metadata-x-aegis-user'] = subject;
  }

  const response = await fetchApi.fetch(url, {
    method: 'POST',
    headers,
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || `Request failed with status ${response.status}`);
  }

  return (await response.json()) as TRes;
};

export const listWorkloads = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  projectId: string,
): Promise<WorkloadDTO[]> => {
  const res = await postJson<{ projectId: string }, ListWorkloadsResponse>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/ListWorkloads',
    { projectId },
    { requireAuth: false },
  );
  return res.items ?? [];
};

export const getWorkload = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  id: string,
): Promise<WorkloadDTO> =>
  postJson<{ id: string }, WorkloadDTO>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/GetWorkload',
    { id },
    { requireAuth: false },
  );

export const createConnectionSession = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  workloadId: string,
  client: 'cli' | 'ssh' | 'vscode' = 'cli',
): Promise<ConnectionSession> =>
  postJson<{ workloadId: string; client: string }, ConnectionSession>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/CreateConnectionSession',
    { workloadId, client },
    { requireAuth: true },
  );

export const renewConnectionSession = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  sessionId: string,
): Promise<ConnectionSession> =>
  postJson<{ sessionId: string }, ConnectionSession>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/RenewConnectionSession',
    { sessionId },
    { requireAuth: true },
  );

export const revokeConnectionSession = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  sessionId: string,
): Promise<void> => {
  await postJson<{ sessionId: string }, unknown>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/RevokeConnectionSession',
    { sessionId },
    { requireAuth: true },
  );
};

export const submitWorkspace = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  identityApi: IdentityApi,
  req: SubmitWorkspaceRequest,
): Promise<WorkloadDTO> => {
  const workspaceInput = req.workspace ?? {};
  const command = normalizeCommand(workspaceInput.command);
  const ports = ensureWorkspacePorts(workspaceInput.ports);
  const env = mergeWorkspaceEnv(workspaceInput.env);
  const maxDuration =
    typeof workspaceInput.maxDurationSeconds === 'number' &&
    Number.isFinite(workspaceInput.maxDurationSeconds)
      ? Math.floor(workspaceInput.maxDurationSeconds)
      : undefined;

  const body = {
    workload: {
      ...(req.id ? { id: req.id } : {}),
      projectId: req.projectId,
      ...(req.queue ? { queue: req.queue } : {}),
      workspace: {
        flavor: workspaceInput.flavor,
        image: workspaceInput.image,
        interactive: true,
        command,
        ports,
        ...(env ? { env } : {}),
        ...(maxDuration ? { maxDurationSeconds: maxDuration } : {}),
      },
    },
  };
  return postJson<typeof body, WorkloadDTO>(
    fetchApi,
    discoveryApi,
    identityApi,
    'aegis.v1.AegisPlatform/SubmitWorkload',
    body,
    { requireAuth: true },
  );
};

export const getFlavor = (w: WorkloadDTO): string =>
  w?.workspace?.flavor ?? w?.training?.flavor ?? '';

export const isTerminalStatus = (status?: string): boolean =>
  status === 'SUCCEEDED' || status === 'FAILED';

type DisplayStatus = {
  label: string;
  color: 'ok' | 'warning' | 'error' | 'progress';
};

export const mapDisplayStatus = (raw?: string): DisplayStatus => {
  switch (raw) {
    case 'RUNNING':
      return { label: 'Running', color: 'progress' };
    case 'SUCCEEDED':
      return { label: 'Succeeded', color: 'ok' };
    case 'FAILED':
      return { label: 'Failed', color: 'error' };
    case 'QUEUED_BY_KUEUE':
      return { label: 'Queued by Kueue', color: 'warning' };
    case 'SUBMITTED':
      return { label: 'Submitted', color: 'warning' };
    case 'PLACED':
    default:
      return { label: 'Queued', color: 'warning' };
  }
};

export type KubernetesLocation = {
  namespace: string;
  kind: string;
  name: string;
};

export const parseKubernetesUrl = (
  url?: string,
): KubernetesLocation | undefined => {
  if (!url || !url.startsWith('k8s://')) {
    return undefined;
  }
  // Format: k8s://namespace/kind/name (kind usually job)
  const parts = url.replace('k8s://', '').split('/');
  if (parts.length < 3) {
    return undefined;
  }
  const [namespace, kind, ...rest] = parts;
  return { namespace, kind, name: rest.join('/') };
};

export const buildKubectlDescribeCommand = (
  loc?: KubernetesLocation,
): string => {
  if (!loc) {
    return '';
  }
  return `kubectl -n ${loc.namespace} describe ${loc.kind} ${loc.name}`;
};
