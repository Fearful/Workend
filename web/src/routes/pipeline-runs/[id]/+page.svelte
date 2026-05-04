<script lang="ts">
  import { invalidateAll } from '$app/navigation';
  import { onMount, onDestroy } from 'svelte';
  import Breadcrumb from '$lib/components/Breadcrumb.svelte';
  import StatusPill from '$lib/components/StatusPill.svelte';
  import TimeAgo from '$lib/components/TimeAgo.svelte';
  import { formatDuration } from '$lib/utils';

  let { data } = $props();

  let pollHandle: ReturnType<typeof setInterval> | null = null;

  function maybeStartPolling() {
    const transient = data.pipelineRun.status === 'queued' || data.pipelineRun.status === 'running' ||
      (data.pipelineRun.children ?? []).some((c) => c.status === 'queued' || c.status === 'running');
    if (transient && !pollHandle) {
      pollHandle = setInterval(() => invalidateAll(), 2000);
    } else if (!transient && pollHandle) {
      clearInterval(pollHandle);
      pollHandle = null;
    }
  }

  $effect(() => {
    maybeStartPolling();
  });

  onMount(maybeStartPolling);
  onDestroy(() => { if (pollHandle) clearInterval(pollHandle); });
</script>

<style>
  .hero {
    display: flex;
    align-items: flex-start;
    gap: var(--space-4);
    margin-bottom: var(--space-5);
    flex-wrap: wrap;
  }
  .hero-main { display: flex; flex-direction: column; gap: var(--space-2); min-width: 0; flex: 1; }
  .hero-line {
    display: flex;
    align-items: center;
    gap: var(--space-2);
    flex-wrap: wrap;
  }
  h1 {
    font-size: var(--fs-xl);
    margin: 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.02em;
    line-height: var(--lh-tight);
  }
  .hero-meta {
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: baseline;
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }
  .hero-meta-item { display: inline-flex; align-items: baseline; gap: 0.375rem; }
  .hero-meta-label {
    color: var(--text-dim);
    font-size: 0.6875rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
  }
  .hero-meta-item strong { color: var(--text); font-weight: var(--fw-medium); }
  .hero-sep { color: var(--border-strong); user-select: none; }

  .step-list {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    overflow: hidden;
  }
  .step-row {
    display: grid;
    grid-template-columns: 2.5rem auto 1fr auto auto;
    gap: var(--space-3);
    align-items: center;
    padding: var(--space-3) var(--space-4);
    border-bottom: 1px solid var(--border);
    text-decoration: none;
    color: inherit;
  }
  .step-row:hover { background: var(--bg-hover); text-decoration: none; }
  .step-row:last-child { border-bottom: none; }
  .step-pos {
    font-family: var(--font-mono);
    color: var(--text-dim);
    font-size: var(--fs-xs);
  }
  .step-task { font-family: var(--font-mono); font-weight: var(--fw-medium); }
  .step-meta { color: var(--text-dim); font-size: var(--fs-xs); font-family: var(--font-mono); }
  .arrow {
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: var(--fs-md);
  }
</style>

<Breadcrumb segments={[
  { label: 'workspaces', href: '/' },
  { label: `pipeline run ${data.pipelineRun.id.slice(0, 8)}` }
]} />

<div class="hero">
  <div class="hero-main">
    <div class="hero-line">
      <StatusPill status={data.pipelineRun.status} />
      <h1>Pipeline run</h1>
    </div>
    <div class="hero-meta">
      <span class="hero-meta-item">
        <span class="hero-meta-label">Steps</span>
        <strong>{(data.pipelineRun.children ?? []).length}</strong>
      </span>
      <span class="hero-sep">·</span>
      <span class="hero-meta-item">
        <span class="hero-meta-label">Duration</span>
        <strong>{formatDuration(data.pipelineRun.started_at, data.pipelineRun.finished_at)}</strong>
      </span>
      <span class="hero-sep">·</span>
      <span class="hero-meta-item">
        <span class="hero-meta-label">Started</span>
        <strong><TimeAgo value={data.pipelineRun.started_at ?? data.pipelineRun.created_at} /></strong>
      </span>
    </div>
  </div>
</div>

<section class="step-list">
  {#each (data.pipelineRun.children ?? []) as c (c.run_id)}
    <a href={`/runs/${c.run_id}`} class="step-row">
      <span class="step-pos">#{c.step + 1}</span>
      <StatusPill status={c.status} size="sm" />
      <span class="step-task">{c.task_name || c.run_id.slice(0, 8)}</span>
      <span class="step-meta">view log →</span>
      <span class="arrow">→</span>
    </a>
  {/each}
  {#if (data.pipelineRun.children ?? []).length === 0}
    <div style="padding: var(--space-5); color: var(--text-dim); text-align: center;">No steps run yet.</div>
  {/if}
</section>
