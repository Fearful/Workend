<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import Badge from '$lib/components/Badge.svelte';

  let { data } = $props();
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
