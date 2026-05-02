import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const POST: RequestHandler = async ({ cookies, request }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<{ url: string }>('/api/auth/github/start', {
    method: 'POST',
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined,
    userAgent: request.headers.get('user-agent') || ''
  });
  if (!result.ok || !result.data) {
    return json({ error: result.error || 'failed' }, { status: result.status });
  }
  // Forward the state cookie set by the API.
  if (result.setCookie) {
    const m = result.setCookie.match(/^workend_oauth_state=([^;]+)/);
    if (m) {
      cookies.set('workend_oauth_state', m[1], {
        path: '/',
        httpOnly: true,
        sameSite: 'lax',
        maxAge: 600
      });
    }
  }
  return json({ url: result.data.url });
};
