<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    title,
    actions,
    children
  }: {
    title?: string;
    actions?: Snippet;
    children?: Snippet;
  } = $props();
</script>

<div class="header-row">
  <h1>{#if children}{@render children()}{:else}{title}{/if}</h1>
  {#if actions}<div class="actions">{@render actions()}</div>{/if}
</div>

<style>
  .header-row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: var(--space-4);
    gap: var(--space-3);
    flex-wrap: wrap;
  }
  h1 {
    font-size: 1.5rem;
    margin: 0;
    font-weight: 600;
    letter-spacing: -0.02em;
    line-height: 1.3;
  }
  .actions {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    flex-wrap: wrap;
  }
</style>
