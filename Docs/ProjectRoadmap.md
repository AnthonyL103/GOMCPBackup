# GoMCP Hosted Multi-User Roadmap (8 Part-Time Weeks)

## Goal
Build a hosted, multi-user web product around the existing GoMCP core so non-coders can use it through a simple UI.

## Constraints
- Total duration: 8 part-time weeks
- Sprint length: 2 weeks
- Priority: deploy quickly and reduce complexity

## Product Scope
Frontend targets:
- Minimalistic, non-technical UX
- Chat textbox interface
- Sign in and sign out
- Secrets configuration UI for cloud/provider credentials
- Drag-and-drop diagram builder as a future feature (not initial priority)

Backend and platform targets:
- Web APIs needed by frontend
- Existing MCP architecture exposed as services
- Tool generation support
- Remote API tool configuration support
- Cloud architecture generation support
- Hosted multi-user deployment

## Fast-Deploy Stack
- Frontend: React + Vite + TypeScript
- Backend API: Go (reuse current codebase)
- Auth and database: Supabase (Auth + Postgres)
- Secret storage: encrypted secrets in Postgres + server-side encryption key from environment
- Hosting: Render or Fly.io for backend, Vercel or Netlify for frontend
- Observability: basic structured logs + error tracking (Sentry optional)

## Architecture Target (End of Week 8)
- Web frontend calls Go API
- Go API manages user sessions and chat sessions
- Go API runs provider requests and MCP tool execution
- Go API supports generated tool servers and remote API tools
- Auth-protected user accounts and per-user secret configuration
- Production deployment with monitoring and rollback path

## Sprint Plan

### Sprint 1 (Weeks 1-2): Web MVP Foundation
Objective:
Ship a hosted MVP where a user can open a webpage, type a message, and get a response from the agent.

Deliverables:
- Minimal frontend with chat transcript and textbox
- Go HTTP API wrapper over current CLI flow
- Session-scoped chat state in backend memory
- Shared server-managed provider key for MVP
- Basic deployment pipeline for frontend and backend
- Health endpoint and basic logging

Implementation tasks:
- Add API endpoints:
  - POST /api/chat/message
  - POST /api/chat/session
  - GET /api/health
- Refactor run loop logic into reusable service methods
- Add CORS and request validation
- Build minimal UI pages:
  - Chat page
  - Basic error/loading states
- Deploy first hosted version

Definition of done:
- A non-technical user can send and receive messages from the browser
- Deployment URL works from another machine
- No terminal interaction required for end users

### Sprint 2 (Weeks 3-4): Multi-User Core + Auth + Secrets
Objective:
Introduce account-based usage and secure per-user configuration.

Deliverables:
- Sign up, sign in, sign out
- Per-user chat sessions persisted in database
- Secrets settings page for provider/cloud credentials
- Server-side secret encryption/decryption
- Authorization checks on all user data endpoints

Implementation tasks:
- Integrate Supabase Auth
- Add DB schema:
  - users (managed by auth)
  - chat_sessions
  - chat_messages
  - user_secrets
- Build frontend auth flows:
  - login
  - logout
  - protected routes
- Add settings page for secret management
- Add backend middleware for auth token validation

Definition of done:
- Users can log in and see only their own chats and secrets
- Secrets are never exposed in frontend logs or API responses
- Hosted app supports at least 5 concurrent student-test users

### Sprint 3 (Weeks 5-6): MCP Power Features
Objective:
Expose core differentiators: generated tools, remote API tools, and cloud architecture generation.

Deliverables:
- API and UI flow for tool generation
- API and UI flow for remote API tool configuration
- Initial cloud architecture generation capability
- Async job status tracking for long operations

Implementation tasks:
- Add endpoints:
  - POST /api/tools/generate
  - POST /api/tools/remote/register
  - POST /api/architecture/generate
  - GET /api/jobs/:id
- Connect existing servergeneration package to authenticated web APIs
- Build frontend pages:
  - Tool generation wizard (form-based)
  - Remote API tool setup page
  - Architecture generation page (prompt + output)
- Add job state store and progress polling
- Add safety limits (timeouts, max generated servers, per-user quotas)

Definition of done:
- Authenticated user can generate a tool server and invoke it
- Authenticated user can configure at least one remote API tool
- User can request and receive a generated cloud architecture output

### Sprint 4 (Weeks 7-8): Production Hardening + Launch + Diagram Foundation
Objective:
Make the system reliable enough for demos/users and set up the future diagram workflow.

Deliverables:
- Production readiness pass
- Monitoring, backup, and rollback procedures
- Final UX polish for non-coders
- Phase 2 technical foundation for drag-and-drop diagrams

Implementation tasks:
- Add guardrails:
  - rate limiting
  - audit logging for admin actions
  - stricter input validation
- Add reliability checks:
  - smoke tests
  - critical API integration tests
- Improve UX copy and onboarding hints for first-time users
- Define and implement diagram data model draft:
  - node types
  - edge schema
  - export format to tool generation payloads
- Do a tiny spike using React Flow (or equivalent) without full feature commitment

Definition of done:
- Public demo can handle realistic class/project traffic
- Recovery steps are documented and tested
- Diagram feature has validated technical direction and schema for next phase

### Future Phase (After Week 8): Full Drag-and-Drop Builder
Priority order:
1. Visual node editor for tools, APIs, and cloud components
2. Canvas to generated config/code pipeline
3. Template library for common architectures
4. Collaboration and sharing

## Weekly Time Budget Guidance (Part-Time)
- 10-15 hours per week target
- 60% implementation
- 20% testing and bug fixes
- 20% docs, deployment, and demos

## Risks and Mitigations
- Risk: Scope too large for part-time schedule
  - Mitigation: Freeze advanced UI and keep first 4 weeks strictly MVP + auth + secrets
- Risk: Security mistakes in secret handling
  - Mitigation: Server-side encryption, never return secret plaintext after save, add secret redaction in logs
- Risk: Tool generation reliability issues
  - Mitigation: Make generation async, enforce timeouts, provide clear failure states and retry paths
- Risk: Hosting complexity
  - Mitigation: Use managed platforms and avoid Kubernetes in this phase

## Milestone Summary
- End of Week 2: Hosted chat MVP
- End of Week 4: Multi-user auth + secrets
- End of Week 6: Tool generation + remote API config + cloud architecture generation
- End of Week 8: Hardened launch + diagram foundation ready for next phase
