<script lang="ts">
  import StatusPill from './StatusPill.svelte';
  import TimeAgo from './TimeAgo.svelte';

  type Run = {
    id: string;
    status: string;
    task_name: string;
    task_source: string;
    workspace_name?: string;
    project_name?: string;
    exit_code: number | null;
    created_at: string;
  };

  let { run, showLocation = true }: { run: Run; showLocation?: boolean } = $props();
</script>

<a href={`/runs/${run.id}`} class="run-row">
  <StatusPill status={run.status} size="sm" />
  <div class="run-info">
    <div class="run-task">
      {run.task_name} <span class="dim">({run.task_source})</span>
    </div>
    {#if showLocation && run.workspace_name && run.project_name}
      <div class="run-loc">{run.workspace_name} / {run.project_name}</div>
    {/if}
  </div>
  {#if run.exit_code != null && run.exit_code !== 0}
    <span class="run-meta exit-bad">exit {run.exit_code}</span>
  {:else}
    <span class="run-meta"></span>
  {/if}
  <span class="run-meta"><TimeAgo value={run.created_at} /></span>
</a>

<style>
  .run-row {
    display: grid;
    grid-template-columns: auto 1fr auto auto;
    align-items: center;
    gap: var(--space-3);
    padding: var(--space-2) var(--space-4);
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-md);
    text-decoration: none;
    color: inherit;
    margin-bottom: 0.25rem;
    font-size: var(--fs-md);
    transition: border-color 120ms ease, box-shadow 120ms ease;
  }
  .run-row:hover {
    border-color: var(--border-strong);
    text-decoration: none;
    box-shadow: var(--shadow-card);
  }
  .run-task {
    font-family: var(--font-mono);
    color: var(--text);
    font-weight: var(--fw-medium);
  }
  .dim { color: var(--text-dim); font-weight: var(--fw-regular); }
  .run-loc {
    color: var(--text-muted);
    font-size: var(--fs-xs);
  }
  .run-meta {
    color: var(--text-dim);
    font-size: var(--fs-xs);
    font-family: var(--font-mono);
  }
  .exit-bad { color: var(--status-danger-fg); }

  @media (max-width: 640px) {
    .run-row {
      grid-template-columns: auto 1fr auto;
      gap: var(--space-2);
    }
    .run-meta:nth-of-type(1) { display: none; }
  }
</style>
