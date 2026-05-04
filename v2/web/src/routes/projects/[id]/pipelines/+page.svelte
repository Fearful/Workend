<script lang="ts">
  import { enhance } from '$app/forms';
  import Panel from '$lib/components/Panel.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Modal from '$lib/components/Modal.svelte';

  let { data, form } = $props();

  let creatorOpen = $state(false);
  let pipelineName = $state('');
  let selectedTaskIDs = $state<string[]>([]);
  let deleteModal = $state<{ id: string; name: string } | null>(null);

  function toggleTask(id: string) {
    if (selectedTaskIDs.includes(id)) {
      selectedTaskIDs = selectedTaskIDs.filter((t) => t !== id);
    } else {
      selectedTaskIDs = [...selectedTaskIDs, id];
    }
  }
  function moveTask(idx: number, dir: -1 | 1) {
    const next = [...selectedTaskIDs];
    const swap = idx + dir;
    if (swap < 0 || swap >= next.length) return;
    [next[idx], next[swap]] = [next[swap], next[idx]];
    selectedTaskIDs = next;
  }
  function taskByID(id: string) {
    return data.tasks.find((t) => t.id === id);
  }
</script>

<style>
  .pipeline-row {
    display: grid;
    grid-template-columns: 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border);
  }
  .pipeline-row:last-child { border-bottom: none; }
  .pipeline-name { font-size: var(--fs-md); font-weight: var(--fw-medium); }
  .pipeline-steps {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-xs);
    margin-top: 0.25rem;
    display: flex;
    align-items: center;
    gap: 0.375rem;
    flex-wrap: wrap;
  }
  .step-pill {
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    padding: 0.0625rem 0.375rem;
  }
  .step-arrow { color: var(--text-dim); }

  .builder { display: flex; flex-direction: column; gap: var(--space-3); }
  .builder-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: var(--space-4);
  }
  .field-label {
    display: block;
    margin-bottom: 0.25rem;
    font-size: var(--fs-md);
    color: var(--text-muted);
  }
  @media (max-width: 768px) {
    .builder-grid { grid-template-columns: 1fr; }
  }
  .task-list {
    display: flex;
    flex-direction: column;
    gap: 2px;
    max-height: 320px;
    overflow-y: auto;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 0.25rem;
  }
  .task-item {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.375rem 0.5rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: var(--fs-sm);
  }
  .task-item:hover { background: var(--bg-hover); }
  .task-item input { width: auto; margin: 0; }
  .task-source {
    font-family: var(--font-mono);
    color: var(--text-dim);
    font-size: var(--fs-xs);
  }

  .selected-list {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: 0.5rem;
    min-height: 120px;
  }
  .selected-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto;
    gap: 0.5rem;
    align-items: center;
    padding: 0.375rem 0.5rem;
    background: var(--bg-page);
    border: 1px solid var(--border);
    border-radius: var(--radius-sm);
    font-size: var(--fs-sm);
  }
  .selected-empty {
    color: var(--text-dim);
    font-size: var(--fs-sm);
    padding: var(--space-3);
    text-align: center;
  }
  .pos {
    font-family: var(--font-mono);
    color: var(--text-dim);
    font-size: var(--fs-xs);
    width: 1.5rem;
  }
  .arrow-btn {
    background: transparent;
    border: 1px solid var(--border);
    color: var(--text-dim);
    padding: 0.125rem 0.375rem;
    border-radius: var(--radius-sm);
    cursor: pointer;
    font-size: var(--fs-xs);
    line-height: 1;
  }
  .arrow-btn:hover { color: var(--text); border-color: var(--border-strong); }
  .arrow-btn:disabled { opacity: 0.3; cursor: not-allowed; }
  .remove-btn {
    background: transparent;
    border: none;
    color: var(--text-dim);
    cursor: pointer;
    font-size: var(--fs-md);
    line-height: 1;
    padding: 0 0.25rem;
  }
  .remove-btn:hover { color: var(--danger-text); }

  .builder-footer {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-top: var(--space-3);
  }
  .builder-hint { color: var(--text-dim); font-size: var(--fs-sm); }

  .pipeline-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
</style>

<SectionHeader title="Pipelines">
  {#snippet actions()}
    <button type="button" onclick={() => { creatorOpen = true; pipelineName = ''; selectedTaskIDs = []; }}>
      New pipeline
    </button>
  {/snippet}
</SectionHeader>

<p style="color: var(--text-dim); font-size: var(--fs-sm); margin: 0 0 var(--space-3);">
  Chain tasks to run sequentially. Each step is a real run with its own log; the chain stops on first failure.
</p>

{#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}

<Panel title="Existing pipelines">
  {#if data.pipelines.length === 0}
    <EmptyState message="No pipelines yet. Click 'New pipeline' to chain tasks together." />
  {:else}
    {#each data.pipelines as p (p.id)}
      <div class="pipeline-row">
        <div>
          <div class="pipeline-name">{p.name}</div>
          <div class="pipeline-steps">
            {#each (p.steps ?? []) as s, i (s.position)}
              <span class="step-pill">{s.task_source || ''}<span style="opacity:0.6">{s.task_source ? ':' : ''}</span>{s.task_name || s.task_id.slice(0, 8)}</span>
              {#if i < (p.steps?.length ?? 0) - 1}<span class="step-arrow">→</span>{/if}
            {/each}
          </div>
        </div>
        <span style="color: var(--text-dim); font-size: var(--fs-xs);">created <TimeAgo value={p.created_at} /></span>
        <div class="pipeline-actions">
          <form method="POST" action="?/run" class="inline-form">
            <input type="hidden" name="id" value={p.id} />
            <button type="submit" disabled={(p.steps ?? []).length === 0}>Run</button>
          </form>
          <button type="button" class="ghost" onclick={() => deleteModal = { id: p.id, name: p.name }}>Delete</button>
        </div>
      </div>
    {/each}
  {/if}
</Panel>

<!-- Builder modal -->
<Modal open={creatorOpen} title="New pipeline" width={720} onClose={() => creatorOpen = false}>
  <form method="POST" action="?/create" class="builder"
        use:enhance={() => {
          return async ({ result, update }) => {
            await update();
            if (result.type === 'success' || result.type === 'redirect') {
              creatorOpen = false;
              pipelineName = '';
              selectedTaskIDs = [];
            }
          };
        }}>
    <div class="field">
      <label for="pipeline-name">Name</label>
      <input id="pipeline-name" name="name" required bind:value={pipelineName} placeholder="e.g. ci, deploy, release" />
    </div>

    <div class="builder-grid">
      <div class="field">
        <span class="field-label">Available tasks</span>
        {#if data.tasks.length === 0}
          <p style="color: var(--text-dim); font-size: var(--fs-sm);">No tasks detected for this project.</p>
        {:else}
          <div class="task-list">
            {#each data.tasks as t (t.id)}
              <label class="task-item">
                <input type="checkbox"
                       checked={selectedTaskIDs.includes(t.id)}
                       onchange={() => toggleTask(t.id)} />
                <span>{t.name}</span>
                <span class="task-source">({t.source})</span>
              </label>
            {/each}
          </div>
        {/if}
      </div>

      <div class="field">
        <span class="field-label">Pipeline steps (use arrows to reorder)</span>
        <div class="selected-list">
          {#if selectedTaskIDs.length === 0}
            <div class="selected-empty">Select tasks on the left to add them here.</div>
          {:else}
            {#each selectedTaskIDs as id, idx (id)}
              {@const t = taskByID(id)}
              <div class="selected-row">
                <span class="pos">{idx + 1}</span>
                <span>{t?.name ?? id} <span class="task-source">({t?.source ?? '?'})</span></span>
                <button type="button" class="arrow-btn" disabled={idx === 0} onclick={() => moveTask(idx, -1)} title="Move up">↑</button>
                <button type="button" class="arrow-btn" disabled={idx === selectedTaskIDs.length - 1} onclick={() => moveTask(idx, 1)} title="Move down">↓</button>
                <button type="button" class="remove-btn" onclick={() => toggleTask(id)} title="Remove">×</button>
                <input type="hidden" name="task_id" value={id} />
              </div>
            {/each}
          {/if}
        </div>
      </div>
    </div>

    <div class="builder-footer">
      <span class="builder-hint">{selectedTaskIDs.length} task{selectedTaskIDs.length === 1 ? '' : 's'} selected</span>
      <div style="display:flex; gap: var(--space-2);">
        <button type="button" class="ghost" onclick={() => creatorOpen = false}>Cancel</button>
        <button type="submit" disabled={!pipelineName.trim() || selectedTaskIDs.length === 0}>Create pipeline</button>
      </div>
    </div>
  </form>
</Modal>

<Modal open={deleteModal !== null} title="Delete pipeline?" width={420} onClose={() => deleteModal = null}>
  {#if deleteModal}
    <p style="margin: 0 0 var(--space-3) 0;">
      Permanently delete pipeline <strong>{deleteModal.name}</strong>? Past pipeline runs are kept; only the chain definition is removed.
    </p>
    <div style="display:flex; gap: var(--space-2); justify-content:flex-end;">
      <button type="button" class="ghost" onclick={() => (deleteModal = null)}>Cancel</button>
      <form method="POST" action="?/delete" class="inline-form">
        <input type="hidden" name="id" value={deleteModal.id} />
        <button type="submit" class="danger">Delete</button>
      </form>
    </div>
  {/if}
</Modal>
