<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    icon,
    message,
    actionHref,
    actionLabel,
    children
  }: {
    icon?: string;
    message?: string;
    actionHref?: string;
    actionLabel?: string;
    children?: Snippet;
  } = $props();
</script>

<div class="empty">
  {#if icon}<div class="icon">{icon}</div>{/if}
  {#if children}
    {@render children()}
  {:else}
    {#if message}<p class="message">{message}</p>{/if}
    {#if actionHref}
      <a href={actionHref} class="action">{actionLabel ?? 'Get started'}</a>
    {/if}
  {/if}
</div>

<style>
  .empty {
    text-align: center;
    padding: 3rem var(--space-4);
    color: var(--text-muted);
  }
  .icon {
    font-size: 2rem;
    opacity: 0.4;
    margin-bottom: var(--space-2);
  }
  .message {
    margin: 0 0 var(--space-3) 0;
  }
  .action {
    display: inline-block;
    margin-top: var(--space-2);
    padding: var(--space-2) var(--space-4);
    background: var(--accent);
    color: white;
    border-radius: var(--radius-md);
    font-size: 0.875rem;
    font-weight: 500;
  }
  .action:hover {
    background: var(--accent-hover);
    text-decoration: none;
  }
</style>
