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

interface OutboundWebhook {
  id: string;
  url: string;
  events: string[];
  secret_hash: string;
  created_at: string;
}

interface PushSubscriptionView {
  id: string;
  endpoint: string;
  created_at: string;
}

interface VapidKey {
  public_key: string;
}

export const load = async ({ locals, cookies, url }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [connResult, sshResult, patResult, webhookResult, pushSubsResult, vapidResult] = await Promise.all([
    apiFetch<ConnectionView[]>('/api/me/connections', { cookie: cookieHeader }),
    apiFetch<SSHKey[]>('/api/me/ssh-keys', { cookie: cookieHeader }),
    apiFetch<PATCredential[]>('/api/me/pat-credentials', { cookie: cookieHeader }),
    apiFetch<OutboundWebhook[]>('/api/me/outbound-webhooks', { cookie: cookieHeader }),
    apiFetch<PushSubscriptionView[]>('/api/me/push/subscriptions', { cookie: cookieHeader }),
    apiFetch<VapidKey>('/api/me/push/vapid', { cookie: cookieHeader })
  ]);
  // The connections endpoint isn't registered when no providers are configured.
  const providersConfigured = connResult.status !== 404;

  return {
    connections: connResult.ok ? (connResult.data ?? []) : [],
    providersConfigured,
    sshKeys: sshResult.ok ? (sshResult.data ?? []) : [],
    patCredentials: patResult.ok ? (patResult.data ?? []) : [],
    webhooks: webhookResult.ok ? (webhookResult.data ?? []) : [],
    pushSubscriptions: pushSubsResult.ok ? (pushSubsResult.data ?? []) : [],
    vapidPublicKey: vapidResult.ok ? (vapidResult.data?.public_key ?? null) : null,
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
  },

  addWebhook: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const url = String(data.get('url') || '').trim();
    const secret = String(data.get('secret') || '').trim();
    const events = data.getAll('events').map(String).filter(Boolean);
    if (!url || events.length === 0) {
      return fail(400, { webhookError: 'URL and at least one event are required', webhookUrl: url });
    }
    const body: Record<string, unknown> = { url, events };
    if (secret) body.secret = secret;
    const result = await apiFetch('/api/me/outbound-webhooks', {
      method: 'POST',
      body,
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { webhookError: result.error || 'failed', webhookUrl: url });
    return { webhookAdded: true };
  },

  deleteWebhook: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { webhookError: 'id required' });
    const result = await apiFetch(`/api/me/outbound-webhooks/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { webhookError: result.error || 'delete failed' });
    return { webhookDeleted: true };
  },

  testWebhook: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { webhookError: 'id required' });
    const result = await apiFetch(`/api/me/outbound-webhooks/${id}/test`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { webhookError: result.error || 'test failed' });
    return { webhookTested: true };
  },

  pushSubscribe: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const subscriptionJson = String(data.get('subscription') || '');
    if (!subscriptionJson) return fail(400, { pushError: 'subscription data required' });
    try {
      const subscription = JSON.parse(subscriptionJson);
      const result = await apiFetch('/api/me/push/subscribe', {
        method: 'POST',
        body: { subscription },
        cookie: cookieHeader
      });
      if (!result.ok) return fail(result.status, { pushError: result.error || 'subscribe failed' });
      return { pushSubscribed: true };
    } catch {
      return fail(400, { pushError: 'invalid subscription data' });
    }
  },

  pushUnsubscribe: async ({ cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch('/api/me/push/subscribe', {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { pushError: result.error || 'unsubscribe failed' });
    return { pushUnsubscribed: true };
  },

  deletePushSubscription: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const id = String((await request.formData()).get('id') || '');
    if (!id) return fail(400, { pushError: 'id required' });
    const result = await apiFetch(`/api/me/push/subscriptions/${id}`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { pushError: result.error || 'delete failed' });
    return { pushSubDeleted: true };
  }
};
;null as any as Actions;