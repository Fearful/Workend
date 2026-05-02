<script lang="ts">
  let { data } = $props();
</script>

<style>
  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1.5rem;
  }

  h1 {
    font-size: 1.5rem;
    margin: 0;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: 1rem;
  }

  .card {
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    padding: 1.25rem;
  }

  .card h3 {
    margin: 0 0 0.5rem 0;
    font-size: 1rem;
  }

  .card p {
    color: #9ca3af;
    font-size: 0.875rem;
    margin: 0 0 0.75rem 0;
    min-height: 1.25rem;
  }

  .card .meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .empty {
    text-align: center;
    padding: 3rem 1rem;
    color: #9ca3af;
  }

  .error {
    color: #ef4444;
    font-size: 0.875rem;
    background: #2a1414;
    border: 1px solid #5b1a1a;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }
</style>

<div class="header-row">
  <h1>Workspaces</h1>
  <a href="/workspaces/new"><button>New workspace</button></a>
</div>

{#if data.error}
  <div class="error">{data.error}</div>
{/if}

{#if data.workspaces.length === 0}
  <div class="empty">
    <p>No workspaces yet. Create your first one.</p>
  </div>
{:else}
  <div class="grid">
    {#each data.workspaces as ws (ws.id)}
      <div class="card">
        <h3>{ws.name}</h3>
        <p>{ws.description || '—'}</p>
        <div class="meta">created {new Date(ws.created_at).toLocaleDateString()}</div>
      </div>
    {/each}
  </div>
{/if}
