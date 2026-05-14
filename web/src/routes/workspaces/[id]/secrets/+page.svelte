<script lang="ts">
  import { formatRelative } from '$lib/utils';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data, form } = $props();

  let editingSecretId = $state<string | null>(null);
  let showAddForm = $state(false);
  let confirmDeleteId = $state<string | null>(null);

  function startEdit(id: string) {
    editingSecretId = id;
  }

  function cancelEdit() {
    editingSecretId = null;
  }

  function confirmDelete(id: string) {
    confirmDeleteId = id;
  }

  function cancelDelete() {
    confirmDeleteId = null;
  }

  let deletingSecret = $derived(
    confirmDeleteId ? data.secrets.find((s: { id: string }) => s.id === confirmDeleteId) : null
  );

  // Reset UI states after successful actions
  $effect(() => {
    if (form?.created) showAddForm = false;
    if (form?.updated) editingSecretId = null;
    if (form?.deleted) confirmDeleteId = null;
  });
</script>

<style>
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
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
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

  .toolbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: var(--space-4);
  }
  .toolbar-info {
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }

  .secret-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .secret-row {
    display: grid;
    grid-template-columns: 1fr auto auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-5);
    border-bottom: 1px solid var(--border);
    font-size: var(--fs-sm);
  }
  .secret-row:last-child { border-bottom: none; }
  .secret-row:hover { background: var(--bg-hover); }

  .secret-name-col { min-width: 0; }
  .secret-name {
    font-weight: var(--fw-medium);
    font-family: var(--font-mono);
    color: var(--text);
    font-size: var(--fs-sm);
  }
  .secret-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    margin-top: 2px;
  }

  .secret-value-mask {
    font-family: var(--font-mono);
    color: var(--text-dim);
    font-size: var(--fs-xs);
    letter-spacing: 0.1em;
  }

  .secret-actions {
    display: flex;
    gap: var(--space-1);
  }

  .add-form {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
    margin-bottom: var(--space-4);
  }
  .add-form h3 {
    margin: 0 0 var(--space-3);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
  }
  .form-grid {
    display: grid;
    grid-template-columns: 1fr 2fr;
    gap: var(--space-3);
    align-items: start;
  }
  .form-grid .field { margin: 0; }
  .form-grid textarea {
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
    resize: vertical;
    min-height: 60px;
  }
  .form-footer {
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
    margin-top: var(--space-3);
  }

  /* Inline edit form */
  .edit-form {
    grid-column: 1 / -1;
    padding: var(--space-3) var(--space-5);
    background: var(--bg-page);
    border-top: 1px solid var(--border);
  }
  .edit-form-inner {
    display: flex;
    gap: var(--space-3);
    align-items: end;
  }
  .edit-form-inner .field {
    margin: 0;
    flex: 1;
  }
  .edit-form-inner textarea {
    font-family: var(--font-mono);
    font-size: var(--fs-sm);
    resize: vertical;
    min-height: 48px;
    width: 100%;
  }

  @media (max-width: 768px) {
    .secret-row { grid-template-columns: 1fr auto; }
    .secret-value-mask { display: none; }
    .form-grid { grid-template-columns: 1fr; }
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: data.workspace.name, href: `/workspaces/${data.workspace.id}` },
  { label: 'secrets' }
]} />

<div class="header-row">
  <h1>{data.workspace.name} <span class="role-tag">{data.workspace.my_role}</span></h1>
</div>

<nav class="ws-tabs" aria-label="Workspace sections">
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}`}>Overview</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/dashboard`}>Dashboard</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/activity`}>Activity</a>
  <a class="ws-tab active" href={`/workspaces/${data.workspace.id}/secrets`}>Secrets</a>
  <a class="ws-tab" href={`/workspaces/${data.workspace.id}/roles`}>Roles</a>
</nav>

{#if form?.error}
  <FlashMessage type="error">{form.error}</FlashMessage>
{/if}
{#if form?.created}
  <FlashMessage type="success">Secret created.</FlashMessage>
{/if}
{#if form?.updated}
  <FlashMessage type="success">Secret updated.</FlashMessage>
{/if}
{#if form?.deleted}
  <FlashMessage type="success">Secret deleted.</FlashMessage>
{/if}

{#if data.secretsError}
  <FlashMessage type="error">{data.secretsError}</FlashMessage>
{/if}

<div class="toolbar">
  <span class="toolbar-info">{data.secrets.length} secret{data.secrets.length !== 1 ? 's' : ''}</span>
  <button onclick={() => { showAddForm = !showAddForm; }}>
    {showAddForm ? 'Cancel' : 'Add secret'}
  </button>
</div>

{#if showAddForm}
  <div class="add-form">
    <h3>New secret</h3>
    <form method="POST" action="?/createSecret">
      <div class="form-grid">
        <div class="field">
          <label for="secret-name">Name</label>
          <input id="secret-name" name="name" type="text" required
                 placeholder="DATABASE_URL"
                 pattern="[A-Za-z_][A-Za-z0-9_]*"
                 title="Letters, digits, and underscores only"
                 value={form?.name || ''} />
        </div>
        <div class="field">
          <label for="secret-value">Value</label>
          <textarea id="secret-value" name="value" required placeholder="Secret value..."></textarea>
        </div>
      </div>
      <div class="form-footer">
        <button type="button" class="ghost" onclick={() => { showAddForm = false; }}>Cancel</button>
        <button type="submit">Create secret</button>
      </div>
    </form>
  </div>
{/if}

{#if data.secrets.length === 0 && !showAddForm}
  <EmptyState icon="⬡" message="No secrets yet. Secrets are encrypted key-value pairs available to pipelines." />
{:else}
  <div class="secret-list">
    {#each data.secrets as secret (secret.id)}
      <div class="secret-row">
        <div class="secret-name-col">
          <div class="secret-name">{secret.name}</div>
          <div class="secret-meta">
            by {secret.created_by} &middot; updated <TimeAgo value={secret.updated_at} />
          </div>
        </div>
        <span class="secret-value-mask">********</span>
        <div class="secret-actions">
          <button class="ghost" onclick={() => startEdit(secret.id)}>Update</button>
          <button class="ghost" onclick={() => confirmDelete(secret.id)}>Delete</button>
        </div>
      </div>
      {#if editingSecretId === secret.id}
        <div class="edit-form">
          <form method="POST" action="?/updateSecret">
            <input type="hidden" name="secret_id" value={secret.id} />
            <div class="edit-form-inner">
              <div class="field">
                <label for="edit-value-{secret.id}">New value for <strong>{secret.name}</strong></label>
                <textarea id="edit-value-{secret.id}" name="value" required placeholder="New secret value..."></textarea>
              </div>
              <button type="submit">Save</button>
              <button type="button" class="ghost" onclick={cancelEdit}>Cancel</button>
            </div>
          </form>
        </div>
      {/if}
    {/each}
  </div>
{/if}

<!-- Delete confirmation modal -->
<Modal open={confirmDeleteId !== null} title="Delete secret" onClose={cancelDelete}>
  {#if deletingSecret}
    <p>Delete <strong style="font-family: var(--font-mono);">{deletingSecret.name}</strong>? This cannot be undone. Pipelines using this secret will fail.</p>
    <form method="POST" action="?/deleteSecret">
      <input type="hidden" name="secret_id" value={deletingSecret.id} />
      <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);">
        <button type="button" class="ghost" onclick={cancelDelete}>Cancel</button>
        <button type="submit" style="background: var(--danger); border-color: var(--danger);">Delete</button>
      </div>
    </form>
  {/if}
</Modal>
