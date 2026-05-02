import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Project {
  id: string;
  workspace_id: string;
  name: string;
  git_url: string;
  default_branch: string | null;
  local_path: string | null;
  status: string;
  last_commit_sha: string | null;
  last_commit_message: string | null;
  last_commit_author: string | null;
  last_synced_at: string | null;
  created_at: string;
  updated_at: string;
  webhook_token?: string;
}

interface Workspace {
  id: string;
  name: string;
}

interface Task {
  id: string;
  project_id: string;
  source: string;
  name: string;
  raw_command: string;
  detected_at: string;
}

interface Run {
  id: string;
  task_id: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  created_at: string;
}

interface Lang {
  files: number;
  lines: number;
  code: number;
  comments: number;
  blanks: number;
}

interface Stats {
  computed_at: string;
  total_files: number;
  total_lines: number;
  total_code: number;
  languages: Record<string, Lang>;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const projResult = await apiFetch<Project>(`/api/projects/${params.id}`, { cookie: cookieHeader });
  if (projResult.status === 404) throw error(404, 'project not found');
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || 'failed to load');

  const [wsResult, tasksResult, runsResult, statsResult] = await Promise.all([
    apiFetch<Workspace>(`/api/workspaces/${projResult.data.workspace_id}`, { cookie: cookieHeader }),
    apiFetch<Task[]>(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch<Run[]>(`/api/projects/${params.id}/runs`, { cookie: cookieHeader }),
    apiFetch<Stats | null>(`/api/projects/${params.id}/stats`, { cookie: cookieHeader })
  ]);

  return {
    project: projResult.data,
    workspace: wsResult.ok ? (wsResult.data ?? null) : null,
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : [],
    runs: runsResult.ok ? (runsResult.data ?? []) : [],
    stats: statsResult.ok ? (statsResult.data ?? null) : null
  };
};

export const actions: Actions = {
  sync: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/sync`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) throw error(result.status, result.error || 'sync failed');
    return { synced: true };
  },
  delete: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const projResult = await apiFetch<Project>(`/api/projects/${params.id}`, {
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    const wsID = projResult.data?.workspace_id;

    const result = await apiFetch(`/api/projects/${params.id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) throw error(result.status, result.error || 'delete failed');
    throw redirect(303, wsID ? `/workspaces/${wsID}` : '/');
  },
  run: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    if (!taskID) return fail(400, { error: 'task_id required' });

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch<{ id: string }>(`/api/tasks/${taskID}/runs`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'run failed to start' });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  }
};
