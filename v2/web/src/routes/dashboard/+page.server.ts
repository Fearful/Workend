import { redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Card {
  project_id: string;
  project_name: string;
  workspace_id: string;
  workspace_name: string;
  status: string;
  last_synced_at: string | null;
  last_commit_sha: string | null;
  last_commit_message: string | null;
  total_lines: number | null;
  top_language: string | null;
  task_count: number;
  last_run_status: string | null;
  last_run_at: string | null;
  last_run_task_name: string | null;
  has_failing_recent_run: boolean;
}

export const load: PageServerLoad = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<Card[]>('/api/me/dashboard', {
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
  });
  return {
    cards: result.ok ? (result.data ?? []) : [],
    error: result.ok ? null : (result.error || 'failed to load dashboard')
  };
};
