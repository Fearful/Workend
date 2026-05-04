import { redirect } from '@sveltejs/kit';
import type { Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const actions: Actions = {
  default: async ({ cookies }) => {
    const token = cookies.get(SESSION_COOKIE);
    if (token) {
      await apiFetch('/api/auth/logout', {
        method: 'POST',
        cookie: `${SESSION_COOKIE}=${token}`
      });
      cookies.delete(SESSION_COOKIE, { path: '/' });
    }
    throw redirect(303, '/login');
  }
};
