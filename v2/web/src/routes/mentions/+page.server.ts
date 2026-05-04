import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Mention {
  id: number;
  run_id: string;
  comment_id: string;
  actor_id: string;
  actor_name: string;
  body: string;
  task_name: string;
  project_name: string;
  created_at: string;
  read_at: string | null;
}

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
  const unreadOnly = url.searchParams.get('unread') === 'true';
  const r = await apiFetch<Mention[]>(
    `/api/me/mentions?limit=100${unreadOnly ? '&unread_only=true' : ''}`,
    { cookie: cookieHeader }
  );
  return {
    mentions: r.ok ? (r.data ?? []) : [],
    unreadOnly
  };
};

export const actions: Actions = {
  markRead: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const idRaw = data.get('id');
    const body = idRaw ? { id: Number(idRaw) } : {};
    const r = await apiFetch('/api/me/mentions/read', {
      method: 'POST',
      body,
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || 'failed' });
    return { ok: true };
  }
};
