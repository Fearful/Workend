<script lang="ts">
  import PageHeader from '$lib/components/PageHeader.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import FlashMessage from '$lib/components/FlashMessage.svelte';
  import RunRow from '$lib/components/RunRow.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';

  let { data } = $props();
</script>

<style>
  .section { margin-top: var(--space-6); }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
    gap: var(--space-4);
  }

  .card {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5);
    text-decoration: none;
    color: inherit;
    display: block;
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .card:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }
  .card h3 {
    margin: 0 0 var(--space-2) 0;
    font-size: var(--fs-lg);
    font-weight: var(--fw-semibold);
    letter-spacing: -0.01em;
  }
  .card p {
    color: var(--text-muted);
    font-size: var(--fs-md);
    margin: 0 0 var(--space-3) 0;
    min-height: 1.25rem;
  }
  .meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
  }
</style>

<PageHeader title="Workspaces">
  {#snippet actions()}
    <a href="/workspaces/new"><button>New workspace</button></a>
  {/snippet}
</PageHeader>

{#if data.error}
  <FlashMessage type="error">{data.error}</FlashMessage>
{/if}

{#if data.workspaces.length === 0}
  <EmptyState
    icon="◇"
    message="No workspaces yet. Create your first one to get started."
    actionHref="/workspaces/new"
    actionLabel="New workspace" />
{:else}
  <div class="grid">
    {#each data.workspaces as ws (ws.id)}
      <a href={`/workspaces/${ws.id}`} class="card">
        <h3>{ws.name}</h3>
        <p>{ws.description || '—'}</p>
        <div class="meta">created <TimeAgo value={ws.created_at} /></div>
      </a>
    {/each}
  </div>
{/if}

{#if data.recentRuns.length > 0}
  <div class="section">
    <SectionHeader title="Recent runs" />
    {#each data.recentRuns as r (r.id)}
      <RunRow run={r} />
    {/each}
  </div>
{/if}
