import { FC, FormEvent, useEffect, useState } from 'react';
import {
  Page,
  Content,
  Header,
  ContentHeader,
  Progress,
  WarningPanel,
} from '@backstage/core-components';
import {
  useApi,
  fetchApiRef,
  alertApiRef,
  discoveryApiRef,
} from '@backstage/core-plugin-api';
import {
  Button,
  Grid,
  TextField,
  Card,
  CardContent,
  Typography,
} from '@material-ui/core';

type WorkloadResponse = {
  id?: string;
  projectId?: string;
  queue?: string;
  clusterId?: string;
  status?: string;
  url?: string;
  message?: string;
};

export const SubmitWorkloadPage: FC = () => {
  const fetchApi = useApi(fetchApiRef);
  const alertApi = useApi(alertApiRef);
  const discoveryApi = useApi(discoveryApiRef);

  // Defaults aimed at the local dev environment
  const [projectId, setProjectId] = useState('p-demo');
  const [queue, setQueue] = useState('default');
  const [flavor, setFlavor] = useState('a10-mig-1g');
  const [image, setImage] = useState('alpine:3.19');
  const [command, setCommand] = useState('echo Hello from Aegis; sleep 2');
  const [maxDurationSeconds, setMaxDurationSeconds] = useState<number | ''>(600);

  const [submitting, setSubmitting] = useState(false);
  const [result, setResult] = useState<WorkloadResponse | null>(null);
  const [error, setError] = useState<string | null>(null);

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault();
    setSubmitting(true);
    setError(null);
    setResult(null);

    const trimmed = command.trim();
    const cmd = trimmed.length ? ['sh', '-c', trimmed] : ['sh', '-c', 'echo'];
    const payload = {
      workload: {
        projectId,
        queue,
        workspace: {
          flavor,
          image,
          command: cmd,
          ...(maxDurationSeconds
            ? { maxDurationSeconds: Number(maxDurationSeconds) }
            : {}),
        },
      },
    };

    try {
      const proxyBase = await discoveryApi.getBaseUrl('proxy');
      const res = await fetchApi.fetch(
        `${proxyBase}/aegis/aegis.v1.AegisPlatform/SubmitWorkload`,
        {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload),
        },
      );

      if (!res.ok) {
        const txt = await res.text();
        const msg = txt || `HTTP ${res.status}`;
        setError(msg);
        alertApi.post({ message: `Submit failed: ${msg}`, severity: 'error' });
      } else {
        const json = (await res.json()) as WorkloadResponse;
        setResult(json);
        alertApi.post({
          message: `Workload submitted: ${json.id ?? '(no id)'}`,
          severity: 'success',
        });
      }
  } catch (err: any) {
      const msg = err?.message ?? String(err);
      setError(msg);
      alertApi.post({ message: `Submit failed: ${msg}`, severity: 'error' });
    } finally {
      setSubmitting(false);
    }
  };

  useEffect(() => {
    let timer: ReturnType<typeof setInterval> | undefined;
    let cancelled = false;

    const shouldPoll =
      Boolean(result?.id) &&
      result?.status !== 'SUCCEEDED' &&
      result?.status !== 'FAILED';
    const targetId = result?.id;

    if (shouldPoll) {
      const startPolling = async () => {
        try {
          const proxyBase = await discoveryApi.getBaseUrl('proxy');
          if (cancelled) {
            return;
          }
          timer = setInterval(async () => {
            try {
              const res = await fetchApi.fetch(
                `${proxyBase}/aegis/aegis.v1.AegisPlatform/GetWorkload`,
                {
                  method: 'POST',
                  headers: { 'Content-Type': 'application/json' },
                  body: JSON.stringify({ id: targetId }),
                },
              );
              if (!res.ok) {
                return;
              }
              const json = (await res.json()) as WorkloadResponse;
              setResult(prev => {
                if (!prev) {
                  return json;
                }
                if (
                  prev.status === json.status &&
                  prev.clusterId === json.clusterId &&
                  prev.url === json.url &&
                  prev.message === json.message
                ) {
                  return prev;
                }
                return json;
              });
            } catch {
              /* ignore polling errors */
            }
          }, 2000);
        } catch (err) {
          if (cancelled) {
            return;
          }
          const msg = (err as Error)?.message ?? String(err);
          setError(prev => prev ?? msg);
        }
      };
      startPolling();
    }

    return () => {
      cancelled = true;
      if (timer) {
        clearInterval(timer);
      }
    };
  }, [result?.id, result?.status, discoveryApi, fetchApi]);

  return (
    <Page themeId="tool">
      <Header title="Aegis — Submit Workload" subtitle="Workspace (MVP)" />
      <Content>
        <ContentHeader title="Workspace Form" />

        <form onSubmit={onSubmit}>
          <Grid container spacing={3}>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Project ID"
                value={projectId}
                onChange={e => setProjectId(e.target.value)}
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Queue"
                value={queue}
                onChange={e => setQueue(e.target.value)}
                required
              />
            </Grid>

            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Flavor"
                value={flavor}
                onChange={e => setFlavor(e.target.value)}
                required
              />
            </Grid>
            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Image"
                value={image}
                onChange={e => setImage(e.target.value)}
                required
              />
            </Grid>

            <Grid item xs={12}>
              <TextField
                fullWidth
                label='Command (wrapped as ["sh","-c", ...])'
                value={command}
                onChange={e => setCommand(e.target.value)}
                helperText="Example: echo Hello; sleep 2"
              />
            </Grid>

            <Grid item xs={12} md={6}>
              <TextField
                fullWidth
                label="Max Duration Seconds"
                type="number"
                value={maxDurationSeconds}
                onChange={e =>
                  setMaxDurationSeconds(
                    e.target.value === '' ? '' : Number(e.target.value),
                  )
                }
                helperText="Optional; defaults to queue/server if omitted"
              />
            </Grid>

            <Grid item xs={12}>
              <Button
                type="submit"
                color="primary"
                variant="contained"
                disabled={submitting}
              >
                Submit
              </Button>
            </Grid>
          </Grid>
        </form>

        {submitting && <Progress />}

        {error && (
          <WarningPanel title="Submission Error" severity="error">
            <Typography
              variant="body2"
              component="pre"
              style={{ whiteSpace: 'pre-wrap' }}
            >
              {error}
            </Typography>
          </WarningPanel>
        )}

        {result && (
          <Card style={{ marginTop: 24 }}>
            <CardContent>
              <Typography variant="h6">Workload Created</Typography>
              <Typography variant="body2">
                <b>ID:</b> {result.id}
              </Typography>
              <Typography variant="body2">
                <b>Status:</b> {result.status}
              </Typography>
              <Typography variant="body2">
                <b>Cluster:</b> {result.clusterId}
              </Typography>
              {result.message && (
                <Typography variant="body2">
                  <b>Message:</b> {result.message}
                </Typography>
              )}
              {result.url && (
                <Typography variant="body2">
                  <b>URL:</b> {result.url}
                </Typography>
              )}
            </CardContent>
          </Card>
        )}
      </Content>
    </Page>
  );
};
