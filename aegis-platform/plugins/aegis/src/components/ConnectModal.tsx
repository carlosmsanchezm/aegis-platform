import { FC, useMemo } from 'react';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Box,
  Typography,
  Link,
} from '@material-ui/core';
import { CopyTextButton, Progress, WarningPanel } from '@backstage/core-components';
import { ConnectionSession } from '../api/aegisClient';

type Props = {
  open: boolean;
  onClose: () => void;
  loading: boolean;
  error?: string | null;
  session: ConnectionSession | null;
  pendingSession: boolean;
  helperInstalled: boolean;
  onConfirmHelper: () => void;
  systemAcked: boolean;
  onAcknowledgeSystemUse: () => void;
  rulesAcked: boolean;
  onAcknowledgeRules: () => void;
  onRequestSession: (client: 'cli' | 'vscode') => void;
  onRenew: () => void;
  onRevoke: () => void;
  workloadId: string;
};

const helperDocUrl = 'https://docs.yourorg.dev/aegis/connect-helper';

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

const buildSshCommand = (session: ConnectionSession): string => {
  const user = session.sshUser && session.sshUser.trim() !== '' ? session.sshUser : 'aegis';
  return `ssh ${user}@${session.sshHostAlias} -o ProxyCommand="aegis-connect --proxy=${session.proxyUrl} --token=${session.token}"`;
};

export const ConnectModal: FC<Props> = ({
  open,
  onClose,
  loading,
  error,
  session,
  pendingSession,
  helperInstalled,
  onConfirmHelper,
  systemAcked,
  onAcknowledgeSystemUse,
  rulesAcked,
  onAcknowledgeRules,
  onRequestSession,
  onRenew,
  onRevoke,
  workloadId,
}) => {
  const sshCommand = useMemo(
    () => (session ? buildSshCommand(session) : ''),
    [session],
  );

  const hasSession = Boolean(session);
  const needsSystemAck = !systemAcked;
  const needsRulesAck = systemAcked && !rulesAcked;
  const needsHelper = systemAcked && rulesAcked && !helperInstalled;
  const readyToMint = systemAcked && rulesAcked && helperInstalled && !hasSession;

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="md">
      <DialogTitle>Secure Workspace Access</DialogTitle>
      <DialogContent dividers>
        {loading && <Progress />}

        {!loading && error && (
          <WarningPanel severity="error" title="Failed to mint session">
            {error}
          </WarningPanel>
        )}

        {!loading && !error && (
          <Box display="flex" flexDirection="column" gridGap={16}>
            {needsSystemAck && (
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Typography variant="h6">System Use Notification</Typography>
                <Typography variant="body2">
                  You are accessing a U.S. Federal information system. Usage is monitored and
                  recorded. Unauthorized use may result in disciplinary action and prosecution. By
                  clicking Continue, you acknowledge that you are an authorized user and consent to
                  monitoring in accordance with agency policy.
                </Typography>
                <Box display="flex" justifyContent="flex-end">
                  <Button color="primary" variant="contained" onClick={onAcknowledgeSystemUse}>
                    I Understand and Accept
                  </Button>
                </Box>
              </Box>
            )}

            {needsRulesAck && (
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Typography variant="h6">Rules of Behavior</Typography>
                <Typography variant="body2">
                  Remote sessions must comply with Aegis acceptable use policies. Do not transfer
                  sensitive data to unauthorized systems, and ensure all activity remains within the
                  approved workspace boundary. You are responsible for protecting credentials and
                  reporting suspected incidents immediately.
                </Typography>
                <Box display="flex" justifyContent="flex-end">
                  <Button color="primary" variant="contained" onClick={onAcknowledgeRules}>
                    I Agree to the Rules of Behavior
                  </Button>
                </Box>
              </Box>
            )}

            {needsHelper && (
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Typography variant="h6">Install the Aegis Connect Helper</Typography>
                <Typography variant="body2">
                  The helper binary brokers the one-time token into an SSH proxy command. Download
                  the latest release for your platform and verify the published checksum before
                  installing.
                </Typography>
                <Link href={helperDocUrl} target="_blank" rel="noopener">
                  View installation instructions and checksums
                </Link>
                <Box display="flex" justifyContent="flex-end">
                  <Button color="primary" variant="contained" onClick={onConfirmHelper}>
                    I Installed the Helper
                  </Button>
                </Box>
              </Box>
            )}

            {readyToMint && (
              <Box display="flex" flexDirection="column" gridGap={12}>
                <Typography variant="h6">Mint a Connection Session</Typography>
                <Typography variant="body2">
                  Generate a one-time connection token for workload <strong>{workloadId}</strong>.
                  Tokens expire within minutes and are invalidated immediately after use.
                </Typography>
                <Box display="flex" gridGap={8}>
                  <Button
                    color="primary"
                    variant="contained"
                    onClick={() => onRequestSession('cli')}
                    disabled={loading}
                  >
                    Generate CLI Session
                  </Button>
                  <Button
                    color="primary"
                    onClick={() => onRequestSession('vscode')}
                    disabled={loading}
                  >
                    Generate VS Code Session
                  </Button>
                </Box>
                {pendingSession && (
                  <Typography variant="caption" color="textSecondary">
                    Complete the steps above to mint a session.
                  </Typography>
                )}
              </Box>
            )}

            {hasSession && session && (
              <Box display="flex" flexDirection="column" gridGap={16}>
                <Box>
                  <Typography variant="body1">
                    Session <strong>{session.sessionId}</strong> expires at {formatExpiry(session.expiresAtUtc)}.
                    This token is single-use and cannot be reused once the proxy validates it.
                  </Typography>
                </Box>

                {session.vscodeUri && (
                  <Box display="flex" gridGap={8}>
                    <Button
                      color="primary"
                      variant="contained"
                      onClick={() => window.open(session.vscodeUri, '_blank')}
                    >
                      Open in VS Code
                    </Button>
                  </Box>
                )}

                {session.sshConfig && (
                  <Box>
                    <Box display="flex" alignItems="center" justifyContent="space-between">
                      <Typography variant="subtitle1">Ephemeral SSH Config</Typography>
                      <CopyTextButton text={session.sshConfig} tooltip="Copy SSH config" />
                    </Box>
                    <Box
                      component="pre"
                      bgcolor="#0e1117"
                      color="#f8f8f2"
                      padding={2}
                      borderRadius={4}
                      style={{ overflowX: 'auto' }}
                    >
                      {session.sshConfig}
                    </Box>
                  </Box>
                )}

                <Box>
                  <Box display="flex" alignItems="center" justifyContent="space-between">
                    <Typography variant="subtitle1">SSH command</Typography>
                    <CopyTextButton text={sshCommand} tooltip="Copy SSH command" />
                  </Box>
                  <Box
                    component="pre"
                    bgcolor="#0e1117"
                    color="#f8f8f2"
                    padding={2}
                    borderRadius={4}
                    style={{ overflowX: 'auto' }}
                  >
                    {sshCommand}
                  </Box>
                </Box>

                <Box>
                  <Box display="flex" alignItems="center" justifyContent="space-between">
                    <Typography variant="subtitle1">Session token</Typography>
                    <CopyTextButton text={session.token} tooltip="Copy token" />
                  </Box>
                  <Box
                    component="pre"
                    bgcolor="#0e1117"
                    color="#f8f8f2"
                    padding={2}
                    borderRadius={4}
                    style={{ overflowX: 'auto' }}
                  >
                    {session.token}
                  </Box>
                </Box>

                <Box display="flex" gridGap={8}>
                  <Button color="primary" variant="outlined" onClick={onRenew} disabled={loading}>
                    Renew Session
                  </Button>
                  <Button color="secondary" onClick={onRevoke} disabled={loading}>
                    Revoke Session
                  </Button>
                </Box>
              </Box>
            )}
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose} color="primary" disabled={loading}>
          Close
        </Button>
      </DialogActions>
    </Dialog>
  );
};
