import { fail, redirect } from '@sveltejs/kit';
import type { Actions, PageServerLoad } from './$types';
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

interface PinnedTask {
  task_id: string;
  task_name: string;
  task_source: string;
  task_command: string;
  project_id: string;
  project_name: string;
  workspace_id: string;
  workspace_name: string;
}

interface FlakyTask {
  task_id: string;
  task_name: string;
  task_source: string;
  project_id: string;
  project_name: string;
  workspace_name: string;
  total: number;
  succeeded: number;
  failed: number;
  flips: number;
  flip_rate: number;
  success_rate: number;
}

interface RunPulseRun {
  id: string;
  task_name: string;
  project_name: string;
  status: string;
  started_at: string;
  workspace_name: string;
}

interface RunPulseResponse {
  runs: RunPulseRun[];
}

interface HeatmapBucket {
  day_of_week: number;
  hour: number;
  failures: number;
}

interface FailureHeatmapResponse {
  days: number;
  buckets: HeatmapBucket[];
}

interface PendingApproval {
  run_id: string;
  task_name: string;
  project_name: string;
  created_at: string;
}

interface ExpiringSandbox {
  sandbox_id: string;
  branch: string;
  minutes_left: number;
}

interface MyQueueResponse {
  pending_approvals: PendingApproval[];
  expiring_sandboxes: ExpiringSandbox[];
  unread_mentions: number;
}

interface VelocityWeek {
  runs: number;
  passed: number;
  failed: number;
  avg_duration_ms: number;
}

interface SprintVelocityResponse {
  this_week: VelocityWeek;
  last_week: VelocityWeek;
  deltas: VelocityWeek;
}

interface SandboxEntry {
  id: string;
  branch: string;
  status: string;
  minutes_left: number;
  project_name: string;
}

interface SandboxStatusResponse {
  sandboxes: SandboxEntry[];
}

interface QuotaWorkspace {
  workspace_id: string;
  workspace_name: string;
  projects: number;
  runs_30d: number;
  usage_percent: number;
}

interface QuotaMeterResponse {
  workspaces: QuotaWorkspace[];
}

export const load: PageServerLoad = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [
    cardsResult,
    pinsResult,
    flakyResult,
    runPulseResult,
    heatmapResult,
    myQueueResult,
    velocityResult,
    sandboxResult,
    quotaResult
  ] = await Promise.all([
    apiFetch<Card[]>('/api/me/dashboard', { cookie: cookieHeader }),
    apiFetch<PinnedTask[]>('/api/me/pinned-tasks', { cookie: cookieHeader }),
    apiFetch<FlakyTask[]>('/api/me/flaky-tasks', { cookie: cookieHeader }),
    apiFetch<RunPulseResponse>('/api/widgets/dashboard/run-pulse', { cookie: cookieHeader }),
    apiFetch<FailureHeatmapResponse>('/api/widgets/dashboard/failure-heatmap?days=7', { cookie: cookieHeader }),
    apiFetch<MyQueueResponse>('/api/widgets/dashboard/my-queue', { cookie: cookieHeader }),
    apiFetch<SprintVelocityResponse>('/api/widgets/dashboard/sprint-velocity', { cookie: cookieHeader }),
    apiFetch<SandboxStatusResponse>('/api/widgets/dashboard/sandbox-status', { cookie: cookieHeader }),
    apiFetch<QuotaMeterResponse>('/api/widgets/dashboard/quota-meter', { cookie: cookieHeader })
  ]);

  const defaultQueue: MyQueueResponse = { pending_approvals: [], expiring_sandboxes: [], unread_mentions: 0 };
  const defaultVelocity: SprintVelocityResponse = {
    this_week: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 },
    last_week: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 },
    deltas: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 }
  };

  return {
    cards: cardsResult.ok ? (cardsResult.data ?? []) : [],
    pinned: pinsResult.ok ? (pinsResult.data ?? []) : [],
    flaky: flakyResult.ok ? (flakyResult.data ?? []) : [],
    error: cardsResult.ok ? null : (cardsResult.error || 'failed to load dashboard'),
    runPulse: runPulseResult.ok ? (runPulseResult.data?.runs ?? []) : [],
    heatmap: heatmapResult.ok ? (heatmapResult.data?.buckets ?? []) : [],
    myQueue: myQueueResult.ok ? (myQueueResult.data ?? defaultQueue) : defaultQueue,
    velocity: velocityResult.ok ? (velocityResult.data ?? defaultVelocity) : defaultVelocity,
    sandboxes: sandboxResult.ok ? (sandboxResult.data?.sandboxes ?? []) : [],
    quota: quotaResult.ok ? (quotaResult.data?.workspaces ?? []) : []
  };
};

export const actions: Actions = {
  runPinned: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    if (!taskID) return fail(400, { error: 'task_id required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch<{ id: string }>(`/api/tasks/${taskID}/runs`, {
      method: 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'run failed to start' });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  unpin: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    if (!taskID) return fail(400, { error: 'task_id required' });
    const cookie = cookies.get(SESSION_COOKIE);
    await apiFetch(`/api/tasks/${taskID}/pin`, {
      method: 'DELETE',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    return { unpinned: true };
  }
};
