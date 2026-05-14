import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface BlameCommit {
  sha: string;
  author: string;
  author_email: string;
  message: string;
  committed_at: string;
  files_changed: number;
  insertions: number;
  deletions: number;
  run_count: number;
  pass_count: number;
  fail_count: number;
  avg_duration_ms: number;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const blameResult = await apiFetch<BlameCommit[]>(
    `/api/projects/${params.id}/blame-timeline?limit=50`,
    { cookie: cookieHeader }
  );

  return {
    blameTimeline: blameResult.ok ? (blameResult.data ?? []) : []
  };
};

export const actions: Actions = {
  syncCommits: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const result = await apiFetch<{ synced: number; inserted: number }>(
      `/api/projects/${params.id}/sync-commits`,
      { method: 'POST', cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { error: result.error || 'commit sync failed' });
    return { synced: result.data?.synced ?? 0, inserted: result.data?.inserted ?? 0 };
  }
};
