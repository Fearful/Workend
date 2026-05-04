<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { shortSha as _shortSha } from '$lib/utils';
  import Badge from '$lib/components/Badge.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';

  let { data, form } = $props();

  let dragName = $state<string | null>(null);
  let dropName = $state<string | null>(null);
  let prModal = $state<{ source: string; target: string; title: string; body: string } | null>(null);
  let switchModal = $state<{ name: string } | null>(null);
  let switchPending = $state(false);
  let switchError = $state<string | null>(null);
  let prDragError = $state<string | null>(null);

  let filter = $state('');
  let showBots = $state(false);

  function isBot(name: string): boolean {
    return name.startsWith('dependabot/') || name.startsWith('renovate/');
  }

  let filteredBranches = $derived.by(() => {
    const q = filter.trim().toLowerCase();
    return data.branches.filter((b) => {
      if (q && !b.name.toLowerCase().includes(q)) return false;
      if (!showBots && !q && isBot(b.name)) return false;
      return true;
    });
  });

  let botCount = $derived(data.branches.filter((b) => isBot(b.name)).length);

  function shortSha(sha: string): string {
    return _shortSha(sha, 8);
  }

  function onDragStart(name: string, e: DragEvent) {
    dragName = name;
    if (e.dataTransfer) {
      e.dataTransfer.effectAllowed = 'move';
      e.dataTransfer.setData('text/plain', name);
    }
  }

  function onDragOver(name: string, e: DragEvent) {
    if (!dragName || dragName === name) return;
    e.preventDefault();
    if (e.dataTransfer) e.dataTransfer.dropEffect = 'move';
    dropName = name;
  }

  function onDragLeave(name: string) {
    if (dropName === name) dropName = null;
  }

  function onDrop(target: string, e: DragEvent) {
    e.preventDefault();
    const source = dragName;
    dragName = null;
    dropName = null;
    if (!source || source === target) return;
    if (!data.canCreatePR) {
      prDragError = 'Pull/merge requests require an OAuth provider connection. Connect one in Settings.';
      setTimeout(() => { prDragError = null; }, 4000);
      return;
    }
    prModal = {
      source,
      target,
      title: `Merge ${source} into ${target}`,
      body: ''
    };
  }

  function onDragEnd() {
    dragName = null;
    dropName = null;
  }

  function requestSwitch(name: string) {
    if (switchPending) return;
    switchError = null;
    switchModal = { name };
  }

  async function confirmSwitch() {
    if (!switchModal || switchPending) return;
    const name = switchModal.name;
    switchPending = true;
    try {
      const r = await fetch(`/projects/${data.project.id}/branches?/switch`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: new URLSearchParams({ name })
      });
      if (!r.ok) {
        switchError = 'Failed to switch branch.';
        return;
      }
      window.location.href = `/projects/${data.project.id}`;
    } finally {
      switchPending = false;
    }
  }
</script>

<style>
  .source-pill {
    display: inline-block;
    padding: 0.125rem 0.5rem;
    border-radius: var(--radius-sm);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    background: var(--bg-hover);
  }

  .hint {
    color: var(--text-dim);
    font-size: 0.8125rem;
    margin: 0 0 var(--space-4) 0;
  }

  .branch-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: 0;
    overflow: hidden;
  }

  .branch-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto;
    align-items: center;
    gap: 0.875rem;
    padding: 0.625rem 0.875rem;
    border-bottom: 1px solid var(--border);
    cursor: grab;
    transition: background 80ms ease, border-color 80ms ease;
  }
  .branch-row:active { cursor: grabbing; }
  .branch-row:last-child { border-bottom: none; }
  .branch-row:hover { background: var(--bg-hover); }
  .branch-row.dragging { opacity: 0.45; }
  .branch-row.drop-target {
    background: rgba(96, 165, 250, 0.1);
    outline: 2px dashed var(--link);
    outline-offset: -2px;
  }

  .grip {
    width: 14px;
    color: var(--text-dim);
    font-size: 1rem;
    user-select: none;
  }

  .name {
    font-family: var(--font-mono);
    font-size: 0.9375rem;
  }

  .sha {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 0.75rem;
  }

  .pr-modal-form {
    display: contents;
  }

  .filter-bar {
    display: flex;
    align-items: center;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
    flex-wrap: wrap;
  }
  .filter-input {
    flex: 1;
    max-width: 280px;
    padding: 0.375rem 0.625rem;
    font-size: var(--fs-sm);
  }
  .show-bots {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    color: var(--text-dim);
    font-size: var(--fs-sm);
    margin: 0;
    cursor: pointer;
  }
  .show-bots input { width: auto; margin: 0; }
  .count {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
    margin-left: auto;
  }

  @media (max-width: 640px) {
    .branch-row {
      grid-template-columns: auto 1fr auto;
    }
    .branch-row > :nth-child(3),
    .branch-row > :nth-child(4) { display: none; }
  }
</style>

<SectionHeader title="Branches">
  {#snippet actions()}
    <span class="source-pill">{data.branchesSource === 'provider' ? 'OAuth' : 'ls-remote'}</span>
  {/snippet}
</SectionHeader>

<p class="hint">
  Click a branch to switch the project to it (re-clones in the background).
  {#if data.canCreatePR}
    Drag a branch onto another to open a pull/merge request from source → target.
  {:else}
    Connect an OAuth provider in <a href="/settings">Settings</a> to enable drag-to-PR.
  {/if}
</p>

{#if data.branchesError}<FlashMessage type="error">{data.branchesError}</FlashMessage>{/if}
{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}
{#if prDragError}<FlashMessage type="error">{prDragError}</FlashMessage>{/if}
{#if form?.prURL}
  <FlashMessage type="success">
    PR #{form.prNumber} opened.
    <a href={form.prURL} target="_blank" rel="noopener">View on provider →</a>
  </FlashMessage>
{/if}

<div class="filter-bar">
  <input type="search" class="filter-input" placeholder="Filter branches…" bind:value={filter} />
  {#if botCount > 0 && !filter.trim()}
    <label class="show-bots">
      <input type="checkbox" bind:checked={showBots} />
      <span>Show {botCount} bot branches</span>
    </label>
  {/if}
  <span class="count">{filteredBranches.length} of {data.branches.length}</span>
</div>

<section class="branch-list">
  {#if data.branches.length === 0}
    <EmptyState message="No branches found." />
  {:else if filteredBranches.length === 0}
    <EmptyState message="No branches match." />
  {:else}
    {#each filteredBranches as b (b.name)}
      <div class="branch-row {dragName === b.name ? 'dragging' : ''} {dropName === b.name ? 'drop-target' : ''}"
           draggable="true"
           role="button"
           tabindex="0"
           ondragstart={(e) => onDragStart(b.name, e)}
           ondragover={(e) => onDragOver(b.name, e)}
           ondragleave={() => onDragLeave(b.name)}
           ondrop={(e) => onDrop(b.name, e)}
           ondragend={onDragEnd}
           onclick={() => requestSwitch(b.name)}
           onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') requestSwitch(b.name); }}>
        <span class="grip" aria-hidden="true">⋮⋮</span>
        <span class="name">{b.name}</span>
        {#if b.default}<Badge variant="success" size="sm">default</Badge>{:else}<span></span>{/if}
        {#if b.protected}<Badge variant="warning" size="sm">protected</Badge>{:else}<span></span>{/if}
        <Tooltip text={b.commit_sha}><span class="sha">{shortSha(b.commit_sha)}</span></Tooltip>
      </div>
    {/each}
  {/if}
</section>

<Modal open={switchModal !== null} title="Switch branch?" width={420} onClose={() => switchModal = null}>
  {#if switchModal}
    <p style="margin: 0 0 var(--space-3) 0;">
      Switch this project to <strong>{switchModal.name}</strong>? The repo will be re-cloned.
    </p>
    {#if switchError}
      <p style="color: var(--danger-text); font-size: 0.875rem; margin: 0 0 var(--space-3) 0;">{switchError}</p>
    {/if}
    <div style="display:flex; gap: var(--space-2); justify-content:flex-end;">
      <button type="button" class="ghost" onclick={() => (switchModal = null)} disabled={switchPending}>Cancel</button>
      <button type="button" onclick={confirmSwitch} disabled={switchPending}>
        {switchPending ? 'Switching…' : 'Switch'}
      </button>
    </div>
  {/if}
</Modal>

<Modal open={prModal !== null} title={prModal ? `Open PR: ${prModal.source} → ${prModal.target}` : ''} width={560} onClose={() => prModal = null}>
  {#if prModal}
    <form method="POST" action="?/createPR" class="pr-modal-form"
          onsubmit={() => { setTimeout(() => { prModal = null; invalidateAll(); }, 0); }}>
      <input type="hidden" name="source" value={prModal.source} />
      <input type="hidden" name="target" value={prModal.target} />
      <div class="field">
        <label for="pr-title">Title</label>
        <input id="pr-title" name="title" bind:value={prModal.title} required />
      </div>
      <div class="field">
        <label for="pr-body">Description (optional)</label>
        <textarea id="pr-body" name="body" rows="6" bind:value={prModal.body}></textarea>
      </div>
      <div style="display:flex; gap: var(--space-2); justify-content:flex-end;">
        <button type="button" class="ghost" onclick={() => (prModal = null)}>Cancel</button>
        <button type="submit">Open PR</button>
      </div>
    </form>
  {/if}
</Modal>
