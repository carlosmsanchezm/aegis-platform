# One-Prompt Templates for Autonomous Execution

Use these exact prompts to trigger fully autonomous ticket implementation. No babysitting required.

## Basic Jira Ticket Implementation

```
Implement Jira ticket <TICKET-KEY>
```

**What happens automatically:**
1. jira-implementer agent activates
2. Fetches ticket from Jira
3. Transitions to "In Progress"
4. Adds codex-in-progress label
5. Implements all changes
6. Runs build + tests (3 retries with GPT-5 escalation)
7. Opens PR
8. Monitors CI (3 retries with GPT-5 escalation)
9. Updates Jira with results
10. Done!

**Example:**
```
Implement Jira ticket MVP-42
```

---

## Jira Ticket with NIST Controls

```
Implement Jira ticket <TICKET-KEY> which includes NIST control <CONTROL-ID>
```

**What happens automatically:**
1. jira-implementer activates
2. Fetches ticket
3. Sees NIST control mentioned
4. Looks up control via mcp__nist__get_control
5. Implements security control requirements
6. Documents in .aegis/controls/
7. Continues with normal workflow (build, test, PR, CI, Jira wrap-up)

**Example:**
```
Implement Jira ticket MVP-43 which includes NIST control AC-2 for account management
```

---

## Just NIST Control (No Jira)

```
Implement NIST control <CONTROL-ID>
```

**What happens automatically:**
1. nist-implementer agent activates
2. Looks up control requirements
3. Checks FedRAMP baseline inclusion
4. Implements code changes
5. Creates control documentation
6. Writes evidence JSON

**Example:**
```
Implement NIST control SC-8(1) for transmission confidentiality
```

---

## Multiple Tickets (Batch)

```
Implement Jira tickets <KEY-1>, <KEY-2>, <KEY-3> in sequence
```

**What happens automatically:**
1. jira-implementer processes first ticket completely
2. Then processes second ticket completely
3. Then processes third ticket completely
4. Each goes through full workflow autonomously

**Example:**
```
Implement Jira tickets MVP-40, MVP-41, MVP-42 in sequence
```

---

## Test-Only (No Implementation)

```
Run tests for current changes and monitor CI
```

**What happens automatically:**
1. ci-sheriff agent activates
2. Runs build
3. Runs tests (3-retry with GPT-5 escalation)
4. Opens PR if not exists
5. Monitors CI (3-retry with GPT-5 escalation)
6. Reports results

**Example:**
```
Run tests for current changes on branch aegis-ci/MVP-42-auth-fix and monitor CI
```

---

## Advanced: With Specific Constraints

```
Implement Jira ticket <TICKET-KEY>

Constraints:
- Only modify files in: services/auth/, pkg/rbac/
- Must implement NIST controls: AC-2, AC-3, AU-2
- Use existing test patterns from services/auth/auth_test.go
```

**What happens automatically:**
1. jira-implementer activates with additional context
2. Respects file constraints
3. Implements multiple NIST controls
4. Follows specified test patterns
5. Full autonomous execution

---

## Tips for Maximum Autonomy

### DO:
✅ Give a clear starting point (ticket key or control ID)
✅ Mention any NIST controls in the prompt
✅ Specify constraints if you have them (file paths, patterns)
✅ Trust the agents to handle retries and escalation

### DON'T:
❌ Ask "Can you implement MVP-42?" → Just say "Implement MVP-42"
❌ Break it into steps → Agents handle all steps
❌ Wait to give next command → Agents run to completion
❌ Micromanage retry attempts → Agents handle retries automatically

---

## What "Autonomous" Means

When you use these prompts, the agent will:

1. **Make decisions**: Chooses how to fix failures without asking
2. **Retry automatically**: Up to 3 attempts for build, tests, and CI
3. **Escalate smartly**: Calls GPT-5 after 3rd failure automatically
4. **Coordinate**: Uses NIST tools and MCP servers as needed
5. **Document**: Updates Jira, creates evidence, writes control docs
6. **Report**: Only stops when done OR truly blocked

**You only intervene if:**
- Agent is blocked (missing credentials, unclear requirements)
- 3 retries + GPT-5 advice failed (unrecoverable error)
- Agent asks a clarifying question (rare)

---

## Common Patterns

### Pattern 1: Regular Feature
```
Implement MVP-45
```
Agent: Jira → Code → Build → Test → PR → CI → Jira ✅

### Pattern 2: Security Control
```
Implement MVP-46 with NIST control AC-7
```
Agent: Jira → NIST Lookup → Code → Build → Test → PR → CI → Jira ✅

### Pattern 3: Multiple Controls
```
Implement MVP-47 which requires AC-2, AC-3, and AU-2
```
Agent: Jira → 3 NIST Lookups → Code → Build → Test → PR → CI → Jira ✅

### Pattern 4: Test Failure Recovery
```
Implement MVP-48
```
Agent tries → Test fails (attempt 1) → Auto-fix → Test fails (attempt 2) → Auto-fix → Test fails (attempt 3) → GPT-5 advice → Apply → Test passes → PR → CI ✅

---

## Monitoring Progress

While agent runs, you'll see:
- "Fetching Jira ticket..."
- "Transitioning to In Progress..."
- "Implementing changes..."
- "Running build (attempt 1)..."
- "Running tests (attempt 1)..."
- "Opening PR..."
- "Monitoring CI..."
- "Updating Jira..."
- "Complete!"

**You don't need to respond unless agent asks a question.**

---

## Example Session

```
You: Implement MVP-42

Agent: Fetching Jira ticket MVP-42...
Agent: Transitioning to In Progress and adding codex-in-progress label...
Agent: Acceptance criteria: [Add MFA support, Update login flow, Add audit logs]
Agent: Allow paths: services/auth/, pkg/mfa/
Agent: Implementing changes to services/auth/login.go...
Agent: Implementing changes to pkg/mfa/totp.go...
Agent: Running build...
Agent: ✅ Build passed
Agent: Running tests (attempt 1)...
Agent: ❌ Tests failed: undefined: mfa.ValidateTOTP
Agent: Fixing import...
Agent: Running tests (attempt 2)...
Agent: ✅ Tests passed
Agent: Creating branch aegis-ci/MVP-42-mfa-support...
Agent: Opening PR...
Agent: PR created: https://github.com/org/repo/pull/567
Agent: Waiting for CI...
Agent: Monitoring CI run...
Agent: ✅ CI passed
Agent: Updating Jira with results...
Agent: Complete! PR: https://github.com/org/repo/pull/567

You: [nothing - agent is done!]
```

---

## Troubleshooting

**Agent keeps asking questions:**
→ Check if ticket has clear acceptance criteria and allow_paths

**Agent stops after 1 retry:**
→ Verify mcp__gpt5__advise is working (check OPENAI_API_KEY)

**Agent doesn't seem autonomous:**
→ Make sure you're using the updated jira-implementer.md with "fully autonomous" in description

**Tests keep failing:**
→ Agent will try 3 times + GPT-5, then ask you. This is expected for hard problems.
