import { error, fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Workspace {
  id: string;
  name: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<Workspace>(`/api/workspaces/${params.id}`, {
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
  });
  if (result.status === 404) throw error(404, 'workspace not found');
  if (!result.ok || !result.data) throw error(500, result.error || 'failed to load workspace');
  return { workspace: result.data };
};

export const actions: Actions = {
  default: async ({ params, request, cookies }) => {
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    const git_url = String(data.get('git_url') || '').trim();
    const branch = String(data.get('branch') || '').trim();

    if (!name || !git_url) {
      return fail(400, { name, git_url, branch, error: 'name and git_url required' });
    }

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}/projects`, {
      method: 'POST',
      body: { name, git_url, branch },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });

    if (!result.ok) {
      return fail(result.status, { name, git_url, branch, error: result.error || 'create failed' });
    }

    throw redirect(303, `/workspaces/${params.id}`);
  }
};
