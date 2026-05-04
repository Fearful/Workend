<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title,
    subtitle,
    actions,
    level = 'h2',
    children
  }: {
    title?: string;
    subtitle?: string;
    actions?: Snippet;
    level?: 'h1' | 'h2' | 'h3';
    children?: Snippet;
  } = $props();
</script>

<div class="sh">
  <div class="sh-text">
    {#if level === 'h1'}
      <h1 class="sh-title sh-title-h1">{#if children}{@render children()}{:else}{title}{/if}</h1>
    {:else if level === 'h3'}
      <h3 class="sh-title sh-title-h3">{#if children}{@render children()}{:else}{title}{/if}</h3>
    {:else}
      <h2 class="sh-title sh-title-h2">{#if children}{@render children()}{:else}{title}{/if}</h2>
    {/if}
    {#if subtitle}<span class="sh-sub">{subtitle}</span>{/if}
  </div>
  {#if actions}<div class="sh-actions">{@render actions()}</div>{/if}
</div>

<style>
  .sh {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: var(--space-3);
    margin-bottom: var(--space-3);
    flex-wrap: wrap;
  }
  .sh-text { display: flex; align-items: baseline; gap: var(--space-3); flex-wrap: wrap; min-width: 0; }
  .sh-title {
    margin: 0;
    font-weight: var(--fw-semibold);
    letter-spacing: -0.01em;
    line-height: var(--lh-tight);
  }
  .sh-title-h1 { font-size: var(--fs-2xl); }
  .sh-title-h2 { font-size: var(--fs-lg); }
  .sh-title-h3 { font-size: var(--fs-md); }
  .sh-sub {
    color: var(--text-dim);
    font-size: var(--fs-sm);
  }
  .sh-actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    flex-shrink: 0;
  }
</style>
