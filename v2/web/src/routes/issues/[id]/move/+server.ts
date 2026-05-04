import type { RequestHandler } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

/**
 * Bridges the kanban drag/drop fetch (form-encoded POST) to the API's
 * JSON PATCH. SvelteKit form actions can't drive arbitrary PATCH requests
 * without a page mount, so the bridge keeps the board's drag handler
 * dependency-free.
 */
export const POST: RequestHandler = async ({ params, request, cookies }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
  const data = await request.formData();
  const to = String(data.get('to_column') || '').trim();
  if (!to) return new Response('to_column required', { status: 400 });
  const r = await apiFetch(`/api/issues/${params.id}/move`, {
    method: 'PATCH',
    body: { to_column: to },
    cookie: cookieHeader
  });
  return new Response(r.error ?? '', { status: r.ok ? 204 : r.status });
};
