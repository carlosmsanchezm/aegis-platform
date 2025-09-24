import { DiscoveryApi, FetchApi } from '@backstage/core-plugin-api';

export type WorkspaceSpec = {
  flavor?: string;
  image?: string;
  command?: string[];
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
  method: string,
  body: TReq,
): Promise<TRes> => {
  const url = await buildProxyUrl(discoveryApi, method);
  const response = await fetchApi.fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
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
  projectId: string,
): Promise<WorkloadDTO[]> => {
  const res = await postJson<{ projectId: string }, ListWorkloadsResponse>(
    fetchApi,
    discoveryApi,
    'aegis.v1.AegisPlatform/ListWorkloads',
    { projectId },
  );
  return res.items ?? [];
};

export const getWorkload = async (
  fetchApi: FetchApi,
  discoveryApi: DiscoveryApi,
  id: string,
): Promise<WorkloadDTO> =>
  postJson<{ id: string }, WorkloadDTO>(
    fetchApi,
    discoveryApi,
    'aegis.v1.AegisPlatform/GetWorkload',
    { id },
  );

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

export const parseKubernetesUrl = (url?: string): KubernetesLocation | undefined => {
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

export const buildKubectlDescribeCommand = (loc?: KubernetesLocation): string => {
  if (!loc) {
    return '';
  }
  return `kubectl -n ${loc.namespace} describe ${loc.kind} ${loc.name}`;
};
