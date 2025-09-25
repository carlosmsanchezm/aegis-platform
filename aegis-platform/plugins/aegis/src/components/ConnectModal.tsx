import { FC, useMemo } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
} from '@material-ui/core';
import { CopyTextButton, Progress, WarningPanel } from '@backstage/core-components';
import { ConnectionDetails } from '../api/aegisClient';

type Props = {
  open: boolean;
  onClose: () => void;
  loading: boolean;
  error?: string | null;
  details?: ConnectionDetails | null;
  workloadId: string;
};

const formatExpiry = (expiresAtUtc?: string): string => {
  if (!expiresAtUtc) {
    return 'Unknown';
  }
  const date = new Date(expiresAtUtc);
  if (Number.isNaN(date.getTime())) {
    return expiresAtUtc;
  }
  return `${date.toLocaleString()} (${date.toISOString()})`;
};

const buildClientSnippet = (details: ConnectionDetails): string => {
  const user = details.sshUsername && details.sshUsername.trim() !== '' ? details.sshUsername : 'vscode';
  return `Host ${details.sshHostAlias}
  HostName ${details.sshHostAlias}
  User ${user}
  Port ${details.destPort}
  ProxyCommand aegis-proxy-client --proxy ${details.proxyUrl} --token '${details.token}'
`;
};

const buildFallbackSnippet = (details: ConnectionDetails, workloadId: string): string => {
  let hostHeader = details.proxyUrl;
  let connectTarget = details.proxyUrl;
  let serverName = details.proxyUrl;
  let path = `/proxy/${workloadId}`;

  try {
    const parsed = new URL(details.proxyUrl);
    hostHeader = parsed.host;
    serverName = parsed.hostname;
    const port = parsed.port || (parsed.protocol === 'https:' ? '443' : '80');
    connectTarget = `${serverName}:${port}`;
    path = parsed.pathname || path;
  } catch (e) {
    // Ignore parsing errors and fall back to best-effort defaults.
  }

  const user = details.sshUsername && details.sshUsername.trim() !== '' ? details.sshUsername : 'vscode';
  return `Host ${details.sshHostAlias}
  HostName ${details.sshHostAlias}
  User ${user}
  Port ${details.destPort}
  ProxyCommand /bin/sh -lc 'printf "CONNECT ${path} HTTP/1.1\\r\\nHost: ${hostHeader}\\r\\nAuthorization: Bearer ${details.token}\\r\\n\\r\\n"; cat' | openssl s_client -quiet -connect ${connectTarget} -servername ${serverName}
`;
};

export const ConnectModal: FC<Props> = ({
  open,
  onClose,
  loading,
  error,
  details,
  workloadId,
}) => {
  const clientSnippet = useMemo(() => (details ? buildClientSnippet(details) : ''), [details]);
  const fallbackSnippet = useMemo(
    () => (details ? buildFallbackSnippet(details, workloadId) : ''),
    [details, workloadId],
  );

  const expiresLabel = details ? formatExpiry(details.expiresAtUtc) : '';
  const token = details?.token ?? '';

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Connect to Workspace</DialogTitle>
      <DialogContent dividers>
        {loading && <Progress />}

        {!loading && error && (
          <WarningPanel severity="error" title="Failed to mint token">
            {error}
          </WarningPanel>
        )}

        {!loading && !error && details && (
          <Box display="flex" flexDirection="column" gridGap={16}>
            <Box>
              <Typography variant="body1">
                Token expires at {expiresLabel}. Tokens are single-use; request a new one if
                a connection fails.
              </Typography>
            </Box>

            <Box>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Typography variant="subtitle1">SSH config (with aegis-proxy-client)</Typography>
                <CopyTextButton text={clientSnippet} tooltip="Copy SSH config" />
              </Box>
              <Box
                component="pre"
                bgcolor="#0e1117"
                color="#f8f8f2"
                padding={2}
                borderRadius={4}
                style={{ overflowX: 'auto' }}
              >
                {clientSnippet}
              </Box>
            </Box>

            <Box>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Typography variant="subtitle1">Fallback (OpenSSL CONNECT)</Typography>
                <CopyTextButton text={fallbackSnippet} tooltip="Copy fallback config" />
              </Box>
              <Box
                component="pre"
                bgcolor="#0e1117"
                color="#f8f8f2"
                padding={2}
                borderRadius={4}
                style={{ overflowX: 'auto' }}
              >
                {fallbackSnippet}
              </Box>
            </Box>

            <Box>
              <Box display="flex" alignItems="center" justifyContent="space-between">
                <Typography variant="subtitle1">JWT token</Typography>
                <CopyTextButton text={token} tooltip="Copy token" />
              </Box>
              <Box
                component="pre"
                bgcolor="#0e1117"
                color="#f8f8f2"
                padding={2}
                borderRadius={4}
                style={{ overflowX: 'auto' }}
              >
                {token}
              </Box>
            </Box>
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} color="primary">
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};
