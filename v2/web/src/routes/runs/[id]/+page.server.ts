import { error, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
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

  // Recent runs of the same task — used by the "Compare with…" dropdown.
  const recentResult = await apiFetch<Run[]>(`/api/projects/${runResult.data.project_id}/runs`, { cookie: cookieHeader });
  const recentRuns = recentResult.ok && recentResult.data
    ? recentResult.data.filter((r) => r.task_id === runResult.data!.task_id && r.id !== params.id).slice(0, 10)
    : [];

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

  return { run: runResult.data, log, recentRuns };
};

export const actions: Actions = {
  rerun: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

    // Look up the original run to get its task_id, then start a new run
    // against that task. Note: this re-runs at *current* repo HEAD, not
    // the original commit — see PLAN.md Stage 7 (deferred follow-up).
    const original = await apiFetch<Run>(`/api/runs/${params.id}`, { cookie: cookieHeader });
    if (!original.ok || !original.data) throw error(500, 'failed to load original run');

    const newRun = await apiFetch<{ id: string }>(`/api/tasks/${original.data.task_id}/runs`, {
      method: 'POST',
      cookie: cookieHeader
    });
    if (!newRun.ok || !newRun.data) throw error(newRun.status, newRun.error || 'rerun failed');
    throw redirect(303, `/runs/${newRun.data.id}`);
  }
};
