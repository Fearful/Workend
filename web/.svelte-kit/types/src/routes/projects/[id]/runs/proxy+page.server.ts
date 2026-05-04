// @ts-nocheck
import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Run {
  id: string;
  task_id: string;
  project_id: string;
  commit_sha: string | null;
  branch: string | null;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  created_at: string;
}

const allowedStatuses = new Set(['succeeded', 'failed', 'cancelled', 'running', 'queued']);

export const load = async ({ params, url, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const status = url.searchParams.get('status') || '';
  const filter = allowedStatuses.has(status) ? status : '';

  const path = filter
    ? `/api/projects/${params.id}/runs?status=${encodeURIComponent(filter)}`
    : `/api/projects/${params.id}/runs`;

  const runsResult = await apiFetch<Run[]>(path, { cookie: cookieHeader });

  return {
    runs: runsResult.ok ? (runsResult.data ?? []) : [],
    filter
  };
};
