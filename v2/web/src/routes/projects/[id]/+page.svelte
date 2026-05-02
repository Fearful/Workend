<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';

  let { data, form } = $props();

  function statusColor(status: string): string {
    switch (status) {
      case 'ready':
      case 'succeeded':
        return '#22c55e';
      case 'cloning':
      case 'pending':
      case 'queued':
      case 'running':
        return '#eab308';
      case 'error':
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

  onMount(maybeStartPolling);
  onDestroy(() => {
    if (pollHandle) clearInterval(pollHandle);
  });

  let tasksBySource = $derived.by(() => {
    const groups: Record<string, typeof data.tasks> = {};
    for (const t of data.tasks) {
      if (!groups[t.source]) groups[t.source] = [];
      groups[t.source].push(t);
    }
    return groups;
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

  h2 {
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
    margin: 2rem 0 1rem 0;
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

  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    gap: 1rem;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
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

  .commit-msg {
    font-family: inherit;
  }

  .task-group {
    margin-bottom: 1rem;
  }

  .task-source-label {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    color: #60a5fa;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: 0.5rem;
    padding: 0.125rem 0.5rem;
    background: rgba(96, 165, 250, 0.1);
    border-radius: 4px;
    display: inline-block;
  }

  .task-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    align-items: center;
    gap: 1rem;
    padding: 0.625rem 0.75rem;
    border-bottom: 1px solid #1f2429;
  }

  .task-row:last-child {
    border-bottom: none;
  }

  .task-name {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.875rem;
  }

  .task-cmd {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .empty {
    color: #6b7280;
    font-size: 0.875rem;
    text-align: center;
    padding: 1.5rem;
  }

  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto;
    align-items: center;
    gap: 1rem;
    padding: 0.5rem 0.75rem;
    border-bottom: 1px solid #1f2429;
    text-decoration: none;
    color: inherit;
    font-size: 0.875rem;
  }

  .run-row:hover {
    background: #1a1f25;
    text-decoration: none;
  }

  .run-row:last-child {
    border-bottom: none;
  }

  .run-name {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .run-meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .error-banner {
    color: #ef4444;
    font-size: 0.875rem;
    background: #2a1414;
    border: 1px solid #5b1a1a;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a>
  {#if data.workspace}
    / <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a>
  {/if}
  / {data.project.name}
</div>

<div class="header-row">
  <h1>
    <span class="dot" style="background: {statusColor(data.project.status)}"></span>
    {data.project.name}
  </h1>
  <div class="actions">
    <form method="POST" action="?/sync" style="margin: 0;">
      <button type="submit">Sync</button>
    </form>
    <form method="POST" action="?/delete" style="margin: 0;" onsubmit={(e) => !confirm('Delete this project?') && e.preventDefault()}>
      <button type="submit" class="ghost">Delete</button>
    </form>
  </div>
</div>

{#if form?.error}
  <div class="error-banner">{form.error}</div>
{/if}

<section class="panel">
  <h2>Repository</h2>
  <div class="row"><span class="label">Git URL</span><span class="value">{data.project.git_url}</span></div>
  <div class="row"><span class="label">Branch</span><span class="value">{data.project.default_branch || '—'}</span></div>
  <div class="row"><span class="label">Status</span><span class="value">{data.project.status}</span></div>
  <div class="row"><span class="label">Local path</span><span class="value">{data.project.local_path || '—'}</span></div>
</section>

<section class="panel">
  <h2>Latest commit</h2>
  <div class="row"><span class="label">SHA</span><span class="value">{shortSha(data.project.last_commit_sha)}</span></div>
  <div class="row"><span class="label">Author</span><span class="value commit-msg">{data.project.last_commit_author || '—'}</span></div>
  <div class="row"><span class="label">Message</span><span class="value commit-msg">{data.project.last_commit_message || '—'}</span></div>
  <div class="row"><span class="label">Synced</span><span class="value">{data.project.last_synced_at ? new Date(data.project.last_synced_at).toLocaleString() : '—'}</span></div>
</section>

<section class="panel">
  <h2>Tasks</h2>
  {#if data.tasks.length === 0}
    <div class="empty">
      {#if data.project.status === 'ready'}
        No runnable tasks detected. Add a <code>package.json</code> with a <code>scripts</code> section, or a <code>justfile</code>.
      {:else}
        Tasks will appear after the repo finishes cloning.
      {/if}
    </div>
  {:else}
    {#each Object.entries(tasksBySource) as [source, tasks] (source)}
      <div class="task-group">
        <div class="task-source-label">{source}</div>
        {#each tasks as t (t.id)}
          <div class="task-row">
            <span class="task-name">{t.name}</span>
            <span class="task-cmd">{t.raw_command}</span>
            <form method="POST" action="?/run" style="margin: 0;">
              <input type="hidden" name="task_id" value={t.id} />
              <button type="submit" disabled={data.project.status !== 'ready'}>Run</button>
            </form>
          </div>
        {/each}
      </div>
    {/each}
  {/if}
</section>

{#if data.runs.length > 0}
  <section class="panel">
    <h2 style="display:flex; align-items:center; justify-content:space-between;">
      <span>Recent runs</span>
      <a href={`/projects/${data.project.id}/runs`} style="font-size:0.75rem; text-transform:none; letter-spacing:0; color:#60a5fa; font-weight:400;">View all →</a>
    </h2>
    {#each data.runs.slice(0, 10) as r (r.id)}
      <a href={`/runs/${r.id}`} class="run-row">
        <span class="dot" style="background: {statusColor(r.status)}"></span>
        <span class="run-name">{r.task_name} <span style="color:#6b7280">({r.task_source})</span></span>
        <span class="run-meta">{r.status}</span>
        <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
        <span class="run-meta">{new Date(r.created_at).toLocaleString()}</span>
      </a>
    {/each}
  </section>
{/if}
