// @ts-nocheck
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
  description: string;
  my_role: string;
}

interface SlowTask {
  task_id: string;
  task_name: string;
  task_source: string;
  project_id: string;
  project_name: string;
  avg_seconds: number;
  runs: number;
}

interface FailingTask {
  task_id: string;
  task_name: string;
  task_source: string;
  project_id: string;
  project_name: string;
  failures: number;
  total: number;
  failure_rate: number;
}

interface DayBucket {
  date: string;
  total: number;
  failed: number;
  succeeded: number;
}

interface ProjectHealth {
  project_id: string;
  project_name: string;
  status: string;
  last_run_status: string | null;
  last_run_at: string | null;
  failure_rate: number;
  runs_in_window: number;
}

interface WorkspaceSummary {
  workspace_id: string;
  window_days: number;
  totals: {
    projects: number;
    members: number;
    runs_in_window: number;
    failed_in_window: number;
    success_rate: number;
    active_now: number;
  };
  slowest_tasks: SlowTask[];
  failing_tasks: FailingTask[];
  daily_runs: DayBucket[];
  projects_health: ProjectHealth[];
}

export const load = async ({ params, url, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const days = Math.max(1, Math.min(365, parseInt(url.searchParams.get('days') || '30', 10) || 30));

  const [wsResp, sumResp] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch<WorkspaceSummary>(`/api/workspaces/${params.id}/dashboard?days=${days}`, { cookie: cookieHeader })
  ]);

  if (wsResp.status === 404) throw error(404, 'workspace not found');
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || 'failed to load workspace');

  return {
    workspace: wsResp.data,
    summary: sumResp.ok ? (sumResp.data ?? null) : null,
    summaryError: sumResp.ok ? null : (sumResp.error || 'failed to load dashboard'),
    days
  };
};
