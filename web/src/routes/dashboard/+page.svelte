<script lang="ts">
  import { statusColor, formatRelative, freshnessClass } from '$lib/utils';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data } = $props();

  let groupedByWorkspace = $derived.by(() => {
    const groups: Record<string, typeof data.cards> = {};
    for (const c of data.cards) {
      if (!groups[c.workspace_name]) groups[c.workspace_name] = [];
      groups[c.workspace_name].push(c);
    }
    return groups;
  });

  let stats = $derived.by(() => {
    const running = data.cards.filter((c) => c.last_run_status === 'running' || c.last_run_status === 'queued').length;
    const failing = data.cards.filter((c) => c.has_failing_recent_run).length;
    return {
      total: data.cards.length,
      running,
      failing
    };
  });

  const DAY_LABELS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  const HOUR_LABELS = Array.from({ length: 24 }, (_, i) => {
    if (i === 0) return '12a';
    if (i < 12) return `${i}a`;
    if (i === 12) return '12p';
    return `${i - 12}p`;
  });

  let heatmapGrid = $derived.by(() => {
    const grid: number[][] = Array.from({ length: 7 }, () => Array(24).fill(0));
    let max = 0;
    for (const b of data.heatmap) {
      const row = b.day_of_week;
      const col = b.hour;
      if (row >= 0 && row < 7 && col >= 0 && col < 24) {
        grid[row][col] = b.failures;
        if (b.failures > max) max = b.failures;
      }
    }
    return { grid, max };
  });

  let queueItemCount = $derived(
    data.myQueue.pending_approvals.length +
    data.myQueue.expiring_sandboxes.length +
    (data.myQueue.unread_mentions > 0 ? 1 : 0)
  );

  let hasWidgets = $derived(
    data.runPulse.length > 0 ||
    data.heatmap.length > 0 ||
    queueItemCount > 0 ||
    data.velocity.this_week.runs > 0 ||
    data.velocity.last_week.runs > 0 ||
    data.sandboxes.length > 0 ||
    data.quota.length > 0
  );

  function formatMs(ms: number): string {
    if (ms <= 0) return '0s';
    const sec = Math.round(ms / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function passRate(passed: number, total: number): string {
    if (total === 0) return '—';
    return `${Math.round((passed / total) * 100)}%`;
  }

  function deltaClass(delta: number, higherIsBetter: boolean): string {
    if (delta === 0) return '';
    const positive = delta > 0;
    return (positive === higherIsBetter) ? 'delta-good' : 'delta-bad';
  }

  function deltaArrow(delta: number): string {
    if (delta > 0) return '↑';
    if (delta < 0) return '↓';
    return '';
  }

</script>

<style>
  /* ── Widgets section ── */
  .widgets-section {
    margin-bottom: var(--space-6);
  }

  .widgets-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(340px, 1fr));
    gap: var(--space-4);
  }

  .widget-heatmap {
    grid-column: span 2;
  }

  @media (max-width: 767px) {
    .widgets-grid {
      grid-template-columns: 1fr;
    }
    .widget-heatmap {
      grid-column: span 1;
    }
  }

  /* ── Run Pulse ── */
  .pulse-list {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .pulse-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0.375rem 0.5rem;
    border-radius: var(--radius-md);
    font-size: var(--fs-sm);
    color: inherit;
    text-decoration: none;
    transition: background 120ms ease;
  }
  .pulse-row:hover {
    background: rgba(107, 114, 128, 0.08);
    text-decoration: none;
  }

  .pulse-task {
    font-family: var(--font-mono);
    font-weight: var(--fw-medium);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pulse-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    white-space: nowrap;
  }

  /* ── Sprint Velocity ── */
  .velocity-table {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  .vel-header,
  .vel-row {
    display: grid;
    grid-template-columns: 5.5rem 1fr 1fr 1fr;
    gap: var(--space-2);
    align-items: center;
    padding: 0.25rem 0.5rem;
  }

  .vel-header {
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-dim);
    padding-bottom: 0.375rem;
    border-bottom: 1px solid var(--border);
  }

  .vel-row {
    font-size: var(--fs-sm);
    border-radius: var(--radius-md);
    transition: background 120ms ease;
  }
  .vel-row:hover {
    background: rgba(107, 114, 128, 0.05);
  }

  .vel-label {
    color: var(--text-muted);
    font-size: var(--fs-xs);
    font-weight: var(--fw-medium);
  }

  .vel-val {
    font-family: var(--font-mono);
    text-align: right;
  }

  .vel-col {
    text-align: right;
  }

  .vel-delta {
    font-weight: var(--fw-semibold);
  }

  .delta-good {
    color: var(--success);
  }

  .delta-bad {
    color: var(--danger-text);
  }

  /* ── My Queue ── */
  .queue-list {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .queue-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0.375rem 0.5rem;
    border-radius: var(--radius-md);
    font-size: var(--fs-sm);
    color: inherit;
    text-decoration: none;
    transition: background 120ms ease;
  }
  a.queue-item:hover {
    background: rgba(107, 114, 128, 0.08);
    text-decoration: none;
  }

  .queue-task {
    font-family: var(--font-mono);
    font-weight: var(--fw-medium);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .queue-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    white-space: nowrap;
  }

  .queue-urgent {
    color: var(--warning);
    font-weight: var(--fw-semibold);
  }

  /* ── Failure Heatmap ── */
  .heatmap-wrap {
    overflow-x: auto;
    -webkit-overflow-scrolling: touch;
  }

  .heatmap-grid {
    display: grid;
    grid-template-columns: 2rem repeat(24, 1fr);
    grid-template-rows: auto repeat(7, 1fr);
    gap: 2px;
    min-width: 480px;
  }

  .heatmap-corner {
    grid-column: 1;
    grid-row: 1;
  }

  .heatmap-hour-label {
    font-size: 0.5625rem;
    color: var(--text-dim);
    text-align: center;
    line-height: 1;
    padding-bottom: 0.25rem;
  }

  .heatmap-day-label {
    font-size: 0.625rem;
    color: var(--text-dim);
    display: flex;
    align-items: center;
    padding-right: 0.25rem;
    line-height: 1;
  }

  .heatmap-cell {
    aspect-ratio: 1;
    border-radius: 2px;
    background: rgba(239, 68, 68, calc(0.08 + var(--intensity) * 0.82));
    transition: opacity 120ms ease;
    min-width: 12px;
    min-height: 12px;
  }
  .heatmap-cell:hover {
    opacity: 0.75;
    outline: 1px solid var(--text-dim);
  }

  .heatmap-legend {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    justify-content: flex-end;
    margin-top: var(--space-2);
    padding-right: 0.25rem;
  }

  .heatmap-legend-label {
    font-size: 0.5625rem;
    color: var(--text-dim);
  }

  .heatmap-legend-cell {
    width: 12px;
    height: 12px;
    border-radius: 2px;
    background: rgba(239, 68, 68, calc(0.08 + var(--intensity) * 0.82));
  }

  /* ── Sandbox Status ── */
  .sandbox-list {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .sandbox-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding: 0.375rem 0.5rem;
    border-radius: var(--radius-md);
    font-size: var(--fs-sm);
    transition: background 120ms ease;
  }
  .sandbox-row:hover {
    background: rgba(107, 114, 128, 0.05);
  }

  .sandbox-branch {
    font-family: var(--font-mono);
    font-weight: var(--fw-medium);
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .sandbox-project {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    white-space: nowrap;
  }

  /* ── Quota Meter ── */
  .quota-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-3);
  }

  .quota-row {
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
  }

  .quota-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: var(--space-2);
  }

  .quota-name {
    font-size: var(--fs-sm);
    font-weight: var(--fw-medium);
  }

  .quota-stats {
    font-size: var(--fs-xs);
    color: var(--text-dim);
    white-space: nowrap;
  }

  .quota-bar-track {
    height: 6px;
    background: rgba(107, 114, 128, 0.15);
    border-radius: var(--radius-full);
    overflow: hidden;
  }

  .quota-bar-fill {
    height: 100%;
    background: var(--accent);
    border-radius: var(--radius-full);
    transition: width 300ms ease;
  }
  .quota-bar-fill.quota-warn {
    background: var(--warning);
  }
  .quota-bar-fill.quota-danger {
    background: var(--danger-text);
  }

  .quota-pct {
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
    color: var(--text-muted);
    text-align: right;
  }

  /* ── Existing layout ── */
  .layout {
    display: grid;
    grid-template-columns: 1fr;
    gap: var(--space-6);
  }

  @media (min-width: 1280px) {
    .layout {
      grid-template-columns: 1fr 320px;
    }
    .sidebar { order: 2; position: sticky; top: var(--space-4); align-self: start; }
    .main-content { order: 1; min-width: 0; }
  }

  h2 {
    font-size: 0.8125rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    font-weight: 600;
    margin: var(--space-6) 0 var(--space-3) 0;
  }
  h2:first-child { margin-top: 0; }
  h2.warn { color: var(--warning); }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-3);
  }

  .card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) 1.125rem;
    text-decoration: none;
    color: inherit;
    display: flex;
    flex-direction: column;
    gap: 0.625rem;
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .card:hover {
    border-color: var(--accent);
    text-decoration: none;
    box-shadow: var(--shadow-card-hover);
  }
  .card.alert { border-color: rgba(239, 68, 68, 0.3); }
  .card.alert:hover { border-color: var(--danger-text); }

  .card-head {
    display: flex;
    align-items: center;
    gap: 0.625rem;
    font-weight: 500;
    font-size: 0.95rem;
  }
  .card-name { flex: 1; }

  .meta-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.375rem var(--space-3);
    font-size: 0.75rem;
  }
  .label { color: var(--text-dim); }
  .value { color: var(--text); font-family: var(--font-mono); }

  .commit-line {
    font-size: 0.75rem;
    color: var(--text-muted);
    font-style: italic;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .pin-card,
  .flaky-card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-3) 0.875rem;
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .pin-card:hover,
  .flaky-card:hover { box-shadow: var(--shadow-card-hover); }
  .flaky-card { border-color: rgba(234, 179, 8, 0.4); text-decoration: none; color: inherit; }

  .pin-head,
  .flaky-head {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: 0.875rem;
  }
  .pin-task,
  .flaky-task {
    font-family: var(--font-mono);
    font-weight: 500;
    flex: 1;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pin-source,
  .flaky-source { color: var(--text-dim); font-size: 0.75rem; }

  .pin-loc,
  .flaky-loc {
    color: var(--text-muted);
    font-size: 0.75rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pin-loc a { color: inherit; }

  .pin-actions { display: flex; gap: 0.375rem; }
  .pin-actions form { margin: 0; }
  .pin-run-form { flex: 1; }
  .pin-run-form button { width: 100%; }

  .flaky-stats {
    display: flex;
    gap: var(--space-4);
    font-size: 0.75rem;
    color: var(--text-dim);
  }
  .flaky-flips { color: var(--warning); font-family: var(--font-mono); font-weight: 600; }

  .pinned-grid,
  .flaky-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 0.625rem;
    margin-bottom: var(--space-2);
  }

  @media (min-width: 1280px) {
    .pinned-grid,
    .flaky-grid {
      grid-template-columns: 1fr;
    }
  }

  .stats-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: var(--space-2);
  }
  .stat-tile {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-3);
    text-align: center;
  }
  .stat-num {
    font-size: 1.5rem;
    font-weight: 600;
    font-family: var(--font-mono);
    line-height: 1;
  }
  .stat-num.danger { color: var(--danger-text); }
  .stat-num.warning { color: var(--warning); }
  .stat-label {
    font-size: 0.6875rem;
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-top: 0.25rem;
  }

  @media (max-width: 767px) {
    .pinned-grid,
    .flaky-grid {
      grid-auto-flow: column;
      grid-template-columns: none;
      grid-auto-columns: 85vw;
      overflow-x: auto;
      scroll-snap-type: x mandatory;
      scroll-padding: var(--space-4);
      padding-bottom: var(--space-2);
      -webkit-overflow-scrolling: touch;
    }
    .pinned-grid > *,
    .flaky-grid > * {
      scroll-snap-align: start;
    }
  }
</style>

<PageHeader title="Dashboard" />

{#if hasWidgets}
<div class="widgets-section">
  <div class="widgets-grid">

    {#if data.runPulse.length > 0}
      <div class="widget widget-run-pulse">
        <Panel title="Run Pulse" padding="compact">
          <div class="pulse-list">
            {#each data.runPulse as run (run.id)}
              <a href={`/runs/${run.id}`} class="pulse-row">
                <StatusPill status={run.status} size="sm" />
                <span class="pulse-task">{run.task_name}</span>
                <span class="pulse-meta">{run.project_name}</span>
                <span class="pulse-meta">{run.workspace_name}</span>
                <TimeAgo value={run.started_at} />
              </a>
            {/each}
          </div>
        </Panel>
      </div>
    {/if}

    {#if data.velocity.this_week.runs > 0 || data.velocity.last_week.runs > 0}
      <div class="widget widget-velocity">
        <Panel title="Sprint Velocity" padding="compact">
          <div class="velocity-table">
            <div class="vel-header">
              <span class="vel-label"></span>
              <span class="vel-col">This week</span>
              <span class="vel-col">Last week</span>
              <span class="vel-col">Delta</span>
            </div>
            <div class="vel-row">
              <span class="vel-label">Runs</span>
              <span class="vel-val">{data.velocity.this_week.runs}</span>
              <span class="vel-val">{data.velocity.last_week.runs}</span>
              <span class="vel-val vel-delta {deltaClass(data.velocity.deltas.runs, true)}">
                {deltaArrow(data.velocity.deltas.runs)} {Math.abs(data.velocity.deltas.runs)}
              </span>
            </div>
            <div class="vel-row">
              <span class="vel-label">Pass rate</span>
              <span class="vel-val">{passRate(data.velocity.this_week.passed, data.velocity.this_week.runs)}</span>
              <span class="vel-val">{passRate(data.velocity.last_week.passed, data.velocity.last_week.runs)}</span>
              <span class="vel-val vel-delta {deltaClass(data.velocity.deltas.passed, true)}">
                {deltaArrow(data.velocity.deltas.passed)} {Math.abs(data.velocity.deltas.passed)}
              </span>
            </div>
            <div class="vel-row">
              <span class="vel-label">Failed</span>
              <span class="vel-val">{data.velocity.this_week.failed}</span>
              <span class="vel-val">{data.velocity.last_week.failed}</span>
              <span class="vel-val vel-delta {deltaClass(data.velocity.deltas.failed, false)}">
                {deltaArrow(data.velocity.deltas.failed)} {Math.abs(data.velocity.deltas.failed)}
              </span>
            </div>
            <div class="vel-row">
              <span class="vel-label">Avg duration</span>
              <span class="vel-val">{formatMs(data.velocity.this_week.avg_duration_ms)}</span>
              <span class="vel-val">{formatMs(data.velocity.last_week.avg_duration_ms)}</span>
              <span class="vel-val vel-delta {deltaClass(data.velocity.deltas.avg_duration_ms, false)}">
                {deltaArrow(data.velocity.deltas.avg_duration_ms)} {formatMs(Math.abs(data.velocity.deltas.avg_duration_ms))}
              </span>
            </div>
          </div>
        </Panel>
      </div>
    {/if}

    {#if queueItemCount > 0}
      <div class="widget widget-queue">
        <Panel title="My Queue" padding="compact">
          <div class="queue-list">
            {#each data.myQueue.pending_approvals as approval (approval.run_id)}
              <a href={`/runs/${approval.run_id}`} class="queue-item">
                <Badge variant="accent" size="sm">approval</Badge>
                <span class="queue-task">{approval.task_name}</span>
                <span class="queue-meta">{approval.project_name}</span>
                <TimeAgo value={approval.created_at} />
              </a>
            {/each}
            {#each data.myQueue.expiring_sandboxes as sb (sb.sandbox_id)}
              <div class="queue-item">
                <Badge variant="warning" size="sm">expiring</Badge>
                <span class="queue-task">{sb.branch}</span>
                <span class="queue-meta queue-urgent">{sb.minutes_left}m left</span>
              </div>
            {/each}
            {#if data.myQueue.unread_mentions > 0}
              <div class="queue-item">
                <Badge variant="info" size="sm">mentions</Badge>
                <span class="queue-task">{data.myQueue.unread_mentions} unread</span>
              </div>
            {/if}
          </div>
        </Panel>
      </div>
    {/if}

    {#if data.heatmap.length > 0}
      <div class="widget widget-heatmap">
        <Panel title="Failure Heatmap" padding="compact">
          <div class="heatmap-wrap">
            <div class="heatmap-grid">
              <div class="heatmap-corner"></div>
              {#each HOUR_LABELS as h, i}
                {#if i % 3 === 0}
                  <span class="heatmap-hour-label">{h}</span>
                {:else}
                  <span class="heatmap-hour-label"></span>
                {/if}
              {/each}
              {#each heatmapGrid.grid as dayRow, dayIdx}
                <span class="heatmap-day-label">{DAY_LABELS[dayIdx]}</span>
                {#each dayRow as count, hourIdx}
                  <div
                    class="heatmap-cell"
                    style="--intensity: {heatmapGrid.max > 0 ? count / heatmapGrid.max : 0}"
                    title="{DAY_LABELS[dayIdx]} {HOUR_LABELS[hourIdx]}: {count} failure{count !== 1 ? 's' : ''}"
                  ></div>
                {/each}
              {/each}
            </div>
            <div class="heatmap-legend">
              <span class="heatmap-legend-label">Less</span>
              <div class="heatmap-legend-cell" style="--intensity: 0"></div>
              <div class="heatmap-legend-cell" style="--intensity: 0.25"></div>
              <div class="heatmap-legend-cell" style="--intensity: 0.5"></div>
              <div class="heatmap-legend-cell" style="--intensity: 0.75"></div>
              <div class="heatmap-legend-cell" style="--intensity: 1"></div>
              <span class="heatmap-legend-label">More</span>
            </div>
          </div>
        </Panel>
      </div>
    {/if}

    {#if data.sandboxes.length > 0}
      <div class="widget widget-sandboxes">
        <Panel title="Sandbox Status" padding="compact">
          <div class="sandbox-list">
            {#each data.sandboxes as sb (sb.id)}
              <div class="sandbox-row">
                <StatusPill status={sb.status} size="sm" />
                <span class="sandbox-branch">{sb.branch}</span>
                <span class="sandbox-project">{sb.project_name}</span>
                {#if sb.minutes_left <= 10}
                  <Badge variant="danger" size="sm">{sb.minutes_left}m left</Badge>
                {:else if sb.minutes_left <= 30}
                  <Badge variant="warning" size="sm">{sb.minutes_left}m left</Badge>
                {:else}
                  <Badge variant="muted" size="sm">{sb.minutes_left}m left</Badge>
                {/if}
              </div>
            {/each}
          </div>
        </Panel>
      </div>
    {/if}

    {#if data.quota.length > 0}
      <div class="widget widget-quota">
        <Panel title="Quota Meter" padding="compact">
          <div class="quota-list">
            {#each data.quota as ws (ws.workspace_id)}
              <div class="quota-row">
                <div class="quota-head">
                  <span class="quota-name">{ws.workspace_name}</span>
                  <span class="quota-stats">{ws.projects} projects &middot; {ws.runs_30d} runs / 30d</span>
                </div>
                <div class="quota-bar-track">
                  <div
                    class="quota-bar-fill {ws.usage_percent >= 90 ? 'quota-danger' : ws.usage_percent >= 70 ? 'quota-warn' : ''}"
                    style="width: {Math.min(ws.usage_percent, 100)}%"
                  ></div>
                </div>
                <span class="quota-pct">{Math.round(ws.usage_percent)}%</span>
              </div>
            {/each}
          </div>
        </Panel>
      </div>
    {/if}

  </div>
</div>
{/if}

<div class="layout">
  <div class="main-content">
    {#if data.cards.length === 0}
      <EmptyState
        icon="◇"
        message="No projects yet. Create a workspace and add some."
        actionHref="/"
        actionLabel="Browse workspaces" />
    {:else}
      {#each Object.entries(groupedByWorkspace) as [wsName, cards] (wsName)}
        <h2>{wsName}</h2>
        <div class="grid">
          {#each cards as c (c.project_id)}
            <a href={`/projects/${c.project_id}`} class="card {c.has_failing_recent_run ? 'alert' : ''}">
              <div class="card-head">
                <StatusPill status={c.status} size="sm" />
                <span class="card-name">{c.project_name}</span>
                {#if c.has_failing_recent_run}
                  <Badge variant="danger" size="sm">recent failure</Badge>
                {/if}
              </div>

              {#if c.last_commit_message}
                <div class="commit-line">{c.last_commit_message}</div>
              {/if}

              <div class="meta-grid">
                <span class="label">Status</span><span class="value">{c.status}</span>

                <span class="label">Tasks</span><span class="value">{c.task_count}</span>

                <span class="label">Lang</span><span class="value">{c.top_language || '—'}</span>

                <span class="label">Lines</span><span class="value">{c.total_lines ? c.total_lines.toLocaleString() : '—'}</span>

                <span class="label">Last sync</span>
                {#if freshnessClass(c.last_synced_at) === 'fresh-good'}
                  <Badge variant="success" size="sm">{formatRelative(c.last_synced_at)}</Badge>
                {:else if freshnessClass(c.last_synced_at) === 'fresh-warn'}
                  <Badge variant="warning" size="sm">{formatRelative(c.last_synced_at)}</Badge>
                {:else}
                  <Badge variant="muted" size="sm">{formatRelative(c.last_synced_at)}</Badge>
                {/if}

                <span class="label">Last run</span>
                {#if c.last_run_status}
                  <span class="value" style="color: {statusColor(c.last_run_status)};">
                    {c.last_run_task_name} · {c.last_run_status}
                  </span>
                {:else}
                  <span class="value">never</span>
                {/if}
              </div>
            </a>
          {/each}
        </div>
      {/each}
    {/if}
  </div>

  <aside class="sidebar">
    {#if data.cards.length > 0}
      <h2>Overview</h2>
      <div class="stats-grid">
        <div class="stat-tile">
          <div class="stat-num">{stats.total}</div>
          <div class="stat-label">Projects</div>
        </div>
        <div class="stat-tile">
          <div class="stat-num warning">{stats.running}</div>
          <div class="stat-label">Running</div>
        </div>
        <div class="stat-tile">
          <div class="stat-num danger">{stats.failing}</div>
          <div class="stat-label">Failing</div>
        </div>
      </div>
    {/if}

    {#if data.pinned.length > 0}
      <h2>Pinned tasks</h2>
      <div class="pinned-grid">
        {#each data.pinned as p (p.task_id)}
          <div class="pin-card">
            <div class="pin-head">
              <span class="pin-task">{p.task_name}</span>
              <span class="pin-source">{p.task_source}</span>
            </div>
            <div class="pin-loc">
              <a href={`/projects/${p.project_id}`}>{p.workspace_name} / {p.project_name}</a>
            </div>
            <div class="pin-actions">
              <form method="POST" action="?/runPinned" class="pin-run-form">
                <input type="hidden" name="task_id" value={p.task_id} />
                <button type="submit">Run</button>
              </form>
              <form method="POST" action="?/unpin">
                <input type="hidden" name="task_id" value={p.task_id} />
                <button type="submit" class="ghost" title="Unpin" aria-label="Unpin">★</button>
              </form>
            </div>
          </div>
        {/each}
      </div>
    {/if}

    {#if data.flaky.length > 0}
      <h2 class="warn">Flaky tasks</h2>
      <div class="flaky-grid">
        {#each data.flaky as f (f.task_id)}
          <a href={`/projects/${f.project_id}/trends?task_id=${f.task_id}`} class="flaky-card">
            <div class="flaky-head">
              <span class="flaky-task">{f.task_name}</span>
              <span class="flaky-source">{f.task_source}</span>
            </div>
            <div class="flaky-loc">{f.workspace_name} / {f.project_name}</div>
            <div class="flaky-stats">
              <span><strong class="flaky-flips">{f.flips}</strong> flips / {f.total} runs</span>
              <span>{Math.round(f.success_rate * 100)}% success</span>
            </div>
          </a>
        {/each}
      </div>
    {/if}
  </aside>
</div>
