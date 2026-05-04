import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Step {
  position: number;
  task_id: string;
  task_name?: string;
  task_source?: string;
}

interface Pipeline {
  id: string;
  project_id: string;
  name: string;
  created_at: string;
  steps?: Step[];
}

interface Task {
  id: string;
  source: string;
  name: string;
  raw_command: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [pipelinesResult, tasksResult] = await Promise.all([
    apiFetch<Pipeline[]>(`/api/projects/${params.id}/pipelines`, { cookie: cookieHeader }),
    apiFetch<Task[]>(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader })
  ]);

  return {
    pipelines: pipelinesResult.ok ? (pipelinesResult.data ?? []) : [],
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : []
  };
};

export const actions: Actions = {
  create: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    const taskIDs = data.getAll('task_id').map((v) => String(v)).filter(Boolean);

    if (!name) return fail(400, { error: 'name required', name });
    if (taskIDs.length === 0) return fail(400, { error: 'pick at least one task', name });

    const result = await apiFetch(`/api/projects/${params.id}/pipelines`, {
      method: 'POST',
      body: { name, task_ids: taskIDs },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'create failed', name });
    return { created: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get('id') || '');
    if (!id) return fail(400, { error: 'id required' });
    const result = await apiFetch(`/api/pipelines/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'delete failed' });
    return { deleted: true };
  },
  run: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const id = String(data.get('id') || '');
    if (!id) return fail(400, { error: 'id required' });
    const result = await apiFetch<{ pipeline_run_id: string }>(`/api/pipelines/${id}/runs`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'pipeline run failed to start' });
    }
    throw redirect(303, `/pipeline-runs/${result.data.pipeline_run_id}`);
  }
};
