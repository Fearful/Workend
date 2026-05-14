// @ts-nocheck
import { error, fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface User {
  id: string;
  email: string;
  display_name: string;
  is_admin: boolean;
  created_at: string;
  workspace_count: number;
  project_count: number;
  run_count: number;
}

interface AuditEntry {
  id: number;
  occurred_at: string;
  actor_id: string | null;
  actor_email: string | null;
  action: string;
  target_kind: string | null;
  target_id: string | null;
  ip: string | null;
  metadata: unknown;
}

export const load = async ({ locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  if (!locals.user.is_admin) throw error(403, 'admin only');

  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [usersResult, auditResult] = await Promise.all([
    apiFetch<User[]>('/api/admin/users', { cookie: cookieHeader }),
    apiFetch<AuditEntry[]>('/api/admin/audit-log', { cookie: cookieHeader })
  ]);

  return {
    users: usersResult.ok ? (usersResult.data ?? []) : [],
    audit: auditResult.ok ? (auditResult.data ?? []) : []
  };
};

export const actions = {
  setRetention: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const days = parseInt(String(data.get('days') || ''), 10);
    if (isNaN(days) || days < 1) {
      return fail(400, { retentionError: 'A valid number of days is required (minimum 1)' });
    }
    const result = await apiFetch('/api/admin/audit-log/retention', {
      method: 'POST',
      body: { days },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { retentionError: result.error || 'failed to set retention' });
    return { retentionSet: true, retentionDays: days };
  }
};
;null as any as Actions;