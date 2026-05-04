<script lang="ts">
  import { onMount } from "svelte";

  let iframeUrl = $state("");

  onMount(() => {
    iframeUrl = `${window.location.protocol}//${window.location.hostname}:4318`;
  });

  $effect(() => {
    const main = document.querySelector('main');
    main?.classList.add('fullbleed');
    return () => main?.classList.remove('fullbleed');
  });
</script>

<svelte:head>
  <title>Telemetry | Workend</title>
</svelte:head>

<style>
  .iframe-panel {
    background: var(--bg-panel);
    border: 1px solid var(--border);
    border-radius: var(--radius-lg);
    height: calc(100vh - 120px);
    padding: 0;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    margin-bottom: var(--space-4);
  }
  iframe {
    width: 100%;
    height: 100%;
    border: none;
    flex-grow: 1;
  }
</style>

<div class="iframe-panel">
  {#if iframeUrl}
    <iframe
      src={iframeUrl}
      title="OpenTelemetry GUI"
      allow="fullscreen"
    ></iframe>
  {:else}
    <div class="empty">Loading Telemetry...</div>
  {/if}
</div>
