<script lang="ts">
  import { statusColor, formatRelative, freshnessClass } from '$lib/utils';
  import PageHeader from '$lib/components/PageHeader.svelte';
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

</script>

<style>
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
