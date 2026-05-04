<script lang="ts">
  let {
    status,
    size = 'md',
    label
  }: {
    status: string;
    size?: 'sm' | 'md';
    label?: string;
  } = $props();

  type Tone = 'success' | 'warning' | 'danger' | 'info' | 'neutral';

  function tone(s: string): Tone {
    const lc = s.toLowerCase();
    if (lc === 'ready' || lc === 'succeeded' || lc === 'success' || lc === 'ok' || lc === 'running-ok') return 'success';
    if (lc === 'running' || lc === 'starting') return 'info';
    if (lc === 'cloning' || lc === 'pending' || lc === 'queued' || lc === 'pause' || lc === 'paused') return 'warning';
    if (lc === 'failed' || lc === 'failure' || lc === 'error') return 'danger';
    return 'neutral';
  }

  let t = $derived(tone(status));
  let display = $derived(label ?? status);
</script>

<span class="pill pill-{t} pill-{size}">
  <span class="dot"></span>
  <span class="text">{display}</span>
</span>

<style>
  .pill {
    display: inline-flex;
    align-items: center;
    gap: 0.375rem;
    border-radius: var(--radius-full);
    font-weight: var(--fw-medium);
    line-height: 1;
    border: 1px solid;
    white-space: nowrap;
  }
  .pill-md {
    padding: 0.25rem 0.625rem;
    font-size: var(--fs-xs);
  }
  .pill-sm {
    padding: 0.125rem 0.5rem;
    font-size: 0.6875rem;
  }
  .dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }
  .pill-success { background: var(--status-success-bg); color: var(--status-success-fg); border-color: var(--status-success-border); }
  .pill-success .dot { background: var(--status-success-fg); }
  .pill-warning { background: var(--status-warning-bg); color: var(--status-warning-fg); border-color: var(--status-warning-border); }
  .pill-warning .dot { background: var(--status-warning-fg); animation: pulse 1.6s ease-in-out infinite; }
  .pill-danger { background: var(--status-danger-bg); color: var(--status-danger-fg); border-color: var(--status-danger-border); }
  .pill-danger .dot { background: var(--status-danger-fg); }
  .pill-info { background: var(--status-info-bg); color: var(--status-info-fg); border-color: var(--status-info-border); }
  .pill-info .dot { background: var(--status-info-fg); animation: pulse 1.6s ease-in-out infinite; }
  .pill-neutral { background: var(--status-neutral-bg); color: var(--status-neutral-fg); border-color: var(--status-neutral-border); }
  .pill-neutral .dot { background: var(--status-neutral-fg); }

  @keyframes pulse {
    0%, 100% { opacity: 1; }
    50% { opacity: 0.45; }
  }
</style>
