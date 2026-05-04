<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    open,
    title,
    onClose,
    children
  }: {
    open: boolean;
    title?: string;
    onClose: () => void;
    children: Snippet;
  } = $props();
</script>

{#if open}
  <div class="sheet-backdrop"
       role="presentation"
       onclick={(e) => { if (e.target === e.currentTarget) onClose(); }}
       onkeydown={(e) => { if (e.key === 'Escape') onClose(); }}>
    <div class="sheet" role="dialog" aria-modal="true" aria-label={title ?? 'Actions'}>
      <div class="grab" aria-hidden="true"></div>
      {#if title}<h2 class="title">{title}</h2>{/if}
      <div class="body">{@render children()}</div>
    </div>
  </div>
{/if}

<style>
  .sheet-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.5);
    z-index: 200;
    display: flex;
    align-items: flex-end;
    justify-content: center;
    animation: backdrop 160ms ease-out;
  }
  .sheet {
    width: 100%;
    max-width: 640px;
    background: var(--bg-panel);
    border-top-left-radius: var(--radius-lg);
    border-top-right-radius: var(--radius-lg);
    border-top: 1px solid var(--border-strong);
    padding: var(--space-3) var(--space-4) calc(var(--space-5) + env(safe-area-inset-bottom));
    max-height: 80vh;
    overflow-y: auto;
    animation: slide-up 200ms ease-out;
  }
  .grab {
    width: 40px;
    height: 4px;
    border-radius: 2px;
    background: var(--border-strong);
    margin: 0 auto var(--space-3);
  }
  .title {
    margin: 0 0 var(--space-3);
    font-size: var(--fs-md);
    font-weight: var(--fw-semibold);
    color: var(--text);
  }
  .body {
    display: flex;
    flex-direction: column;
    gap: var(--space-1);
  }
  @keyframes slide-up {
    from { transform: translateY(100%); }
    to { transform: translateY(0); }
  }
  @keyframes backdrop {
    from { opacity: 0; }
    to { opacity: 1; }
  }
  @media (min-width: 769px) {
    .sheet-backdrop {
      align-items: center;
    }
    .sheet {
      max-width: 480px;
      border-radius: var(--radius-lg);
      border-top: 1px solid var(--border-strong);
      animation: fade-in 160ms ease-out;
      padding: var(--space-5) var(--space-5) var(--space-4);
    }
    .grab { display: none; }
    @keyframes fade-in {
      from { opacity: 0; transform: scale(0.97); }
      to { opacity: 1; transform: scale(1); }
    }
  }
</style>
