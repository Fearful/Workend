import { json } from '@sveltejs/kit';
import type { RequestHandler } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

export const POST: RequestHandler = async ({ params, cookies }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch(`/api/runs/${params.id}/cancel`, {
    method: 'POST',
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
  });
  if (!result.ok) return json({ error: result.error || 'cancel failed' }, { status: result.status });
  return json({ ok: true });
};
