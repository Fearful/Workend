<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';

  let { data } = $props();

  function statusColor(status: string): string {
    switch (status) {
      case 'succeeded':
        return '#22c55e';
      case 'queued':
      case 'running':
        return '#eab308';
      case 'failed':
        return '#ef4444';
      case 'cancelled':
        return '#6b7280';
      default:
        return '#6b7280';
    }
  }

  function shortSha(sha: string | null): string {
    return sha ? sha.slice(0, 7) : '—';
  }

  function formatDuration(start: string | null, end: string | null): string {
    if (!start) return '—';
    const startMs = new Date(start).getTime();
    const endMs = end ? new Date(end).getTime() : Date.now();
    const sec = Math.round((endMs - startMs) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  function setFilter(status: string) {
    const url = new URL(page.url);
    if (status) url.searchParams.set('status', status);
    else url.searchParams.delete('status');
    goto(url, { replaceState: true });
  }
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 0.5rem 0;
  }

  .breadcrumb {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .filters {
    display: flex;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    align-items: center;
    font-size: 0.875rem;
  }

  .filters .label {
    color: #6b7280;
  }

  .pill {
    padding: 0.25rem 0.75rem;
    background: transparent;
    color: #9ca3af;
    border: 1px solid #2d3540;
    border-radius: 999px;
    cursor: pointer;
    font-size: 0.8125rem;
    font-family: inherit;
  }

  .pill.active {
    background: #2563eb;
    color: white;
    border-color: #2563eb;
  }

  .pill:hover:not(.active) {
    background: #1a1f25;
    color: #e8eaed;
  }

  .empty {
    text-align: center;
    padding: 3rem;
    color: #9ca3af;
    background: #14181d;
    border: 1px dashed #1f2429;
    border-radius: 8px;
  }

  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto auto;
    align-items: center;
    gap: 1rem;
    padding: 0.625rem 1rem;
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 6px;
    text-decoration: none;
    color: inherit;
    margin-bottom: 0.375rem;
    font-size: 0.875rem;
    transition: border-color 100ms ease;
  }

  .run-row:hover {
    border-color: #2563eb;
    text-decoration: none;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .run-name {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .run-meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a>
  {#if data.workspace}
    / <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a>
  {/if}
  / <a href={`/projects/${data.project.id}`}>{data.project.name}</a> / runs
</div>

<h1>Run history</h1>

<div class="filters">
  <span class="label">Filter:</span>
  <button class="pill {data.filter === '' ? 'active' : ''}" onclick={() => setFilter('')}>All</button>
  <button class="pill {data.filter === 'succeeded' ? 'active' : ''}" onclick={() => setFilter('succeeded')}>Succeeded</button>
  <button class="pill {data.filter === 'failed' ? 'active' : ''}" onclick={() => setFilter('failed')}>Failed</button>
  <button class="pill {data.filter === 'cancelled' ? 'active' : ''}" onclick={() => setFilter('cancelled')}>Cancelled</button>
  <button class="pill {data.filter === 'running' ? 'active' : ''}" onclick={() => setFilter('running')}>Running</button>
</div>

{#if data.runs.length === 0}
  <div class="empty">
    {#if data.filter}
      No runs matching <strong>{data.filter}</strong>.
    {:else}
      No runs yet.
    {/if}
  </div>
{:else}
  {#each data.runs as r (r.id)}
    <a href={`/runs/${r.id}`} class="run-row">
      <span class="dot" style="background: {statusColor(r.status)}"></span>
      <span class="run-name">{r.task_name} <span style="color:#6b7280">({r.task_source})</span></span>
      <span class="run-meta">{shortSha(r.commit_sha)}</span>
      <span class="run-meta">{r.exit_code ?? '—'}</span>
      <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
      <span class="run-meta">{new Date(r.created_at).toLocaleString()}</span>
    </a>
  {/each}
{/if}
