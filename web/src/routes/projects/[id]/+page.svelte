<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import { shortSha as _shortSha, formatDuration } from '$lib/utils';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import TaskSparkline from '$lib/components/TaskSparkline.svelte';

  let { data, form } = $props();

  function shortSha(sha: string | null): string {
    return _shortSha(sha, 12);
  }

  let pollHandle: ReturnType<typeof setInterval> | null = null;

  function maybeStartPolling() {
    const transient = data.project.status === 'cloning' || data.project.status === 'pending' ||
      data.runs.some((r) => r.status === 'queued' || r.status === 'running');
    if (transient && !pollHandle) {
      pollHandle = setInterval(() => invalidateAll(), 2000);
    } else if (!transient && pollHandle) {
      clearInterval(pollHandle);
      pollHandle = null;
    }
  }

  $effect(() => {
    maybeStartPolling();
  });

  let collapsedPrefixes = $state(new Set<string>());
  let taskFilter = $state('');
  let openMenuTaskID = $state<string | null>(null);

  // Group recent runs by task so each task row can render a sparkline of its
  // own history without an extra API round-trip. data.runs is the project's
  // recent-runs list (default ~50).
  let runsByTask = $derived.by(() => {
    const out: Record<string, { id: string; status: string; created_at: string }[]> = {};
    for (const r of data.runs) {
      if (!out[r.task_id]) out[r.task_id] = [];
      out[r.task_id].push({ id: r.id, status: r.status, created_at: r.created_at });
    }
    return out;
  });

  type TaskGroup = { ungrouped: typeof data.tasks; prefixed: Record<string, typeof data.tasks> };
  let tasksBySource = $derived.by(() => {
    const raw: Record<string, { all: typeof data.tasks; byPrefix: Record<string, typeof data.tasks> }> = {};
    for (const t of data.tasks) {
      if (taskFilter && !t.name.toLowerCase().includes(taskFilter.toLowerCase())) continue;
      if (!raw[t.source]) raw[t.source] = { all: [], byPrefix: {} };
      const sepIdx = Math.max(t.name.indexOf(':'), t.name.indexOf('.'));
      if (sepIdx > 0) {
        const prefix = t.name.slice(0, sepIdx + 1);
        if (!raw[t.source].byPrefix[prefix]) raw[t.source].byPrefix[prefix] = [];
        raw[t.source].byPrefix[prefix].push(t);
      } else {
        raw[t.source].all.push(t);
      }
    }
    const result: Record<string, TaskGroup> = {};
    for (const [source, g] of Object.entries(raw)) {
      result[source] = { ungrouped: [...g.all], prefixed: {} };
      for (const [prefix, tasks] of Object.entries(g.byPrefix)) {
        if (tasks.length >= 2) {
          result[source].prefixed[prefix] = tasks;
        } else {
          result[source].ungrouped.push(...tasks);
        }
      }
    }
    return result;
  });

  let topLangs = $derived.by(() => {
    if (!data.stats) return [];
    return Object.entries(data.stats.languages)
      .sort((a, b) => b[1].lines - a[1].lines)
      .slice(0, 6);
  });

  let totalLangLines = $derived(data.stats?.total_lines || 1);

  function formatTimeout(sec: number | null | undefined): string {
    if (!sec || sec <= 0) return 'default';
    if (sec < 60) return `${sec}s`;
    if (sec < 3600) return `${Math.round(sec / 60)}m`;
    return `${Math.round(sec / 360) / 10}h`;
  }

  async function startCompose() {
    const inputs = document.querySelectorAll<HTMLInputElement>('.compose-env-input');
    const envVars: Record<string, string> = {};
    inputs.forEach((input) => {
      const key = input.dataset.envKey;
      if (key) envVars[key] = input.value;
    });
    await fetch(`/api/projects/${data.project.id}/compose/start`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ env_vars: envVars }),
      credentials: 'same-origin'
    });
    await invalidateAll();
  }

  async function stopCompose() {
    await fetch(`/api/projects/${data.project.id}/compose/stop`, {
      method: 'POST',
      credentials: 'same-origin'
    });
    await invalidateAll();
  }

  let paramsModal = $state<{ taskID: string; taskName: string; env: string; args: string } | null>(null);
  let runError = $state<string | null>(null);
  let timeoutModal = $state<{ taskID: string; taskName: string; value: string; error: string | null } | null>(null);
  let retryModal = $state<{ taskID: string; taskName: string; max: string; backoff: string; error: string | null } | null>(null);
  let concurrencyModal = $state<{ taskID: string; taskName: string; max: string; policy: 'queue' | 'cancel-old' | 'reject'; error: string | null } | null>(null);
  let branchModal = $state<{ taskID: string; taskName: string; branches: { name: string; commit_sha: string }[] | null; selected: string; loading: boolean; error: string | null } | null>(null);
  let servicesModal = $state<{ taskID: string; taskName: string; raw: string; error: string | null } | null>(null);
  let artifactsModal = $state<{ taskID: string; taskName: string; raw: string; error: string | null } | null>(null);

  // -- Blame timeline state --
  let syncingCommits = $state(false);
  let blameDetailModal = $state<{
    sha: string;
    loading: boolean;
    commit: { sha: string; author: string; author_email: string; message: string; committed_at: string; files_changed: number; insertions: number; deletions: number } | null;
    runs: { id: string; status: string; started_at: string | null; finished_at: string | null; duration_ms: number }[];
    error: string | null;
  } | null>(null);

  async function syncCommits() {
    if (syncingCommits) return;
    syncingCommits = true;
    try {
      const r = await fetch(`/projects/${data.project.id}?/syncCommits`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      syncingCommits = false;
    }
  }

  async function openBlameDetail(sha: string) {
    blameDetailModal = { sha, loading: true, commit: null, runs: [], error: null };
    try {
      const r = await fetch(`/api/commits/${sha}`, { credentials: 'same-origin' });
      if (!r.ok) {
        if (blameDetailModal) blameDetailModal = { ...blameDetailModal, loading: false, error: `HTTP ${r.status}` };
        return;
      }
      const json = await r.json() as { commit: typeof blameDetailModal.commit; runs: typeof blameDetailModal.runs };
      if (blameDetailModal) {
        blameDetailModal = { ...blameDetailModal, loading: false, commit: json.commit, runs: json.runs ?? [] };
      }
    } catch (err) {
      if (blameDetailModal) {
        blameDetailModal = { ...blameDetailModal, loading: false, error: err instanceof Error ? err.message : 'failed' };
      }
    }
  }

  function blameRowColor(c: { run_count: number; pass_count: number; fail_count: number }): string {
    if (c.run_count === 0) return '';
    if (c.fail_count > 0) return 'blame-fail';
    return 'blame-pass';
  }

  // -- Monorepo state --
  let detectingPackages = $state(false);
  let autoMapping = $state(false);
  let packageDetailModal = $state<{
    id: string;
    loading: boolean;
    pkg: { id: string; name: string; path: string; pkg_type: string } | null;
    scopes: { id: string; task_id: string; task_name: string }[];
    error: string | null;
  } | null>(null);

  async function detectPackages() {
    if (detectingPackages) return;
    detectingPackages = true;
    try {
      const r = await fetch(`/projects/${data.project.id}?/detectPackages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      detectingPackages = false;
    }
  }

  async function autoMapTasks() {
    if (autoMapping) return;
    autoMapping = true;
    try {
      const r = await fetch(`/projects/${data.project.id}?/autoMapTasks`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      autoMapping = false;
    }
  }

  async function openPackageDetail(pkgID: string) {
    packageDetailModal = { id: pkgID, loading: true, pkg: null, scopes: [], error: null };
    try {
      const r = await fetch(`/api/packages/${pkgID}`, { credentials: 'same-origin' });
      if (!r.ok) {
        if (packageDetailModal) packageDetailModal = { ...packageDetailModal, loading: false, error: `HTTP ${r.status}` };
        return;
      }
      const json = await r.json() as { id: string; name: string; path: string; pkg_type: string; scopes: { id: string; task_id: string; task_name: string }[] };
      if (packageDetailModal) {
        packageDetailModal = { ...packageDetailModal, loading: false, pkg: json, scopes: json.scopes ?? [] };
      }
    } catch (err) {
      if (packageDetailModal) {
        packageDetailModal = { ...packageDetailModal, loading: false, error: err instanceof Error ? err.message : 'failed' };
      }
    }
  }

  async function addTaskScope(pkgID: string, taskID: string) {
    const r = await fetch(`/api/packages/${pkgID}/task-scopes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ task_id: taskID }),
      credentials: 'same-origin'
    });
    if (r.ok) {
      await openPackageDetail(pkgID);
      await invalidateAll();
    }
  }

  async function removeTaskScope(scopeID: string) {
    const r = await fetch(`/api/task-scopes/${scopeID}`, {
      method: 'DELETE',
      credentials: 'same-origin'
    });
    if (r.ok && packageDetailModal) {
      await openPackageDetail(packageDetailModal.id);
      await invalidateAll();
    }
  }

  function pkgTypeBadgeVariant(t: string): 'info' | 'success' | 'warning' | 'danger' | 'muted' | 'accent' {
    switch (t) {
      case 'npm': return 'accent';
      case 'go': return 'info';
      case 'cargo': return 'warning';
      case 'python': return 'success';
      case 'gradle': return 'danger';
      default: return 'muted';
    }
  }

  // -- Preview state --
  let newPreviewBranch = $state('');
  let previewActionPending = $state<string | null>(null);

  async function redeployPreview(previewID: string) {
    previewActionPending = previewID;
    try {
      await fetch(`/api/previews/${previewID}/redeploy`, { method: 'POST', credentials: 'same-origin' });
      await invalidateAll();
    } finally {
      previewActionPending = null;
    }
  }

  async function stopPreview(previewID: string) {
    previewActionPending = previewID;
    try {
      await fetch(`/api/previews/${previewID}/stop`, { method: 'POST', credentials: 'same-origin' });
      await invalidateAll();
    } finally {
      previewActionPending = null;
    }
  }

  let previewDeleteConfirm = $state<string | null>(null);

  async function deletePreview(previewID: string) {
    previewActionPending = previewID;
    try {
      await fetch(`/api/previews/${previewID}`, { method: 'DELETE', credentials: 'same-origin' });
      previewDeleteConfirm = null;
      await invalidateAll();
    } finally {
      previewActionPending = null;
    }
  }

  async function toggleAutoDeploy(previewID: string, current: boolean) {
    await fetch(`/api/previews/${previewID}/auto-deploy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: !current }),
      credentials: 'same-origin'
    });
    await invalidateAll();
  }

  // -- Task metrics state --
  let metricsModal = $state<{
    taskID: string;
    taskName: string;
    loading: boolean;
    metrics: { avg_duration_ms: number; p50_ms: number; p95_ms: number; p99_ms: number; total_runs: number; success_rate: number; last_30d_runs: number } | null;
    trends: { date: string; count: number; passed: number; failed: number; avg_ms: number }[];
    error: string | null;
  } | null>(null);

  async function openTaskMetrics(taskID: string, taskName: string) {
    metricsModal = { taskID, taskName, loading: true, metrics: null, trends: [], error: null };
    try {
      const [metricsRes, trendsRes] = await Promise.all([
        fetch(`/api/tasks/${taskID}/metrics`, { credentials: 'same-origin' }),
        fetch(`/api/tasks/${taskID}/metrics/trends`, { credentials: 'same-origin' })
      ]);
      if (!metricsRes.ok) {
        if (metricsModal) metricsModal = { ...metricsModal, loading: false, error: `HTTP ${metricsRes.status}` };
        return;
      }
      const metricsData = await metricsRes.json();
      const trendsData = trendsRes.ok ? await trendsRes.json() : [];
      if (metricsModal) {
        metricsModal = { ...metricsModal, loading: false, metrics: metricsData, trends: Array.isArray(trendsData) ? trendsData : [] };
      }
    } catch (err) {
      if (metricsModal) {
        metricsModal = { ...metricsModal, loading: false, error: err instanceof Error ? err.message : 'failed' };
      }
    }
  }

  function formatMs(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  }

  // -- Alert rule state --
  let alertFormOpen = $state(false);
  let alertFormTaskID = $state('');
  let alertFormType = $state('duration');
  let alertFormThreshold = $state('');
  let alertFormComparison = $state('gt');

  function openParams(taskID: string, taskName: string) {
    runError = null;
    paramsModal = { taskID, taskName, env: '', args: '' };
    openMenuTaskID = null;
  }

  function openTimeoutEditor(taskID: string, taskName: string, current: number | null) {
    timeoutModal = { taskID, taskName, value: current ? String(current) : '', error: null };
    openMenuTaskID = null;
  }

  function openRetryEditor(taskID: string, taskName: string, currentMax: number, currentBackoff: number) {
    retryModal = {
      taskID, taskName,
      max: String(currentMax),
      backoff: String(currentBackoff),
      error: null
    };
    openMenuTaskID = null;
  }

  function openConcurrencyEditor(taskID: string, taskName: string, currentMax: number, currentPolicy: 'queue' | 'cancel-old' | 'reject') {
    concurrencyModal = {
      taskID, taskName,
      max: String(currentMax),
      policy: currentPolicy,
      error: null
    };
    openMenuTaskID = null;
  }

  async function submitConcurrency() {
    if (!concurrencyModal) return;
    const max = parseInt(concurrencyModal.max.trim(), 10);
    if (!Number.isFinite(max) || max < 0 || max > 100) {
      concurrencyModal.error = 'Max concurrent runs must be 0–100 (0 = use single-run guard).';
      return;
    }
    const r = await fetch(`/projects/${data.project.id}?/setConcurrency`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        task_id: concurrencyModal.taskID,
        max_concurrency: String(max),
        supersede_policy: concurrencyModal.policy
      })
    });
    if (!r.ok) {
      concurrencyModal.error = 'Failed to update concurrency policy.';
      return;
    }
    concurrencyModal = null;
    await invalidateAll();
  }

  function formatConcurrency(max: number, policy: string): string {
    if (max <= 0) return 'single';
    return `${max}× (${policy})`;
  }

  function openServicesEditor(taskID: string, taskName: string, current: string[]) {
    servicesModal = { taskID, taskName, raw: current.join(', '), error: null };
    openMenuTaskID = null;
  }

  function openArtifactsEditor(taskID: string, taskName: string, current: string[]) {
    artifactsModal = { taskID, taskName, raw: current.join('\n'), error: null };
    openMenuTaskID = null;
  }

  async function submitArtifacts() {
    if (!artifactsModal) return;
    const r = await fetch(`/projects/${data.project.id}?/setArtifactPatterns`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        task_id: artifactsModal.taskID,
        artifact_patterns: artifactsModal.raw
      })
    });
    if (!r.ok) {
      artifactsModal.error = 'Failed to update artifact patterns.';
      return;
    }
    artifactsModal = null;
    await invalidateAll();
  }

  function formatArtifacts(arr: string[]): string {
    if (!arr || arr.length === 0) return 'none';
    return `${arr.length} pattern${arr.length === 1 ? '' : 's'}`;
  }

  async function submitServices() {
    if (!servicesModal) return;
    const r = await fetch(`/projects/${data.project.id}?/setNeedsServices`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        task_id: servicesModal.taskID,
        needs_services: servicesModal.raw
      })
    });
    if (!r.ok) {
      servicesModal.error = 'Failed to update services.';
      return;
    }
    servicesModal = null;
    await invalidateAll();
  }

  function formatServices(arr: string[]): string {
    if (!arr || arr.length === 0) return 'none';
    if (arr.length <= 2) return arr.join(', ');
    return `${arr.slice(0, 2).join(', ')} +${arr.length - 2}`;
  }

  async function openBranchPicker(taskID: string, taskName: string) {
    branchModal = { taskID, taskName, branches: null, selected: '', loading: true, error: null };
    openMenuTaskID = null;
    try {
      const r = await fetch(`/api/projects/${data.project.id}/branches`, { credentials: 'same-origin' });
      if (!r.ok) {
        if (branchModal) branchModal.error = `Failed to list branches (HTTP ${r.status})`;
      } else {
        const json = await r.json() as { branches: { name: string; commit_sha: string; default?: boolean }[] };
        const list = json.branches ?? [];
        // Default selection: project default branch if present, else first.
        const def = list.find((b) => b.default) ?? list[0];
        if (branchModal) {
          branchModal.branches = list;
          branchModal.selected = def?.name ?? '';
        }
      }
    } catch (err) {
      if (branchModal) branchModal.error = `Failed to list branches: ${err instanceof Error ? err.message : err}`;
    } finally {
      if (branchModal) branchModal.loading = false;
    }
  }

  async function submitWithParams() {
    if (!paramsModal) return;
    const env: Record<string, string> = {};
    for (const line of paramsModal.env.split('\n')) {
      const trimmed = line.trim();
      if (!trimmed) continue;
      const eq = trimmed.indexOf('=');
      if (eq < 0) {
        runError = `Bad env line "${trimmed}" — use KEY=VALUE`;
        return;
      }
      env[trimmed.slice(0, eq).trim()] = trimmed.slice(eq + 1);
    }
    const args = paramsModal.args
      .split('\n')
      .map((s) => s.trim())
      .filter((s) => s.length > 0);

    const r = await fetch(`/projects/${data.project.id}?/runWithParams`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        task_id: paramsModal.taskID,
        env: JSON.stringify(env),
        args: JSON.stringify(args)
      })
    });
    if (!r.ok) {
      runError = `Failed to start run (HTTP ${r.status})`;
      return;
    }
    const result = (await r.json()) as { type: string; data?: string; location?: string };
    if (result.location) {
      window.location.href = result.location;
      return;
    }
    if (result.type === 'success' && result.data) {
      const parsed = JSON.parse(result.data) as { runID?: string };
      if (parsed.runID) {
        window.location.href = `/runs/${parsed.runID}`;
        return;
      }
    }
    paramsModal = null;
    await invalidateAll();
  }

  async function submitTimeout() {
    if (!timeoutModal) return;
    const trimmed = timeoutModal.value.trim();
    let payload = '';
    if (trimmed === '') {
      payload = '';
    } else {
      const n = parseInt(trimmed, 10);
      if (!Number.isFinite(n) || n < 1 || n > 86400) {
        timeoutModal.error = 'Enter 1–86400 seconds, or leave blank for the default.';
        return;
      }
      payload = String(n);
    }
    const r = await fetch(`/projects/${data.project.id}?/setTimeout`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ task_id: timeoutModal.taskID, timeout_seconds: payload })
    });
    if (!r.ok) {
      timeoutModal.error = 'Failed to update timeout.';
      return;
    }
    timeoutModal = null;
    await invalidateAll();
  }

  async function submitRetry() {
    if (!retryModal) return;
    const max = parseInt(retryModal.max.trim(), 10);
    if (!Number.isFinite(max) || max < 0 || max > 10) {
      retryModal.error = 'Max retries must be 0–10.';
      return;
    }
    let backoff = parseInt(retryModal.backoff.trim(), 10);
    if (max > 0) {
      if (!Number.isFinite(backoff) || backoff < 1 || backoff > 3600) {
        retryModal.error = 'Backoff must be 1–3600 seconds.';
        return;
      }
    } else if (!Number.isFinite(backoff)) {
      backoff = 0;
    }
    const r = await fetch(`/projects/${data.project.id}?/setRetry`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({
        task_id: retryModal.taskID,
        retry_max: String(max),
        retry_backoff_sec: String(backoff)
      })
    });
    if (!r.ok) {
      retryModal.error = 'Failed to update retry policy.';
      return;
    }
    retryModal = null;
    await invalidateAll();
  }

  function formatRetry(max: number, backoff: number): string {
    if (max <= 0) return 'no retry';
    return `${max}× / ${backoff}s`;
  }

  function toggleTaskMenu(id: string) {
    openMenuTaskID = openMenuTaskID === id ? null : id;
  }

  function closeMenuOnClickOutside(e: MouseEvent) {
    const t = e.target as HTMLElement;
    if (!t.closest('.task-menu') && !t.closest('.task-menu-trigger')) {
      openMenuTaskID = null;
    }
  }

  // ---- Panel reordering ----
  const PANEL_KEYS = ['tasks', 'recent-runs', 'repository', 'latest-commit', 'code-stats', 'about', 'contributors', 'pipelines', 'pipeline-configs', 'compose', 'blame-timeline', 'monorepo', 'previews', 'alert-rules', 'impact-radar', 'dependency-tree', 'live-run'] as const;
  type PanelKey = typeof PANEL_KEYS[number];
  const STORAGE_KEY = 'workend.project-overview-layout';
  const LEGACY_STORAGE_KEY = 'workend.project-overview-order';
  const MAX_COLS = 3;

  type PanelLayout = PanelKey[][];

  let panelLayout = $state<PanelLayout>(PANEL_KEYS.map((k) => [k]));
  let dragSource = $state<{ row: number; col: number } | null>(null);
  let dropTarget = $state<{ row: number; col: number; side: 'left' | 'right' } | null>(null);
  let dropRowGap = $state<number | null>(null); // row index where new row would be inserted
  let editLayoutMode = $state(false);

  function isValidLayout(v: unknown): v is PanelKey[][] {
    if (!Array.isArray(v)) return false;
    const known = new Set<string>(PANEL_KEYS);
    for (const row of v) {
      if (!Array.isArray(row)) return false;
      if (row.length === 0 || row.length > MAX_COLS) return false;
      for (const k of row) if (typeof k !== 'string' || !known.has(k)) return false;
    }
    return true;
  }

  function flatten(layout: PanelLayout): PanelKey[] {
    return layout.flat();
  }

  function ensureAllPanels(layout: PanelLayout): PanelLayout {
    const present = new Set<PanelKey>(flatten(layout));
    const missing = PANEL_KEYS.filter((k) => !present.has(k));
    if (missing.length === 0) return layout;
    return [...layout, ...missing.map((k) => [k] as PanelKey[])];
  }

  function loadLayout() {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved) {
        const parsed = JSON.parse(saved);
        if (isValidLayout(parsed)) {
          panelLayout = ensureAllPanels(parsed);
          return;
        }
      }
      const legacy = localStorage.getItem(LEGACY_STORAGE_KEY);
      if (legacy) {
        const parsed = JSON.parse(legacy);
        if (Array.isArray(parsed)) {
          const known = new Set<string>(PANEL_KEYS);
          const flat: PanelKey[] = [];
          for (const k of parsed) if (known.has(k) && !flat.includes(k as PanelKey)) flat.push(k as PanelKey);
          for (const k of PANEL_KEYS) if (!flat.includes(k)) flat.push(k);
          panelLayout = flat.map((k) => [k]);
          persistLayout();
        }
      }
    } catch { /* ignore */ }
  }

  function persistLayout() {
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(panelLayout)); } catch { /* ignore */ }
  }

  function resetLayout() {
    panelLayout = PANEL_KEYS.map((k) => [k]);
    try { localStorage.removeItem(STORAGE_KEY); localStorage.removeItem(LEGACY_STORAGE_KEY); } catch { /* ignore */ }
  }

  function clearDrag() {
    dragSource = null;
    dropTarget = null;
    dropRowGap = null;
  }

  function panelHasData(key: PanelKey): boolean {
    switch (key) {
      case 'tasks':
      case 'repository':
      case 'latest-commit':
        return true;
      case 'recent-runs':
        return data.runs.length > 0;
      case 'code-stats':
        return data.stats != null;
      case 'pipelines':
        return data.remotePipelines.length > 0;
      case 'pipeline-configs':
        return data.pipelineConfigs.length > 0;
      case 'compose':
        return !!data.compose?.compose_file;
      case 'about':
        return !!data.overview?.readme;
      case 'contributors':
        return (data.overview?.contributors.length ?? 0) > 0;
      case 'blame-timeline':
        return data.blameTimeline.length > 0;
      case 'monorepo':
        return data.monorepoPackages.length > 0;
      case 'previews':
        return data.previews.length > 0;
      case 'alert-rules':
        return data.alertRules.length > 0 || data.widgetAlertRules.length > 0;
      case 'impact-radar':
        return data.impactRadar != null && data.impactRadar.tasks.length > 0;
      case 'dependency-tree':
        return data.dependencyTree != null &&
          (data.dependencyTree.upstream.length > 0 || data.dependencyTree.downstream.length > 0);
      case 'live-run':
        return data.liveRun != null;
    }
  }

  function visibleRow(row: PanelKey[]): PanelKey[] {
    if (editLayoutMode) return row;
    return row.filter(panelHasData);
  }

  function panelEmptyLabel(key: PanelKey): string {
    return ({
      tasks: 'Tasks',
      'recent-runs': 'Recent runs',
      repository: 'Repository',
      'latest-commit': 'Latest commit',
      'code-stats': 'Code statistics',
      pipelines: 'CI/CD pipelines',
      'pipeline-configs': 'CI/CD configuration',
      compose: 'Docker Compose',
      about: 'About',
      contributors: 'Contributors',
      'blame-timeline': 'Blame timeline',
      monorepo: 'Monorepo packages',
      previews: 'Deploy previews',
      'alert-rules': 'Alert rules',
      'impact-radar': 'Impact radar',
      'dependency-tree': 'Dependency tree',
      'live-run': 'Live run'
    } as Record<PanelKey, string>)[key];
  }

  function panelEmptyKind(key: PanelKey): string {
    return ({
      tasks: 'tasks',
      'recent-runs': 'runs',
      repository: 'repository data',
      'latest-commit': 'commits',
      'code-stats': 'code statistics',
      pipelines: 'CI runs',
      'pipeline-configs': 'CI configs',
      compose: 'a Docker Compose file',
      about: 'a README',
      contributors: 'commit history',
      'blame-timeline': 'synced commits',
      monorepo: 'detected packages',
      previews: 'deploy previews',
      'alert-rules': 'alert rules',
      'impact-radar': 'impact data',
      'dependency-tree': 'dependency data',
      'live-run': 'a running task'
    } as Record<PanelKey, string>)[key];
  }

  function onPanelDragStart(row: number, col: number, e: DragEvent) {
    if (!editLayoutMode) return;
    dragSource = { row, col };
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', `${row},${col}`);
    }
  }

  function onPanelDragOver(row: number, col: number, e: DragEvent) {
    if (!dragSource) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    const target = e.currentTarget as HTMLElement;
    const rect = target.getBoundingClientRect();
    const side: 'left' | 'right' = e.clientX < rect.left + rect.width / 2 ? 'left' : 'right';
    if (dragSource.row === row && (dragSource.col === col || (side === 'right' && dragSource.col === col + 1) || (side === 'left' && dragSource.col === col - 1))) {
      dropTarget = null;
      dropRowGap = null;
      return;
    }
    const sameRow = dragSource.row === row;
    if (!sameRow && panelLayout[row].length >= MAX_COLS) {
      dropTarget = null;
      return;
    }
    dropTarget = { row, col, side };
    dropRowGap = null;
  }

  function onPanelDragLeave(_row: number, _col: number) {
    // No-op: the next dragover on a new target will replace dropTarget.
  }

  function onPanelDrop(row: number, col: number, e: DragEvent) {
    e.preventDefault();
    if (!dragSource || !dropTarget) { clearDrag(); return; }
    const src = dragSource;
    const side = dropTarget.side;
    applyMove(src.row, src.col, row, col, side);
    clearDrag();
  }

  function onRowGapDragOver(rowIndex: number, e: DragEvent) {
    if (!dragSource) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    dropRowGap = rowIndex;
    dropTarget = null;
  }

  function onRowGapDrop(rowIndex: number, e: DragEvent) {
    e.preventDefault();
    if (!dragSource) { clearDrag(); return; }
    applyMoveToNewRow(dragSource.row, dragSource.col, rowIndex);
    clearDrag();
  }

  function applyMove(srcRow: number, srcCol: number, tgtRow: number, tgtCol: number, side: 'left' | 'right') {
    const next: PanelLayout = panelLayout.map((row) => [...row]);
    const key = next[srcRow][srcCol];
    next[srcRow].splice(srcCol, 1);
    let r = tgtRow;
    let c = tgtCol;
    if (srcRow === r && srcCol < c) c -= 1; // adjust after splice
    if (next[srcRow].length === 0) {
      next.splice(srcRow, 1);
      if (r > srcRow) r -= 1;
    }
    const insertAt = side === 'left' ? c : c + 1;
    if (next[r].length >= MAX_COLS) return; // safety
    next[r].splice(insertAt, 0, key);
    panelLayout = next;
    persistLayout();
  }

  function applyMoveToNewRow(srcRow: number, srcCol: number, gapIndex: number) {
    const next: PanelLayout = panelLayout.map((row) => [...row]);
    const key = next[srcRow][srcCol];
    next[srcRow].splice(srcCol, 1);
    let g = gapIndex;
    if (next[srcRow].length === 0) {
      next.splice(srcRow, 1);
      if (g > srcRow) g -= 1;
    }
    next.splice(g, 0, [key]);
    panelLayout = next;
    persistLayout();
  }

  function onPanelDragEnd() {
    clearDrag();
  }

  onMount(() => {
    maybeStartPolling();
    loadLayout();
    document.addEventListener('click', closeMenuOnClickOutside);
  });
  onDestroy(() => {
    if (pollHandle) clearInterval(pollHandle);
    if (typeof document !== 'undefined') document.removeEventListener('click', closeMenuOnClickOutside);
  });
</script>

<style>
  .layout-toolbar {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
    font-size: 0.8125rem;
    color: var(--text-dim);
  }
  @media (max-width: 768px) {
    .layout-toolbar { display: none; }
  }
  .layout-toolbar .spacer { flex: 1; }
  .layout-toolbar button {
    padding: 0.375rem 0.625rem;
    font-size: 0.8125rem;
  }

  .panel-stack {
    display: flex;
    flex-direction: column;
  }
  .panel-row {
    display: grid;
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }
  /* On mobile, force single column regardless of saved layout */
  @media (max-width: 768px) {
    .panel-row {
      grid-template-columns: 1fr !important;
    }
  }
  .panel-shell {
    position: relative;
    transition: outline-color 80ms ease, background 80ms ease;
    min-width: 0;
  }
  .panel-shell.editing {
    outline: 1px dashed var(--border-strong);
    outline-offset: 4px;
    border-radius: var(--radius-lg);
  }
  .panel-shell.dragging { opacity: 0.45; }
  .panel-shell.drop-left::before,
  .panel-shell.drop-right::after {
    content: '';
    position: absolute;
    top: -4px;
    bottom: -4px;
    width: 3px;
    background: var(--accent);
    border-radius: 2px;
    box-shadow: 0 0 8px var(--accent);
  }
  .panel-shell.drop-left::before { left: -8px; }
  .panel-shell.drop-right::after { right: -8px; }
  .panel-shell > :global(section) {
    margin-bottom: 0;
  }

  .row-gap {
    height: 12px;
    margin: -2px 0;
    border-radius: var(--radius-sm);
    transition: background 80ms ease, height 80ms ease;
  }
  .panel-stack.editing .row-gap {
    border: 1px dashed transparent;
  }
  .row-gap.active {
    background: rgba(96, 165, 250, 0.18);
    border-color: var(--accent) !important;
    height: 24px;
  }

  .panel-empty-hint {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-style: italic;
    padding: var(--space-2) 0;
  }

  .readme-body {
    white-space: pre-wrap;
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
    line-height: var(--lh-normal);
    color: var(--text);
    max-height: 320px;
    overflow-y: auto;
    padding: var(--space-2) 0;
  }
  .readme-truncated {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-style: italic;
    margin: var(--space-2) 0;
  }
  .readme-badges {
    display: flex;
    gap: var(--space-2);
    flex-wrap: wrap;
    margin-top: var(--space-3);
    padding-top: var(--space-3);
    border-top: 1px solid var(--border);
  }

  .contrib-row {
    display: grid;
    grid-template-columns: 32px 1fr 100px auto;
    align-items: center;
    gap: var(--space-3);
    padding: 0.375rem 0;
    font-size: var(--fs-sm);
  }
  .contrib-rank {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }
  .contrib-name { color: var(--text); }
  .contrib-bar {
    height: 6px;
    background: var(--border);
    border-radius: 3px;
    overflow: hidden;
  }
  .contrib-bar-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), var(--link));
    border-radius: 3px;
  }
  .contrib-count {
    color: var(--text-muted);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
    text-align: right;
  }
  .recent-files-head {
    margin-top: var(--space-3);
    color: var(--text-dim);
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .recent-files {
    list-style: none;
    margin: var(--space-2) 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }
  .recent-file-path {
    font-family: var(--font-mono);
    color: var(--text-muted);
    font-size: var(--fs-xs);
    word-break: break-all;
  }

  .drag-handle {
    position: absolute;
    top: 6px;
    right: 6px;
    z-index: 5;
    width: 28px;
    height: 28px;
    display: flex;
    align-items: center;
    justify-content: center;
    background: var(--bg-page);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-sm);
    color: var(--text-muted);
    cursor: grab;
    user-select: none;
    font-size: 0.875rem;
    line-height: 1;
  }
  .drag-handle:active { cursor: grabbing; }
  .drag-handle:hover { background: var(--bg-hover); color: var(--text); }

  .row {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .row .label { color: var(--text-dim); }
  .row .value {
    color: var(--text);
    font-family: var(--font-mono);
    word-break: break-all;
  }
  .commit-msg { font-family: inherit; }
  .webhook-value { display: flex; align-items: center; gap: var(--space-2); }
  .webhook-code {
    background: var(--bg-page);
    padding: 0.25rem 0.5rem;
    border-radius: var(--radius-sm);
    flex: 1;
    overflow-x: auto;
  }

  @media (max-width: 640px) {
    .row { grid-template-columns: 1fr; gap: 0.25rem; }
  }

  .task-group { margin-bottom: var(--space-4); }
  .task-group:last-child { margin-bottom: 0; }

  .pipeline-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .pipeline-row:last-child { border-bottom: none; }
  .pipeline-info { display: flex; gap: var(--space-2); align-items: center; flex-wrap: wrap; }
  .pipeline-name { font-weight: 500; }
  .pipeline-sha { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-dim); }
  .pipeline-time { font-size: 0.75rem; color: var(--text-dim); }
  .pipeline-view {
    padding: 0.375rem 0.625rem;
    font-size: 0.75rem;
  }

  .prefix-toggle {
    display: flex;
    align-items: center;
    gap: 0.375rem;
    width: 100%;
    padding: 0.375rem 0.75rem;
    background: transparent;
    border: none;
    cursor: pointer;
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    color: var(--text-dim);
    border-bottom: 1px solid var(--border);
  }
  .prefix-toggle:hover { background: var(--bg-hover); }
  .prefix-chevron { font-size: 0.625rem; }
  .prefix-label { color: var(--text); }
  .prefix-count { color: var(--text-dim); font-size: 0.75rem; }

  .task-source-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .task-source-label {
    font-family: var(--font-mono);
    color: var(--text-muted);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .task-source-count {
    color: var(--text-dim);
    font-size: 0.75rem;
  }

  .task-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid var(--border);
    position: relative;
  }
  .task-row:last-child { border-bottom: none; }
  .task-row:hover { background: var(--bg-hover); }

  .task-name-cell { min-width: 0; }
  .task-name {
    font-family: var(--font-mono);
    font-size: 0.875rem;
    display: flex;
    align-items: center;
    gap: 0.375rem;
    flex-wrap: wrap;
  }
  .task-cmd {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
    margin-top: 0.125rem;
  }

  .pin-btn {
    background: transparent;
    border: none;
    cursor: pointer;
    padding: 0.25rem;
    color: var(--text-dim);
    font-size: 1rem;
    line-height: 1;
  }
  .pin-btn:hover { color: var(--text); }
  .pin-btn.pinned { color: var(--warning); }

  .task-meta {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 0.6875rem;
  }
  .task-meta .gate-on { color: var(--warning); }
  .task-meta .meta-sep { opacity: 0.4; }

  .task-actions {
    display: inline-flex;
    gap: 0.25rem;
    align-items: center;
  }
  .menu-btn {
    background: transparent;
    border: 1px solid var(--border-strong);
    color: var(--text-muted);
    padding: 0.375rem 0.5rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: 1rem;
    line-height: 1;
  }
  .menu-btn:hover { background: var(--bg-hover); color: var(--text); }
  .run-btn { padding: 0.375rem 0.875rem; }

  .task-menu {
    position: absolute;
    top: calc(100% - 4px);
    right: 0.75rem;
    background: var(--bg-panel);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    padding: 0.25rem;
    min-width: 220px;
    box-shadow: var(--shadow-card-hover, 0 8px 24px rgba(0,0,0,0.25));
    z-index: 30;
    display: flex;
    flex-direction: column;
    gap: 0.125rem;
  }
  .menu-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem 0.625rem;
    background: transparent;
    border: none;
    color: var(--text);
    text-align: left;
    cursor: pointer;
    border-radius: var(--radius-sm);
    font-size: 0.8125rem;
    width: 100%;
    font-family: inherit;
  }
  .menu-item:hover { background: var(--bg-hover); }
  .menu-item-value { color: var(--text-dim); font-family: var(--font-mono); font-size: 0.75rem; }

  .approval-tag {
    font-size: 0.625rem;
    padding: 0.0625rem 0.375rem;
    border-radius: var(--radius-sm);
    background: rgba(234, 179, 8, 0.15);
    color: var(--warning);
    vertical-align: middle;
  }
  .task-spark {
    display: inline-flex;
    align-items: center;
    margin-left: 0.5rem;
    opacity: 0.7;
  }

  @media (max-width: 768px) {
    .task-row {
      grid-template-columns: auto 1fr auto;
    }
    .task-row > .task-meta { display: none; }
  }

  .filter-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  .filter-input {
    flex: 1;
    max-width: 320px;
    padding: 0.375rem 0.625rem;
    font-size: 0.8125rem;
  }
  .filter-count {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .lang-row {
    display: grid;
    grid-template-columns: 110px 1fr 80px 50px;
    align-items: center;
    gap: var(--space-3);
    padding: 0.375rem 0;
    font-size: 0.8125rem;
  }
  .lang-name {
    color: var(--text);
    font-family: var(--font-mono);
  }
  .lang-bar {
    height: 6px;
    background: var(--border);
    border-radius: 3px;
    overflow: hidden;
  }
  .lang-bar-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--accent), var(--link));
    border-radius: 3px;
  }
  .lang-num {
    color: var(--text-muted);
    font-size: 0.75rem;
    font-family: var(--font-mono);
    text-align: right;
  }

  .stats-summary {
    display: flex;
    gap: var(--space-6);
    padding-bottom: var(--space-3);
    margin-bottom: var(--space-2);
    border-bottom: 1px solid var(--border);
  }
  .stat-block { flex: 1; }
  .stat-num {
    font-size: 1.5rem;
    font-weight: 600;
    color: var(--text);
    font-family: var(--font-mono);
  }
  .stat-label {
    color: var(--text-dim);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-top: 0.125rem;
  }

  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0.75rem;
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    color: inherit;
    font-size: var(--fs-md);
  }
  .run-row:hover {
    background: var(--bg-hover);
    text-decoration: none;
  }
  .run-row:last-child { border-bottom: none; }
  .run-name { font-family: var(--font-mono); }
  .dim { color: var(--text-dim); }
  .run-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
  }

  @media (max-width: 768px) {
    .run-row { grid-template-columns: auto 1fr auto; }
    .run-row > :nth-child(3) { display: none; }
  }

  .ci-config-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .ci-config-row:last-child { border-bottom: none; }
  .ci-path { font-family: var(--font-mono); color: var(--text); }

  .compose-env-form { margin-top: var(--space-2); }
  .compose-env-form h3 {
    font-size: 0.8125rem;
    color: var(--text-muted);
    margin: var(--space-3) 0 var(--space-2) 0;
  }
  .compose-env-form .field { margin-bottom: var(--space-2); }
  .compose-env-form label {
    font-family: var(--font-mono);
    font-size: 0.8125rem;
  }
  .compose-status-value {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .compose-actions { margin-top: var(--space-3); }
  .compose-error { color: var(--danger-text); font-size: 0.875rem; }
  .compose-info { color: var(--text-dim); font-size: 0.875rem; }

  .section-link {
    font-size: 0.75rem;
    text-transform: none;
    letter-spacing: 0;
    color: var(--link);
    font-weight: 400;
  }

  .modal-error {
    color: var(--danger-text);
    font-size: 0.875rem;
    margin: 0 0 var(--space-3) 0;
  }
  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }
  .modal-hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin: 0.25rem 0 0 0;
  }
  .modal-fields-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }

  /* Favorite button */
  .favorite-row {
    display: flex;
    align-items: center;
    margin-bottom: var(--space-3);
  }
  .favorite-btn {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    background: transparent;
    border: 1px solid var(--border-strong);
    color: var(--text-muted);
    padding: 0.375rem 0.75rem;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: 0.875rem;
    line-height: 1;
    transition: color 80ms ease, border-color 80ms ease;
  }
  .favorite-btn:hover {
    color: var(--warning);
    border-color: var(--warning);
  }
  .favorite-btn.favorited {
    color: var(--warning);
    border-color: rgba(234, 179, 8, 0.4);
    background: rgba(234, 179, 8, 0.08);
  }
  .favorite-label { font-size: 0.8125rem; }

  /* Blame timeline */
  .blame-table-wrap {
    overflow-x: auto;
    margin: 0 calc(-1 * var(--space-6));
    padding: 0 var(--space-6);
  }
  .blame-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
  }
  .blame-table th {
    text-align: left;
    color: var(--text-dim);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    padding: 0.375rem 0.5rem;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .blame-table th.num,
  .blame-table td.num {
    text-align: right;
  }
  .blame-table td {
    padding: 0.5rem 0.5rem;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .blame-row {
    cursor: pointer;
    transition: background 80ms ease;
  }
  .blame-row:hover {
    background: var(--bg-hover);
  }
  .blame-row.blame-pass {
    border-left: 3px solid var(--success);
  }
  .blame-row.blame-fail {
    border-left: 3px solid var(--danger-text);
  }
  .blame-author {
    max-width: 120px;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .blame-msg {
    max-width: 240px;
    overflow: hidden;
    text-overflow: ellipsis;
    color: var(--text);
  }
  .blame-date {
    color: var(--text-dim);
    font-size: 0.75rem;
  }
  .ins { color: var(--success); }
  .del { color: var(--danger-text); }
  .blame-runs {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
  }
  .pass-count { color: var(--success); font-size: 0.6875rem; }
  .fail-count { color: var(--danger-text); font-size: 0.6875rem; }

  /* Monorepo packages */
  .pkg-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    cursor: pointer;
  }
  .pkg-row:hover { background: var(--bg-hover); }
  .pkg-row:last-child { border-bottom: none; }
  .pkg-info { display: flex; flex-direction: column; gap: 0.125rem; min-width: 0; }
  .pkg-name { font-family: var(--font-mono); font-size: 0.875rem; color: var(--text); }
  .pkg-path { font-family: var(--font-mono); font-size: 0.75rem; color: var(--text-dim); overflow: hidden; text-overflow: ellipsis; }
  .pkg-tasks { font-size: 0.75rem; color: var(--text-muted); font-family: var(--font-mono); white-space: nowrap; }

  .scope-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .scope-row:last-child { border-bottom: none; }
  .scope-name { font-family: var(--font-mono); }
  .scope-add-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    margin-top: var(--space-2);
  }

  /* Deploy previews */
  .preview-create-form {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--border);
  }
  .preview-branch-input {
    flex: 1;
    max-width: 240px;
    font-size: 0.8125rem;
  }
  .preview-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    flex-wrap: wrap;
  }
  .preview-row:last-child { border-bottom: none; }
  .preview-info {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex: 1;
    min-width: 0;
  }
  .preview-url {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--link);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .preview-meta {
    display: flex;
    align-items: center;
    gap: var(--space-3);
  }
  .auto-deploy-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    cursor: pointer;
    font-size: 0.75rem;
    color: var(--text-muted);
  }
  .auto-deploy-label { user-select: none; }
  .preview-actions {
    display: flex;
    gap: 0.25rem;
    align-items: center;
    flex-shrink: 0;
  }

  /* Alert rules */
  .alert-create-form {
    margin-bottom: var(--space-3);
    padding-bottom: var(--space-3);
    border-bottom: 1px solid var(--border);
  }
  .alert-form-row {
    display: flex;
    gap: var(--space-2);
    align-items: flex-end;
    flex-wrap: wrap;
  }
  .alert-form-row .field { flex: 1; min-width: 120px; }
  .alert-form-row .field label { font-size: 0.75rem; }
  .alert-form-row select,
  .alert-form-row input { font-size: 0.8125rem; }
  .alert-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    gap: var(--space-2);
  }
  .alert-row:last-child { border-bottom: none; }
  .alert-info { display: flex; flex-direction: column; gap: 0.125rem; }
  .alert-task-name { font-family: var(--font-mono); font-size: 0.875rem; color: var(--text); }
  .alert-desc { font-size: 0.75rem; color: var(--text-dim); }
  .alert-actions { display: flex; align-items: center; gap: var(--space-2); }

  /* Impact radar */
  .impact-row {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
  }
  .impact-row:last-child { border-bottom: none; }
  .impact-name {
    font-family: var(--font-mono);
    font-size: 0.875rem;
    color: var(--text);
    flex-shrink: 0;
  }
  .impact-patterns {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
    flex: 1;
    justify-content: flex-end;
  }

  /* Dependency tree */
  .dep-section {
    margin-bottom: var(--space-3);
  }
  .dep-section:last-child { margin-bottom: 0; }
  .dep-heading {
    display: block;
    font-size: 0.75rem;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-bottom: var(--space-1);
  }
  .dep-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--border);
    color: inherit;
    text-decoration: none;
    font-size: 0.875rem;
  }
  .dep-row:hover { background: var(--bg-hover); text-decoration: none; }
  .dep-row:last-child { border-bottom: none; }
  .dep-name { font-family: var(--font-mono); color: var(--link); }

  /* Live run */
  .live-run-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
  }
  .live-run-name {
    font-family: var(--font-mono);
    font-size: 0.875rem;
    color: var(--link);
  }
  .live-log-tail {
    margin: 0;
    padding: var(--space-3);
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    font-family: var(--font-mono);
    font-size: 0.75rem;
    line-height: 1.5;
    color: var(--text-muted);
    max-height: 160px;
    overflow-y: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  /* Task metrics modal */
  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: var(--space-3);
  }
  .metric-card {
    text-align: center;
    padding: var(--space-3);
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
  }
  .metric-num {
    font-size: 1.25rem;
    font-weight: 600;
    color: var(--text);
    font-family: var(--font-mono);
  }
  .metric-label {
    font-size: 0.6875rem;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-top: 0.125rem;
  }
  .success-rate-bar {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .success-rate-label {
    display: flex;
    justify-content: space-between;
    font-size: 0.8125rem;
    color: var(--text-muted);
  }
  .success-rate-pct {
    font-family: var(--font-mono);
    font-weight: 600;
    color: var(--text);
  }
  .rate-track {
    height: 8px;
    background: var(--border);
    border-radius: 4px;
    overflow: hidden;
  }
  .rate-fill {
    height: 100%;
    background: linear-gradient(90deg, var(--success), #34d399);
    border-radius: 4px;
    transition: width 200ms ease;
  }
  .trend-chart {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 64px;
    padding: var(--space-1) 0;
  }
  .trend-bar-group {
    flex: 1;
    min-width: 0;
    height: 100%;
    display: flex;
    align-items: flex-end;
  }
  .trend-bar {
    width: 100%;
    display: flex;
    flex-direction: column-reverse;
    border-radius: 2px;
    overflow: hidden;
  }
  .trend-pass {
    background: var(--success);
    min-height: 1px;
  }
  .trend-fail {
    background: var(--danger-text);
    min-height: 1px;
  }

  @media (max-width: 768px) {
    .metrics-grid { grid-template-columns: repeat(2, 1fr); }
    .blame-table-wrap { margin: 0 calc(-1 * var(--space-4)); padding: 0 var(--space-4); }
    .alert-form-row { flex-direction: column; }
    .alert-form-row .field { min-width: 100%; }
    .preview-row { flex-direction: column; align-items: flex-start; }
    .preview-actions { width: 100%; }
  }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<div class="favorite-row">
  <form method="POST" action="?/toggleFavorite" class="inline-form">
    <input type="hidden" name="is_favorited" value={String(data.isFavorited)} />
    <button type="submit"
            class="favorite-btn {data.isFavorited ? 'favorited' : ''}"
            title={data.isFavorited ? 'Remove from favorites' : 'Add to favorites'}>
      {data.isFavorited ? '★' : '☆'}
      <span class="favorite-label">{data.isFavorited ? 'Favorited' : 'Favorite'}</span>
    </button>
  </form>
</div>

<div class="layout-toolbar">
  <span class="spacer"></span>
  {#if editLayoutMode}
    <button type="button" class="ghost" onclick={resetLayout} title="Reset to default order">Reset</button>
    <button type="button" onclick={() => editLayoutMode = false}>Done</button>
  {:else}
    <button type="button" class="ghost" onclick={() => editLayoutMode = true}>Customize layout</button>
  {/if}
</div>

{#snippet taskRow(t: typeof data.tasks[0], isPinned: boolean)}
  <div class="task-row">
    <form method="POST" action="?/togglePin" class="inline-form">
      <input type="hidden" name="task_id" value={t.id} />
      <input type="hidden" name="pinned" value={String(isPinned)} />
      <button type="submit"
              class="pin-btn {isPinned ? 'pinned' : ''}"
              title={isPinned ? 'Unpin from dashboard' : 'Pin to dashboard'}>
        {isPinned ? '★' : '☆'}
      </button>
    </form>
    <div class="task-name-cell">
      <div class="task-name">
        {t.name}
        {#if t.requires_approval}<span class="approval-tag">approval</span>{/if}
        <span class="task-spark"><TaskSparkline runs={runsByTask[t.id] ?? []} /></span>
      </div>
      <div class="task-cmd" title={t.raw_command}>{t.raw_command}</div>
    </div>
    <span class="task-meta">
      <span title="Timeout">⏱ {formatTimeout(t.timeout_seconds)}</span>
      <span class="meta-sep">·</span>
      <span title="Retry policy">↻ {formatRetry(t.retry_max, t.retry_backoff_sec)}</span>
      {#if t.requires_approval}
        <span class="meta-sep">·</span>
        <span class="gate-on" title="Approval required">gate</span>
      {/if}
    </span>
    <span class="task-actions">
      <button type="button" class="menu-btn task-menu-trigger"
              title="More actions"
              aria-haspopup="menu"
              aria-expanded={openMenuTaskID === t.id}
              onclick={(e) => { e.stopPropagation(); toggleTaskMenu(t.id); }}>⋯</button>
      <form method="POST" action="?/run" class="inline-form">
        <input type="hidden" name="task_id" value={t.id} />
        <button type="submit" class="run-btn" disabled={data.project.status !== 'ready'}>Run</button>
      </form>
    </span>
    {#if openMenuTaskID === t.id}
      <div class="task-menu" role="menu">
        <button class="menu-item" type="button" role="menuitem"
                disabled={data.project.status !== 'ready'}
                onclick={() => openParams(t.id, t.name)}>
          <span>Run with parameters…</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                disabled={data.project.status !== 'ready'}
                onclick={() => openBranchPicker(t.id, t.name)}>
          <span>Run on branch…</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => openTimeoutEditor(t.id, t.name, t.timeout_seconds)}>
          <span>Edit timeout</span>
          <span class="menu-item-value">{formatTimeout(t.timeout_seconds)}</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => openRetryEditor(t.id, t.name, t.retry_max, t.retry_backoff_sec)}>
          <span>Edit retry policy</span>
          <span class="menu-item-value">{formatRetry(t.retry_max, t.retry_backoff_sec)}</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => openConcurrencyEditor(t.id, t.name, t.max_concurrency, t.supersede_policy)}>
          <span>Concurrency policy</span>
          <span class="menu-item-value">{formatConcurrency(t.max_concurrency, t.supersede_policy)}</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => openServicesEditor(t.id, t.name, t.needs_services ?? [])}>
          <span>Required services</span>
          <span class="menu-item-value">{formatServices(t.needs_services ?? [])}</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => openArtifactsEditor(t.id, t.name, t.artifact_patterns ?? [])}>
          <span>Artifact patterns</span>
          <span class="menu-item-value">{formatArtifacts(t.artifact_patterns ?? [])}</span>
        </button>
        <button class="menu-item" type="button" role="menuitem"
                onclick={() => { openMenuTaskID = null; openTaskMetrics(t.id, t.name); }}>
          <span>View metrics</span>
        </button>
        <form method="POST" action="?/toggleApproval" class="inline-form" style="display:contents;"
              onsubmit={() => { openMenuTaskID = null; }}>
          <input type="hidden" name="task_id" value={t.id} />
          <input type="hidden" name="next" value={String(!t.requires_approval)} />
          <button class="menu-item" type="submit" role="menuitem">
            <span>{t.requires_approval ? 'Disable approval gate' : 'Require approval'}</span>
            <span class="menu-item-value">{t.requires_approval ? 'on' : 'off'}</span>
          </button>
        </form>
      </div>
    {/if}
  </div>
{/snippet}

{#snippet panelTasks()}
  <Panel title="Tasks">
    {#snippet actions()}
      {#if data.tasks.length > 0}
        <span class="filter-count">{data.tasks.length} total</span>
      {/if}
    {/snippet}
    {#if data.tasks.length === 0}
      <div class="empty">
        {#if data.project.status === 'ready'}
          No runnable tasks detected. Add a <code>package.json</code> with a <code>scripts</code> section, or a <code>justfile</code>.
        {:else}
          Tasks will appear after the repo finishes cloning.
        {/if}
      </div>
    {:else}
      {#if data.tasks.length > 8}
        <div class="filter-row">
          <input type="search" class="filter-input" placeholder="Filter tasks…" bind:value={taskFilter} />
        </div>
      {/if}
      {#each Object.entries(tasksBySource) as [source, group] (source)}
        <div class="task-group">
          <div class="task-source-header">
            <span class="task-source-label">{source}</span>
            <span class="task-source-count">
              {group.ungrouped.length + Object.values(group.prefixed).reduce((s, t) => s + t.length, 0)}
            </span>
          </div>
          {#each group.ungrouped as t (t.id)}
            {@const isPinned = data.pinnedTaskIDs.includes(t.id)}
            {@render taskRow(t, isPinned)}
          {/each}
          {#each Object.entries(group.prefixed) as [prefix, prefixTasks] (prefix)}
            {@const prefixKey = source + '/' + prefix}
            {@const isOpen = !collapsedPrefixes.has(prefixKey)}
            <button type="button" class="prefix-toggle" onclick={() => { const s = new Set(collapsedPrefixes); if (s.has(prefixKey)) s.delete(prefixKey); else s.add(prefixKey); collapsedPrefixes = s; }}>
              <span class="prefix-chevron">{isOpen ? '▾' : '▸'}</span>
              <span class="prefix-label">{prefix}</span>
              <span class="prefix-count">({prefixTasks.length})</span>
            </button>
            {#if isOpen}
              {#each prefixTasks as t (t.id)}
                {@const isPinned = data.pinnedTaskIDs.includes(t.id)}
                {@render taskRow(t, isPinned)}
              {/each}
            {/if}
          {/each}
        </div>
      {/each}
    {/if}
  </Panel>
{/snippet}

{#snippet panelRecentRuns()}
  {#if data.runs.length > 0}
    <Panel title="Recent runs">
      {#snippet actions()}
        <a href={`/projects/${data.project.id}/runs`} class="section-link">View all →</a>
      {/snippet}
      {#each data.runs.slice(0, 10) as r (r.id)}
        <a href={`/runs/${r.id}`} class="run-row">
          <StatusPill status={r.status} size="sm" />
          <span class="run-name">{r.task_name} <span class="dim">({r.task_source})</span></span>
          <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
          <span class="run-meta"><TimeAgo value={r.created_at} /></span>
        </a>
      {/each}
    </Panel>
  {/if}
{/snippet}

{#snippet panelRepository()}
  <Panel title="Repository">
    <div class="row"><span class="label">Git URL</span><span class="value">{data.project.git_url}</span></div>
    <div class="row"><span class="label">Branch</span><span class="value">{data.project.default_branch || '—'}</span></div>
    <div class="row"><span class="label">Status</span><span class="value">{data.project.status}</span></div>
    <div class="row"><span class="label">Local path</span><span class="value">{data.project.local_path || '—'}</span></div>
    {#if data.project.webhook_token}
      <div class="row">
        <span class="label">Webhook URL</span>
        <span class="value webhook-value">
          <code class="webhook-code">{`${typeof window !== 'undefined' ? window.location.origin : ''}/api/webhooks/projects/${data.project.webhook_token}`}</code>
          <button type="button" class="ghost" onclick={() => navigator.clipboard?.writeText(`${window.location.origin}/api/webhooks/projects/${data.project.webhook_token}`)}>Copy</button>
        </span>
      </div>
    {/if}
  </Panel>
{/snippet}

{#snippet panelLatestCommit()}
  <Panel title="Latest commit">
    <div class="row"><span class="label">SHA</span><span class="value">{shortSha(data.project.last_commit_sha)}</span></div>
    <div class="row"><span class="label">Author</span><span class="value commit-msg">{data.project.last_commit_author || '—'}</span></div>
    <div class="row"><span class="label">Message</span><span class="value commit-msg">{data.project.last_commit_message || '—'}</span></div>
    <div class="row"><span class="label">Synced</span><span class="value">{data.project.last_synced_at ? new Date(data.project.last_synced_at).toLocaleString() : '—'}</span></div>
  </Panel>
{/snippet}

{#snippet panelCodeStats()}
  {#if data.stats}
    <Panel title="Code statistics">
      <div class="stats-summary">
        <div class="stat-block">
          <div class="stat-num">{data.stats.total_files.toLocaleString()}</div>
          <div class="stat-label">Files</div>
        </div>
        <div class="stat-block">
          <div class="stat-num">{data.stats.total_lines.toLocaleString()}</div>
          <div class="stat-label">Lines</div>
        </div>
        <div class="stat-block">
          <div class="stat-num">{data.stats.total_code.toLocaleString()}</div>
          <div class="stat-label">Code</div>
        </div>
      </div>
      {#each topLangs as [name, l] (name)}
        <div class="lang-row">
          <span class="lang-name">{name}</span>
          <div class="lang-bar"><div class="lang-bar-fill" style="width: {(l.lines / totalLangLines) * 100}%"></div></div>
          <span class="lang-num">{l.lines.toLocaleString()} lines</span>
          <span class="lang-num">{l.files} files</span>
        </div>
      {/each}
    </Panel>
  {/if}
{/snippet}

{#snippet panelPipelines()}
  {#if data.remotePipelines.length > 0}
    <Panel title="CI/CD pipelines">
      {#each data.remotePipelines as p (p.id)}
        {@const mappedStatus = p.status === 'success' ? 'succeeded' : p.status === 'failure' ? 'failed' : p.status === 'running' ? 'running' : 'pending'}
        <div class="pipeline-row">
          <StatusPill status={mappedStatus} size="sm" />
          <div class="pipeline-info">
            <span class="pipeline-name">{p.workflow_name || 'pipeline'}</span>
            {#if p.branch}<Badge variant="info" size="sm">{p.branch}</Badge>{/if}
            {#if p.commit_sha}<Tooltip text={p.commit_sha}><span class="pipeline-sha">{p.commit_sha.slice(0, 7)}</span></Tooltip>{/if}
          </div>
          <span class="pipeline-time">{p.started_at ? formatDuration(p.started_at, p.finished_at) : ''}</span>
          {#if p.html_url}
            <a href={p.html_url} target="_blank" rel="noopener noreferrer" class="ghost pipeline-view">View</a>
          {/if}
        </div>
      {/each}
    </Panel>
  {/if}
{/snippet}

{#snippet panelPipelineConfigs()}
  {#if data.pipelineConfigs.length > 0}
    <Panel title="CI/CD configuration">
      {#each data.pipelineConfigs as cfg (cfg.id)}
        <div class="ci-config-row">
          <Badge variant="info" size="sm">{cfg.ci_system}</Badge>
          <span class="ci-path">{cfg.file_path}</span>
        </div>
      {/each}
    </Panel>
  {/if}
{/snippet}

{#snippet panelCompose()}
  {#if data.compose?.compose_file}
    <Panel title="Docker Compose">
      <div class="row"><span class="label">File</span><span class="value">{data.compose.compose_file}</span></div>
      <div class="row">
        <span class="label">Status</span>
        <span class="value compose-status-value">
          <StatusPill status={data.compose.status === 'running' ? 'ready' : data.compose.status === 'starting' ? 'cloning' : data.compose.status === 'error' ? 'failed' : 'pending'} size="sm" label={data.compose.status} />
        </span>
      </div>
      {#if data.compose.services.length > 0}
        <div class="row"><span class="label">Services</span><span class="value">{data.compose.services.join(', ')}</span></div>
      {/if}
      {#if data.compose.status === 'stopped' && data.compose.env_vars.length > 0}
        <form class="compose-env-form" onsubmit={(e) => { e.preventDefault(); startCompose(); }}>
          <h3>Environment variables</h3>
          {#each data.compose.env_vars as envVar, i}
            <div class="field">
              <label for={`env-${i}`}>{envVar.key}</label>
              <input id={`env-${i}`} type="text" value={envVar.default_value} data-env-key={envVar.key} class="compose-env-input" />
            </div>
          {/each}
          <button type="submit" disabled={data.project.status !== 'ready'}>Start</button>
        </form>
      {:else if data.compose.status === 'stopped'}
        <div class="compose-actions">
          <button type="button" disabled={data.project.status !== 'ready'} onclick={() => startCompose()}>Start</button>
        </div>
      {:else if data.compose.status === 'running'}
        <div class="compose-actions">
          <button type="button" onclick={() => stopCompose()}>Stop</button>
        </div>
      {:else if data.compose.status === 'starting'}
        <div class="compose-actions compose-info">Starting services...</div>
      {:else if data.compose.status === 'error'}
        <div class="compose-actions compose-error">Failed to start. Check logs and retry.</div>
        <button type="button" class="compose-actions" onclick={() => startCompose()}>Retry</button>
      {/if}
    </Panel>
  {/if}
{/snippet}

{#snippet panelAbout()}
  {#if data.overview?.readme}
    {@const readme = data.overview.readme}
    <Panel title="About">
      {#snippet actions()}
        <span style="color: var(--text-dim); font-family: var(--font-mono); font-size: var(--fs-xs);">{readme.path}</span>
      {/snippet}
      <div class="readme-body">{readme.content}</div>
      {#if readme.truncated}
        <div class="readme-truncated">Truncated — open the full file in your repo for the rest.</div>
      {/if}
      <div class="readme-badges">
        {#if data.overview.license}
          <Badge variant="info" size="sm">License: {data.overview.license.kind}</Badge>
        {/if}
        {#if data.overview.has_contributing}
          <Badge variant="muted" size="sm">CONTRIBUTING ✓</Badge>
        {/if}
        {#if data.overview.has_changelog}
          <Badge variant="muted" size="sm">CHANGELOG ✓</Badge>
        {/if}
      </div>
    </Panel>
  {/if}
{/snippet}

{#snippet panelContributors()}
  {#if (data.overview?.contributors.length ?? 0) > 0}
    <Panel title="Contributors">
      {#snippet actions()}
        <span style="color: var(--text-dim); font-family: var(--font-mono); font-size: var(--fs-xs);">top {data.overview?.contributors.length ?? 0}</span>
      {/snippet}
      {#each data.overview?.contributors ?? [] as c, i (c.name + i)}
        {@const max = data.overview?.contributors[0]?.commits ?? 1}
        <div class="contrib-row">
          <span class="contrib-rank">#{i + 1}</span>
          <span class="contrib-name">{c.name}</span>
          <div class="contrib-bar">
            <div class="contrib-bar-fill" style="width: {(c.commits / max) * 100}%"></div>
          </div>
          <span class="contrib-count">{c.commits.toLocaleString()} commit{c.commits === 1 ? '' : 's'}</span>
        </div>
      {/each}
      {#if data.overview?.recent_files && data.overview.recent_files.length > 0}
        <div class="recent-files-head">Recent files</div>
        <ul class="recent-files">
          {#each data.overview.recent_files.slice(0, 6) as f (f.path)}
            <li><span class="recent-file-path">{f.path}</span></li>
          {/each}
        </ul>
      {/if}
    </Panel>
  {/if}
{/snippet}

{#snippet panelBlameTimeline()}
  <Panel title="Blame timeline">
    {#snippet actions()}
      <button type="button" class="ghost" style="font-size: 0.75rem;" onclick={syncCommits} disabled={syncingCommits}>
        {syncingCommits ? 'Syncing...' : 'Sync commits'}
      </button>
    {/snippet}
    {#if data.blameTimeline.length === 0}
      <div class="empty">No commits synced yet. Click "Sync commits" to populate the timeline.</div>
    {:else}
      <div class="blame-table-wrap">
        <table class="blame-table">
          <thead>
            <tr>
              <th>SHA</th>
              <th>Author</th>
              <th>Message</th>
              <th>Date</th>
              <th class="num">Files</th>
              <th class="num">+/-</th>
              <th class="num">Runs</th>
            </tr>
          </thead>
          <tbody>
            {#each data.blameTimeline as c (c.sha)}
              <tr class="blame-row {blameRowColor(c)}" onclick={() => openBlameDetail(c.sha)} role="button" tabindex="0">
                <td class="mono">{c.sha.slice(0, 8)}</td>
                <td class="blame-author">{c.author}</td>
                <td class="blame-msg">{c.message.split('\n')[0].slice(0, 72)}</td>
                <td class="blame-date"><TimeAgo value={c.committed_at} /></td>
                <td class="num">{c.files_changed}</td>
                <td class="num"><span class="ins">+{c.insertions}</span> <span class="del">-{c.deletions}</span></td>
                <td class="num">
                  {#if c.run_count > 0}
                    <span class="blame-runs">
                      {c.run_count}
                      {#if c.pass_count > 0}<span class="pass-count">{c.pass_count}p</span>{/if}
                      {#if c.fail_count > 0}<span class="fail-count">{c.fail_count}f</span>{/if}
                    </span>
                  {:else}
                    <span class="dim">--</span>
                  {/if}
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </Panel>
{/snippet}

{#snippet panelMonorepo()}
  <Panel title="Monorepo packages">
    {#snippet actions()}
      <button type="button" class="ghost" style="font-size: 0.75rem;" onclick={detectPackages} disabled={detectingPackages}>
        {detectingPackages ? 'Detecting...' : 'Detect packages'}
      </button>
      {#if data.monorepoPackages.length > 0}
        <button type="button" class="ghost" style="font-size: 0.75rem;" onclick={autoMapTasks} disabled={autoMapping}>
          {autoMapping ? 'Mapping...' : 'Auto-map tasks'}
        </button>
      {/if}
    {/snippet}
    {#if data.monorepoPackages.length === 0}
      <div class="empty">No packages detected. Click "Detect packages" to scan for monorepo packages.</div>
    {:else}
      {#each data.monorepoPackages as pkg (pkg.id)}
        <div class="pkg-row" onclick={() => openPackageDetail(pkg.id)} onkeydown={(e) => { if (e.key === 'Enter') openPackageDetail(pkg.id); }} role="button" tabindex="0">
          <div class="pkg-info">
            <span class="pkg-name">{pkg.name}</span>
            <span class="pkg-path">{pkg.path}</span>
          </div>
          <Badge variant={pkgTypeBadgeVariant(pkg.pkg_type)} size="sm">{pkg.pkg_type}</Badge>
          <span class="pkg-tasks">{pkg.task_count} task{pkg.task_count === 1 ? '' : 's'}</span>
        </div>
      {/each}
    {/if}
  </Panel>
{/snippet}

{#snippet panelPreviews()}
  <Panel title="Deploy previews">
    {#snippet actions()}
      <span class="filter-count">{data.previews.length} preview{data.previews.length === 1 ? '' : 's'}</span>
    {/snippet}
    <form method="POST" action="?/createPreview" class="preview-create-form">
      <input type="text" name="branch" placeholder="Branch name" class="preview-branch-input" bind:value={newPreviewBranch} />
      <button type="submit" disabled={!newPreviewBranch.trim()}>New preview</button>
    </form>
    {#if data.previews.length === 0}
      <div class="empty" style="padding: var(--space-4) 0;">No deploy previews yet.</div>
    {:else}
      {#each data.previews as p (p.id)}
        <div class="preview-row">
          <div class="preview-info">
            <StatusPill status={p.status === 'deployed' ? 'ready' : p.status === 'deploying' ? 'cloning' : p.status === 'stopped' ? 'cancelled' : p.status === 'failed' ? 'failed' : 'pending'} size="sm" />
            <Badge variant="info" size="sm">{p.branch}</Badge>
            {#if p.url}
              <a href={p.url} target="_blank" rel="noopener noreferrer" class="preview-url">{p.url}</a>
            {/if}
          </div>
          <div class="preview-meta">
            {#if p.last_deployed_at}
              <span class="dim"><TimeAgo value={p.last_deployed_at} /></span>
            {/if}
            <label class="auto-deploy-toggle" title="Auto-deploy on push">
              <input type="checkbox"
                     checked={p.auto_deploy}
                     onchange={() => toggleAutoDeploy(p.id, p.auto_deploy)} />
              <span class="auto-deploy-label">Auto</span>
            </label>
          </div>
          <div class="preview-actions">
            <button type="button" class="ghost" style="font-size: 0.75rem;"
                    onclick={() => redeployPreview(p.id)}
                    disabled={previewActionPending === p.id}>Redeploy</button>
            {#if p.status === 'deployed' || p.status === 'deploying'}
              <button type="button" class="ghost" style="font-size: 0.75rem;"
                      onclick={() => stopPreview(p.id)}
                      disabled={previewActionPending === p.id}>Stop</button>
            {/if}
            {#if previewDeleteConfirm === p.id}
              <button type="button" class="danger" style="font-size: 0.75rem;"
                      onclick={() => deletePreview(p.id)}
                      disabled={previewActionPending === p.id}>Confirm delete</button>
              <button type="button" class="ghost" style="font-size: 0.75rem;"
                      onclick={() => (previewDeleteConfirm = null)}>Cancel</button>
            {:else}
              <button type="button" class="ghost danger" style="font-size: 0.75rem;"
                      onclick={() => (previewDeleteConfirm = p.id)}>Delete</button>
            {/if}
          </div>
        </div>
      {/each}
    {/if}
  </Panel>
{/snippet}

{#snippet panelAlertRules()}
  <Panel title="Alert rules">
    {#snippet actions()}
      <button type="button" class="ghost" style="font-size: 0.75rem;" onclick={() => (alertFormOpen = !alertFormOpen)}>
        {alertFormOpen ? 'Cancel' : 'New rule'}
      </button>
    {/snippet}
    {#if alertFormOpen}
      <form method="POST" action="?/createAlertRule" class="alert-create-form" onsubmit={() => { alertFormOpen = false; }}>
        <div class="alert-form-row">
          <div class="field">
            <label for="alert-task">Task</label>
            <select id="alert-task" name="task_id" bind:value={alertFormTaskID} required>
              <option value="">Select task...</option>
              {#each data.tasks as t (t.id)}
                <option value={t.id}>{t.name}</option>
              {/each}
            </select>
          </div>
          <div class="field">
            <label for="alert-type">Type</label>
            <select id="alert-type" name="type" bind:value={alertFormType}>
              <option value="duration">Duration</option>
              <option value="failure_rate">Failure rate</option>
              <option value="consecutive_failures">Consecutive failures</option>
            </select>
          </div>
          <div class="field">
            <label for="alert-comparison">Operator</label>
            <select id="alert-comparison" name="comparison" bind:value={alertFormComparison}>
              <option value="gt">greater than</option>
              <option value="gte">greater or equal</option>
              <option value="lt">less than</option>
              <option value="lte">less or equal</option>
            </select>
          </div>
          <div class="field">
            <label for="alert-threshold">Threshold</label>
            <input id="alert-threshold" type="number" name="threshold" bind:value={alertFormThreshold} required
                   placeholder={alertFormType === 'duration' ? 'ms' : alertFormType === 'failure_rate' ? '0-100%' : 'count'} />
          </div>
          <button type="submit" disabled={!alertFormTaskID}>Create</button>
        </div>
      </form>
    {/if}
    {#if data.alertRules.length === 0 && data.widgetAlertRules.length === 0}
      <div class="empty" style="padding: var(--space-4) 0;">No alert rules configured for this project's tasks.</div>
    {:else}
      {#each data.alertRules as rule (rule.id)}
        <div class="alert-row">
          <div class="alert-info">
            <span class="alert-task-name">{rule.task_name}</span>
            <span class="alert-desc">{rule.type} {rule.comparison} {rule.threshold}{rule.type === 'failure_rate' ? '%' : rule.type === 'duration' ? 'ms' : ''}</span>
          </div>
          <div class="alert-actions">
            {#if rule.last_triggered_at}
              <span class="dim" style="font-size: 0.75rem;"><TimeAgo value={rule.last_triggered_at} /></span>
            {/if}
            <form method="POST" action="?/toggleAlertRule" class="inline-form">
              <input type="hidden" name="rule_id" value={rule.id} />
              <button type="submit" class="ghost" style="font-size: 0.75rem;"
                      title={rule.enabled ? 'Disable rule' : 'Enable rule'}>
                {rule.enabled ? 'Enabled' : 'Disabled'}
              </button>
            </form>
            <form method="POST" action="?/deleteAlertRule" class="inline-form">
              <input type="hidden" name="rule_id" value={rule.id} />
              <button type="submit" class="ghost danger" style="font-size: 0.75rem;">Delete</button>
            </form>
          </div>
        </div>
      {/each}
    {/if}
  </Panel>
{/snippet}

{#snippet panelImpactRadar()}
  {#if data.impactRadar && data.impactRadar.tasks.length > 0}
    <Panel title="Impact radar">
      {#snippet actions()}
        {#if data.impactRadar && data.impactRadar.unmapped_tasks > 0}
          <span class="dim" style="font-size: 0.75rem;">{data.impactRadar.unmapped_tasks} unmapped</span>
        {/if}
      {/snippet}
      {#each data.impactRadar.tasks as t (t.task_id)}
        <div class="impact-row">
          <StatusPill status={t.last_status === 'succeeded' ? 'succeeded' : t.last_status === 'failed' ? 'failed' : 'pending'} size="sm" />
          <span class="impact-name">{t.task_name}</span>
          <div class="impact-patterns">
            {#each t.patterns.slice(0, 3) as p}
              <Badge variant="muted" size="sm">{p}</Badge>
            {/each}
            {#if t.patterns.length > 3}
              <span class="dim" style="font-size: 0.6875rem;">+{t.patterns.length - 3}</span>
            {/if}
          </div>
        </div>
      {/each}
    </Panel>
  {/if}
{/snippet}

{#snippet panelDependencyTree()}
  {#if data.dependencyTree && (data.dependencyTree.upstream.length > 0 || data.dependencyTree.downstream.length > 0)}
    <Panel title="Dependency tree">
      {#if data.dependencyTree.upstream.length > 0}
        <div class="dep-section">
          <span class="dep-heading">Upstream ({data.dependencyTree.upstream.length})</span>
          {#each data.dependencyTree.upstream as d (d.project_id)}
            <a href={`/projects/${d.project_id}`} class="dep-row">
              <span class="dep-name">{d.project_name}</span>
              <Badge variant="muted" size="sm">{d.dep_type}</Badge>
            </a>
          {/each}
        </div>
      {/if}
      {#if data.dependencyTree.downstream.length > 0}
        <div class="dep-section">
          <span class="dep-heading">Downstream ({data.dependencyTree.downstream.length})</span>
          {#each data.dependencyTree.downstream as d (d.project_id)}
            <a href={`/projects/${d.project_id}`} class="dep-row">
              <span class="dep-name">{d.project_name}</span>
              <Badge variant="muted" size="sm">{d.dep_type}</Badge>
            </a>
          {/each}
        </div>
      {/if}
    </Panel>
  {/if}
{/snippet}

{#snippet panelLiveRun()}
  {#if data.liveRun}
    <Panel title="Live run">
      <div class="live-run-header">
        <StatusPill status={data.liveRun.run.status === 'running' ? 'running' : 'queued'} size="sm" />
        <a href={`/runs/${data.liveRun.run.id}`} class="live-run-name">{data.liveRun.run.task_name}</a>
        <span class="dim" style="font-size: 0.75rem;"><TimeAgo value={data.liveRun.run.started_at} /></span>
      </div>
      {#if data.liveRun.log_tail}
        <pre class="live-log-tail">{data.liveRun.log_tail}</pre>
      {/if}
    </Panel>
  {/if}
{/snippet}

{#snippet renderPanel(key: PanelKey)}
  {#if key === 'tasks'}{@render panelTasks()}
  {:else if key === 'recent-runs'}{@render panelRecentRuns()}
  {:else if key === 'repository'}{@render panelRepository()}
  {:else if key === 'latest-commit'}{@render panelLatestCommit()}
  {:else if key === 'code-stats'}{@render panelCodeStats()}
  {:else if key === 'about'}{@render panelAbout()}
  {:else if key === 'contributors'}{@render panelContributors()}
  {:else if key === 'pipelines'}{@render panelPipelines()}
  {:else if key === 'pipeline-configs'}{@render panelPipelineConfigs()}
  {:else if key === 'compose'}{@render panelCompose()}
  {:else if key === 'blame-timeline'}{@render panelBlameTimeline()}
  {:else if key === 'monorepo'}{@render panelMonorepo()}
  {:else if key === 'previews'}{@render panelPreviews()}
  {:else if key === 'alert-rules'}{@render panelAlertRules()}
  {:else if key === 'impact-radar'}{@render panelImpactRadar()}
  {:else if key === 'dependency-tree'}{@render panelDependencyTree()}
  {:else if key === 'live-run'}{@render panelLiveRun()}
  {/if}
{/snippet}

<div class="panel-stack" class:editing={editLayoutMode}>
  {#if editLayoutMode}
    <div class="row-gap"
         class:active={dropRowGap === 0 && dragSource !== null}
         ondragover={(e) => onRowGapDragOver(0, e)}
         ondrop={(e) => onRowGapDrop(0, e)}
         role="presentation"></div>
  {/if}
  {#each panelLayout as row, rowIdx (rowIdx)}
    {@const visible = visibleRow(row)}
    {#if visible.length > 0}
      <div class="panel-row" style="grid-template-columns: repeat({visible.length}, minmax(0, 1fr));">
        {#each row as key, colIdx (key)}
          {#if editLayoutMode || panelHasData(key)}
            <div class="panel-shell"
                 class:editing={editLayoutMode}
                 class:dragging={dragSource && dragSource.row === rowIdx && dragSource.col === colIdx}
                 class:drop-left={dropTarget && dropTarget.row === rowIdx && dropTarget.col === colIdx && dropTarget.side === 'left'}
                 class:drop-right={dropTarget && dropTarget.row === rowIdx && dropTarget.col === colIdx && dropTarget.side === 'right'}
                 role={editLayoutMode ? 'group' : undefined}
                 aria-label={editLayoutMode ? `Panel ${key} (drag to reorder)` : undefined}
                 draggable={editLayoutMode}
                 ondragstart={(e) => onPanelDragStart(rowIdx, colIdx, e)}
                 ondragover={(e) => onPanelDragOver(rowIdx, colIdx, e)}
                 ondragleave={() => onPanelDragLeave(rowIdx, colIdx)}
                 ondrop={(e) => onPanelDrop(rowIdx, colIdx, e)}
                 ondragend={onPanelDragEnd}>
              {#if editLayoutMode}
                <span class="drag-handle" aria-hidden="true">⋮⋮</span>
              {/if}
              {#if editLayoutMode && !panelHasData(key)}
                <Panel title={panelEmptyLabel(key)}>
                  <div class="panel-empty-hint">No data yet — this slot stays in your layout for when {panelEmptyKind(key)} appear.</div>
                </Panel>
              {:else}
                {@render renderPanel(key)}
              {/if}
            </div>
          {/if}
        {/each}
      </div>
    {/if}
    {#if editLayoutMode}
      <div class="row-gap"
           class:active={dropRowGap === rowIdx + 1 && dragSource !== null}
           ondragover={(e) => onRowGapDragOver(rowIdx + 1, e)}
           ondrop={(e) => onRowGapDrop(rowIdx + 1, e)}
           role="presentation"></div>
    {/if}
  {/each}
</div>

<!-- Modals -->

<Modal open={paramsModal !== null} title={paramsModal ? `Run with parameters: ${paramsModal.taskName}` : ''} width={560} onClose={() => paramsModal = null}>
  {#if paramsModal}
    <div class="field">
      <label for="env-input">Env vars (one per line, KEY=VALUE)</label>
      <textarea id="env-input" rows="4"
                placeholder={'NODE_ENV=production\nLOG_LEVEL=debug'}
                bind:value={paramsModal.env}></textarea>
    </div>
    <div class="field">
      <label for="args-input">Extra args (one per line)</label>
      <textarea id="args-input" rows="4"
                placeholder={'--watch\n--reporter=verbose'}
                bind:value={paramsModal.args}></textarea>
    </div>
    {#if runError}
      <p class="modal-error">{runError}</p>
    {/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (paramsModal = null)}>Cancel</button>
      <button type="button" onclick={submitWithParams}>Run</button>
    </div>
  {/if}
</Modal>

<Modal open={timeoutModal !== null} title={timeoutModal ? `Timeout: ${timeoutModal.taskName}` : ''} width={420} onClose={() => timeoutModal = null}>
  {#if timeoutModal}
    <div class="field">
      <label for="timeout-input">Timeout in seconds</label>
      <input id="timeout-input" type="number" min="1" max="86400"
             placeholder="e.g. 1800 = 30m, 7200 = 2h"
             bind:value={timeoutModal.value} />
      <p class="modal-hint">Leave blank for the server default. 1–86400 seconds (24h).</p>
    </div>
    {#if timeoutModal.error}
      <p class="modal-error">{timeoutModal.error}</p>
    {/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (timeoutModal = null)}>Cancel</button>
      <button type="button" onclick={submitTimeout}>Save</button>
    </div>
  {/if}
</Modal>

<Modal open={retryModal !== null} title={retryModal ? `Retry policy: ${retryModal.taskName}` : ''} width={460} onClose={() => retryModal = null}>
  {#if retryModal}
    <div class="modal-fields-grid">
      <div class="field">
        <label for="retry-max-input">Max retries</label>
        <input id="retry-max-input" type="number" min="0" max="10" bind:value={retryModal.max} />
        <p class="modal-hint">0 = no retry, up to 10.</p>
      </div>
      <div class="field">
        <label for="retry-backoff-input">Backoff (seconds)</label>
        <input id="retry-backoff-input" type="number" min="1" max="3600"
               disabled={parseInt(retryModal.max, 10) <= 0}
               bind:value={retryModal.backoff} />
        <p class="modal-hint">Wait between retries.</p>
      </div>
    </div>
    {#if retryModal.error}
      <p class="modal-error">{retryModal.error}</p>
    {/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (retryModal = null)}>Cancel</button>
      <button type="button" onclick={submitRetry}>Save</button>
    </div>
  {/if}
</Modal>

<Modal open={artifactsModal !== null} title={artifactsModal ? `Artifact patterns: ${artifactsModal.taskName}` : ''} width={560} onClose={() => artifactsModal = null}>
  {#if artifactsModal}
    <p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">
      One glob pattern per line. After each run finishes, matching files are captured and listed on the run detail page.
    </p>
    <div class="field">
      <label for="artifacts-input">Patterns</label>
      <textarea id="artifacts-input" rows="6"
                placeholder={'**/junit.xml\ncoverage/*.xml\nbuild/output.tar.gz'}
                bind:value={artifactsModal.raw}></textarea>
      <p class="modal-hint">Supported: literal paths, basename globs (<code>*.xml</code>), and <code>**/</code> for any-depth.</p>
    </div>
    {#if artifactsModal.error}<p class="modal-error">{artifactsModal.error}</p>{/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (artifactsModal = null)}>Cancel</button>
      <button type="button" onclick={submitArtifacts}>Save</button>
    </div>
  {/if}
</Modal>

<Modal open={servicesModal !== null} title={servicesModal ? `Required services: ${servicesModal.taskName}` : ''} width={520} onClose={() => servicesModal = null}>
  {#if servicesModal}
    <p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">
      Comma-separated compose service names. Workend will refuse to run this task unless the project's compose stack is up.
    </p>
    <div class="field">
      <label for="services-input">Services</label>
      <input id="services-input" type="text" placeholder="postgres, redis"
             bind:value={servicesModal.raw} />
      <p class="modal-hint">Match service keys from your <code>compose.yaml</code>. Leave blank to remove the requirement.</p>
    </div>
    {#if servicesModal.error}<p class="modal-error">{servicesModal.error}</p>{/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (servicesModal = null)}>Cancel</button>
      <button type="button" onclick={submitServices}>Save</button>
    </div>
  {/if}
</Modal>

<Modal open={branchModal !== null} title={branchModal ? `Run on branch: ${branchModal.taskName}` : ''} width={480} onClose={() => branchModal = null}>
  {#if branchModal}
    <p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">
      Workend will fetch this branch's HEAD into an ephemeral checkout — your project's main checkout is not switched.
    </p>
    {#if branchModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading branches…</div>
    {:else if branchModal.error}
      <p class="modal-error">{branchModal.error}</p>
    {:else if branchModal.branches && branchModal.branches.length > 0}
      <form method="POST" action="?/runOnBranch" class="inline-form" onsubmit={() => { branchModal = null; }}>
        <input type="hidden" name="task_id" value={branchModal.taskID} />
        <div class="field">
          <label for="branch-pick">Branch</label>
          <select id="branch-pick" name="branch" bind:value={branchModal.selected} required>
            {#each branchModal.branches as b (b.name)}
              <option value={b.name}>{b.name}</option>
            {/each}
          </select>
          <p class="modal-hint">{branchModal.branches.length} branches available.</p>
        </div>
        <div class="modal-actions">
          <button type="button" class="ghost" onclick={() => (branchModal = null)}>Cancel</button>
          <button type="submit">Run</button>
        </div>
      </form>
    {:else}
      <p style="color: var(--text-dim);">No branches available.</p>
    {/if}
  {/if}
</Modal>

<Modal open={concurrencyModal !== null} title={concurrencyModal ? `Concurrency: ${concurrencyModal.taskName}` : ''} width={520} onClose={() => concurrencyModal = null}>
  {#if concurrencyModal}
    <div class="modal-fields-grid">
      <div class="field">
        <label for="conc-max">Max concurrent runs</label>
        <input id="conc-max" type="number" min="0" max="100" bind:value={concurrencyModal.max} />
        <p class="modal-hint">0 = use the legacy single-run guard. 1+ allows that many in flight.</p>
      </div>
      <div class="field">
        <label for="conc-policy">When at limit</label>
        <select id="conc-policy" bind:value={concurrencyModal.policy} disabled={parseInt(concurrencyModal.max, 10) <= 0}>
          <option value="queue">queue (reject new for now)</option>
          <option value="cancel-old">cancel-old (supersede the oldest)</option>
          <option value="reject">reject (return 429)</option>
        </select>
        <p class="modal-hint">cancel-old is the right choice for fast-pushing branches and `deploy` tasks.</p>
      </div>
    </div>
    {#if concurrencyModal.error}<p class="modal-error">{concurrencyModal.error}</p>{/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (concurrencyModal = null)}>Cancel</button>
      <button type="button" onclick={submitConcurrency}>Save</button>
    </div>
  {/if}
</Modal>

<!-- Blame detail modal -->
<Modal open={blameDetailModal !== null} title={blameDetailModal ? `Commit ${blameDetailModal.sha.slice(0, 8)}` : ''} width={640} onClose={() => blameDetailModal = null}>
  {#if blameDetailModal}
    {#if blameDetailModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading commit details...</div>
    {:else if blameDetailModal.error}
      <p class="modal-error">{blameDetailModal.error}</p>
    {:else if blameDetailModal.commit}
      <div class="row"><span class="label">SHA</span><span class="value">{blameDetailModal.commit.sha}</span></div>
      <div class="row"><span class="label">Author</span><span class="value commit-msg">{blameDetailModal.commit.author} &lt;{blameDetailModal.commit.author_email}&gt;</span></div>
      <div class="row"><span class="label">Date</span><span class="value">{new Date(blameDetailModal.commit.committed_at).toLocaleString()}</span></div>
      <div class="row"><span class="label">Message</span><span class="value commit-msg">{blameDetailModal.commit.message}</span></div>
      <div class="row"><span class="label">Changes</span><span class="value">{blameDetailModal.commit.files_changed} files, <span class="ins">+{blameDetailModal.commit.insertions}</span> <span class="del">-{blameDetailModal.commit.deletions}</span></span></div>
      {#if blameDetailModal.runs.length > 0}
        <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Associated runs ({blameDetailModal.runs.length})</h3>
        {#each blameDetailModal.runs as r (r.id)}
          <a href={`/runs/${r.id}`} class="run-row" style="padding: var(--space-1) 0;">
            <StatusPill status={r.status} size="sm" />
            <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
            <span class="run-meta">{r.started_at ? new Date(r.started_at).toLocaleString() : '--'}</span>
          </a>
        {/each}
      {:else}
        <p style="color: var(--text-dim); font-size: 0.875rem; margin-top: var(--space-3);">No runs associated with this commit.</p>
      {/if}
    {/if}
    <div class="modal-actions" style="margin-top: var(--space-4);">
      <button type="button" class="ghost" onclick={() => (blameDetailModal = null)}>Close</button>
    </div>
  {/if}
</Modal>

<!-- Package detail modal -->
<Modal open={packageDetailModal !== null} title={packageDetailModal?.pkg ? packageDetailModal.pkg.name : 'Package detail'} width={560} onClose={() => packageDetailModal = null}>
  {#if packageDetailModal}
    {#if packageDetailModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading package...</div>
    {:else if packageDetailModal.error}
      <p class="modal-error">{packageDetailModal.error}</p>
    {:else if packageDetailModal.pkg}
      <div class="row"><span class="label">Path</span><span class="value">{packageDetailModal.pkg.path}</span></div>
      <div class="row"><span class="label">Type</span><span class="value"><Badge variant={pkgTypeBadgeVariant(packageDetailModal.pkg.pkg_type)} size="sm">{packageDetailModal.pkg.pkg_type}</Badge></span></div>
      <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Scoped tasks ({packageDetailModal.scopes.length})</h3>
      {#if packageDetailModal.scopes.length > 0}
        {#each packageDetailModal.scopes as scope (scope.id)}
          <div class="scope-row">
            <span class="scope-name">{scope.task_name}</span>
            <button type="button" class="ghost danger" style="font-size: 0.75rem;"
                    onclick={() => removeTaskScope(scope.id)}>Remove</button>
          </div>
        {/each}
      {:else}
        <p style="color: var(--text-dim); font-size: 0.875rem;">No tasks scoped to this package.</p>
      {/if}
      {#if data.tasks.length > 0}
        {@const scopedIDs = new Set(packageDetailModal.scopes.map((s) => s.task_id))}
        {@const availableTasks = data.tasks.filter((t) => !scopedIDs.has(t.id))}
        {#if availableTasks.length > 0}
          <div style="margin-top: var(--space-3); border-top: 1px solid var(--border); padding-top: var(--space-3);">
            <span style="font-size: 0.8125rem; color: var(--text-muted);">Add task scope:</span>
            <div class="scope-add-list">
              {#each availableTasks.slice(0, 10) as t (t.id)}
                <button type="button" class="ghost" style="font-size: 0.75rem;"
                        onclick={() => addTaskScope(packageDetailModal?.id ?? '', t.id)}>+ {t.name}</button>
              {/each}
            </div>
          </div>
        {/if}
      {/if}
    {/if}
    <div class="modal-actions" style="margin-top: var(--space-4);">
      <button type="button" class="ghost" onclick={() => (packageDetailModal = null)}>Close</button>
    </div>
  {/if}
</Modal>

<!-- Task metrics modal -->
<Modal open={metricsModal !== null} title={metricsModal ? `Metrics: ${metricsModal.taskName}` : ''} width={640} onClose={() => metricsModal = null}>
  {#if metricsModal}
    {#if metricsModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading metrics...</div>
    {:else if metricsModal.error}
      <p class="modal-error">{metricsModal.error}</p>
    {:else if metricsModal.metrics}
      <div class="metrics-grid">
        <div class="metric-card">
          <div class="metric-num">{formatMs(metricsModal.metrics.p50_ms)}</div>
          <div class="metric-label">p50</div>
        </div>
        <div class="metric-card">
          <div class="metric-num">{formatMs(metricsModal.metrics.p95_ms)}</div>
          <div class="metric-label">p95</div>
        </div>
        <div class="metric-card">
          <div class="metric-num">{formatMs(metricsModal.metrics.p99_ms)}</div>
          <div class="metric-label">p99</div>
        </div>
        <div class="metric-card">
          <div class="metric-num">{metricsModal.metrics.total_runs.toLocaleString()}</div>
          <div class="metric-label">Total runs</div>
        </div>
      </div>
      <div class="success-rate-bar" style="margin-top: var(--space-4);">
        <div class="success-rate-label">
          <span>Success rate</span>
          <span class="success-rate-pct">{(metricsModal.metrics.success_rate * 100).toFixed(1)}%</span>
        </div>
        <div class="rate-track">
          <div class="rate-fill" style="width: {metricsModal.metrics.success_rate * 100}%;"></div>
        </div>
        <span style="font-size: 0.75rem; color: var(--text-dim);">{metricsModal.metrics.last_30d_runs} runs in last 30 days</span>
      </div>
      {#if metricsModal.trends.length > 0}
        <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Daily trend (last 30 days)</h3>
        {@const trendMax = Math.max(...metricsModal.trends.map((t) => t.count), 1)}
        <div class="trend-chart">
          {#each metricsModal.trends as day (day.date)}
            <div class="trend-bar-group" title="{day.date}: {day.count} runs ({day.passed}p/{day.failed}f) avg {formatMs(day.avg_ms)}">
              <div class="trend-bar">
                {#if day.passed > 0}
                  <div class="trend-pass" style="height: {(day.passed / trendMax) * 100}%;"></div>
                {/if}
                {#if day.failed > 0}
                  <div class="trend-fail" style="height: {(day.failed / trendMax) * 100}%;"></div>
                {/if}
              </div>
            </div>
          {/each}
        </div>
      {/if}
    {/if}
    <div class="modal-actions" style="margin-top: var(--space-4);">
      <button type="button" class="ghost" onclick={() => (metricsModal = null)}>Close</button>
    </div>
  {/if}
</Modal>
