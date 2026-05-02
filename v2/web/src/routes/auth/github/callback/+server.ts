import { redirect } from '@sveltejs/kit';
import type { RequestHandler } from './$types';

const API_URL = process.env.WORKEND_API_URL || 'http://api:8080';
const SESSION_COOKIE = 'workend_session';

export const GET: RequestHandler = async ({ url, cookies }) => {
  const session = cookies.get(SESSION_COOKIE);
  const state = cookies.get('workend_oauth_state');

  const headers: Record<string, string> = {};
  const cookieParts = [];
  if (session) cookieParts.push(`${SESSION_COOKIE}=${session}`);
  if (state) cookieParts.push(`workend_oauth_state=${state}`);
  if (cookieParts.length) headers['Cookie'] = cookieParts.join('; ');

  const res = await fetch(`${API_URL}/api/auth/github/callback?${url.searchParams.toString()}`, {
    headers
  });

  // Always clear the state cookie on the web side
  cookies.delete('workend_oauth_state', { path: '/' });

  if (!res.ok) {
    const txt = await res.text();
    throw redirect(303, `/settings?gh_error=${encodeURIComponent(txt.trim() || 'failed')}`);
  }
  throw redirect(303, '/settings?gh=connected');
};
