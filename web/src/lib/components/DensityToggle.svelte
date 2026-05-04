<script lang="ts">
  import { onMount } from 'svelte';

  let {
    storageKey,
    onChange
  }: {
    storageKey: string;
    onChange?: (v: 'comfortable' | 'compact') => void;
  } = $props();

  let density = $state<'comfortable' | 'compact'>('comfortable');

  function set(d: 'comfortable' | 'compact') {
    density = d;
    try { localStorage.setItem(storageKey, d); } catch { /* ignore */ }
    if (onChange) onChange(d);
  }

  onMount(() => {
    try {
      const saved = localStorage.getItem(storageKey);
      if (saved === 'comfortable' || saved === 'compact') {
        density = saved;
        if (onChange) onChange(saved);
      }
    } catch { /* ignore */ }
  });
</script>

<div class="dt" role="group" aria-label="Density">
  <button type="button"
          class="dt-btn"
          class:active={density === 'comfortable'}
          aria-pressed={density === 'comfortable'}
          onclick={() => set('comfortable')}
          title="Comfortable density">≡</button>
  <button type="button"
          class="dt-btn"
          class:active={density === 'compact'}
          aria-pressed={density === 'compact'}
          onclick={() => set('compact')}
          title="Compact density">☰</button>
</div>

<style>
  .dt {
    display: inline-flex;
    background: var(--bg-panel);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    overflow: hidden;
  }
  .dt-btn {
    background: transparent;
    color: var(--text-dim);
    border: none;
    padding: 0.25rem 0.5rem;
    font-size: var(--fs-md);
    cursor: pointer;
    line-height: 1;
  }
  .dt-btn.active {
    background: var(--bg-hover);
    color: var(--text);
  }
  .dt-btn:hover:not(.active) {
    color: var(--text);
  }
</style>
