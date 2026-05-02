<script lang="ts">
  let { data } = $props();

  function statusColor(value: string | undefined): string {
    if (!value) return '#6b7280';
    if (value === 'ok') return '#22c55e';
    return '#ef4444';
  }
</script>

<style>
  .lede {
    color: #9ca3af;
    margin-bottom: 2rem;
    line-height: 1.5;
  }

  .panel {
    background: #14181d;
    border: 1px solid #1f2429;
    border-radius: 8px;
    padding: 1.25rem 1.5rem;
  }

  .panel h2 {
    margin: 0 0 1rem 0;
    font-size: 0.875rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: #9ca3af;
    font-weight: 600;
  }

  .check {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0;
    border-bottom: 1px solid #1f2429;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.875rem;
  }

  .check:last-child {
    border-bottom: none;
  }

  .dot {
    width: 8px;
    height: 8px;
    border-radius: 50%;
  }

  .check-name {
    flex: 1;
    color: #e8eaed;
  }

  .check-detail {
    color: #6b7280;
    font-size: 0.75rem;
  }

  .error {
    color: #ef4444;
    font-family: ui-monospace, "SF Mono", Menlo, monospace;
    font-size: 0.875rem;
    margin-top: 0.5rem;
  }
</style>

<p class="lede">
  Walking skeleton. Four services should be reachable: web (you're here), api,
  postgres, and the Dagger engine.
</p>

<section class="panel">
  <h2>System Status</h2>

  {#if !data.apiReachable}
    <div class="check">
      <span class="dot" style="background: {statusColor('error')}"></span>
      <span class="check-name">api</span>
      <span class="check-detail">unreachable</span>
    </div>
    <p class="error">{data.error}</p>
  {:else if data.health}
    <div class="check">
      <span class="dot" style="background: {statusColor(data.health.status)}"></span>
      <span class="check-name">overall</span>
      <span class="check-detail">{data.health.status}</span>
    </div>
    {#each Object.entries(data.health.checks) as [name, value]}
      <div class="check">
        <span class="dot" style="background: {statusColor(value as string)}"></span>
        <span class="check-name">{name}</span>
        <span class="check-detail">{value}</span>
      </div>
    {/each}
  {/if}
</section>
