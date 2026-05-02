<script lang="ts">
  let { data } = $props();

  function statusColor(status: string): string {
    switch (status) {
      case 'ready':
        return '#22c55e';
      case 'cloning':
      case 'pending':
        return '#eab308';
      case 'error':
        return '#ef4444';
      default:
        return '#6b7280';
    }
  }

  function shortSha(sha: string | null): string {
    return sha ? sha.slice(0, 7) : '—';
  }
</script>

<style>
  .header-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    margin-bottom: 1.5rem;
    gap: 1rem;
  }

  h1 {
    font-size: 1.5rem;
    margin: 0 0 0.25rem 0;
  }

  .desc {
    color: #9ca3af;
    font-size: 0.875rem;
    margin: 0;
  }

  .actions {
    display: flex;
    gap: 0.5rem;
    flex-shrink: 0;
  }

  h2 {
    font-size: 1rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
    margin: 2rem 0 1rem 0;
  }

  .empty {
    text-align: center;
    padding: 3rem 1rem;
    color: #9ca3af;
    background: #14181d;
    border: 1px dashed #1f2429;
    border-radius: 8px;
  }

  .project-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: 1rem;
    padding: 1rem 1.25rem;
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    text-decoration: none;
    color: inherit;
    margin-bottom: 0.5rem;
    transition: border-color 100ms ease;
  }

  .project-row:hover {
    border-color: #2563eb;
    text-decoration: none;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .project-name {
    font-size: 0.95rem;
    font-weight: 500;
  }

  .project-meta {
    color: #6b7280;
    font-size: 0.75rem;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
  }

  .project-status {
    color: #9ca3af;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
  }

  .breadcrumb {
    color: #6b7280;
    font-size: 0.875rem;
    margin-bottom: 0.5rem;
  }

  .error-banner {
    color: #ef4444;
    font-size: 0.875rem;
    background: #2a1414;
    border: 1px solid #5b1a1a;
    padding: 0.75rem 1rem;
    border-radius: 6px;
    margin-bottom: 1rem;
  }
</style>

<div class="breadcrumb"><a href="/">workspaces</a> / {data.workspace.name}</div>

<div class="header-row">
  <div>
    <h1>{data.workspace.name}</h1>
    <p class="desc">{data.workspace.description || 'No description'}</p>
  </div>
  <div class="actions">
    <a href={`/workspaces/${data.workspace.id}/projects/new`}><button>Add project</button></a>
    <form method="POST" action="?/delete" style="margin: 0;" onsubmit={(e) => !confirm('Delete this workspace and all its projects?') && e.preventDefault()}>
      <button type="submit" class="ghost">Delete workspace</button>
    </form>
  </div>
</div>

<h2>Projects</h2>

{#if data.projectsError}
  <div class="error-banner">{data.projectsError}</div>
{/if}

{#if data.projects.length === 0}
  <div class="empty">
    <p>No projects yet. Add one by git URL.</p>
  </div>
{:else}
  {#each data.projects as p (p.id)}
    <a href={`/projects/${p.id}`} class="project-row">
      <span class="dot" style="background: {statusColor(p.status)}"></span>
      <div>
        <div class="project-name">{p.name}</div>
        <div class="project-meta">{p.git_url}</div>
      </div>
      <div class="project-meta">{shortSha(p.last_commit_sha)}</div>
      <div class="project-status">{p.status}</div>
    </a>
  {/each}
{/if}
