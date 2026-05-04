// @ts-nocheck
import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface ConnectionView {
  provider_id: string;
  provider: string;
  instance_url: string;
  instance_host: string;
  connected: boolean;
  handle?: string;
  scopes?: string;
  connected_at?: string;
  connection_id?: string;
}

interface SSHKey {
  id: string;
  name: string;
  public_key: string;
  created_at: string;
}

interface PATCredential {
  id: string;
  host: string;
  label: string;
  created_at: string;
}

export const load = async ({ locals, cookies, url }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [connResult, sshResult, patResult] = await Promise.all([
    apiFetch<ConnectionView[]>('/api/me/connections', { cookie: cookieHeader }),
    apiFetch<SSHKey[]>('/api/me/ssh-keys', { cookie: cookieHeader }),
    apiFetch<PATCredential[]>('/api/me/pat-credentials', { cookie: cookieHeader })
  ]);
  // The connections endpoint isn't registered when no providers are configured.
  const providersConfigured = connResult.status !== 404;

  return {
    connections: connResult.ok ? (connResult.data ?? []) : [],
    providersConfigured,
    sshKeys: sshResult.ok ? (sshResult.data ?? []) : [],
    patCredentials: patResult.ok ? (patResult.data ?? []) : [],
    flash: {
      connected: url.searchParams.get('connected'),
      error: url.searchParams.get('conn_error')
    }
  };
};

export const actions = {
  disconnect: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { error: 'id required' });
    const result = await apiFetch(`/api/me/connections/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'disconnect failed' });
    return { disconnected: true };
  },
  addSSHKey: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const name = String(data.get('name') || '').trim();
    const privateKey = String(data.get('private_key') || '');
    if (!name || privateKey.length < 50) {
      return fail(400, { sshError: 'name and a PEM private key are required', sshName: name });
    }
    const result = await apiFetch<SSHKey>('/api/me/ssh-keys', {
      method: 'POST',
      body: { name, private_key: privateKey },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { sshError: result.error || 'failed', sshName: name });
    return { sshAdded: true };
  },
  deleteSSHKey: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { sshError: 'id required' });
    const result = await apiFetch(`/api/me/ssh-keys/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { sshError: result.error || 'delete failed' });
    return { sshDeleted: true };
  },
  addPAT: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const host = String(data.get('host') || '').trim();
    const label = String(data.get('label') || '').trim();
    const token = String(data.get('token') || '').trim();
    if (!host || !label || token.length < 8) {
      return fail(400, {
        patError: 'host, label and a token (8+ chars) are required',
        patHost: host,
        patLabel: label
      });
    }
    const result = await apiFetch<PATCredential>('/api/me/pat-credentials', {
      method: 'POST',
      body: { host, label, token },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) {
      return fail(result.status, {
        patError: result.error || 'failed',
        patHost: host,
        patLabel: label
      });
    }
    return { patAdded: true };
  },
  deletePAT: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { patError: 'id required' });
    const result = await apiFetch(`/api/me/pat-credentials/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { patError: result.error || 'delete failed' });
    return { patDeleted: true };
  }
};
;null as any as Actions;