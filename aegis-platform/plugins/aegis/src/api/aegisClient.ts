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
  id: string;
  projectId: string;
  queue?: string;
  workspace: WorkspaceSpec;
};

export type ListWorkloadsResponse = {
  items: WorkloadDTO[];
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
  const body = {
    workload: {
      id: req.id,
      projectId: req.projectId,
      queue: req.queue,
      kind: {
        workspace: {
          ...req.workspace,
        },
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
