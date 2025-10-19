# Agent Log Visibility Guide

Your agents can now **automatically read diagnostic logs** to understand why tools fail. This makes them much more self-sufficient.

## What Changed

### Before (Without Log Visibility):
```
Agent: Running mcp__jira_local__get_issue for MVP-42...
MCP: [Fails silently]
Agent: Tool failed. Let me retry... [guesses what went wrong]
Agent: Failed again. Escalating to user.
```

### After (With Log Visibility):
```
Agent: Running mcp__jira_local__get_issue for MVP-42...
MCP: [Fails]
Agent: Reading MCP logs to diagnose...
Agent: Log shows "403 Forbidden - Invalid API token"
Agent: The Jira API token is invalid or expired. User, please check JIRA_API_TOKEN in .env
```

## Log Locations Agents Know About

### 1. MCP Server Logs
**Path**: `~/Library/Caches/claude-cli-nodejs/-Users-carlossanchez-code-pulumi-provision/mcp-logs-<server>/`

**Logs available:**
- `mcp-logs-jira-local/` - All Jira tool calls
- `mcp-logs-aegis/` - Build/test/evidence operations
- `mcp-logs-gpt5/` - GPT-5 advisor calls
- `mcp-logs-nist/` - NIST control lookups
- `mcp-logs-github/` - GitHub Actions operations
- `mcp-logs-repo-files/` - File read/write/edit operations
- `mcp-logs-atlassian-local/` - Atlassian SSO operations
- `mcp-logs-aws-api/`, `mcp-logs-aws-docs/` - AWS operations

**What's in them:**
- Timestamp of each tool call
- Input parameters
- Output/response
- Error messages (connection, auth, timeout)
- Stack traces for failures

### 2. Aegis Evidence/Audit Logs
**Path**: `.aegis/logs/` and `.aegis/evidence/`

**Files:**
- `.aegis/logs/jira_mcp_audit.jsonl` - Complete audit trail of Jira actions (transitions, comments, field updates)
- `.aegis/evidence/*.json` - Evidence from `mcp__aegis__evidence_write_json` (test results, control implementations)

### 3. CI/GitHub Logs
**Accessed via**: `gh` CLI commands

**Commands agents use:**
```bash
# List recent runs
gh run list --workflow preview-deployment.yml --branch <branch>

# Get full logs
gh run view --workflow preview-deployment.yml --branch <branch> --log

# View specific job
gh run view <run-id> --log --job <job-name>
```

### 4. Test Script Output
**Accessed via**: Re-running test script or reading tool output

**If truncated in MCP tool result:**
```bash
./scripts/test-local-with-tls.sh 2>&1 | tail -500
```

## How Agents Use Logs

### Scenario 1: Jira Tool Fails

**Agent's thought process:**
```
1. mcp__jira_local__get_issue fails
2. Read latest mcp-logs-jira-local/*.txt
3. Parse error: "404 Not Found"
4. Conclusion: Ticket MVP-999 doesn't exist
5. Report: "Ticket MVP-999 not found in Jira. Please verify ticket key."
```

### Scenario 2: Build Fails

**Agent's thought process:**
```
1. mcp__aegis__go_build returns exit code 1
2. Read build output from tool result
3. See: "undefined: crypto.GenerateKey"
4. Read mcp-logs-aegis/*.txt for full context
5. Conclusion: Missing import or typo in function name
6. Fix: Add import or correct function name
7. Retry build
```

### Scenario 3: Tests Time Out

**Agent's thought process:**
```
1. mcp__aegis__run_tests times out after 10 min
2. Read latest mcp-logs-aegis/*.txt
3. See: "Test server failed to start, connection refused"
4. Re-run test script manually to see full output
5. See: Port 8443 already in use
6. Conclusion: Previous test run didn't clean up
7. Fix: Kill process on 8443, retry tests
```

### Scenario 4: GPT-5 Advice Fails

**Agent's thought process:**
```
1. mcp__gpt5__advise returns error
2. Read latest mcp-logs-gpt5/*.txt
3. See: "Request too large: 52000 tokens, max 32000"
4. Conclusion: Payload too large
5. Fix: Truncate diffs to last 1000 lines each, retry
```

### Scenario 5: CI Fails

**Agent's thought process:**
```
1. gh run watch shows failure
2. Run: gh run view --log > ci-full.log
3. Read ci-full.log
4. See: "ImagePullBackOff: Failed to pull image"
5. Conclusion: Image tag incorrect or image not pushed
6. Fix: Check Dockerfile, verify image push step
```

## What This Means for You

**Your agents are now much smarter:**

✅ **Self-diagnose**: Read logs to understand failures
✅ **Targeted fixes**: Fix root cause, not symptoms
✅ **Better reporting**: Give you exact error details when escalating
✅ **Faster iteration**: Don't waste retries on guesses

**You benefit from:**

1. **Less babysitting**: Agents figure out issues themselves
2. **Better error messages**: When they escalate, you get full context
3. **Faster debugging**: Agents read logs before asking you
4. **More autonomous**: Handle transient errors (timeouts, rate limits) automatically

## Examples

### Example 1: Auth Issue

**Without log reading:**
```
Agent: Jira tool failed. Retrying...
Agent: Failed again. User, Jira is not working.
```

**With log reading:**
```
Agent: Jira tool failed. Reading logs...
Agent: Log shows "401 Unauthorized - Token expired"
Agent: User, JIRA_API_TOKEN has expired. Please refresh in .env:
       https://id.atlassian.com/manage-profile/security/api-tokens
```

### Example 2: Network Issue

**Without log reading:**
```
Agent: NIST control lookup failed. Retrying...
Agent: Failed again. Cannot look up AC-2.
```

**With log reading:**
```
Agent: NIST control lookup failed. Reading logs...
Agent: Log shows "Connection timeout after 30s"
Agent: Retrying with longer timeout...
Agent: Success! Control AC-2 retrieved.
```

### Example 3: Configuration Issue

**Without log reading:**
```
Agent: Tests failed. Retrying...
Agent: Failed 3 times. Escalating to GPT-5...
[Wastes GPT-5 call on config issue]
```

**With log reading:**
```
Agent: Tests failed. Reading logs...
Agent: Log shows "File not found: /var/run/test-certs/server.crt"
Agent: Test certificates missing. Regenerating...
Agent: Running: ./scripts/generate-test-certs.sh
Agent: Tests passed!
[No GPT-5 call needed]
```

## Configuration

**No configuration needed!** Agents automatically know:

1. Where MCP logs are (~/Library/Caches/...)
2. Where project logs are (.aegis/logs/)
3. How to read them (Read tool, Glob tool)
4. What to look for (error patterns)

**Just use your agents normally and they'll read logs when needed.**

## Monitoring Agent Behavior

When you see an agent say:
- "Reading MCP logs to diagnose..."
- "Checking latest aegis log..."
- "Reading CI logs for details..."

**That's good!** It means the agent is being smart and diagnosing issues instead of guessing.

## Best Practices

**For you:**
1. ✅ Let agents read logs - they know how
2. ✅ Trust their diagnosis - they have more context than before
3. ✅ Provide feedback if diagnosis is wrong (helps improve patterns)

**For agents (already built in):**
1. ✅ Read logs BEFORE retrying
2. ✅ Read logs BEFORE escalating to GPT-5
3. ✅ Read logs BEFORE asking user
4. ✅ Include log excerpts in escalation reports

## Troubleshooting

**Agent not reading logs:**
- Check if logs exist: `ls ~/Library/Caches/claude-cli-nodejs/-Users-carlossanchez-code-pulumi-provision/`
- Verify MCP servers are running: `/mcp` in Claude Code
- Try: "Read the latest jira-local MCP log to see what happened"

**Logs empty or missing:**
- MCP servers may not have run yet
- Check MCP server status: `/mcp`
- Try running a tool manually first

**Agent misdiagnosing from logs:**
- Report the specific case (helps improve patterns)
- You can override: "Ignore the log, the real issue is X"

## Summary

Your agents now have **full visibility into their own operations** through:
- MCP server logs (tool calls)
- Project logs (evidence, audit trail)
- CI logs (GitHub Actions)
- Test output (local execution)

This makes them **much more autonomous and effective** at diagnosing and fixing issues without your intervention.

**Just give them a ticket and let them work - they'll read logs as needed!**
