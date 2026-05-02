import type { Handle } from '@sveltejs/kit';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const handle: Handle = async ({ event, resolve }) => {
  event.locals.user = null;

  const token = event.cookies.get(SESSION_COOKIE);
  if (token) {
    const result = await apiFetch<App.Locals['user']>('/api/me', {
      cookie: `${SESSION_COOKIE}=${token}`,
      userAgent: event.request.headers.get('user-agent') || ''
    });
    if (result.ok && result.data) {
      event.locals.user = result.data;
    } else if (result.status === 401) {
      // stale cookie — drop it
      event.cookies.delete(SESSION_COOKIE, { path: '/' });
    }
  }

  return resolve(event);
};
