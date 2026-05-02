import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Run {
  id: string;
  task_id: string;
  project_id: string;
  commit_sha: string | null;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  log_path: string;
  created_at: string;
  updated_at: string;
  task_name: string;
  task_source: string;
}

export const load: PageServerLoad = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const runResult = await apiFetch<Run>(`/api/runs/${params.id}`, { cookie: cookieHeader });
  if (runResult.status === 404) throw error(404, 'run not found');
  if (!runResult.ok || !runResult.data) throw error(500, runResult.error || 'failed to load');

  // For terminal runs, fetch the full log up-front. For active runs, the
  // browser will subscribe to the SSE stream.
  let log = '';
  const isTerminal =
    runResult.data.status === 'succeeded' ||
    runResult.data.status === 'failed' ||
    runResult.data.status === 'cancelled';
  if (isTerminal) {
    try {
      const apiURL = process.env.WORKEND_API_URL || 'http://api:8080';
      const r = await fetch(`${apiURL}/api/runs/${params.id}/log`, {
        headers: cookieHeader ? { Cookie: cookieHeader } : {}
      });
      if (r.ok) log = await r.text();
    } catch {
      // ignore
    }
  }

  return { run: runResult.data, log };
};
