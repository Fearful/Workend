<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';

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
    return sha ? sha.slice(0, 12) : '—';
  }

  function formatDuration(start: string | null, end: string | null): string {
    if (!start) return '—';
    const startMs = new Date(start).getTime();
    const endMs = end ? new Date(end).getTime() : Date.now();
    const sec = Math.round((endMs - startMs) / 1000);
    if (sec < 60) return `${sec}s`;
    return `${Math.floor(sec / 60)}m ${sec % 60}s`;
  }

  let pollHandle: ReturnType<typeof setInterval> | null = null;

  function maybeStartPolling() {
    const transient = data.run.status === 'queued' || data.run.status === 'running';
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

  onMount(maybeStartPolling);
  onDestroy(() => {
    if (pollHandle) clearInterval(pollHandle);
  });
</script>

<style>
  h1 {
    font-size: 1.5rem;
    margin: 0 0 0.25rem 0;
    display: flex;
    align-items: center;
    gap: 0.75rem;
  }

  .dot {
    width: 10px;
    height: 10px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .breadcrumb {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .panel {
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    padding: 1.25rem 1.5rem;
    margin-bottom: 1rem;
  }

  .panel h2 {
    margin-top: 0;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
  }

  .row {
    display: grid;
    grid-template-columns: 140px 1fr;
    gap: 0.75rem;
    padding: 0.5rem 0;
    border-bottom: 1px solid #1f2429;
    font-size: 0.875rem;
  }

  .row:last-child {
    border-bottom: none;
  }

  .row .label {
    color: #6b7280;
  }

  .row .value {
    color: #e8eaed;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    word-break: break-all;
  }

  .log-panel {
    padding: 0;
  }

  .log-header {
    padding: 0.875rem 1.25rem;
    border-bottom: 1px solid #1f2429;
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  .log-header h2 {
    margin: 0;
  }

  .log-meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  pre {
    margin: 0;
    padding: 1rem 1.25rem;
    background: #0d0f12;
    color: #cbd5e1;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.8125rem;
    line-height: 1.4;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-word;
    min-height: 8rem;
    max-height: 70vh;
  }

  .empty-log {
    color: #6b7280;
    font-style: italic;
    padding: 1.5rem 1.25rem;
    text-align: center;
  }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a> / <a href={`/projects/${data.run.project_id}`}>back to project</a> / run {data.run.id.slice(0, 8)}
</div>

<h1>
  <span class="dot" style="background: {statusColor(data.run.status)}"></span>
  {data.run.task_name} <span style="color:#6b7280">({data.run.task_source})</span>
</h1>

<section class="panel">
  <h2>Run</h2>
  <div class="row"><span class="label">Status</span><span class="value">{data.run.status}</span></div>
  <div class="row"><span class="label">Exit code</span><span class="value">{data.run.exit_code ?? '—'}</span></div>
  <div class="row"><span class="label">Commit</span><span class="value">{shortSha(data.run.commit_sha)}</span></div>
  <div class="row"><span class="label">Duration</span><span class="value">{formatDuration(data.run.started_at, data.run.finished_at)}</span></div>
  <div class="row"><span class="label">Started</span><span class="value">{data.run.started_at ? new Date(data.run.started_at).toLocaleString() : '—'}</span></div>
  <div class="row"><span class="label">Finished</span><span class="value">{data.run.finished_at ? new Date(data.run.finished_at).toLocaleString() : '—'}</span></div>
</section>

<section class="panel log-panel">
  <div class="log-header">
    <h2>Log</h2>
    <span class="log-meta">{data.log.length} bytes</span>
  </div>
  {#if data.log}
    <pre>{data.log}</pre>
  {:else if data.run.status === 'queued' || data.run.status === 'running'}
    <div class="empty-log">Log will appear as the task runs.</div>
  {:else}
    <div class="empty-log">No output captured.</div>
  {/if}
</section>
