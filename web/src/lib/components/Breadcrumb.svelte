<script lang="ts">
  type Segment = { label: string; href?: string };

  let { segments }: { segments: Segment[] } = $props();
</script>

<div class="breadcrumb">
  {#each segments as seg, i (i)}
    {#if seg.href}<a href={seg.href}>{seg.label}</a>{:else}<span class="current">{seg.label}</span>{/if}
    {#if i < segments.length - 1}<span class="sep">/</span>{/if}
  {/each}
</div>

<style>
  .breadcrumb {
    color: var(--text-dim);
    font-size: 0.875rem;
    margin-bottom: var(--space-4);
    display: flex;
    flex-wrap: wrap;
    gap: var(--space-2);
    align-items: center;
  }
  .sep { color: var(--text-dim); }
  .current { color: var(--text-muted); }
</style>
