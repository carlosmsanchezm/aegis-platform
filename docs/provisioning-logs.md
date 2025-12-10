## Provisioning Logs API

The platform-api exposes provisioning logs and phase timestamps for `ProjectInfra` jobs.

### Endpoints

- `GET /api/v1/provisioning/jobs/{jobId}/logs` — fetch buffered logs with pagination.
- `GET /api/v1/provisioning/jobs/{jobId}/logs/stream` — long-poll for new log lines (waits up to ~20s).

### Query Parameters

- `since` (optional): RFC3339 timestamp cursor. Only entries after this time are returned.
- `limit` (optional): max entries to return (default 200, max 1000).
- `stream` (optional, boolean): when `true`, the handler waits briefly for new lines before responding (useful for live tails).

### Response

```json
{
  "jobId": "infra-aegis-test-us-east-1",
  "projectId": "test",
  "clusterId": "test",
  "phase": "Provisioning",
  "startedAt": "2024-05-07T18:23:10.112345Z",
  "completedAt": "2024-05-07T18:26:44.009876Z",
  "logs": [
    {
      "timestamp": "2024-05-07T18:23:11.102938Z",
      "phase": "Provisioning",
      "type": "progress",
      "message": "Refreshing (aegis-platform-test-us-east-1)..."
    },
    {
      "timestamp": "2024-05-07T18:26:44.009876Z",
      "phase": "Ready",
      "type": "event",
      "message": "pulumi provisioning completed"
    }
  ],
  "nextCursor": "2024-05-07T18:26:44.009876Z"
}
```

### AuthZ

The handler enforces project scoping using the project attached to the provisioning job. Requests without access to the job’s project receive `403 Forbidden`.
