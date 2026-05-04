<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    variant = 'muted',
    size = 'md',
    children
  }: {
    variant?: 'success' | 'warning' | 'danger' | 'muted' | 'info' | 'accent';
    size?: 'sm' | 'md';
    children: Snippet;
  } = $props();
</script>

<span class="badge badge-{variant} size-{size}">{@render children()}</span>

<style>
  .badge {
    display: inline-flex;
    align-items: center;
    border-radius: var(--radius-full);
    font-weight: 500;
    line-height: 1.4;
    white-space: nowrap;
  }
  .size-sm {
    font-size: 0.6875rem;
    padding: 0.0625rem 0.5rem;
  }
  .size-md {
    font-size: 0.75rem;
    padding: 0.125rem 0.625rem;
  }
  .badge-success { background: rgba(34, 197, 94, 0.12); color: var(--success); }
  .badge-warning { background: rgba(234, 179, 8, 0.12); color: var(--warning); }
  .badge-danger  { background: rgba(239, 68, 68, 0.12); color: var(--danger-text); }
  .badge-muted   { background: rgba(107, 114, 128, 0.18); color: var(--text-muted); }
  .badge-info    { background: rgba(96, 165, 250, 0.12); color: var(--link); }
  .badge-accent  { background: rgba(37, 99, 235, 0.12); color: var(--accent); }
</style>
