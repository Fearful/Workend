// @ts-nocheck
import { fail, redirect } from '@sveltejs/kit';
import type { PageServerLoad, Actions } from './$types';
import { apiFetch } from '$lib/api';

const SESSION_COOKIE = 'workend_session';

interface Task {
  id: string;
  project_id: string;
  source: string;
  name: string;
  raw_command: string;
  detected_at: string;
  timeout_seconds: number | null;
  retry_max: number;
  retry_backoff_sec: number;
  requires_approval: boolean;
  max_concurrency: number;
  supersede_policy: 'queue' | 'cancel-old' | 'reject';
  needs_services: string[];
  artifact_patterns: string[];
}

interface PinnedTask {
  task_id: string;
}

interface Run {
  id: string;
  task_id: string;
  status: string;
  started_at: string | null;
  finished_at: string | null;
  exit_code: number | null;
  task_name: string;
  task_source: string;
  created_at: string;
}

interface Lang {
  files: number;
  lines: number;
  code: number;
  comments: number;
  blanks: number;
}

interface Stats {
  computed_at: string;
  total_files: number;
  total_lines: number;
  total_code: number;
  languages: Record<string, Lang>;
}

interface PipelineRun {
  id: string;
  provider_run_id: string;
  status: string;
  branch: string | null;
  commit_sha: string | null;
  workflow_name: string | null;
  html_url: string | null;
  started_at: string | null;
  finished_at: string | null;
  fetched_at: string;
}

interface PipelineConfig {
  id: string;
  project_id: string;
  ci_system: string;
  file_path: string;
  detected_at: string;
}

interface ComposeConfig {
  compose_file: string | null;
  services: string[];
  env_vars: { key: string; default_value: string }[];
  status: string;
  started_at: string | null;
}

interface OverviewBundle {
  readme: { path: string; bytes: number; truncated: boolean; content: string } | null;
  contributors: { name: string; commits: number }[];
  license: { path: string; kind: string } | null;
  has_contributing: boolean;
  has_changelog: boolean;
  recent_files: { path: string; modified: string }[];
}

export const load = async ({ params, locals, cookies }: Parameters<PageServerLoad>[0]) => {
  if (!locals.user) throw redirect(303, '/login');
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : undefined;

  const [tasksResult, runsResult, statsResult, pinsResult, pipelinesResult, ciConfigsResult, composeResult, overviewResult] = await Promise.all([
    apiFetch<Task[]>(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch<Run[]>(`/api/projects/${params.id}/runs`, { cookie: cookieHeader }),
    apiFetch<Stats | null>(`/api/projects/${params.id}/stats`, { cookie: cookieHeader }),
    apiFetch<PinnedTask[]>(`/api/me/pinned-tasks`, { cookie: cookieHeader }),
    apiFetch<PipelineRun[]>(`/api/projects/${params.id}/remote-pipelines`, { cookie: cookieHeader }),
    apiFetch<PipelineConfig[]>(`/api/projects/${params.id}/pipeline-configs`, { cookie: cookieHeader }),
    apiFetch<ComposeConfig>(`/api/projects/${params.id}/compose`, { cookie: cookieHeader }),
    apiFetch<OverviewBundle>(`/api/projects/${params.id}/overview`, { cookie: cookieHeader })
  ]);

  return {
    tasks: tasksResult.ok ? (tasksResult.data ?? []) : [],
    runs: runsResult.ok ? (runsResult.data ?? []) : [],
    stats: statsResult.ok ? (statsResult.data ?? null) : null,
    pinnedTaskIDs: pinsResult.ok ? (pinsResult.data ?? []).map((p) => p.task_id) : [],
    remotePipelines: pipelinesResult.ok ? (pipelinesResult.data ?? []) : [],
    pipelineConfigs: ciConfigsResult.ok ? (ciConfigsResult.data ?? []) : [],
    compose: composeResult.ok ? (composeResult.data ?? null) : null,
    overview: overviewResult.ok ? (overviewResult.data ?? null) : null
  };
};

export const actions = {
  run: async ({ request, cookies }: import('./$types').RequestEvent) => {
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
  runWithParams: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    if (!taskID) return fail(400, { error: 'task_id required' });

    let env: Record<string, string> = {};
    let args: string[] = [];
    try {
      env = JSON.parse(String(data.get('env') || '{}'));
      args = JSON.parse(String(data.get('args') || '[]'));
    } catch {
      return fail(400, { error: 'invalid env/args JSON' });
    }

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch<{ id: string }>(`/api/tasks/${taskID}/runs`, {
      method: 'POST',
      body: { env, args },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'run failed to start' });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  runOnBranch: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const branch = String(data.get('branch') || '').trim();
    if (!taskID) return fail(400, { error: 'task_id required' });
    if (!branch) return fail(400, { error: 'branch required' });

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch<{ id: string }>(`/api/tasks/${taskID}/runs`, {
      method: 'POST',
      body: { branch },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || 'run failed to start' });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  setRetry: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const maxRaw = String(data.get('retry_max') || '').trim();
    const backoffRaw = String(data.get('retry_backoff_sec') || '').trim();
    if (!taskID) return fail(400, { error: 'task_id required' });
    const body: { retry_max?: number; retry_backoff_sec?: number } = {};
    if (maxRaw !== '') {
      const n = parseInt(maxRaw, 10);
      if (!Number.isFinite(n) || n < 0 || n > 10) return fail(400, { error: 'retry_max must be 0..10' });
      body.retry_max = n;
    }
    if (backoffRaw !== '') {
      const n = parseInt(backoffRaw, 10);
      if (!Number.isFinite(n) || n < 1 || n > 3600) return fail(400, { error: 'retry_backoff_sec must be 1..3600' });
      body.retry_backoff_sec = n;
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body,
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'failed' });
    return { ok: true };
  },
  toggleApproval: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const next = String(data.get('next') || '') === 'true';
    if (!taskID) return fail(400, { error: 'task_id required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body: { requires_approval: next },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'failed' });
    return { ok: true };
  },
  togglePin: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const pinned = String(data.get('pinned') || '') === 'true';
    if (!taskID) return fail(400, { error: 'task_id required' });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/pin`, {
      method: pinned ? 'DELETE' : 'POST',
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'pin failed' });
    return { pinned: !pinned };
  },
  setTimeout: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const raw = String(data.get('timeout_seconds') || '').trim();
    if (!taskID) return fail(400, { error: 'task_id required' });

    let body: { timeout_seconds: number | null };
    if (raw === '') {
      body = { timeout_seconds: null };
    } else {
      const n = parseInt(raw, 10);
      if (!Number.isFinite(n) || n < 1 || n > 86400) {
        return fail(400, { error: 'timeout_seconds must be 1..86400 or blank' });
      }
      body = { timeout_seconds: n };
    }

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body,
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'update failed' });
    return { updated: true };
  },
  setArtifactPatterns: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const raw = String(data.get('artifact_patterns') || '').trim();
    if (!taskID) return fail(400, { error: 'task_id required' });

    const patterns = raw === '' ? [] :
      raw.split('\n').map((s) => s.trim()).filter(Boolean);
    if (patterns.length > 32) return fail(400, { error: 'too many patterns (max 32)' });

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body: { artifact_patterns: patterns },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'update failed' });
    return { updated: true };
  },
  setNeedsServices: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const raw = String(data.get('needs_services') || '').trim();
    if (!taskID) return fail(400, { error: 'task_id required' });

    const services = raw === '' ? [] : raw.split(',').map((s) => s.trim()).filter(Boolean);
    if (services.length > 32) return fail(400, { error: 'too many services (max 32)' });

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body: { needs_services: services },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'update failed' });
    return { updated: true };
  },
  setConcurrency: async ({ request, cookies }: import('./$types').RequestEvent) => {
    const data = await request.formData();
    const taskID = String(data.get('task_id') || '');
    const maxRaw = String(data.get('max_concurrency') || '').trim();
    const policy = String(data.get('supersede_policy') || 'queue');
    if (!taskID) return fail(400, { error: 'task_id required' });

    const max = parseInt(maxRaw, 10);
    if (!Number.isFinite(max) || max < 0 || max > 100) {
      return fail(400, { error: 'max_concurrency must be 0..100' });
    }
    if (policy !== 'queue' && policy !== 'cancel-old' && policy !== 'reject') {
      return fail(400, { error: 'supersede_policy must be queue|cancel-old|reject' });
    }

    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}`, {
      method: 'PATCH',
      body: { max_concurrency: max, supersede_policy: policy },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : undefined
    });
    if (!result.ok) return fail(result.status, { error: result.error || 'update failed' });
    return { updated: true };
  }
};
;null as any as Actions;