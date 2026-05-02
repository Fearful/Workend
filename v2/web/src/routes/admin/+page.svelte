<script lang="ts">
  let { data } = $props();
</script>

<style>
  h1 { font-size: 1.5rem; margin: 0 0 1.5rem 0; }
  h2 { font-size: 0.875rem; text-transform: uppercase; letter-spacing: 0.05em;
       color: #9ca3af; font-weight: 600; margin: 2rem 0 0.75rem 0; }

  table {
    width: 100%; border-collapse: collapse;
    background: #14181d; border: 1px solid #1f2429; border-radius: 8px;
    overflow: hidden; font-size: 0.8125rem;
  }
  th, td {
    padding: 0.625rem 0.875rem; text-align: left;
    border-bottom: 1px solid #1f2429;
  }
  th {
    background: #1a1f25; color: #9ca3af; font-weight: 600;
    text-transform: uppercase; letter-spacing: 0.05em; font-size: 0.6875rem;
  }
  tr:last-child td { border-bottom: none; }
  td.mono { font-family: ui-monospace, "SF Mono", Menlo, monospace; color: #e8eaed; }
  td.muted { color: #6b7280; font-family: ui-monospace, "SF Mono", Menlo, monospace; }
  .admin-badge { color: #2563eb; font-size: 0.6875rem; font-weight: 600; margin-left: 0.5rem; }
</style>

<h1>Admin</h1>

<h2>Users ({data.users.length})</h2>
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
        <td>{u.display_name}{#if u.is_admin}<span class="admin-badge">ADMIN</span>{/if}</td>
        <td class="mono">{u.email}</td>
        <td class="muted">{new Date(u.created_at).toLocaleDateString()}</td>
        <td class="mono">{u.workspace_count}</td>
        <td class="mono">{u.project_count}</td>
        <td class="mono">{u.run_count}</td>
      </tr>
    {/each}
  </tbody>
</table>

<h2>Audit log (last {data.audit.length})</h2>
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
