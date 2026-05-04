<script lang="ts" generics="T extends string">
  let {
    options,
    value,
    onChange,
    label
  }: {
    options: { value: T; label: string; count?: number }[];
    value: T;
    onChange: (v: T) => void;
    label?: string;
  } = $props();
</script>

<div class="filters" role="tablist" aria-label={label}>
  {#if label}<span class="label">{label}</span>{/if}
  {#each options as o (o.value)}
    <button type="button" role="tab"
            class="pill {value === o.value ? 'active' : ''}"
            aria-selected={value === o.value}
            onclick={() => onChange(o.value)}>
      {o.label}
      {#if o.count != null}<span class="count">{o.count}</span>{/if}
    </button>
  {/each}
</div>

<style>
  .filters {
    display: flex;
    gap: var(--space-2);
    align-items: center;
    flex-wrap: wrap;
  }
  .label {
    color: var(--text-dim);
    font-size: var(--fs-sm);
    margin-right: var(--space-1);
  }
  .pill {
    padding: 0.25rem 0.625rem;
    background: transparent;
    color: var(--text-muted);
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-full);
    cursor: pointer;
    font-size: var(--fs-sm);
    font-family: inherit;
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    transition: background 80ms, color 80ms, border-color 80ms;
  }
  .pill:hover { background: var(--bg-hover); color: var(--text); }
  .pill.active {
    background: var(--accent);
    color: white;
    border-color: var(--accent);
  }
  .count {
    font-size: 0.6875rem;
    background: rgba(255, 255, 255, 0.1);
    padding: 0 0.375rem;
    border-radius: var(--radius-full);
    min-width: 1rem;
    text-align: center;
  }
  .pill:not(.active) .count {
    background: var(--bg-hover);
    color: var(--text-dim);
  }
</style>
