import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Sandbox {
  id: string;
  project_id: string;
  workspace_id: string;
  branch: string;
  status: string;
  url: string;
  port: number;
  container_id: string;
  created_by: string;
  expires_at: string;
  created_at: string;
  destroyed_at: string | null;
}

interface ShellSession {
  id: string;
  sandbox_id: string;
  status: string;
  created_at: string;
  closed_at: string | null;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const sandboxResult = await apiFetch<Sandbox>(`/api/sandboxes/${params.id}`, { cookie: cookieHeader });
  if (sandboxResult.status === 404) throw error(404, 'sandbox not found');
  if (!sandboxResult.ok || !sandboxResult.data) throw error(500, sandboxResult.error || 'failed to load');

  const sessionsResult = await apiFetch<ShellSession[]>(
    `/api/sandboxes/${params.id}/shell-sessions`,
    { cookie: cookieHeader }
  );
  const sessions = sessionsResult.ok ? (sessionsResult.data ?? []) : [];

  return { sandbox: sandboxResult.data, shellSessions: sessions };
};

export const actions: Actions = {
  createShell: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

    const r = await apiFetch<ShellSession>(`/api/sandboxes/${params.id}/shell`, {
      method: 'POST',
      cookie: cookieHeader
    });

    if (!r.ok) return fail(r.status, { error: r.error || 'failed to create shell session' });
    return { shellCreated: true, sessionId: r.data?.id };
  },

  exec: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();

    const sessionId = String(fd.get('session_id') || '');
    const command = String(fd.get('command') || '').trim();

    if (!sessionId) return fail(400, { error: 'session_id required' });
    if (!command) return fail(400, { error: 'command required', sessionId });

    const r = await apiFetch<{ output: string; exit_code: number }>(
      `/api/shell-sessions/${sessionId}/exec`,
      {
        method: 'POST',
        body: { command },
        cookie: cookieHeader
      }
    );

    if (!r.ok) return fail(r.status, { error: r.error || 'exec failed', sessionId });
    return { execResult: r.data, sessionId, command };
  },

  closeShell: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();
    const sessionId = String(fd.get('session_id') || '');

    if (!sessionId) return fail(400, { error: 'session_id required' });

    const r = await apiFetch(`/api/shell-sessions/${sessionId}/close`, {
      method: 'POST',
      cookie: cookieHeader
    });

    if (!r.ok) return fail(r.status, { error: r.error || 'failed to close session' });
    return { shellClosed: true };
  },

  extend: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();
    const hoursRaw = String(fd.get('hours') || '').trim();

    const hours = parseInt(hoursRaw, 10);
    if (!Number.isFinite(hours) || hours < 1 || hours > 72) {
      return fail(400, { error: 'hours must be 1-72' });
    }

    const r = await apiFetch(`/api/sandboxes/${params.id}/extend`, {
      method: 'POST',
      body: { hours },
      cookie: cookieHeader
    });

    if (!r.ok) return fail(r.status, { error: r.error || 'failed to extend' });
    return { extended: true };
  },

  destroy: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

    const r = await apiFetch(`/api/sandboxes/${params.id}/destroy`, {
      method: 'POST',
      cookie: cookieHeader
    });

    if (!r.ok) return fail(r.status, { error: r.error || 'failed to destroy' });
    throw redirect(303, '/sandboxes');
  }
};
