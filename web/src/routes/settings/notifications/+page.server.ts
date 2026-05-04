import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Notif {
  id: string;
  kind: string;
  target: string;
  trigger: string;
  enabled: boolean;
  created_at: string;
}

interface Subscription {
  id: string;
  config_id: string;
  scope_type: 'global' | 'workspace' | 'project' | 'task';
  scope_id: string | null;
  severity: 'all' | 'failures' | 'off';
  scope_name?: string;
  config_kind?: string;
  created_at: string;
}

interface Workspace { id: string; name: string; }

interface Project { id: string; name: string; workspace_id: string; }

export const load: PageServerLoad = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [notifsResp, subsResp, wsResp] = await Promise.all([
    apiFetch<Notif[]>('/api/me/notifications', { cookie: cookieHeader }),
    apiFetch<Subscription[]>('/api/me/notification-subscriptions', { cookie: cookieHeader }),
    apiFetch<Workspace[]>('/api/workspaces', { cookie: cookieHeader })
  ]);

  // Hydrate per-workspace project lists for the scope picker.
  const workspaces = wsResp.ok ? (wsResp.data ?? []) : [];
  const projectsByWS: Record<string, Project[]> = {};
  await Promise.all(workspaces.map(async (w) => {
    const r = await apiFetch<Project[]>(`/api/workspaces/${w.id}/projects`, { cookie: cookieHeader });
    if (r.ok && r.data) projectsByWS[w.id] = r.data;
  }));

  return {
    notifications: notifsResp.ok ? (notifsResp.data ?? []) : [],
    subscriptions: subsResp.ok ? (subsResp.data ?? []) : [],
    workspaces,
    projectsByWS
  };
};

export const actions: Actions = {
  create: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const kind = String(data.get('kind') || '');
    const target = String(data.get('target') || '').trim();
    const trigger = String(data.get('trigger') || 'on_failure');
    if (!kind || !target) return fail(400, { error: 'kind and target required', target });
    const result = await apiFetch('/api/me/notifications', {
      method: 'POST',
      body: { kind, target, trigger },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'create failed', target });
    return { created: true };
  },
  delete: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    const result = await apiFetch(`/api/me/notifications/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'delete failed' });
    return { deleted: true };
  },
  test: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    const result = await apiFetch(`/api/me/notifications/${id}/test`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'test send failed' });
    return { tested: true };
  },
  subscribe: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const config_id = String(data.get('config_id') || '');
    const scope_type = String(data.get('scope_type') || 'global');
    const scope_id = String(data.get('scope_id') || '').trim();
    const severity = String(data.get('severity') || 'failures');
    const result = await apiFetch('/api/me/notification-subscriptions', {
      method: 'POST',
      body: { config_id, scope_type, scope_id, severity },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'subscribe failed' });
    return { subscribed: true };
  },
  unsubscribe: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    const result = await apiFetch(`/api/me/notification-subscriptions/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'unsubscribe failed' });
    return { unsubscribed: true };
  }
};
