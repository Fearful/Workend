<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let detectPending = $state(false);
  let autoMapPending = $state(false);
  let detailModal = $state<{
    id: string;
    loading: boolean;
    pkg: { id: string; name: string; path: string; pkg_type: string } | null;
    scopes: { id: string; task_id: string; task_name: string }[];
    error: string | null;
  } | null>(null);

  function badgeVariant(t: string): 'info' | 'success' | 'warning' | 'danger' | 'muted' | 'accent' {
    switch (t) {
      case 'npm': return 'accent';
      case 'go': return 'info';
      case 'cargo': return 'warning';
      case 'python': return 'success';
      case 'gradle': return 'danger';
      default: return 'muted';
    }
  }

  async function detectPackages() {
    if (detectPending) return;
    detectPending = true;
    try {
      const r = await fetch(`/projects/${data.project.id}/monorepo?/detectPackages`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      detectPending = false;
    }
  }

  async function autoMapTasks() {
    if (autoMapPending) return;
    autoMapPending = true;
    try {
      const r = await fetch(`/projects/${data.project.id}/monorepo?/autoMap`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: ''
      });
      if (!r.ok) return;
      await invalidateAll();
    } finally {
      autoMapPending = false;
    }
  }

  async function openDetail(pkgID: string) {
    detailModal = { id: pkgID, loading: true, pkg: null, scopes: [], error: null };
    try {
      const r = await fetch(`/api/packages/${pkgID}`, { credentials: 'same-origin' });
      if (!r.ok) {
        if (detailModal) detailModal = { ...detailModal, loading: false, error: `HTTP ${r.status}` };
        return;
      }
      const json = await r.json() as {
        id: string; name: string; path: string; pkg_type: string;
        scopes: { id: string; task_id: string; task_name: string }[];
      };
      if (detailModal) {
        detailModal = { ...detailModal, loading: false, pkg: json, scopes: json.scopes ?? [] };
      }
    } catch (err) {
      if (detailModal) {
        detailModal = { ...detailModal, loading: false, error: err instanceof Error ? err.message : 'failed' };
      }
    }
  }

  async function addTaskScope(pkgID: string, taskID: string) {
    const r = await fetch(`/api/packages/${pkgID}/task-scopes`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ task_id: taskID }),
      credentials: 'same-origin'
    });
    if (r.ok) {
      await openDetail(pkgID);
      await invalidateAll();
    }
  }

  async function removeTaskScope(scopeID: string) {
    const r = await fetch(`/api/task-scopes/${scopeID}`, {
      method: 'DELETE',
      credentials: 'same-origin'
    });
    if (r.ok && detailModal) {
      await openDetail(detailModal.id);
      await invalidateAll();
    }
  }
</script>

<style>
  .page-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-4);
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  h2 { margin: 0; font-size: var(--fs-lg); font-weight: var(--fw-semibold); }
  .header-actions { display: flex; gap: var(--space-2); }

  .pkg-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
    gap: var(--space-3);
  }
  .pkg-card {
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-4);
    background: var(--bg-panel);
    cursor: pointer;
    transition: border-color 80ms ease, box-shadow 80ms ease;
  }
  .pkg-card:hover {
    border-color: var(--border-strong);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  .pkg-card-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-2);
  }
  .pkg-card-name {
    font-family: var(--font-mono);
    font-size: 0.875rem;
    font-weight: 500;
    color: var(--text);
  }
  .pkg-card-path {
    font-family: var(--font-mono);
    font-size: 0.75rem;
    color: var(--text-dim);
    margin-bottom: var(--space-2);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .pkg-card-footer {
    font-size: 0.75rem;
    color: var(--text-muted);
  }

  .empty {
    text-align: center;
    padding: 3rem var(--space-4);
    color: var(--text-muted);
  }

  .detail-row {
    display: grid;
    grid-template-columns: 80px 1fr;
    gap: var(--space-3);
    padding: var(--space-2) 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .detail-row:last-child { border-bottom: none; }
  .detail-label { color: var(--text-dim); }
  .detail-value { color: var(--text); font-family: var(--font-mono); word-break: break-all; }

  .scope-item {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.375rem 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .scope-item:last-child { border-bottom: none; }
  .scope-task-name { font-family: var(--font-mono); }

  .scope-add-section {
    margin-top: var(--space-3);
    border-top: 1px solid var(--border);
    padding-top: var(--space-3);
  }
  .scope-add-list {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-1);
    margin-top: var(--space-2);
  }

  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-4);
  }
  .modal-error { color: var(--danger-text); font-size: 0.875rem; }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<div class="page-header">
  <h2>Monorepo packages</h2>
  <div class="header-actions">
    <button type="button" onclick={detectPackages} disabled={detectPending}>
      {detectPending ? 'Detecting...' : 'Detect packages'}
    </button>
    {#if data.packages.length > 0}
      <button type="button" class="ghost" onclick={autoMapTasks} disabled={autoMapPending}>
        {autoMapPending ? 'Mapping...' : 'Auto-map tasks'}
      </button>
    {/if}
  </div>
</div>

{#if data.packages.length === 0}
  <div class="empty">
    <p>No packages detected. Click "Detect packages" to scan for package manifests (package.json, go.mod, Cargo.toml, etc.).</p>
  </div>
{:else}
  <div class="pkg-grid">
    {#each data.packages as pkg (pkg.id)}
      <div class="pkg-card" onclick={() => openDetail(pkg.id)} onkeydown={(e) => { if (e.key === 'Enter') openDetail(pkg.id); }} role="button" tabindex="0">
        <div class="pkg-card-header">
          <span class="pkg-card-name">{pkg.name}</span>
          <Badge variant={badgeVariant(pkg.pkg_type)} size="sm">{pkg.pkg_type}</Badge>
        </div>
        <div class="pkg-card-path">{pkg.path}</div>
        <div class="pkg-card-footer">{pkg.task_count} task{pkg.task_count === 1 ? '' : 's'} scoped</div>
      </div>
    {/each}
  </div>
{/if}

<Modal open={detailModal !== null} title={detailModal?.pkg ? detailModal.pkg.name : 'Package detail'} width={560} onClose={() => detailModal = null}>
  {#if detailModal}
    {#if detailModal.loading}
      <div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading...</div>
    {:else if detailModal.error}
      <p class="modal-error">{detailModal.error}</p>
    {:else if detailModal.pkg}
      <div class="detail-row"><span class="detail-label">Path</span><span class="detail-value">{detailModal.pkg.path}</span></div>
      <div class="detail-row"><span class="detail-label">Type</span><span class="detail-value"><Badge variant={badgeVariant(detailModal.pkg.pkg_type)} size="sm">{detailModal.pkg.pkg_type}</Badge></span></div>

      <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Scoped tasks ({detailModal.scopes.length})</h3>
      {#if detailModal.scopes.length > 0}
        {#each detailModal.scopes as scope (scope.id)}
          <div class="scope-item">
            <span class="scope-task-name">{scope.task_name}</span>
            <button type="button" class="ghost danger" style="font-size: 0.75rem;"
                    onclick={() => removeTaskScope(scope.id)}>Remove</button>
          </div>
        {/each}
      {:else}
        <p style="color: var(--text-dim); font-size: 0.875rem;">No tasks scoped to this package.</p>
      {/if}

      {#if data.tasks.length > 0}
        {@const scopedIDs = new Set(detailModal.scopes.map((s) => s.task_id))}
        {@const available = data.tasks.filter((t) => !scopedIDs.has(t.id))}
        {#if available.length > 0}
          <div class="scope-add-section">
            <span style="font-size: 0.8125rem; color: var(--text-muted);">Add task scope:</span>
            <div class="scope-add-list">
              {#each available.slice(0, 15) as t (t.id)}
                <button type="button" class="ghost" style="font-size: 0.75rem;"
                        onclick={() => addTaskScope(detailModal?.id ?? '', t.id)}>+ {t.name}</button>
              {/each}
              {#if available.length > 15}
                <span style="font-size: 0.75rem; color: var(--text-dim);">+{available.length - 15} more</span>
              {/if}
            </div>
          </div>
        {/if}
      {/if}
    {/if}
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (detailModal = null)}>Close</button>
    </div>
  {/if}
</Modal>
