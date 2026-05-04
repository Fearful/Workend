<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data } = $props();

  function setDays(d: number) {
    const u = new URL(page.url);
    u.searchParams.set('days', String(d));
    goto(u, { replaceState: true });
  }

  function fmtDur(sec: number): string {
    if (sec < 60) return `${Math.round(sec)}s`;
    if (sec < 3600) return `${Math.floor(sec / 60)}m ${Math.round(sec % 60)}s`;
    return `${Math.floor(sec / 3600)}h ${Math.floor((sec % 3600) / 60)}m`;
  }

  let maxDay = $derived.by(() => {
    let m = 0;
    for (const b of data.summary?.daily_runs ?? []) if (b.total > m) m = b.total;
    return Math.max(1, m);
  });
</script>

<style>
  h1 {
    font-size: var(--fs-2xl);
    margin: 0 0 0.25rem 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
  }
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-2);
    flex-wrap: wrap;
  }
  .ws-tabs {
    border-bottom: 1px solid var(--border);
    margin-bottom: var(--space-5);
    display: flex;
    gap: 0;
  }
  .ws-tab {
    padding: 0.625rem 0.875rem;
    color: var(--text-dim);
    text-decoration: none;
    font-size: var(--fs-md);
    font-weight: var(--fw-medium);
    border-bottom: 2px solid transparent;
    margin-bottom: -1px;
  }
  .ws-tab:hover { color: var(--text); text-decoration: none; }
  .ws-tab.active { color: var(--text); border-bottom-color: var(--accent); }

  .controls {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    margin-bottom: var(--space-4);
  }
  .controls span { color: var(--text-dim); font-size: var(--fs-sm); }
  .pill {
    padding: 0.25rem 0.625rem;
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-full);
    font-size: var(--fs-sm);
    cursor: pointer;
  }
  .pill.active { background: var(--accent); color: white; border-color: var(--accent); }
  .pill:hover:not(.active) { background: var(--bg-hover); color: var(--text); }

  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(140px, 1fr));
    gap: var(--space-3);
    margin-bottom: var(--space-5);
  }
  .stat {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
  }
  .stat-num {
    font-size: var(--fs-xl);
    font-weight: var(--fw-semibold);
    color: var(--text);
    font-family: var(--font-mono);
    line-height: 1.1;
  }
  .stat-label {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-top: 0.125rem;
  }
  .stat .ok  { color: var(--status-success-fg); }
  .stat .bad { color: var(--status-danger-fg); }

  .grid-2 {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
  }
  @media (max-width: 768px) {
    .grid-2 { grid-template-columns: 1fr; }
  }

  .panel {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
  }
  .panel h2 {
    margin: 0 0 var(--space-3);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
  }

  .row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-3);
    padding: 0.5rem 0;
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
    text-decoration: none;
    color: inherit;
  }
  .row:last-child { border-bottom: none; }
  .row:hover { background: var(--bg-hover); }
  .row .name { font-family: var(--font-mono); }
  .row .sub { color: var(--text-dim); font-size: var(--fs-xs); }
  .row .meta { color: var(--text-dim); font-family: var(--font-mono); font-size: var(--fs-xs); }
  .row .meta.bad { color: var(--status-danger-fg); }

  .health-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .health-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    color: inherit;
  }
  .health-row:last-child { border-bottom: none; }
  .health-row:hover { background: var(--bg-hover); }
  .health-name { font-weight: var(--fw-medium); }
  .health-meta { color: var(--text-dim); font-size: var(--fs-xs); font-family: var(--font-mono); }
  .health-meta.bad  { color: var(--status-danger-fg); }
  .health-meta.warn { color: var(--status-warning-fg); }

  .runs-bar-grid {
    display: grid;
    grid-template-columns: 1fr;
    gap: 4px;
    margin: var(--space-2) 0 0;
  }
  .runs-bar-row {
    display: grid;
    grid-template-columns: 80px 1fr 60px;
    gap: var(--space-2);
    align-items: center;
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }
  .runs-bar-track {
    height: 8px;
    background: var(--border);
    border-radius: 4px;
    overflow: hidden;
    display: flex;
  }
  .runs-bar-ok   { background: var(--status-success-fg); }
  .runs-bar-bad  { background: var(--status-danger-fg); }
  .runs-bar-amount { font-family: var(--font-mono); text-align: right; }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name, href: `/workspaces/${data.workspace.id}` },
  { label: 'dashboard' }
]} />

<div class="header-row">
  <h1>{data.workspace.name}</h1>
</div>

<div class="ws-tabs">
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
</div>

<div class="controls">
  <span>Window:</span>
  {#each [7, 30, 90] as d}
    <button class="pill" class:active={data.days === d} onclick={() => setDays(d)}>{d} days</button>
  {/each}
</div>

{#if data.summaryError}<FlashMessage type="error">{data.summaryError}</FlashMessage>{/if}

{#if data.summary}
  <div class="stats">
    <div class="stat">
      <div class="stat-num">{data.summary.totals.projects}</div>
      <div class="stat-label">Projects</div>
    </div>
    <div class="stat">
      <div class="stat-num">{data.summary.totals.members}</div>
      <div class="stat-label">Members</div>
    </div>
    <div class="stat">
      <div class="stat-num">{data.summary.totals.runs_in_window.toLocaleString()}</div>
      <div class="stat-label">Runs</div>
    </div>
    <div class="stat">
      <div class="stat-num"><span class={data.summary.totals.failed_in_window > 0 ? 'bad' : ''}>{data.summary.totals.failed_in_window}</span></div>
      <div class="stat-label">Failed</div>
    </div>
    <div class="stat">
      <div class="stat-num"><span class={data.summary.totals.success_rate >= 0.9 ? 'ok' : 'bad'}>{Math.round((data.summary.totals.success_rate || 0) * 100)}%</span></div>
      <div class="stat-label">Success rate</div>
    </div>
    <div class="stat">
      <div class="stat-num">{data.summary.totals.active_now}</div>
      <div class="stat-label">Active now</div>
    </div>
  </div>

  <div class="grid-2">
    <section class="panel">
      <h2>Slowest tasks</h2>
      {#if data.summary.slowest_tasks.length === 0}
        <div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">Not enough data yet.</div>
      {:else}
        {#each data.summary.slowest_tasks as t (t.task_id)}
          <a class="row" href={`/projects/${t.project_id}/runs`}>
            <span></span>
            <span>
              <span class="name">{t.task_name}</span>
              <div class="sub">{t.project_name} · {t.task_source} · {t.runs} runs</div>
            </span>
            <span class="meta">{fmtDur(t.avg_seconds)} avg</span>
          </a>
        {/each}
      {/if}
    </section>

    <section class="panel">
      <h2>Most failing tasks</h2>
      {#if data.summary.failing_tasks.length === 0}
        <div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">No failures in window.</div>
      {:else}
        {#each data.summary.failing_tasks as t (t.task_id)}
          <a class="row" href={`/projects/${t.project_id}/runs?status=failed`}>
            <span></span>
            <span>
              <span class="name">{t.task_name}</span>
              <div class="sub">{t.project_name} · {t.task_source}</div>
            </span>
            <span class="meta bad">{Math.round(t.failure_rate * 100)}% · {t.failures}/{t.total}</span>
          </a>
        {/each}
      {/if}
    </section>
  </div>

  <section class="panel" style="margin-bottom: var(--space-5);">
    <h2>Run volume</h2>
    {#if data.summary.daily_runs.length === 0}
      <div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">No runs in window.</div>
    {:else}
      <div class="runs-bar-grid">
        {#each data.summary.daily_runs as b (b.date)}
          <div class="runs-bar-row">
            <span>{b.date.slice(5)}</span>
            <div class="runs-bar-track">
              <div class="runs-bar-ok"  style="width: {(b.succeeded / maxDay) * 100}%;"></div>
              <div class="runs-bar-bad" style="width: {(b.failed    / maxDay) * 100}%;"></div>
            </div>
            <span class="runs-bar-amount">{b.total}</span>
          </div>
        {/each}
      </div>
    {/if}
  </section>

  <SectionHeader title="Projects health" />
  <div class="health-list">
    {#each data.summary.projects_health as p (p.project_id)}
      <a class="health-row" href={`/projects/${p.project_id}`}>
        <StatusPill status={p.status} size="sm" />
        <span class="health-name">{p.project_name}</span>
        <span class="health-meta">
          {p.runs_in_window} runs
        </span>
        <span class="health-meta {p.failure_rate >= 0.5 ? 'bad' : (p.failure_rate >= 0.1 ? 'warn' : '')}">
          {p.runs_in_window > 0 ? `${Math.round(p.failure_rate * 100)}% fail` : '—'}
        </span>
        <span class="health-meta">
          {#if p.last_run_at}last <TimeAgo value={p.last_run_at} /> · {p.last_run_status ?? ''}{:else}no runs{/if}
        </span>
      </a>
    {/each}
  </div>
{/if}
