<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import RunRow from '$lib/components/RunRow.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data } = $props();

  let comparisonOpen = $state(false);
  let comparisonDays = $state(7);
  let comparisonFetched = $state<typeof data.comparison | null>(null);
  let comparisonLoading = $state(false);
  let comparisonData = $derived(comparisonFetched ?? data.comparison);

  async function loadComparison(days: number) {
    comparisonDays = days;
    comparisonLoading = true;
    try {
      const r = await fetch(`/api/widgets/workspaces/comparison?days=${days}`);
      if (r.ok) {
        comparisonFetched = await r.json();
      }
    } catch {
      // keep existing data on failure
    } finally {
      comparisonLoading = false;
    }
  }

  function healthColor(failRate: number, totalRuns: number): 'success' | 'warning' | 'danger' | 'muted' {
    if (totalRuns === 0) return 'muted';
    if (failRate < 0.10) return 'success';
    if (failRate <= 0.30) return 'warning';
    return 'danger';
  }

  function healthLabel(failRate: number, totalRuns: number): string {
    if (totalRuns === 0) return 'No data';
    if (failRate < 0.10) return 'Healthy';
    if (failRate <= 0.30) return 'Warning';
    return 'Critical';
  }

  function severityVariant(severity: string): 'danger' | 'warning' | 'info' | 'muted' {
    const s = severity.toLowerCase();
    if (s === 'critical') return 'danger';
    if (s === 'high') return 'warning';
    if (s === 'medium') return 'warning';
    return 'muted';
  }

  function formatDuration(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    if (ms < 60000) return `${(ms / 1000).toFixed(1)}s`;
    return `${(ms / 60000).toFixed(1)}m`;
  }

  function storageMax(items: typeof data.storage): number {
    let m = 0;
    for (const s of items) {
      const total = s.projects + s.tasks + s.runs + s.artifacts;
      if (total > m) m = total;
    }
    return m || 1;
  }

  let activeIncidents = $derived(data.incidents.filter(i => i.incident_count > 0));
</script>

<style>
  .section { margin-top: var(--space-6); }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: var(--space-4);
  }

  .card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5);
    text-decoration: none;
    color: inherit;
    display: block;
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .card:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }
  .card h3 {
    margin: 0 0 var(--space-2) 0;
    font-size: var(--fs-lg);
    font-weight: var(--fw-semibold);
    letter-spacing: -0.01em;
  }
  .card p {
    color: var(--text-muted);
    font-size: var(--fs-md);
    margin: 0 0 var(--space-3) 0;
    min-height: 1.25rem;
  }
  .meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
  }

  /* --- Incident banner --- */
  .incident-banner {
    background: var(--status-danger-bg);
    border: 1px solid var(--status-danger-border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
    margin-bottom: var(--space-4);
  }
  .incident-banner-title {
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--status-danger-fg);
    margin: 0 0 var(--space-2) 0;
  }
  .incident-item {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    font-size: var(--fs-sm);
    color: var(--text);
    padding: var(--space-1) 0;
  }

  /* --- Health grid --- */
  .health-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-3);
  }
  .health-card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    transition: border-color 120ms ease;
  }
  .health-card:hover {
    border-color: var(--border-strong);
  }
  .health-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-3);
  }
  .health-name {
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
  }
  .health-stats {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  .health-stat {
    text-align: center;
  }
  .health-stat-value {
    font-size: var(--fs-lg);
    font-weight: var(--fw-semibold);
    font-family: var(--font-mono);
    color: var(--text);
  }
  .health-stat-label {
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }
  .fail-bar {
    height: 6px;
    background: var(--bg-hover);
    border-radius: var(--radius-full);
    overflow: hidden;
  }
  .fail-bar-fill {
    height: 100%;
    border-radius: var(--radius-full);
    transition: width 300ms ease;
  }
  .fail-bar-fill.success { background: var(--success); }
  .fail-bar-fill.warning { background: var(--warning); }
  .fail-bar-fill.danger { background: var(--danger); }
  .health-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-2);
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }

  /* --- Comparison table --- */
  .collapsible-toggle {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    cursor: pointer;
    background: none;
    border: none;
    color: var(--text-muted);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    padding: 0;
  }
  .collapsible-toggle:hover {
    color: var(--text);
    background: none;
  }
  .toggle-arrow {
    display: inline-block;
    transition: transform 150ms ease;
    font-size: var(--fs-xs);
  }
  .toggle-arrow.open {
    transform: rotate(90deg);
  }
  .days-selector {
    display: flex;
    gap: var(--space-1);
    margin: var(--space-3) 0;
  }
  .days-btn {
    font-size: var(--fs-xs);
    padding: 0.25rem 0.625rem;
    border-radius: var(--radius-full);
    background: var(--bg-hover);
    color: var(--text-muted);
    border: 1px solid var(--border);
    cursor: pointer;
  }
  .days-btn:hover { background: var(--bg-panel); color: var(--text); }
  .days-btn.active { background: var(--accent); color: white; border-color: var(--accent); }
  .days-btn.active:hover { background: var(--accent-hover); }
  .comp-table {
    width: 100%;
    border-collapse: collapse;
    font-size: var(--fs-sm);
  }
  .comp-table th, .comp-table td {
    padding: 0.5rem 0.75rem;
    text-align: left;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  .comp-table th {
    color: var(--text-muted);
    font-weight: var(--fw-semibold);
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .comp-table tr:last-child td { border-bottom: none; }
  .comp-table td.mono { font-family: var(--font-mono); }

  /* --- Storage overview --- */
  .storage-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: var(--space-3);
  }
  .storage-card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
  }
  .storage-name {
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    margin-bottom: var(--space-3);
  }
  .storage-row {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    margin-bottom: var(--space-2);
    font-size: var(--fs-xs);
    color: var(--text-muted);
  }
  .storage-label {
    width: 60px;
    flex-shrink: 0;
  }
  .storage-bar-track {
    flex: 1;
    height: 8px;
    background: var(--bg-hover);
    border-radius: var(--radius-full);
    overflow: hidden;
  }
  .storage-bar-fill {
    height: 100%;
    border-radius: var(--radius-full);
    transition: width 300ms ease;
    min-width: 2px;
  }
  .bar-projects   { background: var(--accent); }
  .bar-tasks      { background: var(--info); }
  .bar-runs       { background: var(--success); }
  .bar-artifacts  { background: var(--warning); }
  .storage-count {
    width: 48px;
    text-align: right;
    font-family: var(--font-mono);
    flex-shrink: 0;
  }

  @media (max-width: 640px) {
    .health-stats { grid-template-columns: 1fr 1fr; }
  }
</style>

<PageHeader title="Workspaces">
  {#snippet actions()}
    <a href="/workspaces/new"><button>New workspace</button></a>
  {/snippet}
</PageHeader>

{#if data.error}
  <FlashMessage type="error">{data.error}</FlashMessage>
{/if}

{#if activeIncidents.length > 0}
  <div class="incident-banner">
    <p class="incident-banner-title">Open incidents</p>
    {#each activeIncidents as inc (inc.workspace_id)}
      <div class="incident-item">
        <Badge variant={severityVariant(inc.worst_severity)} size="sm">{inc.worst_severity}</Badge>
        <span>{inc.workspace_name} — {inc.incident_count} incident{inc.incident_count === 1 ? '' : 's'}</span>
      </div>
    {/each}
  </div>
{/if}

{#if data.health.length > 0}
  <div class="section">
    <SectionHeader title="Workspace health" />
    <div class="health-grid">
      {#each data.health as h (h.workspace_id)}
        <div class="health-card">
          <div class="health-header">
            <span class="health-name">{h.workspace_name}</span>
            <Badge variant={healthColor(h.recent_fail_rate, h.total_runs_7d)} size="sm">{healthLabel(h.recent_fail_rate, h.total_runs_7d)}</Badge>
          </div>
          <div class="health-stats">
            <div class="health-stat">
              <div class="health-stat-value">{h.member_count}</div>
              <div class="health-stat-label">Members</div>
            </div>
            <div class="health-stat">
              <div class="health-stat-value">{h.project_count}</div>
              <div class="health-stat-label">Projects</div>
            </div>
            <div class="health-stat">
              <div class="health-stat-value">{h.active_runs}</div>
              <div class="health-stat-label">Active runs</div>
            </div>
          </div>
          <div class="fail-bar">
            <div class="fail-bar-fill {healthColor(h.recent_fail_rate, h.total_runs_7d)}"
                 style="width: {h.total_runs_7d > 0 ? Math.min(h.recent_fail_rate * 100, 100) : 0}%"></div>
          </div>
          <div class="health-footer">
            <span>7d fail rate: {h.total_runs_7d > 0 ? `${(h.recent_fail_rate * 100).toFixed(1)}%` : '—'}</span>
            <span>{#if h.last_activity_at}<TimeAgo value={h.last_activity_at} />{:else}no runs{/if}</span>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}

<div class="section">
{#if data.workspaces.length === 0}
  <EmptyState
    icon="◇"
    message="No workspaces yet. Create your first one to get started."
    actionHref="/workspaces/new"
    actionLabel="New workspace" />
{:else}
  <div class="grid">
    {#each data.workspaces as ws (ws.id)}
      <a href={`/workspaces/${ws.id}`} class="card">
        <h3>{ws.name}</h3>
        <p>{ws.description || '—'}</p>
        <div class="meta">created <TimeAgo value={ws.created_at} /></div>
      </a>
    {/each}
  </div>
{/if}
</div>

{#if data.recentRuns.length > 0}
  <div class="section">
    <SectionHeader title="Recent runs" />
    {#each data.recentRuns as r (r.id)}
      <RunRow run={r} />
    {/each}
  </div>
{/if}

{#if comparisonData.length > 0}
  <div class="section">
    <button type="button" class="collapsible-toggle" onclick={() => (comparisonOpen = !comparisonOpen)}>
      <span class="toggle-arrow" class:open={comparisonOpen}>&#9654;</span>
      Workspace comparison
    </button>
    {#if comparisonOpen}
      <div class="days-selector">
        {#each [7, 14, 30] as d}
          <button type="button"
                  class="days-btn"
                  class:active={comparisonDays === d}
                  onclick={() => loadComparison(d)}>
            {d}d
          </button>
        {/each}
        {#if comparisonLoading}
          <span style="font-size: var(--fs-xs); color: var(--text-dim); align-self: center; margin-left: var(--space-2);">Loading...</span>
        {/if}
      </div>
      <Panel padding="none">
        <table class="comp-table">
          <thead>
            <tr>
              <th>Workspace</th>
              <th>Runs / day</th>
              <th>Success rate</th>
              <th>Avg duration</th>
            </tr>
          </thead>
          <tbody>
            {#each comparisonData as c (c.workspace_id)}
              <tr>
                <td>{c.workspace_name}</td>
                <td class="mono">{(c.runs_per_day ?? 0).toFixed(1)}</td>
                <td class="mono">{(c.success_rate ?? 0).toFixed(1)}%</td>
                <td class="mono">{formatDuration(c.avg_duration_ms ?? 0)}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </Panel>
    {/if}
  </div>
{/if}

{#if data.storage.length > 0}
  <div class="section">
    <SectionHeader title="Storage overview" />
    <div class="storage-grid">
      {#each data.storage as s (s.workspace_id)}
        {@const maxTotal = storageMax(data.storage)}
        <div class="storage-card">
          <div class="storage-name">{s.workspace_name}</div>
          <div class="storage-row">
            <span class="storage-label">Projects</span>
            <div class="storage-bar-track">
              <div class="storage-bar-fill bar-projects" style="width: {(s.projects / maxTotal) * 100}%"></div>
            </div>
            <span class="storage-count">{s.projects}</span>
          </div>
          <div class="storage-row">
            <span class="storage-label">Tasks</span>
            <div class="storage-bar-track">
              <div class="storage-bar-fill bar-tasks" style="width: {(s.tasks / maxTotal) * 100}%"></div>
            </div>
            <span class="storage-count">{s.tasks}</span>
          </div>
          <div class="storage-row">
            <span class="storage-label">Runs</span>
            <div class="storage-bar-track">
              <div class="storage-bar-fill bar-runs" style="width: {(s.runs / maxTotal) * 100}%"></div>
            </div>
            <span class="storage-count">{s.runs}</span>
          </div>
          <div class="storage-row">
            <span class="storage-label">Artifacts</span>
            <div class="storage-bar-track">
              <div class="storage-bar-fill bar-artifacts" style="width: {(s.artifacts / maxTotal) * 100}%"></div>
            </div>
            <span class="storage-count">{s.artifacts}</span>
          </div>
        </div>
      {/each}
    </div>
  </div>
{/if}
