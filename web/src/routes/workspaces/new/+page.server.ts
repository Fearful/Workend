import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const load: PageServerLoad = async ({ locals }) => {
  if (!locals.user) throw redirect(303, '/login');
  return {};
};

export const actions: Actions = {
  default: async ({ request, cookies }) => {
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    const description = String(data.get('description') || '').trim();

    if (!name) {
      return fail(400, { name, description, error: 'name required' });
    }

    const token = cookies.get(SESSION_COOKIE);
    const result = await apiFetch('/api/workspaces', {
      method: 'POST',
      body: { name, description },
      cookie: token ? `${SESSION_COOKIE}=${token}` : undefined
    });

    if (!result.ok) {
      return fail(result.status, { name, description, error: result.error || 'create failed' });
    }

    throw redirect(303, '/');
  }
};
