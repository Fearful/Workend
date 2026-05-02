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

export const load: PageServerLoad = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');

  const token = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<Workspace[]>('/api/workspaces', {
    cookie: token ? `${SESSION_COOKIE}=${token}` : undefined
  });

  return {
    workspaces: result.ok ? (result.data ?? []) : [],
    error: result.ok ? null : (result.error || 'failed to load workspaces')
  };
};
