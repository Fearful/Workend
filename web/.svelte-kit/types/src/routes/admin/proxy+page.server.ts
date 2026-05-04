// @ts-nocheck
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
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
