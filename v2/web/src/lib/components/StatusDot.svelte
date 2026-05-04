<script lang="ts">
  import { statusColor } from '$lib/utils';

  let { status, size = 8 }: { status: string | null; size?: number } = $props();
</script>

<span class="dot" style="background: {statusColor(status)}; width: {size}px; height: {size}px;"></span>
