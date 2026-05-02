<script lang="ts">
  let { data, form } = $props();
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 0.5rem 0; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 1.5rem 0 1rem 0; }
  .breadcrumb { color: #6b7280; font-size: 0.875rem; margin-bottom: 0.5rem; }

  .panel {
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    padding: 1.25rem 1.5rem; margin-bottom: 1rem;
  }
  .panel h2 { margin-top: 0; }

  .row {
    display: grid; grid-template-columns: 1fr 1fr auto auto auto auto;
    align-items: center; gap: 1rem;
    padding: 0.625rem 0; border-bottom: 1px solid #1f2429; font-size: 0.875rem;
  }
  .row:last-child { border-bottom: none; }
  .mono { font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .muted { color: #6b7280; font-size: 0.75rem; font-family: ui-monospace, "SF Mono", Menlo, monospace; }

  .form-grid { display: grid; grid-template-columns: 2fr 2fr auto auto;
               gap: 0.75rem; align-items: end; }

  .empty { color: #6b7280; text-align: center; padding: 1.5rem; font-size: 0.875rem; }
  .error { color: #ef4444; font-size: 0.875rem; margin-top: 0.5rem; }
  .badge-on { color: #22c55e; font-weight: 600; font-size: 0.6875rem; }
  .badge-off { color: #6b7280; font-weight: 600; font-size: 0.6875rem; }
</style>

<div class="breadcrumb">
  <a href="/">workspaces</a>
  {#if data.workspace}/ <a href={`/workspaces/${data.workspace.id}`}>{data.workspace.name}</a>{/if}
  / <a href={`/projects/${data.project.id}`}>{data.project.name}</a>
  / schedules
</div>

<h1>Schedules</h1>

<section class="panel">
  <h2>Existing schedules</h2>
  {#if data.schedules.length === 0}
    <div class="empty">No schedules yet.</div>
  {:else}
    {#each data.schedules as s (s.id)}
      <div class="row">
        <span class="mono">{s.task_name} <span style="color:#6b7280">({s.task_source})</span></span>
        <span class="mono">{s.cron_expr}</span>
        <span class={s.enabled ? 'badge-on' : 'badge-off'}>{s.enabled ? 'ON' : 'OFF'}</span>
        <span class="muted">next: {s.next_run_at ? new Date(s.next_run_at).toLocaleString() : '—'}</span>
        <form method="POST" action="?/toggle" style="margin: 0;">
          <input type="hidden" name="id" value={s.id} />
          <button type="submit" class="ghost">{s.enabled ? 'Pause' : 'Resume'}</button>
        </form>
        <form method="POST" action="?/delete" style="margin: 0;" onsubmit={(e) => !confirm('Delete schedule?') && e.preventDefault()}>
          <input type="hidden" name="id" value={s.id} />
          <button type="submit" class="ghost">Delete</button>
        </form>
      </div>
    {/each}
  {/if}
</section>

<section class="panel">
  <h2>New schedule</h2>
  {#if data.tasks.length === 0}
    <div class="empty">No tasks detected for this project. Add a package.json or justfile to its repo.</div>
  {:else}
    <form method="POST" action="?/create">
      <div class="form-grid">
        <div class="field">
          <label for="task_id">Task</label>
          <select id="task_id" name="task_id" required style="width:100%; padding: 0.5rem 0.75rem; background:#14181d; color:#e8eaed; border:1px solid #2d3540; border-radius:6px; font-family:inherit; font-size:0.875rem;">
            {#each data.tasks as t (t.id)}
              <option value={t.id}>{t.source}: {t.name}</option>
            {/each}
          </select>
        </div>
        <div class="field">
          <label for="cron_expr">Cron expression</label>
          <input id="cron_expr" name="cron_expr" type="text" required placeholder="0 9 * * 1" value={form?.cron_expr || ''} />
          <small style="color:#6b7280; font-size:0.75rem;">Standard 5-field: m h dom mon dow. <code>0 9 * * 1</code> = 9am every Monday.</small>
        </div>
        <label style="display:flex; align-items:center; gap:0.5rem; padding-bottom:0.5rem;">
          <input type="checkbox" name="enabled" checked />
          <span style="color:#9ca3af; font-size:0.875rem;">Enabled</span>
        </label>
        <button type="submit" style="margin-bottom: 0.125rem;">Create</button>
      </div>
      {#if form?.error}<p class="error">{form.error}</p>{/if}
    </form>
  {/if}
</section>
