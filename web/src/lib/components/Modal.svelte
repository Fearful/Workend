<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    open,
    title,
    width = 420,
    onClose,
    children,
    footer
  }: {
    open: boolean;
    title?: string;
    width?: number;
    onClose: () => void;
    children: Snippet;
    footer?: Snippet;
  } = $props();
</script>

{#if open}
  <div role="dialog"
       aria-modal="true"
       aria-label={title}
       tabindex="-1"
       class="modal-backdrop"
       onclick={(e) => { if (e.target === e.currentTarget) onClose(); }}
       onkeydown={(e) => { if (e.key === 'Escape') onClose(); }}>
    <div class="modal-content" style="width: min({width}px, 90vw);">
      {#if title}<h2 class="modal-title">{title}</h2>{/if}
      {@render children()}
      {#if footer}
        <div class="modal-footer">{@render footer()}</div>
      {/if}
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: var(--space-4);
  }
  .modal-content {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    padding: var(--space-5) var(--space-6);
    max-height: 90vh;
    overflow-y: auto;
  }
  .modal-title {
    margin: 0 0 var(--space-4) 0;
    font-size: 1.05rem;
    font-weight: 600;
  }
  .modal-footer {
    margin-top: var(--space-4);
    display: flex;
    justify-content: flex-end;
    gap: var(--space-2);
  }
</style>
