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
    if (action.startsWith('workspace.')) return '◯';
    if (action.startsWith('user.')) return '☉';
    return '·';
  }

  function eventHref(e: { action: string; target_kind: string; target_id: string }): string | null {
    if (e.target_kind === 'run' && e.target_id) return `/runs/${e.target_id}`;
    if (e.target_kind === 'project' && e.target_id) return `/projects/${e.target_id}`;
    return null;
  }
</script>

<style>
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: var(--space-6);
    gap: var(--space-4);
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

  .desc { color: var(--text-muted); font-size: var(--fs-md); margin: 0; }
  .actions { display: flex; gap: var(--space-2); flex-shrink: 0; flex-wrap: wrap; }

  .ws-tabs {
    border-bottom: 1px solid var(--border);
    display: flex;
    gap: 0;
    margin: var(--space-3) 0 var(--space-5);
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

  @media (max-width: 768px) {
    .member-row { grid-template-columns: 1fr auto auto; }
    .member-email { display: none; }
    .add-member-grid { grid-template-columns: 1fr; }
    .issue-row { grid-template-columns: auto 1fr; }
    .issue-row > :nth-child(3),
    .issue-row > :nth-child(4) { display: none; }
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name }
]} />

<div class="header-row">
  <div>
    <h1>{data.workspace.name} <span class="role-tag">{data.workspace.my_role}</span></h1>
    <p class="desc">{data.workspace.description || 'No description'}</p>
  </div>
  <div class="actions">
    <a href={`/workspaces/${data.workspace.id}/projects/new`}><button>Add project</button></a>
    {#if data.workspace.my_role === 'owner'}
      <form method="POST" action="?/delete" class="inline-form" onsubmit={(e) => !confirm('Delete this workspace and all its projects?') && e.preventDefault()}>
        <button type="submit" class="ghost">Delete workspace</button>
      </form>
    {/if}
  </div>
</div>

<nav class="ws-tabs" aria-label="Workspace sections">
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
</nav>

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
              <span class="sep">·</span>
              <Tooltip text={p.last_commit_sha}><span class="mono">{shortSha(p.last_commit_sha)}</span></Tooltip>
            {/if}
            {#if p.last_synced_at}
              <span class="sep">·</span>
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
    <SectionHeader title="Activity" />
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
