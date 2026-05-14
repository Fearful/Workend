import { error, fail, redirect } from '@sveltejs/kit';
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
  timed_out: boolean;
  log_path: string;
  params: { env: Record<string, string> | null; args: string[] | null };
  attempt: number;
  parent_run_id: string | null;
  created_at: string;
  updated_at: string;
  task_name: string;
  task_source: string;
}

interface Comment {
  id: string;
  run_id: string;
  user_id: string;
  user_display_name: string;
  body: string;
  created_at: string;
}

interface Artifact {
  id: string;
  run_id: string;
  relative_path: string;
  size_bytes: number;
  mime_type: string | null;
  captured_at: string;
}

interface RunShare {
  id: string;
  token: string;
  created_by: string;
  expires_at: string;
  created_at: string;
}

interface RunVerification {
  signature: string;
  key_id: string;
  verified: boolean;
  signed_at: string;
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

  const commentsResult = await apiFetch<Comment[]>(`/api/runs/${params.id}/comments`, { cookie: cookieHeader });
  const comments = commentsResult.ok ? (commentsResult.data ?? []) : [];

  const artifactsResult = await apiFetch<Artifact[]>(`/api/runs/${params.id}/artifacts`, { cookie: cookieHeader });
  const artifacts = artifactsResult.ok ? (artifactsResult.data ?? []) : [];

  // Whether this run's task is currently pinned by the user — drives the
  // pin/unpin button on the run hero.
  const pinsResp = await apiFetch<{ task_id: string }[]>(`/api/me/pinned-tasks`, { cookie: cookieHeader });
  const pinnedTaskIDs = pinsResp.ok ? (pinsResp.data ?? []).map((p) => p.task_id) : [];
  const isTaskPinned = pinnedTaskIDs.includes(runResult.data.task_id);

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

  // Sharing & signing data
  const sharesResult = await apiFetch<RunShare[]>(`/api/runs/${params.id}/shares`, { cookie: cookieHeader });
  const shares = sharesResult.ok ? (sharesResult.data ?? []) : [];

  let verification: RunVerification | null = null;
  if (isTerminal) {
    const verifyResult = await apiFetch<RunVerification>(`/api/runs/${params.id}/verify`, { cookie: cookieHeader });
    verification = verifyResult.ok ? (verifyResult.data ?? null) : null;
  }

  return { run: runResult.data, log, recentRuns, comments, artifacts, isTaskPinned, shares, verification };
};

export const actions: Actions = {
  rerun: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

    // Re-run against current repo HEAD (no `at_commit`). For pinned re-runs
    // see the rerunPinned action below.
    const original = await apiFetch<Run>(`/api/runs/${params.id}`, { cookie: cookieHeader });
    if (!original.ok || !original.data) throw error(500, 'failed to load original run');

    const newRun = await apiFetch<{ id: string }>(`/api/tasks/${original.data.task_id}/runs`, {
      method: 'POST',
      cookie: cookieHeader
    });
    if (!newRun.ok || !newRun.data) throw error(newRun.status, newRun.error || 'rerun failed');
    throw redirect(303, `/runs/${newRun.data.id}`);
  },
  comment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const body = String(data.get('body') || '').trim();
    if (!body) return { commentError: 'comment body required', commentDraft: body };
    const result = await apiFetch<Comment>(`/api/runs/${params.id}/comments`, {
      method: 'POST',
      body: { body },
      cookie: cookieHeader
    });
    if (!result.ok) {
      return { commentError: result.error || 'failed to post comment', commentDraft: body };
    }
    return { commented: true };
  },
  deleteComment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const id = String(data.get('id') || '');
    if (!id) return { commentError: 'id required' };
    const result = await apiFetch(`/api/runs/${params.id}/comments/${id}`, {
      method: 'DELETE',
      cookie: cookieHeader
    });
    if (!result.ok) {
      return { commentError: result.error || 'failed to delete' };
    }
    return { commentDeleted: true };
  },
  approve: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const approved = String(data.get('approved') || '') === 'true';
    const note = String(data.get('note') || '');
    const r = await apiFetch(`/api/runs/${params.id}/approve`, {
      method: 'POST',
      body: { approved, note },
      cookie: cookieHeader
    });
    if (!r.ok) return { approvalError: r.error || 'failed' };
    return { approvalDecided: approved };
  },
  togglePin: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const pinned = String(data.get('pinned') || '') === 'true';
    if (!taskID) return { pinError: 'task_id required' };
    const result = await apiFetch(`/api/tasks/${taskID}/pin`, {
      method: pinned ? 'DELETE' : 'POST',
      cookie: cookieHeader
    });
    if (!result.ok) return { pinError: result.error || 'pin failed' };
    return { pinToggled: !pinned };
  },
  rerunPinned: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

    const original = await apiFetch<Run>(`/api/runs/${params.id}`, { cookie: cookieHeader });
    if (!original.ok || !original.data) throw error(500, 'failed to load original run');
    if (!original.data.commit_sha) throw error(400, 'original run has no recorded commit');

    const newRun = await apiFetch<{ id: string }>(`/api/tasks/${original.data.task_id}/runs`, {
      method: 'POST',
      body: { at_commit: original.data.commit_sha },
      cookie: cookieHeader
    });
    if (!newRun.ok || !newRun.data) throw error(newRun.status, newRun.error || 'rerun failed');
    throw redirect(303, `/runs/${newRun.data.id}`);
  },

  share: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();
    const expiresRaw = String(fd.get('expires_hours') || '').trim();

    const body: { expires_hours?: number } = {};
    if (expiresRaw) {
      const hours = parseInt(expiresRaw, 10);
      if (!Number.isFinite(hours) || hours < 1 || hours > 720) {
        return fail(400, { shareError: 'expires_hours must be 1-720' });
      }
      body.expires_hours = hours;
    }

    const r = await apiFetch<RunShare>(`/api/runs/${params.id}/share`, {
      method: 'POST',
      body,
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { shareError: r.error || 'failed to create share link' });
    return { shareCreated: true };
  },

  revokeShare: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();
    const shareId = String(fd.get('share_id') || '');
    if (!shareId) return fail(400, { shareError: 'share_id required' });

    const r = await apiFetch(`/api/run-shares/${shareId}`, {
      method: 'DELETE',
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { shareError: r.error || 'failed to revoke share' });
    return { shareRevoked: true };
  },

  sign: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;
    const fd = await request.formData();
    const key_id = String(fd.get('key_id') || '').trim();
    const private_key = String(fd.get('private_key') || '').trim();

    if (!key_id) return fail(400, { signError: 'key_id is required' });
    if (!private_key) return fail(400, { signError: 'private_key is required' });

    const r = await apiFetch(`/api/runs/${params.id}/sign`, {
      method: 'POST',
      body: { key_id, private_key },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { signError: r.error || 'failed to sign run' });
    return { signed: true };
  }
};
