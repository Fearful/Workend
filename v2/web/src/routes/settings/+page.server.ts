import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface ConnectionView {
  provider_id: string;
  provider: string;
  instance_url: string;
  instance_host: string;
  connected: boolean;
  handle?: string;
  scopes?: string;
  connected_at?: string;
  connection_id?: string;
}

export const load: PageServerLoad = async ({ locals, cookies, url }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const result = await apiFetch<ConnectionView[]>('/api/me/connections', { cookie: cookieHeader });
  // The endpoint isn't registered when no providers are configured at all.
  const providersConfigured = result.status !== 404;

  return {
    connections: result.ok ? (result.data ?? []) : [],
    providersConfigured,
    flash: {
      connected: url.searchParams.get('connected'),
      error: url.searchParams.get('conn_error')
    }
  };
};

export const actions: Actions = {
  disconnect: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { error: 'id required' });
    const result = await apiFetch(`/api/me/connections/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'disconnect failed' });
    return { disconnected: true };
  }
};
