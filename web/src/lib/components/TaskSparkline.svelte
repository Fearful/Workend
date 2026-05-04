<script lang="ts">
  import Tooltip from './Tooltip.svelte';

  let {
    runs,
    max = 10
  }: {
    runs: { id: string; status: string; created_at: string }[];
    max?: number;
  } = $props();

  // Most-recent first; we render right-to-left so the rightmost dot is the
  // newest run. Take a slice and reverse for left-to-right rendering.
  let display = $derived.by(() => {
    const slice = runs.slice(0, max);
    return slice.reverse();
  });

  function color(status: string): string {
    switch (status) {
      case 'succeeded': return 'var(--status-success-fg)';
      case 'failed':    return 'var(--status-danger-fg)';
      case 'cancelled': return 'var(--text-dim)';
      case 'running':
      case 'queued':    return 'var(--status-warning-fg)';
      default:          return 'var(--text-dim)';
    }
  }

  function tipText(): string {
    if (display.length === 0) return 'No runs yet';
    const latest = runs[0];
    const succeeded = runs.filter((r) => r.status === 'succeeded').length;
    const failed = runs.filter((r) => r.status === 'failed').length;
    return `${runs.length} run${runs.length === 1 ? '' : 's'} · ${succeeded} ok / ${failed} failed · last: ${latest.status}`;
  }
</script>

{#if display.length > 0}
  <Tooltip text={tipText()}>
    <span class="spark" aria-hidden="true">
      {#each display as r (r.id)}
        <span class="dot" style="background: {color(r.status)};"></span>
      {/each}
    </span>
  </Tooltip>
{:else}
  <span class="spark empty"></span>
{/if}

<style>
  .spark {
    display: inline-flex;
    align-items: center;
    gap: 2px;
    height: 8px;
  }
  .spark.empty::before {
    content: '·';
    color: var(--text-dim);
    font-family: var(--font-mono);
    font-size: 0.75rem;
  }
  .dot {
    width: 5px;
    height: 5px;
    border-radius: 50%;
    flex-shrink: 0;
  }
</style>
