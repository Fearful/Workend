<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title,
    actions,
    padding = 'normal',
    children
  }: {
    title?: string;
    actions?: Snippet;
    padding?: 'normal' | 'compact' | 'none';
    children: Snippet;
  } = $props();
</script>

<section class="panel padding-{padding}">
  {#if title || actions}
    <div class="panel-head">
      {#if title}<h2 class="panel-title">{title}</h2>{/if}
      {#if actions}<div class="panel-actions">{@render actions()}</div>{/if}
    </div>
  {/if}
  {@render children()}
</section>

<style>
  .panel {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    margin-bottom: var(--space-4);
  }
  .padding-normal { padding: var(--space-5) var(--space-6); }
  .padding-compact { padding: var(--space-3) var(--space-4); }
  .padding-none { padding: 0; }

  .panel-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-3);
    gap: var(--space-3);
  }

  .panel-title {
    margin: 0;
    font-size: var(--fs-md);
    color: var(--text);
    font-weight: var(--fw-semibold);
    letter-spacing: -0.005em;
    line-height: var(--lh-tight);
  }

  .panel-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
  }
</style>
