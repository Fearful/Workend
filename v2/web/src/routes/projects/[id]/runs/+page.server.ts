import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Project {
  id: string;
  workspace_id: string;
  name: string;
  status: string;
}

interface Workspace {
  id: string;
  name: string;
}

interface Run {
  id: string;
  task_id: string;
  project_id: string;
  commit_sha: string | null;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  created_at: string;
}

const allowedStatuses = new Set(['succeeded', 'failed', 'cancelled', 'running', 'queued']);

export const load: PageServerLoad = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const status = url.searchParams.get('status') || '';
  const filter = allowedStatuses.has(status) ? status : '';

  const path = filter
    ? `/api/projects/${params.id}/runs?status=${encodeURIComponent(filter)}`
    : `/api/projects/${params.id}/runs`;

  const [projResult, runsResult] = await Promise.all([
    apiFetch<Project>(`/api/projects/${params.id}`, { cookie: cookieHeader }),
    apiFetch<Run[]>(path, { cookie: cookieHeader })
  ]);

  if (projResult.status === 404) throw error(404, 'project not found');
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || 'failed to load');

  const wsResult = await apiFetch<Workspace>(`/api/workspaces/${projResult.data.workspace_id}`, { cookie: cookieHeader });

  return {
    project: projResult.data,
    workspace: wsResult.ok ? (wsResult.data ?? null) : null,
    runs: runsResult.ok ? (runsResult.data ?? []) : [],
    filter
  };
};
