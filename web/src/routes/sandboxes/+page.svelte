<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';

  let { data, form } = $props();

  let showCreate = $state(false);
  let extendTarget = $state<{ id: string; branch: string } | null>(null);
  let extendHours = $state('2');
  let destroyTarget = $state<{ id: string; branch: string } | null>(null);

  function isActive(status: string): boolean {
    return status === 'running' || status === 'starting' || status === 'ready';
  }

  function isDead(status: string): boolean {
    return status === 'destroyed' || status === 'expired' || status === 'error';
  }

  function timeRemaining(expiresAt: string): string {
    const diff = new Date(expiresAt).getTime() - Date.now();
    if (diff <= 0) return 'expired';
    const hours = Math.floor(diff / 3_600_000);
    const minutes = Math.floor((diff % 3_600_000) / 60_000);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  }

  let activeSandboxes = $derived(data.sandboxes.filter((s) => !isDead(s.status)));
  let deadSandboxes = $derived(data.sandboxes.filter((s) => isDead(s.status)));
</script>

<style>
  .sandbox-list {
    display: flex;
    flex-direction: column;
    gap: var(--space-4);
  }

  .sandbox-card {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: var(--space-3);
    align-items: start;
    padding: var(--space-4) var(--space-5);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
  }

  .sandbox-card.dimmed {
    opacity: 0.5;
  }

  .card-main {
    display: flex;
    flex-direction: column;
    gap: var(--space-2);
    min-width: 0;
  }

  .card-header {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }

  .card-branch {
    font-family: var(--font-mono);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
  }

  .card-meta {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    color: var(--text-dim);
    font-size: var(--fs-xs);
    flex-wrap: wrap;
  }

  .meta-sep { color: var(--border-strong); user-select: none; }

  .card-url {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    color: var(--link);
    word-break: break-all;
  }

  .card-actions {
    display: flex;
    gap: var(--space-2);
    flex-shrink: 0;
    flex-wrap: wrap;
  }

  .card-actions button {
    padding: 0.375rem 0.75rem;
    font-size: var(--fs-xs);
  }

  .section-label {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    text-transform: uppercase;
    letter-spacing: 0.05em;
    margin-bottom: var(--space-2);
    margin-top: var(--space-4);
  }

  .modal-actions {
    display: flex;
    gap: var(--space-2);
    justify-content: flex-end;
    margin-top: var(--space-4);
  }

  .modal-hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin: 0.25rem 0 0 0;
  }

  .confirm-text {
    color: var(--text-muted);
    font-size: var(--fs-sm);
    margin: 0 0 var(--space-3) 0;
  }

  .confirm-text strong {
    color: var(--text);
    font-family: var(--font-mono);
  }

  .time-remaining {
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
  }

  @media (max-width: 640px) {
    .sandbox-card {
      grid-template-columns: 1fr;
    }
    .card-actions {
      justify-content: flex-start;
    }
  }
</style>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<PageHeader title="Sandboxes">
  {#snippet actions()}
    <button type="button" onclick={() => (showCreate = true)}>New Sandbox</button>
  {/snippet}
</PageHeader>

{#if data.sandboxes.length === 0}
  <EmptyState icon="&#x1F4E6;" message="No sandboxes yet. Create one to get an isolated dev environment." />
{:else}
  {#if activeSandboxes.length > 0}
    <div class="sandbox-list">
      {#each activeSandboxes as s (s.id)}
        <div class="sandbox-card">
          <div class="card-main">
            <div class="card-header">
              <StatusPill status={s.status} />
              <a href={`/sandboxes/${s.id}`} class="card-branch">{s.branch || 'main'}</a>
              <Badge variant="muted" size="sm">{s.project_id.slice(0, 8)}</Badge>
            </div>
            {#if s.url && isActive(s.status)}
              <a href={s.url} target="_blank" rel="noopener noreferrer" class="card-url">{s.url}</a>
            {/if}
            <div class="card-meta">
              <span>Created <TimeAgo value={s.created_at} /></span>
              <span class="meta-sep">*</span>
              <span class="time-remaining">{timeRemaining(s.expires_at)} remaining</span>
              {#if s.port}
                <span class="meta-sep">*</span>
                <span>Port {s.port}</span>
              {/if}
            </div>
          </div>
          <div class="card-actions">
            <button type="button" class="ghost"
                    onclick={() => { extendTarget = { id: s.id, branch: s.branch }; extendHours = '2'; }}>
              Extend
            </button>
            <button type="button" class="danger"
                    onclick={() => { destroyTarget = { id: s.id, branch: s.branch }; }}>
              Destroy
            </button>
          </div>
        </div>
      {/each}
    </div>
  {/if}

  {#if deadSandboxes.length > 0}
    <div class="section-label">Expired / Destroyed</div>
    <div class="sandbox-list">
      {#each deadSandboxes as s (s.id)}
        <div class="sandbox-card dimmed">
          <div class="card-main">
            <div class="card-header">
              <StatusPill status={s.status} />
              <a href={`/sandboxes/${s.id}`} class="card-branch">{s.branch || 'main'}</a>
              <Badge variant="muted" size="sm">{s.project_id.slice(0, 8)}</Badge>
            </div>
            <div class="card-meta">
              <span>Created <TimeAgo value={s.created_at} /></span>
              {#if s.destroyed_at}
                <span class="meta-sep">*</span>
                <span>Destroyed <TimeAgo value={s.destroyed_at} /></span>
              {/if}
            </div>
          </div>
          <div class="card-actions"></div>
        </div>
      {/each}
    </div>
  {/if}
{/if}

<!-- Create Sandbox Modal -->
<Modal open={showCreate} title="New Sandbox" width={480} onClose={() => (showCreate = false)}>
  <form method="POST" action="?/create" onsubmit={() => { showCreate = false; }}>
    <div class="field">
      <label for="sb-project">Project ID</label>
      <input id="sb-project" name="project_id" type="text" required placeholder="e.g. proj_abc123" />
    </div>
    <div class="field">
      <label for="sb-branch">Branch</label>
      <input id="sb-branch" name="branch" type="text" placeholder="main" />
      <p class="modal-hint">Defaults to "main" if blank.</p>
    </div>
    <div class="field">
      <label for="sb-expires">Expires in (hours)</label>
      <input id="sb-expires" name="expires_hours" type="number" min="1" max="72" value="4" />
      <p class="modal-hint">1-72 hours. Defaults to 4.</p>
    </div>
    <div class="modal-actions">
      <button type="button" class="ghost" onclick={() => (showCreate = false)}>Cancel</button>
      <button type="submit">Create</button>
    </div>
  </form>
</Modal>

<!-- Extend Modal -->
<Modal open={extendTarget !== null} title="Extend Sandbox" width={400} onClose={() => (extendTarget = null)}>
  {#if extendTarget}
    <form method="POST" action="?/extend" onsubmit={() => { extendTarget = null; }}>
      <input type="hidden" name="id" value={extendTarget.id} />
      <p class="confirm-text">
        Extend sandbox on branch <strong>{extendTarget.branch || 'main'}</strong>.
      </p>
      <div class="field">
        <label for="ext-hours">Additional hours</label>
        <input id="ext-hours" name="hours" type="number" min="1" max="72" bind:value={extendHours} />
        <p class="modal-hint">1-72 hours to add.</p>
      </div>
      <div class="modal-actions">
        <button type="button" class="ghost" onclick={() => (extendTarget = null)}>Cancel</button>
        <button type="submit">Extend</button>
      </div>
    </form>
  {/if}
</Modal>

<!-- Destroy Confirmation Modal -->
<Modal open={destroyTarget !== null} title="Destroy Sandbox" width={400} onClose={() => (destroyTarget = null)}>
  {#if destroyTarget}
    <form method="POST" action="?/destroy" onsubmit={() => { destroyTarget = null; }}>
      <input type="hidden" name="id" value={destroyTarget.id} />
      <p class="confirm-text">
        This will permanently destroy the sandbox on branch <strong>{destroyTarget.branch || 'main'}</strong>.
        All data inside the container will be lost.
      </p>
      <div class="modal-actions">
        <button type="button" class="ghost" onclick={() => (destroyTarget = null)}>Cancel</button>
        <button type="submit" class="danger">Destroy</button>
      </div>
    </form>
  {/if}
</Modal>
