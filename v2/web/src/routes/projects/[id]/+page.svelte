<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';

  let { data } = $props();

  function statusColor(status: string): string {
    switch (status) {
      case 'ready':
        return '#22c55e';
      case 'cloning':
      case 'pending':
        return '#eab308';
      case 'error':
        return '#ef4444';
      default:
        return '#6b7280';
    }
  }

  function shortSha(sha: string | null): string {
    return sha ? sha.slice(0, 12) : '—';
  }

  // Auto-refresh while clone is in progress
  let pollHandle: ReturnType<typeof setInterval> | null = null;

  function maybeStartPolling() {
    const transient = data.project.status === 'cloning' || data.project.status === 'pending';
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
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
    margin: 0 0 1rem 0;
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

<section class="panel">
  <h2>Repository</h2>
  <div class="row">
    <span class="label">Git URL</span>
    <span class="value">{data.project.git_url}</span>
  </div>
  <div class="row">
    <span class="label">Branch</span>
    <span class="value">{data.project.default_branch || '—'}</span>
  </div>
  <div class="row">
    <span class="label">Status</span>
    <span class="value">{data.project.status}</span>
  </div>
  <div class="row">
    <span class="label">Local path</span>
    <span class="value">{data.project.local_path || '—'}</span>
  </div>
</section>

<section class="panel">
  <h2>Latest commit</h2>
  <div class="row">
    <span class="label">SHA</span>
    <span class="value">{shortSha(data.project.last_commit_sha)}</span>
  </div>
  <div class="row">
    <span class="label">Author</span>
    <span class="value commit-msg">{data.project.last_commit_author || '—'}</span>
  </div>
  <div class="row">
    <span class="label">Message</span>
    <span class="value commit-msg">{data.project.last_commit_message || '—'}</span>
  </div>
  <div class="row">
    <span class="label">Synced</span>
    <span class="value">{data.project.last_synced_at ? new Date(data.project.last_synced_at).toLocaleString() : '—'}</span>
  </div>
</section>
