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

export const load = async ({ locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');

  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : undefined;

  const [workspacesResult, runsResult] = await Promise.all([
    apiFetch<Workspace[]>('/api/workspaces', { cookie: cookieHeader }),
    apiFetch<DashRun[]>('/api/me/runs', { cookie: cookieHeader })
  ]);

  return {
    workspaces: workspacesResult.ok ? (workspacesResult.data ?? []) : [],
    recentRuns: runsResult.ok ? (runsResult.data ?? []) : [],
    error: workspacesResult.ok ? null : (workspacesResult.error || 'failed to load workspaces')
  };
};
