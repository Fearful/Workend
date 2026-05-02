import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Project {
  id: string;
  workspace_id: string;
  name: string;
}

interface Workspace {
  id: string;
  name: string;
}

interface Task {
  id: string;
  source: string;
  name: string;
  raw_command: string;
}

interface Schedule {
  id: string;
  task_id: string;
  cron_expr: string;
  enabled: boolean;
  last_run_at: string | null;
  next_run_at: string | null;
  created_at: string;
  task_name: string;
  task_source: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [projResult, wsListResult, tasksResult, schedulesResult] = await Promise.all([
    apiFetch<Project>(`/api/projects/${params.id}`, { cookie: cookieHeader }),
    apiFetch<Workspace>(`/api/workspaces/$WS$`, { cookie: cookieHeader }), // placeholder, replaced below
    apiFetch<Task[]>(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch<Schedule[]>(`/api/projects/${params.id}/schedules`, { cookie: cookieHeader })
  ]);

  if (projResult.status === 404) throw error(404, 'project not found');
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || 'failed to load');

  const wsResult = await apiFetch<Workspace>(`/api/workspaces/${projResult.data.workspace_id}`, { cookie: cookieHeader });

  return {
    project: projResult.data,
    workspace: wsResult.ok ? (wsResult.data ?? null) : null,
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : [],
    schedules: schedulesResult.ok ? (schedulesResult.data ?? []) : []
  };
};

export const actions: Actions = {
  create: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const task_id = String(data.get('task_id') || '');
    const cron_expr = String(data.get('cron_expr') || '').trim();
    const enabled = data.get('enabled') === 'on' || data.get('enabled') === 'true';

    if (!task_id || !cron_expr) {
      return fail(400, { error: 'task and cron expression required', cron_expr });
    }

    const result = await apiFetch(`/api/projects/${params.id}/schedules`, {
      method: 'POST',
      body: { task_id, cron_expr, enabled },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'create failed', cron_expr });
    return { created: true };
  },
  toggle: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get('id') || '');
    const result = await apiFetch(`/api/schedules/${id}/toggle`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'toggle failed' });
    return { toggled: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get('id') || '');
    const result = await apiFetch(`/api/schedules/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'delete failed' });
    return { deleted: true };
  }
};
