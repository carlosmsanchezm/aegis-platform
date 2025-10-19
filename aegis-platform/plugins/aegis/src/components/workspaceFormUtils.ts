import { WORKSPACE_DEFAULT_ENV } from '../api/aegisClient';

export type EnvMap = Record<string, string>;

const PORT_TOKENIZER = /[,\\s]+/;

export const parsePortsInput = (value: string): number[] =>
  value
    .split(PORT_TOKENIZER)
    .map(token => token.trim())
    .filter(Boolean)
    .map(token => Number.parseInt(token, 10))
    .filter(port => Number.isFinite(port) && port > 0);

export const validatePortsInput = (value: string): string | null => {
  const tokens = value
    .split(PORT_TOKENIZER)
    .map(token => token.trim())
    .filter(Boolean);

  for (const token of tokens) {
    const parsed = Number.parseInt(token, 10);
    if (!Number.isInteger(parsed) || parsed <= 0 || parsed > 65535) {
      return `Port "${token}" must be a positive integer`;
    }
  }

  return null;
};

export const parseEnvInput = (value: string): EnvMap => {
  const result: EnvMap = {};
  value
    .split('\n')
    .map(line => line.trim())
    .filter(Boolean)
    .forEach(line => {
      const [rawKey, ...rest] = line.split('=');
      if (!rawKey) {
        return;
      }
      const key = rawKey.trim();
      if (!key) {
        return;
      }
      result[key] = rest.join('=').trim();
    });
  return result;
};

export const validateEnvInput = (value: string): string | null => {
  const lines = value.split('\n');
  for (const rawLine of lines) {
    const line = rawLine.trim();
    if (!line) {
      continue;
    }
    if (!line.includes('=')) {
      return `Line "${line}" must be in KEY=VALUE format`;
    }
    const [key] = line.split('=');
    if (!key || !key.trim()) {
      return `Environment variable name is required in "${line}"`;
    }
  }
  return null;
};

export const formatEnvMap = (env: EnvMap): string =>
  Object.entries(env)
    .map(([key, value]) => `${key}=${value}`)
    .join('\n');

export const formatDefaultEnv = (): string =>
  formatEnvMap(WORKSPACE_DEFAULT_ENV);
