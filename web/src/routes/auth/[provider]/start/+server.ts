import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';
const STATE_COOKIE_PREFIX = 'workend_oauth_state_';

export const POST: RequestHandler = async ({ params, cookies, request }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<{ url: string }>(`/api/auth/${params.provider}/start`, {
    method: 'POST',
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined,
    userAgent: request.headers.get('user-agent') || ''
  });
  if (!result.ok || !result.data) {
    return json({ error: result.error || 'failed' }, { status: result.status });
  }
  // Forward the per-provider state cookie set by the API.
  if (result.setCookie) {
    const expected = STATE_COOKIE_PREFIX + params.provider;
    const m = result.setCookie.match(new RegExp(`^${expected}=([^;]+)`));
    if (m) {
      cookies.set(expected, m[1], {
        path: '/',
        httpOnly: true,
        sameSite: 'lax',
        maxAge: 600
      });
    }
  }
  return json({ url: result.data.url });
};
