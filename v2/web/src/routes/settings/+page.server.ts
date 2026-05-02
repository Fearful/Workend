import { redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface GHStatus {
  connected: boolean;
  handle?: string;
  scopes?: string;
}

export const load: PageServerLoad = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const gh = await apiFetch<GHStatus>('/api/me/github', { cookie: cookieHeader });
  const ghStatus: GHStatus = gh.ok && gh.data ? gh.data : { connected: false };
  // 404-style "not configured" surfaces as a 404 from the api when the
  // endpoint isn't registered. Treat as "not configured."
  const ghEnabled = gh.status !== 404;

  return { ghStatus, ghEnabled };
};

export const actions: Actions = {
  disconnectGithub: async ({ cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    await apiFetch('/api/me/github', {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    return { disconnected: true };
  }
};
