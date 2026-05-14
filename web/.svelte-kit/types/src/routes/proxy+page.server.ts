// @ts-nocheck
import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
  description: string;
  created_at: string;
  updated_at: string;
}

interface DashRun {
  id: string;
  task_id: string;
  project_id: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  created_at: string;
  project_name: string;
  workspace_name: string;
}

interface WorkspaceHealth {
  workspace_id: string;
  workspace_name: string;
  member_count: number;
  project_count: number;
  active_runs: number;
  recent_fail_rate: number;
  total_runs_7d: number;
  last_activity_at: string | null;
}

interface WorkspaceComparison {
  workspace_id: string;
  workspace_name: string;
  runs_per_day: number;
  success_rate: number;
  avg_duration_ms: number;
}

interface WorkspaceIncident {
  workspace_id: string;
  workspace_name: string;
  incident_count: number;
  worst_severity: string;
}

interface WorkspaceStorage {
  workspace_id: string;
  workspace_name: string;
  projects: number;
  tasks: number;
  runs: number;
  artifacts: number;
}

export const load = async ({ locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');

  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : undefined;

  const [workspacesResult, runsResult, healthResult, comparisonResult, incidentsResult, storageResult] = await Promise.all([
    apiFetch<Workspace[]>('/api/workspaces', { cookie: cookieHeader }),
    apiFetch<DashRun[]>('/api/me/runs', { cookie: cookieHeader }),
    apiFetch<WorkspaceHealth[]>('/api/widgets/workspaces/health', { cookie: cookieHeader }),
    apiFetch<WorkspaceComparison[]>('/api/widgets/workspaces/comparison?days=7', { cookie: cookieHeader }),
    apiFetch<WorkspaceIncident[]>('/api/widgets/workspaces/incidents', { cookie: cookieHeader }),
    apiFetch<WorkspaceStorage[]>('/api/widgets/workspaces/storage', { cookie: cookieHeader })
  ]);

  return {
    workspaces: workspacesResult.ok ? (workspacesResult.data ?? []) : [],
    recentRuns: runsResult.ok ? (runsResult.data ?? []) : [],
    error: workspacesResult.ok ? null : (workspacesResult.error || 'failed to load workspaces'),
    health: healthResult.ok ? (healthResult.data ?? []) : [],
    comparison: comparisonResult.ok ? (comparisonResult.data ?? []) : [],
    incidents: incidentsResult.ok ? (incidentsResult.data ?? []) : [],
    storage: storageResult.ok ? (storageResult.data ?? []) : []
  };
};
