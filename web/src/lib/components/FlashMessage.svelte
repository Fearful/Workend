<script lang="ts">
  import type { Snippet } from 'svelte';

  let {
    type = 'error',
    children
  }: {
    type?: 'error' | 'success';
    children: Snippet;
  } = $props();
</script>

<div class={type === 'error' ? 'error-banner' : 'success-banner'} role={type === 'error' ? 'alert' : 'status'}>
  {@render children()}
</div>
