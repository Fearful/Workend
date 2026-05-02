<script lang="ts">
  let { data } = $props();

  function statusColor(status: string | null): string {
    switch (status) {
      case 'ready': case 'succeeded': return '#22c55e';
      case 'running': case 'queued': case 'cloning': case 'pending': return '#eab308';
      case 'failed': case 'error': return '#ef4444';
      case 'cancelled': return '#6b7280';
      default: return '#6b7280';
    }
  }

  function formatRelative(iso: string | null): string {
    if (!iso) return '—';
    const ms = Date.now() - new Date(iso).getTime();
    const min = Math.floor(ms / 60000);
    if (min < 1) return 'just now';
    if (min < 60) return `${min}m ago`;
    const hr = Math.floor(min / 60);
    if (hr < 24) return `${hr}h ago`;
    const d = Math.floor(hr / 24);
    return `${d}d ago`;
  }

  function freshnessClass(iso: string | null): string {
    if (!iso) return 'fresh-stale';
    const days = (Date.now() - new Date(iso).getTime()) / (24 * 60 * 60 * 1000);
    if (days < 7) return 'fresh-good';
    if (days < 30) return 'fresh-warn';
    return 'fresh-stale';
  }

  let groupedByWorkspace = $derived.by(() => {
    const groups: Record<string, typeof data.cards> = {};
    for (const c of data.cards) {
      if (!groups[c.workspace_name]) groups[c.workspace_name] = [];
      groups[c.workspace_name].push(c);
    }
    return groups;
  });
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 1.5rem 0; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 1.75rem 0 0.75rem 0; }

  .grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(300px, 1fr)); gap: 0.75rem; }

  .card {
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    padding: 1rem 1.125rem; text-decoration: none; color: inherit;
    display: flex; flex-direction: column; gap: 0.625rem;
    transition: border-color 100ms ease;
  }
  .card:hover { border-color: #2563eb; text-decoration: none; }
  .card.alert { border-color: #5b1a1a; }
  .card.alert:hover { border-color: #ef4444; }

  .card-head {
    display: flex; align-items: center; gap: 0.625rem;
    font-weight: 500; font-size: 0.95rem;
  }
  .dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }

  .meta-grid {
    display: grid; grid-template-columns: 1fr 1fr;
    gap: 0.375rem 0.75rem; font-size: 0.75rem;
  }
  .label { color: #6b7280; }
  .value { color: #cbd5e1; font-family: ui-monospace, "SF Mono", Menlo, monospace; }

  .freshness { font-size: 0.6875rem; padding: 0.125rem 0.5rem; border-radius: 999px; }
  .fresh-good  { background: rgba(34, 197, 94, 0.1); color: #22c55e; }
  .fresh-warn  { background: rgba(234, 179, 8, 0.1); color: #eab308; }
  .fresh-stale { background: rgba(107, 114, 128, 0.15); color: #9ca3af; }

  .commit-line {
    font-size: 0.75rem; color: #9ca3af; font-style: italic;
    overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  }

  .empty { text-align: center; padding: 3rem 1rem; color: #9ca3af; }

  .alert-flag {
    color: #ef4444; font-size: 0.6875rem; font-weight: 600;
    background: rgba(239, 68, 68, 0.1); padding: 0.125rem 0.5rem; border-radius: 999px;
  }
</style>

<h1>Dashboard</h1>

{#if data.cards.length === 0}
  <div class="empty">No projects yet. <a href="/">Create a workspace</a> and add some.</div>
{:else}
  {#each Object.entries(groupedByWorkspace) as [wsName, cards] (wsName)}
    <h2>{wsName}</h2>
    <div class="grid">
      {#each cards as c (c.project_id)}
        <a href={`/projects/${c.project_id}`} class="card {c.has_failing_recent_run ? 'alert' : ''}">
          <div class="card-head">
            <span class="dot" style="background: {statusColor(c.status)}"></span>
            <span style="flex:1;">{c.project_name}</span>
            {#if c.has_failing_recent_run}<span class="alert-flag">recent failure</span>{/if}
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
            <span class="freshness {freshnessClass(c.last_synced_at)}">{formatRelative(c.last_synced_at)}</span>

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
