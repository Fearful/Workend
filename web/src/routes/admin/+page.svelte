<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Panel from '$lib/components/Panel.svelte';
  import Badge from '$lib/components/Badge.svelte';

  let { data, form } = $props();

  let exportFrom = $state('');
  let exportTo = $state('');
  let exporting = $state(false);

  async function exportCSV() {
    if (!exportFrom || !exportTo) return;
    exporting = true;
    try {
      const resp = await fetch(`/api/admin/audit-log/export?from=${exportFrom}&to=${exportTo}`);
      if (!resp.ok) {
        alert('Export failed: ' + (await resp.text()));
        return;
      }
      const blob = await resp.blob();
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `audit-log-${exportFrom}-to-${exportTo}.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (err) {
      alert('Export failed');
    } finally {
      exporting = false;
    }
  }
</script>

<style>
  h2 {
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--text-muted);
    font-weight: 600;
    margin: var(--space-8) 0 var(--space-3) 0;
  }

  .table-wrapper {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: auto;
    -webkit-overflow-scrolling: touch;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.8125rem;
  }
  th, td {
    padding: 0.625rem 0.875rem;
    text-align: left;
    border-bottom: 1px solid var(--border);
    white-space: nowrap;
  }
  th {
    background: var(--bg-hover);
    color: var(--text-muted);
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    font-size: 0.6875rem;
  }
  tr:last-child td { border-bottom: none; }
  td.mono { font-family: var(--font-mono); color: var(--text); }
  td.muted { color: var(--text-dim); font-family: var(--font-mono); }
  .name-cell { display: inline-flex; align-items: center; gap: var(--space-2); }

  .export-row {
    display: flex;
    align-items: flex-end;
    gap: var(--space-3);
    flex-wrap: wrap;
  }
  .export-row .field {
    margin-bottom: 0;
    min-width: 160px;
  }
  .export-row label {
    display: block;
    margin-bottom: 0.25rem;
    font-size: 0.875rem;
    color: var(--text-muted);
  }

  .retention-row {
    display: flex;
    align-items: flex-end;
    gap: var(--space-3);
    flex-wrap: wrap;
    margin-top: var(--space-4);
    padding-top: var(--space-4);
    border-top: 1px solid var(--border);
  }
  .retention-row .field {
    margin-bottom: 0;
    min-width: 120px;
  }

  .form-error {
    color: var(--danger-text);
    font-size: 0.875rem;
    margin: var(--space-2) 0 0 0;
  }
  .form-success {
    color: var(--success);
    font-size: 0.875rem;
    margin: var(--space-2) 0 0 0;
  }
  .hint {
    color: var(--text-dim);
    font-size: 0.75rem;
    margin: 0.25rem 0 0 0;
    line-height: 1.5;
  }
</style>

<PageHeader title="Admin" />

<h2>Users ({data.users.length})</h2>
<div class="table-wrapper">
  <table>
    <thead>
      <tr>
        <th>Display name</th><th>Email</th><th>Created</th>
        <th>Workspaces</th><th>Projects</th><th>Runs</th>
      </tr>
    </thead>
    <tbody>
      {#each data.users as u (u.id)}
        <tr>
          <td>
            <span class="name-cell">
              {u.display_name}
              {#if u.is_admin}<Badge variant="accent" size="sm">ADMIN</Badge>{/if}
            </span>
          </td>
          <td class="mono">{u.email}</td>
          <td class="muted">{new Date(u.created_at).toLocaleDateString()}</td>
          <td class="mono">{u.workspace_count}</td>
          <td class="mono">{u.project_count}</td>
          <td class="mono">{u.run_count}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<h2>Audit log (last {data.audit.length})</h2>
<div class="table-wrapper">
  <table>
    <thead>
      <tr>
        <th>When</th><th>Actor</th><th>Action</th><th>Target</th><th>IP</th>
      </tr>
    </thead>
    <tbody>
      {#each data.audit as e (e.id)}
        <tr>
          <td class="muted">{new Date(e.occurred_at).toLocaleString()}</td>
          <td class="mono">{e.actor_email || '—'}</td>
          <td class="mono">{e.action}</td>
          <td class="muted">{e.target_kind ? `${e.target_kind}:${(e.target_id ?? '').slice(0, 8)}` : '—'}</td>
          <td class="muted">{e.ip || '—'}</td>
        </tr>
      {/each}
    </tbody>
  </table>
</div>

<Panel title="Audit log management">
  <p class="hint">Export audit log entries as CSV or configure automatic retention cleanup.</p>

  <div class="export-row">
    <div class="field">
      <label for="export-from">From</label>
      <input id="export-from" type="date" bind:value={exportFrom} />
    </div>
    <div class="field">
      <label for="export-to">To</label>
      <input id="export-to" type="date" bind:value={exportTo} />
    </div>
    <button type="button" onclick={exportCSV} disabled={exporting || !exportFrom || !exportTo} class="ghost">
      {exporting ? 'Exporting...' : 'Export CSV'}
    </button>
  </div>

  <div class="retention-row">
    <form method="POST" action="?/setRetention" style="display: flex; align-items: flex-end; gap: var(--space-3); flex-wrap: wrap;">
      <div class="field">
        <label for="retention-days">Retention period (days)</label>
        <input id="retention-days" name="days" type="number" min="1" max="3650"
               placeholder="e.g. 90, 180, 365"
               style="width: 180px;" />
      </div>
      <button type="submit" class="ghost">Set retention</button>
    </form>
  </div>
  {#if form?.retentionError}<p class="form-error">{form.retentionError}</p>{/if}
  {#if form?.retentionSet}<p class="form-success">Retention set to {form.retentionDays} days.</p>{/if}
</Panel>
