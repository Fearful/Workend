<script lang="ts">
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import Modal from '$lib/components/Modal.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data, form } = $props();

  let deleteModal = $state<{ id: string; taskName: string; cron: string } | null>(null);
</script>

<style>
  .row {
    display: grid;
    grid-template-columns: 1fr 1fr auto auto auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: 0.625rem 0;
    border-bottom: 1px solid var(--border);
    font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .mono { font-family: var(--font-mono); }
  .dim { color: var(--text-dim); }
  .muted {
    color: var(--text-dim);
    font-size: 0.75rem;
    font-family: var(--font-mono);
  }

  .form-grid {
    display: grid;
    grid-template-columns: 2fr 2fr auto auto;
    gap: var(--space-3);
    align-items: end;
  }

  .checkbox-label {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    padding-bottom: var(--space-2);
  }
  .checkbox-label span {
    color: var(--text-muted);
    font-size: 0.875rem;
  }
  .checkbox-label input { width: auto; }

  .schedule-select {
    width: 100%;
    padding: 0.5rem 0.75rem;
    background: var(--bg-panel);
    color: var(--text);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    font-family: inherit;
    font-size: 0.875rem;
  }
  .hint { color: var(--text-dim); font-size: 0.75rem; }

  @media (max-width: 768px) {
    .row {
      grid-template-columns: 1fr;
      gap: var(--space-2);
      padding: var(--space-3) 0;
    }
    .form-grid { grid-template-columns: 1fr; }
  }
</style>

<SectionHeader title="Schedules" />

<Panel title="Existing schedules">
  {#if data.schedules.length === 0}
    <EmptyState message="No schedules yet." />
  {:else}
    {#each data.schedules as s (s.id)}
      <div class="row">
        <span class="mono">{s.task_name} <span class="dim">({s.task_source})</span></span>
        <span class="mono">{s.cron_expr}</span>
        {#if s.enabled}
          <Badge variant="success" size="sm">ON</Badge>
        {:else}
          <Badge variant="muted" size="sm">OFF</Badge>
        {/if}
        <span class="muted">next: <TimeAgo value={s.next_run_at} /></span>
        <form method="POST" action="?/toggle" class="inline-form">
          <input type="hidden" name="id" value={s.id} />
          <button type="submit" class="ghost">{s.enabled ? 'Pause' : 'Resume'}</button>
        </form>
        <button type="button" class="ghost"
                onclick={() => deleteModal = { id: s.id, taskName: s.task_name, cron: s.cron_expr }}>Delete</button>
      </div>
    {/each}
  {/if}
</Panel>

<Panel title="New schedule">
  {#if data.tasks.length === 0}
    <EmptyState message="No tasks detected for this project. Add a package.json or justfile to its repo." />
  {:else}
    <form method="POST" action="?/create">
      <div class="form-grid">
        <div class="field">
          <label for="task_id">Task</label>
          <select id="task_id" name="task_id" required class="schedule-select">
            {#each data.tasks as t (t.id)}
              <option value={t.id}>{t.source}: {t.name}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label for="cron_expr">Cron expression</label>
          <input id="cron_expr" name="cron_expr" type="text" required placeholder="0 9 * * 1" value={form?.cron_expr || ''} />
          <small class="hint">Standard 5-field: m h dom mon dow. <code>0 9 * * 1</code> = 9am every Monday.</small>
        </div>
        <label class="checkbox-label">
          <input type="checkbox" name="enabled" checked />
          <span>Enabled</span>
        </label>
        <button type="submit">Create</button>
      </div>
      {#if form?.error}<FlashMessage type="error">{form.error}</FlashMessage>{/if}
    </form>
  {/if}
</Panel>

<Modal open={deleteModal !== null} title="Delete schedule?" width={420} onClose={() => deleteModal = null}>
  {#if deleteModal}
    <p style="margin: 0 0 var(--space-3) 0;">
      Stop running <strong>{deleteModal.taskName}</strong> on <code>{deleteModal.cron}</code>?
      Past runs are kept; only the schedule is removed.
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
