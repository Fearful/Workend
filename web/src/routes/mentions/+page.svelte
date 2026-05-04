<script lang="ts">
  import { formatRelative } from '$lib/utils';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';

  let { data } = $props();
</script>

<style>
  .toolbar {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    margin-bottom: var(--space-4);
    font-size: 0.875rem;
  }
  .toolbar-spacer { flex: 1; }
  .row {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    padding: var(--space-3) var(--space-4);
    margin-bottom: var(--space-2);
    display: flex;
    flex-direction: column;
    gap: 0.375rem;
    text-decoration: none;
    color: inherit;
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .row:hover { border-color: var(--accent); box-shadow: var(--shadow-card-hover); text-decoration: none; }
  .row.unread { border-left: 3px solid var(--warning); }
  .head {
    display: flex;
    align-items: baseline;
    gap: var(--space-2);
    font-size: 0.8125rem;
    color: var(--text-muted);
    flex-wrap: wrap;
  }
  .actor { color: var(--text); font-weight: 500; }
  .task-name { color: var(--text); font-family: var(--font-mono); }
  .timestamp { margin-left: auto; }
  .preview {
    color: var(--text);
    font-size: 0.875rem;
    overflow: hidden;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    line-clamp: 3;
    -webkit-box-orient: vertical;
  }
</style>

<PageHeader title="Mentions" />

<div class="toolbar">
  {#if data.unreadOnly}
    <a href="/mentions">All mentions</a>
  {:else}
    <a href="/mentions?unread=true">Unread only</a>
  {/if}
  <span class="toolbar-spacer"></span>
  <form method="POST" action="?/markRead" class="inline-form">
    <button type="submit" class="ghost">Mark all read</button>
  </form>
</div>

{#if data.mentions.length === 0}
  <EmptyState icon="@" message="No mentions yet." />
{:else}
  {#each data.mentions as m (m.id)}
    <a href={`/runs/${m.run_id}`} class="row {m.read_at === null ? 'unread' : ''}">
      <div class="head">
        <span class="actor">{m.actor_name || 'someone'}</span>
        <span>mentioned you in</span>
        <span class="task-name">{m.task_name}</span>
        <span>·</span>
        <span>{m.project_name}</span>
        <span class="timestamp">{formatRelative(m.created_at)}</span>
      </div>
      <div class="preview">{m.body}</div>
    </a>
  {/each}
{/if}
