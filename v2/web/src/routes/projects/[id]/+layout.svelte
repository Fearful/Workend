<script lang="ts">
  import { goto, invalidateAll } from '$app/navigation';
  import { shortSha as _shortSha } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import ProjectTabs from '$lib/components/ProjectTabs.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, children } = $props();

  function shortSha(sha: string | null): string {
    return _shortSha(sha, 7);
  }

  let deleteModal = $state(false);
  let deletePending = $state(false);
  let actionError = $state<string | null>(null);
  let syncPending = $state(false);
  let actionsMenuOpen = $state(false);

  async function doSync() {
    if (syncPending) return;
    syncPending = true;
    actionError = null;
    try {
      const r = await fetch(`/api/projects/${data.project.id}/sync`, {
        method: 'POST',
        credentials: 'same-origin'
      });
      if (!r.ok) {
        actionError = 'Sync failed';
        return;
      }
      await invalidateAll();
    } finally {
      syncPending = false;
    }
  }

  async function doDelete() {
    if (deletePending) return;
    deletePending = true;
    try {
      const r = await fetch(`/api/projects/${data.project.id}`, {
        method: 'DELETE',
        credentials: 'same-origin'
      });
      if (!r.ok) {
        actionError = 'Delete failed';
        deletePending = false;
        deleteModal = false;
        return;
      }
      const wsID = data.workspace?.id;
      goto(wsID ? `/workspaces/${wsID}` : '/');
    } catch {
      actionError = 'Delete failed';
      deletePending = false;
      deleteModal = false;
    }
  }
</script>

<style>
  .project-shell { padding-bottom: var(--space-6); }

  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
    gap: var(--space-3);
    flex-wrap: wrap;
  }
  h1 {
    font-size: var(--fs-2xl);
    margin: 0;
    display: flex;
    align-items: center;
    gap: var(--space-3);
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
    min-width: 0;
  }
  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    max-width: 60vw;
  }

  .actions-bar {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
    align-items: center;
  }

  .ribbon {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }
  .ribbon-item {
    display: inline-flex;
    align-items: baseline;
    gap: 0.375rem;
  }
  .ribbon-item .label {
    color: var(--text-dim);
    text-transform: uppercase;
    letter-spacing: 0.04em;
    font-size: 0.6875rem;
  }
  .ribbon-item .val {
    color: var(--text);
    font-weight: var(--fw-medium);
  }
  .ribbon-item .mono {
    font-family: var(--font-mono);
  }
  .ribbon-sep {
    color: var(--border-strong);
    user-select: none;
  }

  .commit-tip { display: flex; flex-direction: column; gap: 0.25rem; max-width: 320px; }
  .commit-tip .author { color: var(--text-muted); font-size: var(--fs-xs); }
  .commit-tip .msg { color: var(--text); font-size: var(--fs-xs); line-height: var(--lh-normal); }

  /* mobile compact */
  @media (max-width: 768px) {
    .ribbon { gap: var(--space-2); font-size: var(--fs-xs); }
    .ribbon-sep { display: none; }
    .ribbon-item .label { display: none; }
    h1 { font-size: var(--fs-xl); }
    .name { max-width: 50vw; }
    .actions-bar .desktop-only { display: none; }
  }
  @media (min-width: 769px) {
    .actions-bar .mobile-only { display: none; }
  }

  .more-menu-wrap { position: relative; }
  .more-btn {
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    padding: 0.4rem 0.6rem;
    border-radius: var(--radius-md);
    cursor: pointer;
    font-size: var(--fs-md);
    line-height: 1;
  }
  .more-btn:hover { background: var(--bg-hover); color: var(--text); }
  .more-menu {
    position: absolute;
    right: 0;
    top: calc(100% + 4px);
    background: var(--bg-panel);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    padding: 0.25rem;
    min-width: 200px;
    box-shadow: var(--shadow-popover);
    z-index: 50;
    display: flex;
    flex-direction: column;
  }
  .more-menu button {
    background: transparent;
    color: var(--text);
    border: none;
    text-align: left;
    padding: 0.5rem 0.625rem;
    font-size: var(--fs-sm);
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-family: inherit;
    width: 100%;
  }
  .more-menu button:hover { background: var(--bg-hover); }
  .more-menu .danger { color: var(--danger-text); }

  .delete-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
  }
</style>

<div class="project-shell">
  <Breadcrumb segments={[
    { label: 'workspaces', href: '/' },
    ...(data.workspace ? [{ label: data.workspace.name, href: `/workspaces/${data.workspace.id}` }] : []),
    { label: data.project.name }
  ]} />

  <div class="header-row">
    <h1>
      <StatusPill status={data.project.status} />
      <span class="name">{data.project.name}</span>
    </h1>
    <div class="actions-bar">
      <button type="button" onclick={doSync} disabled={syncPending} class="desktop-only">
        {syncPending ? 'Syncing…' : 'Sync'}
      </button>
      <div class="more-menu-wrap">
        <button type="button" class="more-btn"
                onclick={(e) => { e.stopPropagation(); actionsMenuOpen = !actionsMenuOpen; }}
                aria-haspopup="menu"
                aria-expanded={actionsMenuOpen}
                title="Project actions">⋯</button>
        {#if actionsMenuOpen}
          <div class="more-menu" role="menu" tabindex="-1"
               onkeydown={(e) => { if (e.key === 'Escape') actionsMenuOpen = false; }}>
            <button class="mobile-only" onclick={doSync} disabled={syncPending}>
              {syncPending ? 'Syncing…' : 'Sync now'}
            </button>
            <button onclick={() => { window.open(data.project.git_url, '_blank'); }}>Open repo in new tab</button>
            <button class="danger" onclick={() => deleteModal = true}>Delete project…</button>
          </div>
        {/if}
      </div>
    </div>
  </div>

  <div class="ribbon">
    <span class="ribbon-item">
      <span class="label">Branch</span>
      <Tooltip text={data.project.default_branch || 'No default branch detected'}>
        <span class="val mono">{data.project.default_branch || '—'}</span>
      </Tooltip>
    </span>
    <span class="ribbon-sep">·</span>
    {#if data.project.last_commit_sha}
      <span class="ribbon-item">
        <span class="label">Commit</span>
        <Tooltip placement="bottom">
          {#snippet content()}
            <div class="commit-tip">
              <div class="mono">{data.project.last_commit_sha}</div>
              {#if data.project.last_commit_author}
                <div class="author">{data.project.last_commit_author}</div>
              {/if}
              {#if data.project.last_commit_message}
                <div class="msg">{data.project.last_commit_message}</div>
              {/if}
            </div>
          {/snippet}
          <span class="val mono">{shortSha(data.project.last_commit_sha)}</span>
        </Tooltip>
      </span>
      <span class="ribbon-sep">·</span>
    {/if}
    {#if data.project.last_synced_at}
      <span class="ribbon-item">
        <span class="label">Synced</span>
        <span class="val"><TimeAgo value={data.project.last_synced_at} /></span>
      </span>
    {/if}
  </div>

  <ProjectTabs projectID={data.project.id} />

  {#if actionError}<FlashMessage type="error">{actionError}</FlashMessage>{/if}

  {@render children()}
</div>

<Modal open={deleteModal} title="Delete project?" width={420} onClose={() => deleteModal = false}>
  <p style="margin: 0 0 var(--space-3) 0;">
    This will permanently delete <strong>{data.project.name}</strong>, its tasks, run history, and any configured schedules.
    The remote git repository is unaffected.
  </p>
  <div class="delete-actions">
    <button type="button" class="ghost" onclick={() => (deleteModal = false)} disabled={deletePending}>Cancel</button>
    <button type="button" class="danger" onclick={doDelete} disabled={deletePending}>
      {deletePending ? 'Deleting…' : 'Delete project'}
    </button>
  </div>
</Modal>

<svelte:window onclick={() => actionsMenuOpen = false} />
