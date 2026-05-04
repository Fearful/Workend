<script lang="ts">
  import Panel from '$lib/components/Panel.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';

  type ActivityCell = {
    date: string;
    total: number;
    succeeded: number;
    failed: number;
    cancelled: number;
  };

  let { data } = $props();

  function taskColor(taskID: string): string {
    let h = 0;
    for (let i = 0; i < taskID.length; i++) h = (h * 31 + taskID.charCodeAt(i)) | 0;
    return `hsl(${Math.abs(h) % 360}, 65%, 60%)`;
  }

  function formatDur(sec: number): string {
    if (sec < 60) return `${sec}s`;
    if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
    return `${Math.floor(sec / 3600)}h ${Math.floor((sec % 3600) / 60)}m`;
  }

  const W = 760;
  const H = 320;
  const pad = { top: 16, right: 16, bottom: 28, left: 48 };
  const innerW = W - pad.left - pad.right;
  const innerH = H - pad.top - pad.bottom;

  let domain = $derived.by(() => {
    if (data.trends.points.length === 0) {
      return { minT: 0, maxT: 1, maxD: 1 };
    }
    let minT = Infinity, maxT = -Infinity, maxD = 0;
    for (const p of data.trends.points) {
      const t = new Date(p.finished_at).getTime();
      if (t < minT) minT = t;
      if (t > maxT) maxT = t;
      if (p.duration_sec > maxD) maxD = p.duration_sec;
    }
    if (minT === maxT) maxT = minT + 1;
    if (maxD === 0) maxD = 1;
    return { minT, maxT, maxD };
  });

  function xPos(tISO: string): number {
    const t = new Date(tISO).getTime();
    return pad.left + ((t - domain.minT) / (domain.maxT - domain.minT)) * innerW;
  }

  function yPos(sec: number): number {
    return pad.top + innerH - (sec / domain.maxD) * innerH;
  }

  let series = $derived.by(() => {
    const byTask = new Map<string, typeof data.trends.points>();
    for (const p of data.trends.points) {
      const list = byTask.get(p.task_id) ?? [];
      list.push(p);
      byTask.set(p.task_id, list);
    }
    return Array.from(byTask.entries()).map(([taskID, pts]) => ({
      taskID,
      taskName: pts[0].task_name,
      taskSource: pts[0].task_source,
      color: taskColor(taskID),
      points: pts
    }));
  });

  function pathFor(pts: typeof data.trends.points): string {
    return pts
      .map((p, i) => `${i === 0 ? 'M' : 'L'} ${xPos(p.finished_at).toFixed(1)} ${yPos(p.duration_sec).toFixed(1)}`)
      .join(' ');
  }

  let yTicks = $derived.by(() => {
    const out: { v: number; y: number }[] = [];
    for (let i = 0; i <= 4; i++) {
      const v = (domain.maxD * i) / 4;
      out.push({ v, y: yPos(v) });
    }
    return out;
  });

  function fmtDate(ms: number): string {
    const d = new Date(ms);
    return `${d.getMonth() + 1}/${d.getDate()}`;
  }

  let heatmap = $derived.by(() => {
    const counts = new Map<string, ActivityCell>();
    for (const c of data.activity.cells) counts.set(c.date, c);

    const today = new Date();
    today.setHours(0, 0, 0, 0);
    const end = new Date(today);
    end.setDate(end.getDate() + (6 - end.getDay()));
    const start = new Date(end);
    start.setDate(start.getDate() - 7 * 53 + 1);

    const cols: { date: string; cell: ActivityCell | null }[][] = [];
    let cursor = new Date(start);
    let week: { date: string; cell: ActivityCell | null }[] = [];
    while (cursor <= end) {
      const iso = cursor.toISOString().slice(0, 10);
      week.push({ date: iso, cell: counts.get(iso) ?? null });
      if (week.length === 7) {
        cols.push(week);
        week = [];
      }
      cursor = new Date(cursor);
      cursor.setDate(cursor.getDate() + 1);
    }
    if (week.length > 0) cols.push(week);
    let max = 0;
    for (const c of data.activity.cells) if (c.total > max) max = c.total;
    return { cols, max };
  });

  function cellColor(c: ActivityCell | null, max: number): string {
    if (!c || c.total === 0) return 'var(--bg-hover)';
    const intensity = Math.min(1, c.total / Math.max(1, max));
    if (c.failed > 0) {
      const a = 0.25 + 0.55 * intensity;
      return `rgba(239, 68, 68, ${a})`;
    }
    const a = 0.25 + 0.55 * intensity;
    return `rgba(34, 197, 94, ${a})`;
  }

  function cellTooltip(d: { date: string; cell: ActivityCell | null }): string {
    if (!d.cell) return `${d.date} — no runs`;
    const parts = [`${d.cell.total} runs`];
    if (d.cell.succeeded) parts.push(`${d.cell.succeeded} ok`);
    if (d.cell.failed) parts.push(`${d.cell.failed} failed`);
    if (d.cell.cancelled) parts.push(`${d.cell.cancelled} cancelled`);
    return `${d.date} — ${parts.join(', ')}`;
  }

  let xTicks = $derived.by(() => {
    const out: { label: string; x: number }[] = [];
    for (let i = 0; i <= 4; i++) {
      const t = domain.minT + ((domain.maxT - domain.minT) * i) / 4;
      out.push({ label: fmtDate(t), x: pad.left + (innerW * i) / 4 });
    }
    return out;
  });

  let hovered = $state<{ p: (typeof data.trends.points)[0]; x: number; y: number } | null>(null);

  function navWith(params: Record<string, string>) {
    const p = new URLSearchParams(window.location.search);
    for (const [k, v] of Object.entries(params)) {
      if (v === '') p.delete(k);
      else p.set(k, v);
    }
    window.location.search = p.toString();
  }
</script>

<style>
  .controls {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
  .controls label {
    display: inline;
    margin: 0;
    color: var(--text-muted);
  }
  .controls select {
    background: var(--bg-input);
    color: var(--text);
    border: 1px solid var(--border-strong);
    padding: 0.375rem 0.625rem;
    border-radius: var(--radius-md);
    font: inherit;
    font-size: 0.875rem;
    width: auto;
  }

  .legend {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem var(--space-5);
    margin-top: 0.75rem;
    font-size: 0.8125rem;
    color: var(--text-muted);
  }
  .legend-item {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
  }
  .swatch {
    width: 10px;
    height: 10px;
    border-radius: 2px;
    display: inline-block;
  }

  .summary {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-4);
  }

  .stat {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
  }
  .stat-num {
    font-size: 1.25rem;
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

  svg {
    display: block;
    max-width: 100%;
    height: auto;
  }

  .grid-line {
    stroke: var(--border);
    stroke-width: 1;
  }

  .axis-label {
    fill: var(--text-dim);
    font-size: 11px;
    font-family: var(--font-mono);
  }

  .tooltip {
    pointer-events: none;
    font-size: 0.75rem;
    background: var(--bg-page);
    color: var(--text);
    border: 1px solid var(--border-strong);
    padding: var(--space-2) 0.625rem;
    border-radius: var(--radius-md);
    font-family: var(--font-mono);
  }

  .heatmap-section { overflow-x: auto; }
  .heatmap-header {
    font-size: 0.75rem;
    color: var(--text-muted);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--space-2);
  }
  .heatmap-grid { display: flex; gap: 2px; }
  .heatmap-week {
    display: grid;
    grid-template-rows: repeat(7, 12px);
    gap: 2px;
  }
  .heatmap-cell {
    width: 12px;
    height: 12px;
    border-radius: 2px;
  }
  .heatmap-legend {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-top: var(--space-2);
    font-size: 0.75rem;
    color: var(--text-dim);
  }
  .heatmap-legend-cell {
    display: inline-block;
    width: 12px;
    height: 12px;
    border-radius: 2px;
  }
  .heatmap-legend-divider { margin-left: var(--space-4); }

  .chart-panel { position: relative; }
  .point-circle { cursor: pointer; }

  @media (min-width: 1280px) {
    .heatmap-cell { width: 14px; height: 14px; }
    .heatmap-week { grid-template-rows: repeat(7, 14px); }
  }
</style>

<SectionHeader title="Run duration trends" />

{#if heatmap.cols.length > 0}
  <section class="panel heatmap-section">
    <div class="heatmap-header">Activity (last 53 weeks)</div>
    <div class="heatmap-grid">
      {#each heatmap.cols as week, wi (wi)}
        <div class="heatmap-week">
          {#each week as day, di (di)}
            <div title={cellTooltip(day)}
                 class="heatmap-cell"
                 style="background: {cellColor(day.cell, heatmap.max)};"></div>
          {/each}
        </div>
      {/each}
    </div>
    <div class="heatmap-legend">
      <span>less</span>
      <span class="heatmap-legend-cell" style="background:var(--bg-hover);"></span>
      <span class="heatmap-legend-cell" style="background:rgba(34,197,94,0.4);"></span>
      <span class="heatmap-legend-cell" style="background:rgba(34,197,94,0.7);"></span>
      <span>more</span>
      <span class="heatmap-legend-divider">·</span>
      <span class="heatmap-legend-cell" style="background:rgba(239,68,68,0.5);"></span>
      <span>has failures</span>
    </div>
  </section>
{/if}

<div class="controls">
  <label for="task">Task:</label>
  <select id="task"
          value={data.filterTaskID}
          onchange={(e) => navWith({ task_id: (e.currentTarget as HTMLSelectElement).value })}>
    <option value="">All tasks</option>
    {#each data.tasks as t (t.id)}
      <option value={t.id}>{t.source} · {t.name}</option>
    {/each}
  </select>
  <label for="days">Window:</label>
  <select id="days"
          value={String(data.days)}
          onchange={(e) => navWith({ days: (e.currentTarget as HTMLSelectElement).value })}>
    <option value="7">7 days</option>
    <option value="30">30 days</option>
    <option value="90">90 days</option>
    <option value="180">180 days</option>
    <option value="365">365 days</option>
  </select>
</div>

{#if data.trends.points.length === 0}
  <Panel>
    <EmptyState message={`No completed runs in the last ${data.days} days.`} />
  </Panel>
{:else}
  {@const total = data.trends.points.length}
  {@const succ = data.trends.points.filter((p) => p.status === 'succeeded').length}
  {@const fail = data.trends.points.filter((p) => p.status === 'failed').length}
  {@const avg = Math.round(data.trends.points.reduce((s, p) => s + p.duration_sec, 0) / total)}
  {@const max = Math.max(...data.trends.points.map((p) => p.duration_sec))}

  <div class="summary">
    <div class="stat"><div class="stat-num">{total}</div><div class="stat-label">Runs</div></div>
    <div class="stat"><div class="stat-num">{Math.round((succ / total) * 100)}%</div><div class="stat-label">Success rate</div></div>
    <div class="stat"><div class="stat-num">{fail}</div><div class="stat-label">Failures</div></div>
    <div class="stat"><div class="stat-num">{formatDur(avg)}</div><div class="stat-label">Average</div></div>
    <div class="stat"><div class="stat-num">{formatDur(max)}</div><div class="stat-label">Slowest</div></div>
  </div>

  <section class="panel chart-panel">
    <svg viewBox="0 0 {W} {H}" role="img" aria-label="Duration trend chart">
      {#each yTicks as t (t.v)}
        <line class="grid-line" x1={pad.left} x2={W - pad.right} y1={t.y} y2={t.y} />
        <text class="axis-label" x={pad.left - 6} y={t.y + 3} text-anchor="end">{formatDur(Math.round(t.v))}</text>
      {/each}
      {#each xTicks as t, i (i)}
        <text class="axis-label" x={t.x} y={H - 8} text-anchor="middle">{t.label}</text>
      {/each}
      {#each series as s (s.taskID)}
        <path d={pathFor(s.points)} fill="none" stroke={s.color} stroke-width="1.5" stroke-linejoin="round" />
        {#each s.points as p (p.run_id)}
          <circle cx={xPos(p.finished_at)} cy={yPos(p.duration_sec)} r="3"
                  fill={p.status === 'succeeded' ? s.color : 'transparent'}
                  stroke={s.color} stroke-width="1.5"
                  role="button" tabindex="0"
                  class="point-circle"
                  aria-label={`${p.task_name} run · ${formatDur(p.duration_sec)} · ${p.status}`}
                  onmouseenter={() => (hovered = { p, x: xPos(p.finished_at), y: yPos(p.duration_sec) })}
                  onmouseleave={() => (hovered = null)}
                  onclick={() => (window.location.href = `/runs/${p.run_id}`)}
                  onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') window.location.href = `/runs/${p.run_id}`; }} />
        {/each}
      {/each}
      {#if hovered}
        <line class="grid-line" x1={hovered.x} x2={hovered.x} y1={pad.top} y2={H - pad.bottom} stroke-dasharray="2 3" />
      {/if}
    </svg>
    {#if hovered}
      <div class="tooltip"
           style="position:absolute; left:{Math.min(hovered.x + 10, W - 200)}px; top:{Math.max(hovered.y - 50, 8)}px;">
        <div><strong>{hovered.p.task_name}</strong> ({hovered.p.task_source})</div>
        <div>{formatDur(hovered.p.duration_sec)} · {hovered.p.status}{hovered.p.timed_out ? ' (timeout)' : ''}</div>
        <div style="color:var(--text-dim);">{new Date(hovered.p.finished_at).toLocaleString()}</div>
      </div>
    {/if}

    <div class="legend">
      {#each series as s (s.taskID)}
        <span class="legend-item">
          <span class="swatch" style="background:{s.color}"></span>
          {s.taskName} <span style="color:var(--text-dim)">({s.taskSource}, {s.points.length})</span>
        </span>
      {/each}
    </div>
  </section>
{/if}
