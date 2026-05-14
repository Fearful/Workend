// @ts-nocheck
import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
  my_role: string;
}

interface Project {
  id: string;
  workspace_id: string;
  name: string;
  git_url: string;
  default_branch: string | null;
  status: string;
  last_commit_sha: string | null;
  last_commit_message: string | null;
  last_commit_author: string | null;
  last_synced_at: string | null;
  created_at: string;
  updated_at: string;
}

interface Member {
  user_id: string;
  email: string;
  display_name: string;
  role: string;
  added_at: string;
}

interface ActivityEvent {
  id: number;
  occurred_at: string;
  actor_id: string | null;
  actor_name: string;
  action: string;
  target_kind: string;
  target_id: string;
  metadata?: unknown;
}

interface RecentIssue {
  id: string;
  project_id: string;
  project_name: string;
  provider_number: number;
  title: string;
  state: string;
  labels: string[];
  author_handle: string;
  html_url: string;
  upstream_updated_at: string | null;
}

interface ProjectGridItem {
  project_id: string;
  project_name: string;
  task_count: number;
  last_run_at: string | null;
  last_status: string;
  recent_passes: number;
  recent_fails: number;
  status_color: string;
}

interface TeamMember {
  user_id: string;
  display_name: string;
  role: string;
  last_active_at: string | null;
  active_runs: number;
  is_active: boolean;
}

interface DepOverview {
  total_projects: number;
  total_deps: number;
  unresolved_deps: number;
  most_connected: { project_id: string; project_name: string; inbound_deps: number; outbound_deps: number }[];
}

interface SecretsSummary {
  workspace_id: string;
  workspace_name: string;
  total_secrets: number;
  oldest_updated_at: string | null;
  newest_updated_at: string | null;
  creator_count: number;
}

interface PipelineRun {
  pipeline_run_id: string;
  pipeline_name: string;
  project_name: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
}

interface PipelineBoard {
  queued: PipelineRun[];
  running: PipelineRun[];
  passed: PipelineRun[];
  failed: PipelineRun[];
}

interface SandboxItem {
  id: string;
  type: string;
  branch: string;
  status: string;
  project_name: string;
  minutes_left?: number;
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const wid = params.id;

  const [
    wsResp, projResp, membersResp, activityResp, issuesResp,
    gridResp, teamResp, depsResp, secretsSumResp, pipeResp, sandboxResp
  ] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${wid}`, { cookie: cookieHeader }),
    apiFetch<Project[]>(`/api/workspaces/${wid}/projects`, { cookie: cookieHeader }),
    apiFetch<Member[]>(`/api/workspaces/${wid}/members`, { cookie: cookieHeader }),
    apiFetch<ActivityEvent[]>(`/api/workspaces/${wid}/activity?limit=30`, { cookie: cookieHeader }),
    apiFetch<RecentIssue[]>(`/api/workspaces/${wid}/recent-issues`, { cookie: cookieHeader }),
    apiFetch<ProjectGridItem[]>(`/api/widgets/workspace/${wid}/project-grid`, { cookie: cookieHeader }),
    apiFetch<TeamMember[]>(`/api/widgets/workspace/${wid}/team-presence`, { cookie: cookieHeader }),
    apiFetch<DepOverview>(`/api/widgets/workspace/${wid}/dependency-overview`, { cookie: cookieHeader }),
    apiFetch<SecretsSummary>(`/api/widgets/workspace/${wid}/secrets-summary`, { cookie: cookieHeader }),
    apiFetch<PipelineBoard>(`/api/widgets/workspace/${wid}/pipeline-board`, { cookie: cookieHeader }),
    apiFetch<SandboxItem[]>(`/api/widgets/workspace/${wid}/sandbox-preview-rack`, { cookie: cookieHeader })
  ]);

  if (wsResp.status === 404) throw error(404, 'workspace not found');
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || 'failed to load workspace');

  return {
    workspace: wsResp.data,
    projects: projResp.ok ? (projResp.data ?? []) : [],
    projectsError: projResp.ok ? null : (projResp.error || 'failed to load projects'),
    members: membersResp.ok ? (membersResp.data ?? []) : [],
    activity: activityResp.ok ? (activityResp.data ?? []) : [],
    recentIssues: issuesResp.ok ? (issuesResp.data ?? []) : [],
    projectGrid: gridResp.ok ? (gridResp.data ?? []) : [],
    teamPresence: teamResp.ok ? (teamResp.data ?? []) : [],
    depOverview: depsResp.ok ? (depsResp.data ?? null) : null,
    secretsSummary: secretsSumResp.ok ? (secretsSumResp.data ?? null) : null,
    pipelineBoard: pipeResp.ok ? (pipeResp.data ?? null) : null,
    sandboxItems: sandboxResp.ok ? (sandboxResp.data ?? []) : []
  };
};

export const actions = {
  delete: async ({ params, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) throw error(result.status, result.error || 'delete failed');
    throw redirect(303, '/');
  },
  addMember: async ({ params, request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const email = String(data.get('email') || '').trim();
    const role = String(data.get('role') || 'member');
    if (!email) return fail(400, { memberError: 'email required', email });
    const result = await apiFetch(`/api/workspaces/${params.id}/members`, {
      method: 'POST',
      body: { email, role },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || 'add failed', email });
    return { memberAdded: true };
  },
  removeMember: async ({ params, request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const userID = String((await request.formData()).get('user_id') || '');
    if (!userID) return fail(400, { memberError: 'user_id required' });
    const result = await apiFetch(`/api/workspaces/${params.id}/members/${userID}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || 'remove failed' });
    return { memberRemoved: true };
  }
};
;null as any as Actions;