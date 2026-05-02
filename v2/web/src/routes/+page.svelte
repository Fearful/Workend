<script lang="ts">
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

  function formatRelative(iso: string): string {
    const ms = Date.now() - new Date(iso).getTime();
    const min = Math.floor(ms / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const d = Math.floor(hr / 24);
    return `${d}d ago`;
  }
</script>

<style>
  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  h1 {
    font-size: 1.5rem;
    margin: 0;
  }

  h2 {
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
    margin: 2.5rem 0 1rem 0;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 1rem;
  }

  .card {
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    padding: 1.25rem;
    text-decoration: none;
    color: inherit;
    display: block;
    transition: border-color 100ms ease;
  }

  .card:hover {
    border-color: #2563eb;
    text-decoration: none;
  }

  .card h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
  }

  .card p {
    color: #9ca3af;
    font-size: 0.875rem;
    margin: 0 0 0.75rem 0;
    min-height: 1.25rem;
  }

  .meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .empty {
    text-align: center;
    padding: 3rem 1rem;
    color: #9ca3af;
  }

  .error {
    color: #ef4444;
    font-size: 0.875rem;
    background: #2a1414;
    border: 1px solid #5b1a1a;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }

  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: 1rem;
    padding: 0.5rem 1rem;
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

  .run-loc {
    color: #9ca3af;
    font-size: 0.8125rem;
  }

  .run-task {
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    color: #e8eaed;
  }

  .run-meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }
</style>

<div class="header-row">
  <h1>Workspaces</h1>
  <a href="/workspaces/new"><button>New workspace</button></a>
</div>

{#if data.error}
  <div class="error">{data.error}</div>
{/if}

{#if data.workspaces.length === 0}
  <div class="empty">
    <p>No workspaces yet. Create your first one.</p>
  </div>
{:else}
  <div class="grid">
    {#each data.workspaces as ws (ws.id)}
      <a href={`/workspaces/${ws.id}`} class="card">
        <h3>{ws.name}</h3>
        <p>{ws.description || '—'}</p>
        <div class="meta">created {new Date(ws.created_at).toLocaleDateString()}</div>
      </a>
    {/each}
  </div>
{/if}

{#if data.recentRuns.length > 0}
  <h2>Recent runs</h2>
  {#each data.recentRuns as r (r.id)}
    <a href={`/runs/${r.id}`} class="run-row">
      <span class="dot" style="background: {statusColor(r.status)}"></span>
      <div>
        <div class="run-task">{r.task_name} <span style="color:#6b7280">({r.task_source})</span></div>
        <div class="run-loc">{r.workspace_name} / {r.project_name}</div>
      </div>
      <span class="run-meta">{r.status}{r.exit_code !== null ? ` · exit ${r.exit_code}` : ''}</span>
      <span class="run-meta">{formatRelative(r.created_at)}</span>
    </a>
  {/each}
{/if}
