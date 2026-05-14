<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data, form } = $props();

  let newBranch = $state('');
  let actionPending = $state<string | null>(null);
  let deleteConfirmID = $state<string | null>(null);

  function previewStatus(status: string): string {
    switch (status) {
      case 'deployed': return 'ready';
      case 'deploying': return 'cloning';
      case 'stopped': return 'cancelled';
      case 'failed': return 'failed';
      default: return 'pending';
    }
  }

  async function redeployPreview(previewID: string) {
    actionPending = previewID;
    try {
      await fetch(`/api/previews/${previewID}/redeploy`, {
        method: 'POST',
        credentials: 'same-origin'
      });
      await invalidateAll();
    } finally {
      actionPending = null;
    }
  }

  async function stopPreview(previewID: string) {
    actionPending = previewID;
    try {
      await fetch(`/api/previews/${previewID}/stop`, {
        method: 'POST',
        credentials: 'same-origin'
      });
      await invalidateAll();
    } finally {
      actionPending = null;
    }
  }

  async function toggleAutoDeploy(previewID: string, current: boolean) {
    await fetch(`/api/previews/${previewID}/auto-deploy`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: !current }),
      credentials: 'same-origin'
    });
    await invalidateAll();
  }
</script>

<style>
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-4);
  }
  h2 {
    margin: 0;
    font-size: var(--fs-lg);
    font-weight: var(--fw-semibold);
  }

  .create-form {
    display: flex;
    gap: var(--space-2);
    margin-bottom: var(--space-4);
    padding-bottom: var(--space-4);
    border-bottom: 1px solid var(--border);
  }
  .branch-input {
    flex: 1;
    max-width: 300px;
    font-size: 0.875rem;
  }

  .preview-card {
    display: grid;
    grid-template-columns: auto 1fr auto;
    align-items: start;
    gap: var(--space-3);
    padding: var(--space-4);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    margin-bottom: var(--space-3);
    background: var(--bg-panel);
    transition: border-color 80ms ease;
  }
  .preview-card:hover {
    border-color: var(--border-strong);
  }

  .card-body { display: flex; flex-direction: column; gap: var(--space-1); min-width: 0; }
  .card-branch {
    display: flex;
    align-items: center;
    gap: var(--space-2);
  }
  .card-url {
    font-family: var(--font-mono);
    font-size: 0.8125rem;
    color: var(--link);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .card-meta {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    font-size: 0.75rem;
    color: var(--text-dim);
  }

  .card-actions {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
    align-items: flex-end;
  }
  .card-actions button { font-size: 0.8125rem; }

  .auto-deploy-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    cursor: pointer;
    font-size: 0.8125rem;
    color: var(--text-muted);
  }
  .auto-deploy-toggle input { margin: 0; }

  .delete-confirm {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }

  .empty {
    text-align: center;
    padding: 3rem var(--space-4);
    color: var(--text-muted);
  }

  @media (max-width: 768px) {
    .preview-card {
      grid-template-columns: 1fr;
    }
    .card-actions { flex-direction: row; flex-wrap: wrap; }
  }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<div class="page-header">
  <h2>Deploy previews</h2>
</div>

<form method="POST" action="?/createPreview" class="create-form">
  <input type="text" name="branch" class="branch-input" placeholder="Branch name (e.g. feature/new-ui)" bind:value={newBranch} />
  <button type="submit" disabled={!newBranch.trim()}>Create preview</button>
</form>

{#if data.previews.length === 0}
  <div class="empty">
    <p>No deploy previews yet. Create one by entering a branch name above.</p>
  </div>
{:else}
  {#each data.previews as p (p.id)}
    <div class="preview-card">
      <StatusPill status={previewStatus(p.status)} size="sm" />
      <div class="card-body">
        <div class="card-branch">
          <Badge variant="info" size="sm">{p.branch}</Badge>
          <span style="font-size: 0.75rem; color: var(--text-dim);">{p.status}</span>
        </div>
        {#if p.url}
          <a href={p.url} target="_blank" rel="noopener noreferrer" class="card-url">{p.url}</a>
        {/if}
        <div class="card-meta">
          {#if p.last_deployed_at}
            <span>Deployed <TimeAgo value={p.last_deployed_at} /></span>
          {/if}
          <span>Created <TimeAgo value={p.created_at} /></span>
          <label class="auto-deploy-toggle">
            <input type="checkbox"
                   checked={p.auto_deploy}
                   onchange={() => toggleAutoDeploy(p.id, p.auto_deploy)} />
            Auto-deploy
          </label>
        </div>
      </div>
      <div class="card-actions">
        <button type="button" class="ghost"
                onclick={() => redeployPreview(p.id)}
                disabled={actionPending === p.id}>
          {actionPending === p.id ? 'Working...' : 'Redeploy'}
        </button>
        {#if p.status === 'deployed' || p.status === 'deploying'}
          <button type="button" class="ghost"
                  onclick={() => stopPreview(p.id)}
                  disabled={actionPending === p.id}>Stop</button>
        {/if}
        {#if deleteConfirmID === p.id}
          <div class="delete-confirm">
            <form method="POST" action="?/deletePreview" class="inline-form">
              <input type="hidden" name="preview_id" value={p.id} />
              <button type="submit" class="danger" style="font-size: 0.8125rem;">Confirm</button>
            </form>
            <button type="button" class="ghost" style="font-size: 0.8125rem;"
                    onclick={() => (deleteConfirmID = null)}>Cancel</button>
          </div>
        {:else}
          <button type="button" class="ghost danger" onclick={() => (deleteConfirmID = p.id)}>Delete</button>
        {/if}
      </div>
    </div>
  {/each}
{/if}
