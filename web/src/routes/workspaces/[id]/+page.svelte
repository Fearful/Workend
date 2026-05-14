<script lang="ts">
  import { shortSha, formatRelative } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';

  let { data, form } = $props();

  function actionLabel(action: string): string {
    switch (action) {
      case 'workspace.create': return 'created the workspace';
      case 'project.create':   return 'added a project';
      case 'project.delete':   return 'deleted a project';
      case 'project.sync':     return 'synced a project';
      case 'run.start':        return 'started a run';
      case 'run.cancel':       return 'cancelled a run';
      case 'run.complete':     return 'completed a run';
      case 'user.login':       return 'signed in';
      case 'user.signup':      return 'signed up';
      default:                 return action.replace(/\./g, ' ');
    }
  }

  function actionIcon(action: string): string {
    if (action.startsWith('run.')) return '▶';
    if (action.startsWith('project.')) return '◇';
    if (action.startsWith('workspace.')) return '○';
    if (action.startsWith('user.')) return '☉';
    return '·';
  }

  function eventHref(e: { action: string; target_kind: string; target_id: string }): string | null {
    if (e.target_kind === 'run' && e.target_id) return `/runs/${e.target_id}`;
    if (e.target_kind === 'project' && e.target_id) return `/projects/${e.target_id}`;
    return null;
  }

  function pipelineStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'info' {
    switch (status) {
      case 'passed': case 'succeeded': return 'success';
      case 'running': case 'queued': return 'warning';
      case 'failed': return 'danger';
      default: return 'muted';
    }
  }

  function sandboxStatusVariant(status: string): 'success' | 'warning' | 'danger' | 'muted' | 'info' {
    switch (status) {
      case 'running': return 'success';
      case 'provisioning': return 'warning';
      case 'stopped': case 'expired': return 'muted';
      case 'error': return 'danger';
      default: return 'info';
    }
  }
</script>

<style>
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }

  h1 {
    font-size: var(--fs-2xl);
    margin: 0 0 0.25rem 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
  }
  .role-tag {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-left: var(--space-2);
    font-weight: var(--fw-regular);
  }

  .sub-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
  .desc { color: var(--text-muted); font-size: var(--fs-md); margin: 0; }
  .actions { display: flex; gap: var(--space-2); flex-shrink: 0; flex-wrap: wrap; }

  .ws-tabs {
    border-bottom: 1px solid var(--border);
    display: flex;
    gap: 0;
    margin-bottom: var(--space-5);
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

  .section { margin-top: var(--space-6); }

  /* --- Widget Grid --- */
  .widget-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(320px, 1fr));
    gap: var(--space-4);
    margin-bottom: var(--space-6);
  }
  .widget {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
    min-height: 140px;
  }
  .widget-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-3);
    gap: var(--space-2);
  }
  .widget-title {
    margin: 0;
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
    letter-spacing: -0.005em;
  }
  .widget-link {
    font-size: var(--fs-xs);
    color: var(--link);
  }
  .widget-link:hover { text-decoration: underline; }

  /* --- Project Grid Widget --- */
  .proj-card-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .proj-card {
    display: grid;
    grid-template-columns: 4px 1fr auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    text-decoration: none;
    color: inherit;
    transition: border-color 120ms ease;
  }
  .proj-card:hover {
    border-color: var(--border-strong);
    text-decoration: none;
  }
  .proj-color-bar {
    width: 4px;
    height: 28px;
    border-radius: 2px;
  }
  .proj-card-name {
    font-size: var(--fs-sm);
    font-weight: var(--fw-medium);
  }
  .proj-card-sub {
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }
  .proj-card-sparkline {
    display: inline-flex;
    align-items: center;
    gap: 2px;
  }
  .spark-dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  /* --- Team Presence Widget --- */
  .team-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .team-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-1) 0;
    font-size: var(--fs-sm);
  }
  .presence-dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
    flex-shrink: 0;
  }
  .presence-dot.active { background: var(--success); }
  .presence-dot.inactive { background: var(--text-dim); opacity: 0.4; }
  .team-name { color: var(--text); }
  .team-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
    text-align: right;
  }

  /* --- Stats Widget (Secrets Summary / Deps) --- */
  .stat-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-3);
  }
  .stat-cell { min-width: 0; }
  .stat-num {
    font-size: var(--fs-xl);
    font-weight: var(--fw-semibold);
    font-family: var(--font-mono);
    color: var(--text);
    line-height: 1.1;
  }
  .stat-label {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    margin-top: 2px;
  }

  /* --- Pipeline Board Widget --- */
  .pipe-stats {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: var(--space-2);
    margin-bottom: var(--space-3);
  }
  .pipe-stat {
    text-align: center;
    padding: var(--space-2);
    border-radius: var(--radius-md);
    background: var(--bg-page);
  }
  .pipe-stat-num {
    font-size: var(--fs-lg);
    font-weight: var(--fw-semibold);
    font-family: var(--font-mono);
  }
  .pipe-stat-label {
    font-size: var(--fs-xs);
    color: var(--text-dim);
    text-transform: uppercase;
  }
  .pipe-run-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  .pipe-run {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: var(--space-2);
    align-items: center;
    font-size: var(--fs-xs);
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--border);
  }
  .pipe-run:last-child { border-bottom: none; }
  .pipe-run-name {
    font-family: var(--font-mono);
    color: var(--text);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pipe-run-time {
    color: var(--text-dim);
  }

  /* --- Sandbox Widget --- */
  .sandbox-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
  }
  .sandbox-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-3);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    font-size: var(--fs-sm);
  }
  .sandbox-name { font-weight: var(--fw-medium); }
  .sandbox-branch {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }
  .sandbox-ttl {
    font-size: var(--fs-xs);
    color: var(--text-dim);
    font-family: var(--font-mono);
  }

  /* --- Deps Top Connected --- */
  .dep-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    margin-top: var(--space-3);
  }
  .dep-row {
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: var(--fs-sm);
    padding: var(--space-1) 0;
    border-bottom: 1px solid var(--border);
  }
  .dep-row:last-child { border-bottom: none; }
  .dep-edges {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--text-dim);
  }

  /* --- Existing Styles --- */
  .project-row {
    display: grid;
    grid-template-columns: auto 1fr;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    text-decoration: none;
    color: inherit;
    margin-bottom: var(--space-2);
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .project-row:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }
  .project-name {
    font-size: var(--fs-md);
    font-weight: var(--fw-medium);
    color: var(--text);
  }
  .project-sub {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-top: 2px;
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
  }
  .project-sub .mono { font-family: var(--font-mono); }
  .project-sub .sep { opacity: 0.5; }

  .list-card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-2) var(--space-4);
  }

  .member-row {
    display: grid;
    grid-template-columns: 1fr 1fr auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .member-row:last-child { border-bottom: none; }
  .member-name { color: var(--text); }
  .member-email { color: var(--text-dim); font-family: var(--font-mono); font-size: 0.75rem; }
  .role-label {
    font-size: 0.6875rem;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }
  .role-owner { color: var(--link); }
  .role-member { color: var(--text-muted); }

  .add-member-form {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4) var(--space-5);
    margin-top: var(--space-2);
  }
  .add-member-grid {
    display: grid;
    grid-template-columns: 1fr 140px auto;
    gap: var(--space-3);
    align-items: end;
    margin: 0;
  }
  .field-tight { margin: 0; }
  .role-select {
    padding: 0.5rem 0.75rem;
    background: var(--bg-panel);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font: inherit;
    font-size: 0.875rem;
    width: 100%;
  }

  .issue-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
    text-decoration: none;
    color: inherit;
  }
  .issue-row:last-child { border-bottom: none; }
  .issue-row:hover { background: var(--bg-hover); text-decoration: none; }
  .issue-num { font-weight: 500; }
  .issue-labels { display: inline-flex; gap: 0.25rem; margin-left: 0.25rem; vertical-align: middle; }
  .issue-project,
  .issue-time {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .activity-row {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 0.875rem;
    align-items: center;
    padding: var(--space-2) var(--space-4);
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .activity-row:last-child { border-bottom: none; }
  .activity-icon {
    color: var(--text-dim);
    width: 1.25rem;
    text-align: center;
  }
  .activity-actor { font-weight: 500; }
  .activity-action { color: var(--text-muted); }
  .activity-target {
    margin-left: 0.375rem;
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--link);
  }
  .activity-time { color: var(--text-dim); font-size: 0.75rem; }

  .widget-empty {
    color: var(--text-dim);
    font-size: var(--fs-sm);
    padding: var(--space-4) 0;
    text-align: center;
  }

  @media (max-width: 768px) {
    .member-row { grid-template-columns: 1fr auto auto; }
    .member-email { display: none; }
    .add-member-grid { grid-template-columns: 1fr; }
    .issue-row { grid-template-columns: auto 1fr; }
    .issue-row > :nth-child(3),
    .issue-row > :nth-child(4) { display: none; }
    .widget-grid { grid-template-columns: 1fr; }
    .pipe-stats { grid-template-columns: repeat(2, 1fr); }
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name }
]} />

<div class="header-row">
  <h1>{data.workspace.name} <span class="role-tag">{data.workspace.my_role}</span></h1>
</div>

<nav class="ws-tabs" aria-label="Workspace sections">
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/activity`}>Activity</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/secrets`}>Secrets</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/roles`}>Roles</a>
</nav>

<div class="sub-header">
  <p class="desc">{data.workspace.description || 'No description'}</p>
  <div class="actions">
    <a href={`/workspaces/${data.workspace.id}/projects/new`}><button>Add project</button></a>
    {#if data.workspace.my_role === 'owner'}
      <form method="POST" action="?/delete" class="inline-form" onsubmit={(e) => !confirm('Delete this workspace and all its projects?') && e.preventDefault()}>
        <button type="submit" class="ghost">Delete workspace</button>
      </form>
    {/if}
  </div>
</div>

<!-- Widget Grid -->
<div class="widget-grid">
  <!-- Project Grid Widget -->
  <div class="widget">
    <div class="widget-head">
      <h3 class="widget-title">Projects</h3>
      {#if data.projectGrid.length > 0}
        <span class="widget-link" style="color: var(--text-dim); font-size: var(--fs-xs);">{data.projectGrid.length} total</span>
      {/if}
    </div>
    {#if data.projectGrid.length === 0}
      <div class="widget-empty">No projects yet</div>
    {:else}
      <div class="proj-card-list">
        {#each data.projectGrid.slice(0, 6) as p (p.project_id)}
          <a href={`/projects/${p.project_id}`} class="proj-card">
            <div class="proj-color-bar" style="background: {p.status_color || 'var(--text-dim)'};"></div>
            <div>
              <div class="proj-card-name">{p.project_name}</div>
              <div class="proj-card-sub">{p.task_count} tasks</div>
            </div>
            <div class="proj-card-sparkline">
              {#if p.recent_passes > 0 || p.recent_fails > 0}
                <span class="spark-count" style="color: var(--status-success-fg);">{p.recent_passes}</span>
                <span class="spark-sep">/</span>
                <span class="spark-count" style="color: var(--status-danger-fg);">{p.recent_fails}</span>
              {/if}
            </div>
          </a>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Team Presence Widget -->
  <div class="widget">
    <div class="widget-head">
      <h3 class="widget-title">Team</h3>
      {#if data.teamPresence.length > 0}
        {@const activeCount = data.teamPresence.filter(m => m.is_active).length}
        <span style="color: var(--text-dim); font-size: var(--fs-xs);">{activeCount} active</span>
      {/if}
    </div>
    {#if data.teamPresence.length === 0}
      <div class="widget-empty">No team members</div>
    {:else}
      <div class="team-list">
        {#each data.teamPresence.slice(0, 8) as m (m.user_id)}
          <div class="team-row">
            <span class="presence-dot" class:active={m.is_active} class:inactive={!m.is_active}></span>
            <span class="team-name">
              {m.display_name}
              {#if m.active_runs > 0}
                <Badge variant="info" size="sm">{m.active_runs} running</Badge>
              {/if}
            </span>
            <span class="team-meta">
              {#if m.last_active_at}<TimeAgo value={m.last_active_at} />{:else}--{/if}
            </span>
          </div>
        {/each}
      </div>
    {/if}
  </div>

  <!-- Pipeline Board Widget -->
  {#if data.pipelineBoard}
    <div class="widget">
      <div class="widget-head">
        <h3 class="widget-title">Pipelines</h3>
      </div>
      <div class="pipe-stats">
        <div class="pipe-stat">
          <div class="pipe-stat-num" style="color: var(--text-dim);">{data.pipelineBoard.queued.length}</div>
          <div class="pipe-stat-label">Queued</div>
        </div>
        <div class="pipe-stat">
          <div class="pipe-stat-num" style="color: var(--status-info-fg);">{data.pipelineBoard.running.length}</div>
          <div class="pipe-stat-label">Running</div>
        </div>
        <div class="pipe-stat">
          <div class="pipe-stat-num" style="color: var(--status-success-fg);">{data.pipelineBoard.passed.length}</div>
          <div class="pipe-stat-label">Passed</div>
        </div>
        <div class="pipe-stat">
          <div class="pipe-stat-num" style="color: var(--status-danger-fg);">{data.pipelineBoard.failed.length}</div>
          <div class="pipe-stat-label">Failed</div>
        </div>
      </div>
      {#if [...data.pipelineBoard.running, ...data.pipelineBoard.queued, ...data.pipelineBoard.passed, ...data.pipelineBoard.failed].length > 0}
        <div class="pipe-run-list">
          {#each [...data.pipelineBoard.running, ...data.pipelineBoard.queued, ...data.pipelineBoard.passed, ...data.pipelineBoard.failed].slice(0, 5) as run (run.pipeline_run_id)}
            <div class="pipe-run">
              <Badge variant={pipelineStatusVariant(run.status)} size="sm">{run.status}</Badge>
              <span class="pipe-run-name">{run.pipeline_name} / {run.project_name}</span>
              <span class="pipe-run-time"><TimeAgo value={run.started_at} /></span>
            </div>
          {/each}
        </div>
      {:else}
        <div class="widget-empty">No recent runs</div>
      {/if}
    </div>
  {/if}

  <!-- Secrets Summary Widget -->
  {#if data.secretsSummary}
    <div class="widget">
      <div class="widget-head">
        <h3 class="widget-title">Secrets</h3>
        <a href={`/workspaces/${data.workspace.id}/secrets`} class="widget-link">Manage</a>
      </div>
      <div class="stat-grid">
        <div class="stat-cell">
          <div class="stat-num">{data.secretsSummary.total_secrets}</div>
          <div class="stat-label">Total</div>
        </div>
        <div class="stat-cell">
          <div class="stat-num">{data.secretsSummary.creator_count}</div>
          <div class="stat-label">Contributors</div>
        </div>
        <div class="stat-cell">
          <div class="stat-label" style="margin-top: 0;">Newest</div>
          <div style="font-size: var(--fs-sm);">
            {#if data.secretsSummary.newest_updated_at}<TimeAgo value={data.secretsSummary.newest_updated_at} />{:else}--{/if}
          </div>
        </div>
        <div class="stat-cell">
          <div class="stat-label" style="margin-top: 0;">Oldest</div>
          <div style="font-size: var(--fs-sm);">
            {#if data.secretsSummary.oldest_updated_at}<TimeAgo value={data.secretsSummary.oldest_updated_at} />{:else}--{/if}
          </div>
        </div>
      </div>
    </div>
  {/if}

  <!-- Dependency Overview Widget -->
  {#if data.depOverview}
    <div class="widget">
      <div class="widget-head">
        <h3 class="widget-title">Dependencies</h3>
      </div>
      <div class="stat-grid">
        <div class="stat-cell">
          <div class="stat-num">{data.depOverview.total_projects}</div>
          <div class="stat-label">Projects</div>
        </div>
        <div class="stat-cell">
          <div class="stat-num">{data.depOverview.total_deps}</div>
          <div class="stat-label">Edges</div>
        </div>
        <div class="stat-cell">
          <div class="stat-num" style="color: {data.depOverview.unresolved_deps > 0 ? 'var(--status-danger-fg)' : 'var(--text)'};">{data.depOverview.unresolved_deps}</div>
          <div class="stat-label">Unresolved</div>
        </div>
        <div class="stat-cell"></div>
      </div>
      {#if data.depOverview.most_connected.length > 0}
        <div class="dep-list">
          {#each data.depOverview.most_connected.slice(0, 4) as dep (dep.project_id)}
            <div class="dep-row">
              <a href={`/projects/${dep.project_id}`} style="color: var(--text); font-size: var(--fs-sm);">{dep.project_name}</a>
              <span class="dep-edges">{dep.inbound_deps + dep.outbound_deps} edges</span>
            </div>
          {/each}
        </div>
      {/if}
    </div>
  {/if}

  <!-- Sandbox Preview Rack Widget -->
  {#if data.sandboxItems.length > 0}
    <div class="widget">
      <div class="widget-head">
        <h3 class="widget-title">Sandboxes</h3>
        <span style="color: var(--text-dim); font-size: var(--fs-xs);">{data.sandboxItems.length} active</span>
      </div>
      <div class="sandbox-list">
        {#each data.sandboxItems.slice(0, 5) as item (item.id)}
          <div class="sandbox-row">
            <Badge variant={sandboxStatusVariant(item.status)} size="sm">{item.status}</Badge>
            <div>
              <div class="sandbox-name">{item.project_name}</div>
              <div class="sandbox-branch">{item.branch}</div>
            </div>
            <span class="sandbox-ttl">
              {#if item.minutes_left != null}{item.minutes_left}m left{:else}{item.type}{/if}
            </span>
          </div>
        {/each}
      </div>
    </div>
  {/if}
</div>

<!-- Existing sections below widgets -->
<div class="section">
  <SectionHeader title="Projects" />

  {#if data.projectsError}
    <FlashMessage type="error">{data.projectsError}</FlashMessage>
  {/if}

  {#if data.projects.length === 0}
    <EmptyState message="No projects yet. Add one by git URL."
                actionHref={`/workspaces/${data.workspace.id}/projects/new`}
                actionLabel="Add project" />
  {:else}
    {#each data.projects as p (p.id)}
      <a href={`/projects/${p.id}`} class="project-row">
        <StatusPill status={p.status} size="sm" />
        <div>
          <div class="project-name">{p.name}</div>
          <div class="project-sub">
            <span class="mono">{p.git_url.replace(/^https?:\/\//, '').replace(/\.git$/, '')}</span>
            {#if p.last_commit_sha}
              <span class="sep">&middot;</span>
              <Tooltip text={p.last_commit_sha}><span class="mono">{shortSha(p.last_commit_sha)}</span></Tooltip>
            {/if}
            {#if p.last_synced_at}
              <span class="sep">&middot;</span>
              <span>synced <TimeAgo value={p.last_synced_at} /></span>
            {/if}
          </div>
        </div>
      </a>
    {/each}
  {/if}
</div>

<div class="section">
  <SectionHeader title="Members" />
<div class="list-card">
  {#each data.members as m (m.user_id)}
    <div class="member-row">
      <span class="member-name">{m.display_name}</span>
      <span class="member-email">{m.email}</span>
      <span class="role-label {m.role === 'owner' ? 'role-owner' : 'role-member'}">{m.role}</span>
      {#if data.workspace.my_role === 'owner' || m.user_id === data.user?.id}
        <form method="POST" action="?/removeMember" class="inline-form" onsubmit={(e) => !confirm(m.user_id === data.user?.id ? 'Leave this workspace?' : `Remove ${m.display_name}?`) && e.preventDefault()}>
          <input type="hidden" name="user_id" value={m.user_id} />
          <button type="submit" class="ghost">{m.user_id === data.user?.id ? 'Leave' : 'Remove'}</button>
        </form>
      {:else}
        <span></span>
      {/if}
    </div>
  {/each}
</div>

  {#if data.workspace.my_role === 'owner'}
    <div class="add-member-form">
      <form method="POST" action="?/addMember" class="add-member-grid">
        <div class="field field-tight">
          <label for="email">Add member by email</label>
          <input id="email" name="email" type="email" required placeholder="alice@example.com" value={form?.email || ''} />
        </div>
        <div class="field field-tight">
          <label for="role">Role</label>
          <select id="role" name="role" class="role-select">
            <option value="member">member</option>
            <option value="owner">owner</option>
          </select>
        </div>
        <button type="submit">Add</button>
      </form>
      {#if form?.memberError}<FlashMessage type="error">{form.memberError}</FlashMessage>{/if}
    </div>
  {/if}
</div>

{#if data.recentIssues.length > 0}
  <div class="section">
    <SectionHeader title="Recent issues" />
    <div class="list-card">
      {#each data.recentIssues as issue (issue.id)}
        <a href={issue.html_url} target="_blank" rel="noopener noreferrer" class="issue-row">
          <span class="dot" style="background: {issue.state === 'open' ? 'var(--success)' : 'var(--text-dim)'}"></span>
          <div>
            <span class="issue-num">#{issue.provider_number}</span>
            <span>{issue.title}</span>
            {#if issue.labels.length > 0}
              <span class="issue-labels">
                {#each issue.labels as label}
                  <Badge variant="info" size="sm">{label}</Badge>
                {/each}
              </span>
            {/if}
          </div>
          <span class="issue-project">{issue.project_name}</span>
          <span class="issue-time">{issue.upstream_updated_at ? formatRelative(issue.upstream_updated_at) : ''}</span>
        </a>
      {/each}
    </div>
  </div>
{/if}

{#if data.activity.length > 0}
  <div class="section">
    <SectionHeader title="Activity">
      {#snippet actions()}
        <a href={`/workspaces/${data.workspace.id}/activity`} class="widget-link">View all</a>
      {/snippet}
    </SectionHeader>
    <div class="list-card">
      {#each data.activity as e (e.id)}
        {@const href = eventHref(e)}
        <div class="activity-row">
          <span class="activity-icon">{actionIcon(e.action)}</span>
          <span>
            <strong class="activity-actor">{e.actor_name || 'someone'}</strong>
            <span class="activity-action"> {actionLabel(e.action)}</span>
            {#if href}
              <a href={href} class="activity-target">{e.target_id.slice(0, 8)}</a>
            {/if}
          </span>
          <span class="activity-time">{formatRelative(e.occurred_at)}</span>
        </div>
      {/each}
    </div>
  </div>
{/if}
