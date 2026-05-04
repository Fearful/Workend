<script lang="ts">
  import { goto } from '$app/navigation';
  import { page } from '$app/state';
  import { shortSha, formatDuration } from '$lib/utils';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import Tooltip from '$lib/components/Tooltip.svelte';
  import FilterPills from '$lib/components/FilterPills.svelte';
  import SectionHeader from '$lib/components/SectionHeader.svelte';
  import DensityToggle from '$lib/components/DensityToggle.svelte';

  let { data } = $props();
  let density = $state<'comfortable' | 'compact'>('comfortable');

  type Filter = '' | 'succeeded' | 'failed' | 'cancelled' | 'running';

  function setFilter(status: Filter) {
    const url = new URL(page.url);
    if (status) url.searchParams.set('status', status);
    else url.searchParams.delete('status');
    goto(url, { replaceState: true });
  }

  let counts = $derived.by(() => {
    const c: Record<string, number> = { '': data.runs.length };
    for (const r of data.runs) c[r.status] = (c[r.status] || 0) + 1;
    return c;
  });

  let filterOptions = $derived([
    { value: '' as Filter, label: 'All', count: counts[''] || 0 },
    { value: 'succeeded' as Filter, label: 'Succeeded', count: counts['succeeded'] || 0 },
    { value: 'failed' as Filter, label: 'Failed', count: counts['failed'] || 0 },
    { value: 'running' as Filter, label: 'Running', count: counts['running'] || 0 },
    { value: 'cancelled' as Filter, label: 'Cancelled', count: counts['cancelled'] || 0 }
  ]);
</script>

<style>
  .toolbar {
    display: flex;
    gap: var(--space-3);
    align-items: center;
    margin-bottom: var(--space-4);
    flex-wrap: wrap;
  }
  .toolbar-actions {
    margin-left: auto;
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }

  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto auto auto;
    align-items: center;
    gap: var(--space-4);
    padding: var(--space-3) var(--space-4);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    text-decoration: none;
    color: inherit;
    margin-bottom: var(--space-2);
    font-size: var(--fs-md);
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .run-row.compact {
    padding: 0.375rem var(--space-3);
    font-size: var(--fs-sm);
    margin-bottom: 0.25rem;
  }
  .run-row:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }

  .run-name { font-family: var(--font-mono); font-weight: var(--fw-medium); }
  .dim { color: var(--text-dim); font-weight: var(--fw-regular); }
  .run-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
  }
  .exit-bad { color: var(--status-danger-fg); }
  .branch-tag {
    display: inline-block;
    margin-left: 0.375rem;
    padding: 0.0625rem 0.375rem;
    background: var(--status-info-bg);
    color: var(--status-info-fg);
    border: 1px solid var(--status-info-border);
    border-radius: var(--radius-sm);
    font-size: 0.6875rem;
    font-family: var(--font-mono);
    font-weight: var(--fw-regular);
    vertical-align: middle;
  }

  @media (max-width: 768px) {
    .run-row { grid-template-columns: auto 1fr auto; gap: var(--space-2); padding: var(--space-2) var(--space-3); }
    .run-row > :nth-child(3),
    .run-row > :nth-child(4),
    .run-row > :nth-child(5) { display: none; }
    .run-stack-mobile {
      display: flex;
      flex-direction: column;
      gap: 2px;
      min-width: 0;
    }
    .run-stack-mobile .run-meta { font-size: 0.6875rem; }
  }
  @media (min-width: 769px) {
    .mobile-only { display: none; }
  }
</style>

<SectionHeader title="Run history" />

<div class="toolbar">
  <FilterPills options={filterOptions} value={(data.filter || '') as Filter} onChange={setFilter} />
  <div class="toolbar-actions">
    <DensityToggle storageKey="workend.density.runs" onChange={(v) => density = v} />
  </div>
</div>

{#if data.runs.length === 0}
  <EmptyState>
    {#if data.filter}
      No runs matching <strong>{data.filter}</strong>.
    {:else}
      No runs yet.
    {/if}
  </EmptyState>
{:else}
  {#each data.runs as r (r.id)}
    <a href={`/runs/${r.id}`} class="run-row" class:compact={density === 'compact'}>
      <StatusPill status={r.status} size="sm" />
      <span class="run-name">
        {r.task_name} <span class="dim">({r.task_source})</span>
        {#if r.branch}<span class="branch-tag" title={`Ran on branch ${r.branch}`}>{r.branch}</span>{/if}
        <span class="mobile-only run-stack-mobile">
          <span class="run-meta">{formatDuration(r.started_at, r.finished_at)} · <TimeAgo value={r.created_at} /></span>
        </span>
      </span>
      <Tooltip text={r.commit_sha || 'no commit'}>
        <span class="run-meta">{shortSha(r.commit_sha)}</span>
      </Tooltip>
      <span class="run-meta" class:exit-bad={r.exit_code != null && r.exit_code !== 0}>
        {r.exit_code != null ? `exit ${r.exit_code}` : '—'}
      </span>
      <span class="run-meta">{formatDuration(r.started_at, r.finished_at)}</span>
      <span class="run-meta"><TimeAgo value={r.created_at} /></span>
    </a>
  {/each}
{/if}
