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

  // Fetch the log only when the run is in a terminal state. While running,
  // we'll fetch on each poll cycle. (Stage 6 swaps this to SSE.)
  let log = '';
  const logResult = await apiFetch<unknown>(`/api/runs/${params.id}/log`, { cookie: cookieHeader });
  if (logResult.ok && typeof logResult.data === 'string') {
    log = logResult.data;
  } else if (logResult.ok && logResult.status === 200) {
    // apiFetch returns data as parsed JSON; for plain text we need a raw fetch.
    log = '';
  }

  // Plain text fetch since apiFetch parses JSON
  try {
    const apiURL = process.env.WORKEND_API_URL || 'http://api:8080';
    const r = await fetch(`${apiURL}/api/runs/${params.id}/log`, {
      headers: cookieHeader ? { Cookie: cookieHeader } : {}
    });
    if (r.ok) log = await r.text();
  } catch {
    // ignore
  }

  return { run: runResult.data, log };
};
