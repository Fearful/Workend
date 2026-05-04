import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const API_URL = process.env.WORKEND_API_URL || 'http://api:8080';
const SESSION_COOKIE = 'workend_session';
const STATE_COOKIE_PREFIX = 'workend_oauth_state_';

export const GET: RequestHandler = async ({ url, params, cookies }) => {
  const session = cookies.get(SESSION_COOKIE);
  const stateCookieName = STATE_COOKIE_PREFIX + params.provider;
  const state = cookies.get(stateCookieName);

  const headers: Record<string, string> = {};
  const cookieParts: string[] = [];
  if (session) cookieParts.push(`${SESSION_COOKIE}=${session}`);
  if (state) cookieParts.push(`${stateCookieName}=${state}`);
  if (cookieParts.length) headers['Cookie'] = cookieParts.join('; ');

  const res = await fetch(`${API_URL}/api/auth/${params.provider}/callback?${url.searchParams.toString()}`, {
    headers
  });

  cookies.delete(stateCookieName, { path: '/' });

  if (!res.ok) {
    const txt = await res.text();
    throw redirect(303, `/settings?conn_error=${encodeURIComponent(txt.trim() || 'failed')}`);
  }
  throw redirect(303, '/settings?connected=' + encodeURIComponent(params.provider ?? ''));
};
