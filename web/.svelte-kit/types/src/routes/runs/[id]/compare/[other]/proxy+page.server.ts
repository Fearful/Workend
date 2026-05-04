// @ts-nocheck
import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Run {
  id: string;
  task_id: string;
  task_name: string;
  task_source: string;
  status: string;
  exit_code: number | null;
  started_at: string | null;
  finished_at: string | null;
  commit_sha: string | null;
}

interface Hunk {
  op: number; // 0 equal, 1 insert (right-only), 2 delete (left-only)
  text: string;
}

interface CompareResponse {
  left: Run;
  right: Run;
  diff: Hunk[];
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch<CompareResponse>(
    `/api/runs/${params.id}/compare?to=${encodeURIComponent(params.other)}`,
    { cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined }
  );
  if (!result.ok || !result.data) throw error(result.status, result.error || 'compare failed');
  return result.data;
};
